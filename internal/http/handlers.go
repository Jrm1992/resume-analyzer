package http

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jose/resume-analyzer/internal/config"
	"github.com/jose/resume-analyzer/internal/jobs"
	"github.com/jose/resume-analyzer/internal/llm"
	"github.com/jose/resume-analyzer/internal/pdf"
)

const maxJDBytes = 50 * 1024

type Analyzer interface {
	Analyze(ctx context.Context, resume, jd, language string) (*llm.AnalysisResult, error)
}

// allowedLanguages is the set of language codes accepted by handleAnalyze.
// "" means auto-detect from resume.
var allowedLanguages = map[string]bool{
	"":   true,
	"pt": true,
	"en": true,
	"es": true,
}

type Server struct {
	Config    *config.Config
	Templates *Templates
	Store     *jobs.Store
	Queue     *jobs.Queue
	Analyzer  Analyzer
}

func (s *Server) JobHandler(ctx context.Context, j *jobs.Job) {
	s.Store.Update(j.ID, func(j *jobs.Job) { j.Status = jobs.StatusRunning })
	cctx, cancel := context.WithTimeout(ctx, s.Config.LLMTimeout)
	defer cancel()
	start := time.Now()
	res, err := s.Analyzer.Analyze(cctx, j.Resume, j.JD, j.Language)
	dur := time.Since(start)
	if err != nil {
		slog.Error("job failed", "id", j.ID, "dur_ms", dur.Milliseconds(), "model", s.Config.LLMModel, "err", err)
		s.Store.Update(j.ID, func(j *jobs.Job) {
			j.Status = jobs.StatusFailed
			j.Err = err.Error()
		})
		return
	}
	slog.Info("job done", "id", j.ID, "dur_ms", dur.Milliseconds(), "model", s.Config.LLMModel, "score", res.Score)
	s.Store.Update(j.ID, func(j *jobs.Job) {
		j.Status = jobs.StatusDone
		j.Result = res
	})
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.Templates.RenderPage(w, "index", nil); err != nil {
		slog.Error("render index", "err", err)
		http.Error(w, "render error", 500)
	}
}

func (s *Server) handleAnalyze(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(s.Config.MaxPDFBytes + 256*1024); err != nil {
		s.badRequest(w, "Upload too large or malformed.")
		return
	}
	file, hdr, err := r.FormFile("resume")
	if err != nil {
		s.badRequest(w, "Upload a PDF file.")
		return
	}
	defer file.Close()

	if hdr.Size > s.Config.MaxPDFBytes {
		s.badRequest(w, fmt.Sprintf("File too large (limit %d MB).", s.Config.MaxPDFBytes/(1024*1024)))
		return
	}

	var head [5]byte
	if _, err := io.ReadFull(file, head[:]); err != nil {
		s.badRequest(w, "Upload a PDF file.")
		return
	}
	if string(head[:]) != "%PDF-" {
		s.badRequest(w, "Upload a PDF file.")
		return
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		s.writeError(w, 500, "could not read upload")
		return
	}

	jd := strings.TrimSpace(r.FormValue("jd"))
	if jd == "" {
		s.badRequest(w, "Paste a job description.")
		return
	}
	if len(jd) > maxJDBytes {
		s.badRequest(w, "Job description too long (max 50 KB).")
		return
	}

	lang := strings.ToLower(strings.TrimSpace(r.FormValue("lang")))
	if !allowedLanguages[lang] {
		s.badRequest(w, "Unsupported language. Use one of: auto, pt, en, es.")
		return
	}

	resumeText, err := pdf.Parse(file)
	if err != nil {
		if errors.Is(err, pdf.ErrEmptyText) {
			s.badRequest(w, "Could not read PDF text. Is it scanned or encrypted?")
			return
		}
		s.badRequest(w, "Could not read PDF: "+err.Error())
		return
	}

	j := s.Store.Create(resumeText, jd, lang)
	if err := s.Queue.Enqueue(j); err != nil {
		if errors.Is(err, jobs.ErrQueueFull) {
			s.writeError(w, http.StatusServiceUnavailable, "Server busy, try again.")
			return
		}
		s.writeError(w, 500, err.Error())
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.Templates.RenderPartial(w, "queued", j); err != nil {
		slog.Error("render queued", "err", err)
	}
}

func (s *Server) handleJobStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	j, ok := s.Store.Get(id)
	if !ok {
		http.Error(w, "job not found", 404)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	switch j.Status {
	case jobs.StatusQueued, jobs.StatusRunning:
		_ = s.Templates.RenderPartial(w, "queued", j)
	case jobs.StatusDone:
		_ = s.Templates.RenderPartial(w, "done", j)
	case jobs.StatusFailed:
		_ = s.Templates.RenderPartial(w, "failed", j)
	}
}

func (s *Server) handleJobPDF(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	j, ok := s.Store.Get(id)
	if !ok {
		http.Error(w, "job not found", 404)
		return
	}
	if j.Status != jobs.StatusDone {
		http.Error(w, "analysis not complete", http.StatusConflict)
		return
	}
	data, err := pdf.Render(j.Result.Rewritten)
	if err != nil {
		slog.Error("pdf render", "err", err)
		http.Error(w, "render error", 500)
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", `attachment; filename="resume-rewritten.pdf"`)
	_, _ = w.Write(data)
}

func (s *Server) badRequest(w http.ResponseWriter, msg string) {
	s.writeError(w, http.StatusBadRequest, msg)
}

func (s *Server) writeError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(code)
	_, _ = w.Write([]byte(`<div class="error">` + htmlEscape(msg) + `</div>`))
}

func htmlEscape(s string) string {
	var buf bytes.Buffer
	for _, r := range s {
		switch r {
		case '<':
			buf.WriteString("&lt;")
		case '>':
			buf.WriteString("&gt;")
		case '&':
			buf.WriteString("&amp;")
		case '"':
			buf.WriteString("&quot;")
		default:
			buf.WriteRune(r)
		}
	}
	return buf.String()
}
