package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	BaseURL string
	Model   string
	Timeout time.Duration
	HTTP    *http.Client
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
	Format   string        `json:"format,omitempty"`
}

type chatResponse struct {
	Message chatMessage `json:"message"`
	Done    bool        `json:"done"`
	Error   string      `json:"error,omitempty"`
}

// Analyze calls Ollama /api/chat once. On malformed JSON, retries once
// with a stricter reminder prompt. Returns the parsed AnalysisResult.
func (c *Client) Analyze(ctx context.Context, resumeText, jobDescription string) (*AnalysisResult, error) {
	if c.HTTP == nil {
		c.HTTP = &http.Client{}
	}
	ctx, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()

	sys, user := BuildPrompt(resumeText, jobDescription)

	res, err := c.tryOnce(ctx, sys, user)
	if err == nil {
		return res, nil
	}
	if !errors.Is(err, errInvalidJSON) {
		return nil, err
	}
	// Retry once with strict prompt.
	res, err = c.tryOnce(ctx, BuildStrictSystemPrompt(), user)
	if err != nil {
		return nil, fmt.Errorf("llm: invalid response after retry: %w", err)
	}
	return res, nil
}

var errInvalidJSON = errors.New("invalid JSON in LLM response")

func (c *Client) tryOnce(ctx context.Context, system, user string) (*AnalysisResult, error) {
	reqBody := chatRequest{
		Model: c.Model,
		Messages: []chatMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
		Stream: false,
		Format: "json",
	}
	buf, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("llm: marshal: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		strings.TrimRight(c.BaseURL, "/")+"/api/chat", bytes.NewReader(buf))
	if err != nil {
		return nil, fmt.Errorf("llm: new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("llm: http: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("llm: read body: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("llm: ollama status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var chat chatResponse
	if err := json.Unmarshal(body, &chat); err != nil {
		return nil, fmt.Errorf("llm: decode envelope: %w", err)
	}
	if chat.Error != "" {
		return nil, fmt.Errorf("llm: ollama error: %s", chat.Error)
	}
	content := stripFences(chat.Message.Content)

	var out AnalysisResult
	if err := json.Unmarshal([]byte(content), &out); err != nil {
		return nil, fmt.Errorf("%w: %v", errInvalidJSON, err)
	}
	return &out, nil
}

// stripFences removes ```json ... ``` fences and leading prose before first '{'.
func stripFences(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.Index(s, "{"); i > 0 {
		s = s[i:]
	}
	if j := strings.LastIndex(s, "}"); j >= 0 {
		s = s[:j+1]
	}
	return s
}
