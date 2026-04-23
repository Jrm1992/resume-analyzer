package config

import (
	"testing"
	"time"
)

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("OLLAMA_URL", "")
	t.Setenv("OLLAMA_MODEL", "")
	t.Setenv("MAX_PDF_MB", "")
	t.Setenv("LLM_TIMEOUT_SEC", "")
	t.Setenv("WORKERS", "")
	t.Setenv("QUEUE_CAPACITY", "")
	t.Setenv("JOB_TTL_MIN", "")

	c, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.Port != 8080 {
		t.Errorf("Port = %d, want 8080", c.Port)
	}
	if c.OllamaURL != "http://localhost:11434" {
		t.Errorf("OllamaURL = %q", c.OllamaURL)
	}
	if c.OllamaModel != "llama3.1:8b" {
		t.Errorf("OllamaModel = %q", c.OllamaModel)
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
	t.Setenv("PORT", "")
	t.Setenv("OLLAMA_URL", "")
	t.Setenv("OLLAMA_MODEL", "")
	t.Setenv("MAX_PDF_MB", "")
	t.Setenv("LLM_TIMEOUT_SEC", "")
	t.Setenv("WORKERS", "")
	t.Setenv("QUEUE_CAPACITY", "")
	t.Setenv("JOB_TTL_MIN", "")

	t.Setenv("PORT", "9000")
	t.Setenv("OLLAMA_MODEL", "mistral:7b")
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
	if c.OllamaModel != "mistral:7b" {
		t.Errorf("OllamaModel = %q", c.OllamaModel)
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

func TestLoad_InvalidInt(t *testing.T) {
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
	}
	for _, c := range cases {
		t.Run(c.env+"="+c.val, func(t *testing.T) {
			t.Setenv(c.env, c.val)
			if _, err := Load(); err == nil {
				t.Fatalf("%s=%s: expected error, got nil", c.env, c.val)
			}
		})
	}
}

func TestLoad_WhitespaceOllamaURL_FallsBackToDefault(t *testing.T) {
	t.Setenv("OLLAMA_URL", "   ")
	c, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.OllamaURL != "http://localhost:11434" {
		t.Errorf("OllamaURL = %q, want default", c.OllamaURL)
	}
}
