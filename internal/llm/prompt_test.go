package llm

import (
	"strings"
	"testing"
)

func TestBuildPrompt_IncludesResumeAndJD(t *testing.T) {
	sys, user := BuildPrompt("RESUME TEXT HERE", "JOB DESCRIPTION HERE", LangAuto)
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

func TestBuildPrompt_LanguageDirectiveIsExplicit(t *testing.T) {
	cases := map[string]string{
		LangPT:   "PORTUGUESE",
		LangEN:   "ENGLISH",
		LangES:   "SPANISH",
		LangAuto: "Detect the resume's language",
	}
	for lang, want := range cases {
		t.Run("lang="+lang, func(t *testing.T) {
			sys, _ := BuildPrompt("r", "j", lang)
			if !strings.Contains(sys, want) {
				t.Errorf("lang %q: prompt missing %q\n---\n%s", lang, want, sys)
			}
			if !strings.Contains(sys, "Do NOT mix languages") {
				t.Error("prompt missing mix-prevention clause")
			}
		})
	}
}

func TestBuildStrictPrompt_EmphasizesJSONOnly(t *testing.T) {
	sys := BuildStrictSystemPrompt(LangPT)
	if !strings.Contains(strings.ToLower(sys), "only valid json") {
		t.Error("strict prompt should emphasize JSON-only output")
	}
	if !strings.Contains(sys, "PORTUGUESE") {
		t.Error("strict prompt should carry language directive")
	}
}
