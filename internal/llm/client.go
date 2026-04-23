package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// maxLLMRespBytes caps the upstream response body to protect against
// misbehaving providers streaming unbounded payloads. 2MB accommodates
// ~4000 output tokens plus JSON envelope with generous headroom.
const maxLLMRespBytes = 2 << 20

type Client struct {
	BaseURL        string
	APIKey         string
	Model          string
	MaxTokens      int
	ResponseFormat string // "json_object" | "text" | "none" or "" to omit
	Timeout        time.Duration
	HTTP           *http.Client
}

// NewClient builds a Client with an HTTP transport tuned for a single
// upstream LLM endpoint: HTTP/2, keepalive, bounded idle pool.
func NewClient(baseURL, apiKey, model, responseFormat string, maxTokens int, timeout time.Duration) *Client {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          20,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ForceAttemptHTTP2:     true,
	}
	return &Client{
		BaseURL:        baseURL,
		APIKey:         apiKey,
		Model:          model,
		MaxTokens:      maxTokens,
		ResponseFormat: responseFormat,
		Timeout:        timeout,
		HTTP:           &http.Client{Timeout: timeout + 5*time.Second, Transport: transport},
	}
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type responseFormat struct {
	Type string `json:"type"`
}

type chatRequest struct {
	Model          string          `json:"model"`
	Messages       []chatMessage   `json:"messages"`
	Temperature    float64         `json:"temperature"`
	MaxTokens      int             `json:"max_tokens,omitempty"`
	ResponseFormat *responseFormat `json:"response_format,omitempty"`
}

type chatChoice struct {
	Message chatMessage `json:"message"`
}

type errorEnvelope struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

type chatResponse struct {
	Choices []chatChoice   `json:"choices"`
	Error   *errorEnvelope `json:"error,omitempty"`
}

// firstAttemptBudgetRatio caps how much of the overall deadline the initial
// attempt may consume. Reserves the remainder for the retry path so a slow
// first attempt (common with local models generating long JSON) cannot
// starve the stricter-prompt retry.
const firstAttemptBudgetRatio = 0.65

// Analyze calls an OpenAI-compatible /chat/completions endpoint once. On
// malformed JSON in the model's content field, retries once with a stricter
// reminder prompt. The language parameter constrains the output language for
// free-text fields (LangAuto | LangPT | LangEN | LangES).
//
// Deadline budget: if the caller's context has a deadline (or Client.Timeout
// is applied below), the first attempt runs under a sub-context capped at
// firstAttemptBudgetRatio of the remaining time. The retry path then uses
// the parent deadline directly, so the reserved slice is guaranteed to the
// retry regardless of how long the first attempt took.
func (c *Client) Analyze(ctx context.Context, resumeText, jobDescription, language string) (*AnalysisResult, error) {
	if c.HTTP == nil {
		c.HTTP = &http.Client{}
	}
	// Only apply Client.Timeout if the caller hasn't already set a deadline.
	// Avoids double-wrapping and preserves the caller's remaining budget
	// across both the initial call and the retry path.
	if _, ok := ctx.Deadline(); !ok && c.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.Timeout)
		defer cancel()
	}

	sys, user := BuildPrompt(resumeText, jobDescription, language)

	firstCtx, cancelFirst := firstAttemptContext(ctx)
	res, err := c.tryOnce(firstCtx, sys, user)
	cancelFirst()
	if err == nil {
		return res, nil
	}
	if !errors.Is(err, errInvalidJSON) {
		return nil, err
	}
	// Retry once with stricter prompt, using the full parent deadline.
	res, err = c.tryOnce(ctx, BuildStrictSystemPrompt(language), user)
	if err != nil {
		return nil, fmt.Errorf("llm: invalid response after retry: %w", err)
	}
	return res, nil
}

// firstAttemptContext derives a sub-context whose deadline is at most
// firstAttemptBudgetRatio of parent's remaining time. If parent has no
// deadline, returns parent unchanged with a no-op cancel.
func firstAttemptContext(parent context.Context) (context.Context, context.CancelFunc) {
	deadline, ok := parent.Deadline()
	if !ok {
		return parent, func() {}
	}
	remaining := time.Until(deadline)
	if remaining <= 0 {
		return parent, func() {}
	}
	budget := time.Duration(float64(remaining) * firstAttemptBudgetRatio)
	return context.WithTimeout(parent, budget)
}

var errInvalidJSON = errors.New("invalid JSON in LLM response")

func (c *Client) tryOnce(ctx context.Context, system, user string) (*AnalysisResult, error) {
	reqBody := chatRequest{
		Model: c.Model,
		Messages: []chatMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
		Temperature: 0.3,
		MaxTokens:   c.MaxTokens,
	}
	if rf := c.ResponseFormat; rf != "" && rf != "none" {
		reqBody.ResponseFormat = &responseFormat{Type: rf}
	}
	buf, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("llm: marshal: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		strings.TrimRight(c.BaseURL, "/")+"/chat/completions", bytes.NewReader(buf))
	if err != nil {
		return nil, fmt.Errorf("llm: new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("llm: http: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxLLMRespBytes))
	if err != nil {
		return nil, fmt.Errorf("llm: read body: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("llm: upstream status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var chat chatResponse
	if err := json.Unmarshal(body, &chat); err != nil {
		return nil, fmt.Errorf("llm: decode envelope: %w", err)
	}
	if chat.Error != nil && chat.Error.Message != "" {
		return nil, fmt.Errorf("llm: upstream error: %s", chat.Error.Message)
	}
	if len(chat.Choices) == 0 {
		return nil, fmt.Errorf("llm: empty choices array")
	}
	content := stripFences(chat.Choices[0].Message.Content)

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
