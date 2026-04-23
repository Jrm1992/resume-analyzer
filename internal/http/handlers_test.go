package http

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jose/resume-analyzer/internal/config"
	"github.com/jose/resume-analyzer/internal/jobs"
	"github.com/jose/resume-analyzer/internal/llm"
	"github.com/jose/resume-analyzer/internal/pdf"
)

type stubAnalyzer struct {
	result *llm.AnalysisResult
	err    error
	delay  time.Duration
}

func (s *stubAnalyzer) Analyze(ctx context.Context, resume, jd string) (*llm.AnalysisResult, error) {
	if s.delay > 0 {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(s.delay):
		}
	}
	return s.result, s.err
}

func newTestServer(t *testing.T, a Analyzer) *Server {
	t.Helper()
	cfg := &config.Config{MaxPDFBytes: 10 << 20, LLMTimeout: 5 * time.Second, Workers: 1, QueueCapacity: 8, JobTTL: time.Hour}
	tpl, err := LoadTemplates()
	if err != nil {
		t.Fatalf("LoadTemplates: %v", err)
	}
	store := jobs.NewStore()
	queue := jobs.NewQueue(cfg.Workers, cfg.QueueCapacity)
	s := &Server{Config: cfg, Templates: tpl, Store: store, Queue: queue, Analyzer: a}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	queue.Start(ctx, s.JobHandler)
	return s
}

func makePDFUpload(t *testing.T, pdfBytes []byte, jd string) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, _ := mw.CreateFormFile("resume", "resume.pdf")
	_, _ = fw.Write(pdfBytes)
	_ = mw.WriteField("jd", jd)
	_ = mw.Close()
	return &buf, mw.FormDataContentType()
}

func TestIndex_Returns200WithForm(t *testing.T) {
	s := newTestServer(t, &stubAnalyzer{})
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	s.Router().ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "<form") {
		t.Error("missing form")
	}
}

func TestAnalyze_RejectsMissingFile(t *testing.T) {
	s := newTestServer(t, &stubAnalyzer{})
	body, ct := makePDFUpload(t, []byte{}, "some jd")
	req := httptest.NewRequest("POST", "/analyze", body)
	req.Header.Set("Content-Type", ct)
	rec := httptest.NewRecorder()
	s.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d", rec.Code)
	}
}

func TestAnalyze_RejectsNonPDFUpload(t *testing.T) {
	s := newTestServer(t, &stubAnalyzer{})
	body, ct := makePDFUpload(t, []byte("HELLO NOT A PDF"), "some jd")
	req := httptest.NewRequest("POST", "/analyze", body)
	req.Header.Set("Content-Type", ct)
	rec := httptest.NewRecorder()
	s.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Upload a PDF file") {
		t.Errorf("body missing expected message: %s", rec.Body.String())
	}
}

func TestAnalyze_RejectsEmptyJD(t *testing.T) {
	s := newTestServer(t, &stubAnalyzer{})
	body, ct := makePDFUpload(t, makeSimplePDF2(t, "hi"), "")
	req := httptest.NewRequest("POST", "/analyze", body)
	req.Header.Set("Content-Type", ct)
	rec := httptest.NewRecorder()
	s.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d", rec.Code)
	}
}

func TestAnalyze_Valid_EnqueuesAndReturnsSpinner(t *testing.T) {
	stub := &stubAnalyzer{result: sampleResult(), delay: 50 * time.Millisecond}
	s := newTestServer(t, stub)
	body, ct := makePDFUpload(t, makeSimplePDF2(t, "Jane Doe Backend"), "Looking for Go engineer")
	req := httptest.NewRequest("POST", "/analyze", body)
	req.Header.Set("Content-Type", ct)
	rec := httptest.NewRecorder()
	s.Router().ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	b := rec.Body.String()
	if !strings.Contains(b, `hx-get="/jobs/`) {
		t.Errorf("spinner partial missing hx-get: %s", b)
	}
}

