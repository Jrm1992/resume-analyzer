package config

import (
	"strings"
	"testing"
	"time"
)

// clearEnv clears every env var the config reads, so each test starts from a
// deterministic baseline regardless of the host environment.
func clearEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{
		"PORT",
		"LLM_BASE_URL",
		"LLM_API_KEY",
		"LLM_MODEL",
		"LLM_MAX_TOKENS",
		"MAX_PDF_MB",
		"LLM_TIMEOUT_SEC",
		"WORKERS",
		"QUEUE_CAPACITY",
		"JOB_TTL_MIN",
	} {
		t.Setenv(k, "")
	}
}

func TestLoad_Defaults(t *testing.T) {
	clearEnv(t)
	t.Setenv("LLM_API_KEY", "test-key")

	c, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.Port != 8080 {
		t.Errorf("Port = %d", c.Port)
	}
	if c.LLMBaseURL != "https://api.openai.com/v1" {
		t.Errorf("LLMBaseURL = %q", c.LLMBaseURL)
	}
	if c.LLMAPIKey != "test-key" {
		t.Errorf("LLMAPIKey = %q", c.LLMAPIKey)
	}
	if c.LLMModel != "gpt-4o-mini" {
		t.Errorf("LLMModel = %q", c.LLMModel)
	}
	if c.LLMMaxTokens != 4000 {
		t.Errorf("LLMMaxTokens = %d", c.LLMMaxTokens)
	}
	if c.MaxPDFBytes != 10*1024*1024 {
		t.Errorf("MaxPDFBytes = %d", c.MaxPDFBytes)
	}
	if c.LLMTimeout != 120*time.Second {
		t.Errorf("LLMTimeout = %v", c.LLMTimeout)
	}
	if c.Workers != 2 {
		t.Errorf("Workers = %d", c.Workers)
	}
	if c.QueueCapacity != 100 {
		t.Errorf("QueueCapacity = %d", c.QueueCapacity)
	}
	if c.JobTTL != 60*time.Minute {
		t.Errorf("JobTTL = %v", c.JobTTL)
	}
}

func TestLoad_Overrides(t *testing.T) {
	clearEnv(t)
	t.Setenv("LLM_API_KEY", "sk-xyz")
	t.Setenv("LLM_BASE_URL", "https://api.anthropic.com/v1/")
	t.Setenv("LLM_MODEL", "claude-sonnet-4-5")
	t.Setenv("LLM_MAX_TOKENS", "8192")
	t.Setenv("PORT", "9000")
	t.Setenv("MAX_PDF_MB", "5")
	t.Setenv("LLM_TIMEOUT_SEC", "30")
	t.Setenv("WORKERS", "4")

	c, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.Port != 9000 {
		t.Errorf("Port = %d", c.Port)
	}
	if c.LLMBaseURL != "https://api.anthropic.com/v1" {
		t.Errorf("LLMBaseURL = %q (trailing slash should be trimmed)", c.LLMBaseURL)
	}
	if c.LLMAPIKey != "sk-xyz" {
		t.Errorf("LLMAPIKey = %q", c.LLMAPIKey)
	}
	if c.LLMModel != "claude-sonnet-4-5" {
		t.Errorf("LLMModel = %q", c.LLMModel)
	}
	if c.LLMMaxTokens != 8192 {
		t.Errorf("LLMMaxTokens = %d", c.LLMMaxTokens)
	}
	if c.MaxPDFBytes != 5*1024*1024 {
		t.Errorf("MaxPDFBytes = %d", c.MaxPDFBytes)
	}
	if c.LLMTimeout != 30*time.Second {
		t.Errorf("LLMTimeout = %v", c.LLMTimeout)
	}
	if c.Workers != 4 {
		t.Errorf("Workers = %d", c.Workers)
	}
}

