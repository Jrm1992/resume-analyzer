package http

import (
	"bytes"
	"strings"
	"testing"

	"github.com/jose/resume-analyzer/internal/jobs"
	"github.com/jose/resume-analyzer/internal/llm"
)

func TestTemplates_RendersIndex(t *testing.T) {
	tpl, err := LoadTemplates()
	if err != nil {
		t.Fatalf("LoadTemplates: %v", err)
	}
	var buf bytes.Buffer
	if err := tpl.RenderPage(&buf, "index", nil); err != nil {
		t.Fatalf("RenderPage: %v", err)
	}
	s := buf.String()
	if !strings.Contains(s, "<form") {
		t.Error("index missing form")
	}
	if !strings.Contains(s, "name=\"resume\"") {
		t.Error("index missing resume field")
	}
}

func TestTemplates_RendersDonePartial(t *testing.T) {
	tpl, err := LoadTemplates()
	if err != nil {
		t.Fatalf("LoadTemplates: %v", err)
	}
	j := &jobs.Job{
		ID:     "abc",
		Status: jobs.StatusDone,
		Result: &llm.AnalysisResult{
			Score:     77,
			Strengths: []string{"Go"},
			Missing:   []string{"Kubernetes"},
		},
	}
	var buf bytes.Buffer
	if err := tpl.RenderPartial(&buf, "done", j); err != nil {
		t.Fatalf("RenderPartial: %v", err)
	}
	s := buf.String()
	if !strings.Contains(s, "77%") {
		t.Error("missing score")
	}
	if !strings.Contains(s, "Kubernetes") {
		t.Error("missing keyword chip")
	}
}
