package llm

import (
	"strings"
	"testing"
)

func TestBuildPrompt_IncludesResumeAndJD(t *testing.T) {
	sys, user := BuildPrompt("RESUME TEXT HERE", "JOB DESCRIPTION HERE")
	if !strings.Contains(sys, "JSON") {
		t.Error("system prompt should mention JSON output")
	}
	if !strings.Contains(sys, "score") {
		t.Error("system prompt should reference score field")
	}
	if !strings.Contains(user, "RESUME TEXT HERE") {
		t.Error("user prompt missing resume text")
	}
	if !strings.Contains(user, "JOB DESCRIPTION HERE") {
		t.Error("user prompt missing job description")
	}
}

func TestBuildStrictPrompt_EmphasizesJSONOnly(t *testing.T) {
	sys := BuildStrictSystemPrompt()
	if !strings.Contains(strings.ToLower(sys), "only valid json") {
		t.Error("strict prompt should emphasize JSON-only output")
	}
}