func TestJobStatus_Polls_And_ReachesDone(t *testing.T) {
	stub := &stubAnalyzer{result: sampleResult()}
	s := newTestServer(t, stub)

	body, ct := makePDFUpload(t, makeSimplePDF2(t, "Jane Doe Backend"), "Go engineer")
	req := httptest.NewRequest("POST", "/analyze", body)
	req.Header.Set("Content-Type", ct)
	rec := httptest.NewRecorder()
	s.Router().ServeHTTP(rec, req)
	jobID := extractJobID(t, rec.Body.String())

	deadline := time.Now().Add(time.Second)
	var last string
	for time.Now().Before(deadline) {
		r2 := httptest.NewRecorder()
		req2 := httptest.NewRequest("GET", "/jobs/"+jobID, nil)
		s.Router().ServeHTTP(r2, req2)
		last = r2.Body.String()
		if strings.Contains(last, "Match score") {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("never reached done, last body = %s", last)
}

func TestJobPDF_ReturnsPDFWhenDone(t *testing.T) {
	stub := &stubAnalyzer{result: sampleResult()}
	s := newTestServer(t, stub)

	body, ct := makePDFUpload(t, makeSimplePDF2(t, "Jane Doe"), "Go role")
	req := httptest.NewRequest("POST", "/analyze", body)
	req.Header.Set("Content-Type", ct)
	rec := httptest.NewRecorder()
	s.Router().ServeHTTP(rec, req)
	jobID := extractJobID(t, rec.Body.String())

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if j, ok := s.Store.Get(jobID); ok && j.Status == jobs.StatusDone {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	r2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("GET", "/jobs/"+jobID+"/pdf", nil)
	s.Router().ServeHTTP(r2, req2)
	if r2.Code != 200 {
		t.Fatalf("status = %d", r2.Code)
	}
	ct2 := r2.Header().Get("Content-Type")
	if ct2 != "application/pdf" {
		t.Errorf("Content-Type = %q", ct2)
	}
	data, _ := io.ReadAll(r2.Body)
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		t.Error("not a PDF")
	}
}

func TestJobPDF_Returns409WhenRunning(t *testing.T) {
	stub := &stubAnalyzer{result: sampleResult(), delay: 500 * time.Millisecond}
	s := newTestServer(t, stub)

	body, ct := makePDFUpload(t, makeSimplePDF2(t, "Jane"), "jd")
	req := httptest.NewRequest("POST", "/analyze", body)
	req.Header.Set("Content-Type", ct)
	rec := httptest.NewRecorder()
	s.Router().ServeHTTP(rec, req)
	jobID := extractJobID(t, rec.Body.String())

	r2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("GET", "/jobs/"+jobID+"/pdf", nil)
	s.Router().ServeHTTP(r2, req2)
	if r2.Code != http.StatusConflict {
		t.Errorf("status = %d, want 409", r2.Code)
	}
}

// helpers

func sampleResult() *llm.AnalysisResult {
	return &llm.AnalysisResult{
		Score:     80,
		Breakdown: llm.CategoryBreakdown{Skills: 80, Experience: 80, Education: 80},
		Missing:   []string{"Kubernetes"},
		Strengths: []string{"Go"},
		Rewritten: pdf.RewrittenResume{Name: "Jane Doe"},
	}
}

func extractJobID(t *testing.T, body string) string {
	t.Helper()
	const marker = `hx-get="/jobs/`
	i := strings.Index(body, marker)
	if i < 0 {
		t.Fatalf("marker not found in %q", body)
	}
	rest := body[i+len(marker):]
	end := strings.Index(rest, `"`)
	if end < 0 {
		t.Fatal("end quote not found")
	}
	return rest[:end]
}

// Duplicate of pdf.makeSimplePDF to avoid cross-package test helpers.
func makeSimplePDF2(t *testing.T, text string) []byte {
	t.Helper()
	content := "BT /F1 12 Tf 72 720 Td (" + text + ") Tj ET"
	stream := []byte(content)
	var pdfBuf bytes.Buffer
	pdfBuf.WriteString("%PDF-1.4\n")
	objs := []string{
		"1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n",
		"2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n",
		"3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R /Resources << /Font << /F1 5 0 R >> >> >>\nendobj\n",
		"4 0 obj\n<< /Length " + itoa2(len(stream)) + " >>\nstream\n" + string(stream) + "\nendstream\nendobj\n",
		"5 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>\nendobj\n",
	}
	offsets := []int{0}
	for _, o := range objs {
		offsets = append(offsets, pdfBuf.Len())
		pdfBuf.WriteString(o)
	}
	xrefStart := pdfBuf.Len()
	pdfBuf.WriteString("xref\n0 6\n0000000000 65535 f \n")
	for _, off := range offsets[1:] {
		pdfBuf.WriteString(pad102(off) + " 00000 n \n")
	}
	pdfBuf.WriteString("trailer\n<< /Size 6 /Root 1 0 R >>\nstartxref\n" + itoa2(xrefStart) + "\n%%EOF\n")
	return pdfBuf.Bytes()
}
func itoa2(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
func pad102(n int) string {
	s := itoa2(n)
	for len(s) < 10 {
		s = "0" + s
	}
	return s
}
