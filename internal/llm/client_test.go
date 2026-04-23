package llm

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

// Ollama /api/chat response envelope (non-streaming).
type ollamaChatResp struct {
	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
	Done bool `json:"done"`
}

func mustLoad(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(b)
}

func respondWith(content string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chat" {
			http.Error(w, "not found", 404)
			return
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), "RESUME") {
			http.Error(w, "prompt missing resume", 400)
			return
		}
		resp := ollamaChatResp{Done: true}
		resp.Message.Role = "assistant"
		resp.Message.Content = content
		_ = json.NewEncoder(w).Encode(resp)
	}
}

func TestAnalyze_SuccessParsesJSON(t *testing.T) {
	content := mustLoad(t, "../../testdata/llm-responses/ok.json")
	srv := httptest.NewServer(respondWith(content))
	defer srv.Close()

	c := &Client{BaseURL: srv.URL, Model: "test", Timeout: 5 * time.Second, HTTP: srv.Client()}
	res, err := c.Analyze(context.Background(), "my resume", "my jd")
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if res.Score != 82 {
		t.Errorf("score = %d, want 82", res.Score)
	}
	if len(res.Missing) != 2 {
		t.Errorf("missing len = %d", len(res.Missing))
	}
	if res.Rewritten.Name != "Jane Doe" {
		t.Errorf("rewritten name = %q", res.Rewritten.Name)
	}
	if len(res.JobKeywords) != 4 {
		t.Errorf("job_keywords len = %d, want 4", len(res.JobKeywords))
	}
	var present int
	for _, k := range res.JobKeywords {
		if k.Present {
			present++
		}
	}
	if present != 2 {
		t.Errorf("present keywords = %d, want 2", present)
	}
	if res.TitleAlignment == "" {
		t.Error("title_alignment empty")
	}
}

func TestAnalyze_RetriesOnMalformedJSON(t *testing.T) {
	bad := mustLoad(t, "../../testdata/llm-responses/malformed.json")
	good := mustLoad(t, "../../testdata/llm-responses/ok.json")

	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		resp := ollamaChatResp{Done: true}
		resp.Message.Role = "assistant"
		if calls == 1 {
			resp.Message.Content = bad
		} else {
			resp.Message.Content = good
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := &Client{BaseURL: srv.URL, Model: "test", Timeout: 5 * time.Second, HTTP: srv.Client()}
	res, err := c.Analyze(context.Background(), "r", "j")
	if err != nil {
		t.Fatalf("Analyze after retry: %v", err)
	}
	if calls != 2 {
		t.Errorf("expected 2 calls, got %d", calls)
	}
	if res.Score != 82 {
		t.Errorf("score = %d", res.Score)
	}
}

func TestAnalyze_FailsAfterTwoBadResponses(t *testing.T) {
	bad := mustLoad(t, "../../testdata/llm-responses/malformed.json")
	srv := httptest.NewServer(respondWith(bad))
	defer srv.Close()

	c := &Client{BaseURL: srv.URL, Model: "test", Timeout: 5 * time.Second, HTTP: srv.Client()}
	_, err := c.Analyze(context.Background(), "r", "j")
	if err == nil {
		t.Fatal("expected error after two malformed responses")
	}
	if !strings.Contains(err.Error(), "invalid") {
		t.Errorf("err = %v", err)
	}
}

func TestAnalyze_PropagatesContextCancel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(2 * time.Second):
		}
	}))
	defer srv.Close()

	c := &Client{BaseURL: srv.URL, Model: "test", Timeout: 100 * time.Millisecond, HTTP: srv.Client()}
	_, err := c.Analyze(context.Background(), "r", "j")
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestAnalyze_ReturnsErrorOnHTTP404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"model not found"}`, 404)
	}))
	defer srv.Close()

	c := &Client{BaseURL: srv.URL, Model: "missing:1b", Timeout: 2 * time.Second, HTTP: srv.Client()}
	_, err := c.Analyze(context.Background(), "r", "j")
	if err == nil {
		t.Fatal("expected 404 error")
	}
	if !strings.Contains(err.Error(), "model") && !strings.Contains(err.Error(), "404") {
		t.Errorf("err = %v", err)
	}
}
