package http

import (
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jose/resume-analyzer/internal/assets"
)

func (s *Server) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(recoverMiddleware)
	r.Use(loggingMiddleware)
	r.Use(maxBodyMiddleware(s.Config.MaxPDFBytes + 512*1024)) // PDF + JD + form overhead

	r.Get("/", s.handleIndex)
	r.Post("/analyze", s.handleAnalyze)
	r.Get("/jobs/{id}", s.handleJobStatus)
	r.Get("/jobs/{id}/pdf", s.handleJobPDF)
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})

	staticSub, _ := fs.Sub(assets.Static, "static")
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(staticSub))))

	return r
}
