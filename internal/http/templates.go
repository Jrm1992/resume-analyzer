package http

import (
	"fmt"
	"html/template"
	"io"

	"github.com/jose/resume-analyzer/internal/assets"
)

type Templates struct {
	pages    *template.Template
	partials *template.Template
}

func LoadTemplates() (*Templates, error) {
	pages, err := template.ParseFS(assets.Templates,
		"templates/layout.html",
		"templates/index.html",
	)
	if err != nil {
		return nil, fmt.Errorf("templates: pages: %w", err)
	}
	partials, err := template.ParseFS(assets.Templates,
		"templates/partials/queued.html",
		"templates/partials/done.html",
		"templates/partials/failed.html",
	)
	if err != nil {
		return nil, fmt.Errorf("templates: partials: %w", err)
	}
	return &Templates{pages: pages, partials: partials}, nil
}

func (t *Templates) RenderPage(w io.Writer, _ string, data any) error {
	return t.pages.ExecuteTemplate(w, "layout", data)
}

func (t *Templates) RenderPartial(w io.Writer, name string, data any) error {
	return t.partials.ExecuteTemplate(w, name, data)
}