func TestLoad_MissingAPIKey(t *testing.T) {
	clearEnv(t)
	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing LLM_API_KEY")
	}
	if !strings.Contains(err.Error(), "LLM_API_KEY") {
		t.Errorf("err = %v", err)
	}
}

func TestLoad_WhitespaceAPIKey_FailsFast(t *testing.T) {
	clearEnv(t)
	t.Setenv("LLM_API_KEY", "   ")
	_, err := Load()
	if err == nil {
		t.Fatal("expected error for whitespace-only LLM_API_KEY")
	}
}

func TestLoad_InvalidInt(t *testing.T) {
	clearEnv(t)
	t.Setenv("LLM_API_KEY", "k")
	t.Setenv("PORT", "notanumber")
	if _, err := Load(); err == nil {
		t.Fatal("expected error for invalid PORT")
	}
}

func TestLoad_RejectsNonPositiveInts(t *testing.T) {
	cases := []struct{ env, val string }{
		{"PORT", "0"},
		{"PORT", "-1"},
		{"WORKERS", "0"},
		{"QUEUE_CAPACITY", "0"},
		{"MAX_PDF_MB", "0"},
		{"LLM_TIMEOUT_SEC", "0"},
		{"JOB_TTL_MIN", "0"},
		{"LLM_MAX_TOKENS", "0"},
	}
	for _, c := range cases {
		t.Run(c.env+"="+c.val, func(t *testing.T) {
			clearEnv(t)
			t.Setenv("LLM_API_KEY", "k")
			t.Setenv(c.env, c.val)
			if _, err := Load(); err == nil {
				t.Fatalf("%s=%s: expected error, got nil", c.env, c.val)
			}
		})
	}
}

func TestLoad_WhitespaceBaseURL_FallsBackToDefault(t *testing.T) {
	clearEnv(t)
	t.Setenv("LLM_API_KEY", "k")
	t.Setenv("LLM_BASE_URL", "   ")
	c, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.LLMBaseURL != "https://api.openai.com/v1" {
		t.Errorf("LLMBaseURL = %q, want default", c.LLMBaseURL)
	}
}
func TestLoad_ConfigSecretFillsAPIKey(t *testing.T) {
	clearEnv(t)
	t.Setenv("CONFIG_SECRET", `{"LLM_API_KEY":"from-secret","LLM_MODEL":"gpt-5-mini"}`)

	c, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.LLMAPIKey != "from-secret" {
		t.Errorf("LLMAPIKey = %q, want from-secret", c.LLMAPIKey)
	}
	if c.LLMModel != "gpt-5-mini" {
		t.Errorf("LLMModel = %q, want gpt-5-mini", c.LLMModel)
	}
}

func TestLoad_ExplicitEnvWinsOverConfigSecret(t *testing.T) {
	clearEnv(t)
	t.Setenv("CONFIG_SECRET", `{"LLM_API_KEY":"from-secret","LLM_MODEL":"from-secret-model"}`)
	t.Setenv("LLM_API_KEY", "explicit-key")

	c, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.LLMAPIKey != "explicit-key" {
		t.Errorf("LLMAPIKey = %q, want explicit-key", c.LLMAPIKey)
	}
	if c.LLMModel != "from-secret-model" {
		t.Errorf("LLMModel = %q, want from-secret-model", c.LLMModel)
	}
}

func TestLoad_InvalidConfigSecretFails(t *testing.T) {
	clearEnv(t)
	t.Setenv("CONFIG_SECRET", "not-json")

	if _, err := Load(); err == nil {
		t.Fatal("Load succeeded, want error for invalid CONFIG_SECRET JSON")
	}
}

func TestLoad_EmptyConfigSecretIgnored(t *testing.T) {
	clearEnv(t)
	t.Setenv("CONFIG_SECRET", "")
	t.Setenv("LLM_API_KEY", "k")

	c, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.LLMAPIKey != "k" {
		t.Errorf("LLMAPIKey = %q, want k", c.LLMAPIKey)
	}
}
