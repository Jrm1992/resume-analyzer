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

// OpenAI /chat/completions response envelope (non-streaming).
type openaiChoice struct {
	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
}
type openaiChatResp struct {
	Choices []openaiChoice `json:"choices"`
}

func mustLoad(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(b)
}

// respondWith returns a handler that serves /chat/completions with the given
// inner model-message content. It also asserts Bearer auth + path + body shape.
func respondWith(content string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			http.Error(w, "not found", 404)
			return
		}
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
			http.Error(w, "missing bearer", 401)
			return
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), "RESUME") {
			http.Error(w, "prompt missing resume", 400)
			return
		}
		var resp openaiChatResp
		resp.Choices = []openaiChoice{{}}
		resp.Choices[0].Message.Role = "assistant"
		resp.Choices[0].Message.Content = content
		_ = json.NewEncoder(w).Encode(resp)
	}
}

func testClient(baseURL string, httpc *http.Client, timeout time.Duration) *Client {
	return &Client{
		BaseURL:        baseURL,
		APIKey:         "test-key",
		Model:          "test-model",
		MaxTokens:      1024,
		ResponseFormat: "json_object",
		Timeout:        timeout,
		HTTP:           httpc,
	}
}

func TestAnalyze_SuccessParsesJSON(t *testing.T) {
	content := mustLoad(t, "../../testdata/llm-responses/ok.json")
	srv := httptest.NewServer(respondWith(content))
	defer srv.Close()

	c := testClient(srv.URL, srv.Client(), 5*time.Second)
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

func TestAnalyze_SendsBearerTokenAndJSONFormat(t *testing.T) {
	var gotAuth, gotCT string
	var bodyStr string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotCT = r.Header.Get("Content-Type")
		b, _ := io.ReadAll(r.Body)
		bodyStr = string(b)
		var resp openaiChatResp
		resp.Choices = []openaiChoice{{}}
		resp.Choices[0].Message.Content = mustLoad(t, "../../testdata/llm-responses/ok.json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := testClient(srv.URL, srv.Client(), 5*time.Second)
	if _, err := c.Analyze(context.Background(), "rrr", "jjj"); err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if gotAuth != "Bearer test-key" {
		t.Errorf("Authorization = %q", gotAuth)
	}
	if gotCT != "application/json" {
		t.Errorf("Content-Type = %q", gotCT)
	}
	if !strings.Contains(bodyStr, `"response_format":{"type":"json_object"}`) {
		t.Errorf("body missing response_format: %s", bodyStr)
	}
	if !strings.Contains(bodyStr, `"model":"test-model"`) {
		t.Errorf("body missing model: %s", bodyStr)
	}
	if !strings.Contains(bodyStr, `"max_tokens":1024`) {
		t.Errorf("body missing max_tokens: %s", bodyStr)
	}
}

func TestAnalyze_ResponseFormatNone_OmitsField(t *testing.T) {
	var bodyStr string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		bodyStr = string(b)
		var resp openaiChatResp
		resp.Choices = []openaiChoice{{}}
		resp.Choices[0].Message.Content = mustLoad(t, "../../testdata/llm-responses/ok.json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := testClient(srv.URL, srv.Client(), 5*time.Second)
	c.ResponseFormat = "none"
	if _, err := c.Analyze(context.Background(), "r", "j"); err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if strings.Contains(bodyStr, `"response_format"`) {
		t.Errorf("body should NOT contain response_format: %s", bodyStr)
	}
}

func TestAnalyze_RetriesOnMalformedJSON(t *testing.T) {
	bad := mustLoad(t, "../../testdata/llm-responses/malformed.json")
	good := mustLoad(t, "../../testdata/llm-responses/ok.json")

	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		var resp openaiChatResp
		resp.Choices = []openaiChoice{{}}
		resp.Choices[0].Message.Role = "assistant"
		if calls == 1 {
			resp.Choices[0].Message.Content = bad
		} else {
			resp.Choices[0].Message.Content = good
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := testClient(srv.URL, srv.Client(), 5*time.Second)
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

	c := testClient(srv.URL, srv.Client(), 5*time.Second)
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

	c := testClient(srv.URL, srv.Client(), 100*time.Millisecond)
	_, err := c.Analyze(context.Background(), "r", "j")
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestAnalyze_ReturnsErrorOnHTTP401(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":{"message":"invalid api key","type":"auth_error"}}`, 401)
	}))
	defer srv.Close()

	c := testClient(srv.URL, srv.Client(), 2*time.Second)
	_, err := c.Analyze(context.Background(), "r", "j")
	if err == nil {
		t.Fatal("expected 401 error")
	}
	if !strings.Contains(err.Error(), "401") && !strings.Contains(err.Error(), "api key") {
		t.Errorf("err = %v", err)
	}
}

func TestAnalyze_ReturnsErrorOnUpstreamErrorField(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"error":{"message":"rate limit exceeded","type":"rate_limit"}}`))
	}))
	defer srv.Close()

	c := testClient(srv.URL, srv.Client(), 2*time.Second)
	_, err := c.Analyze(context.Background(), "r", "j")
	if err == nil {
		t.Fatal("expected error from upstream error field")
	}
	if !strings.Contains(err.Error(), "rate limit") {
		t.Errorf("err = %v", err)
	}
}
