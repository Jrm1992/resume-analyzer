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
	BaseURL   string
	APIKey    string
	Model     string
	MaxTokens int
	Timeout   time.Duration
	HTTP      *http.Client
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

// Analyze calls an OpenAI-compatible /chat/completions endpoint once. On
// malformed JSON in the model's content field, retries once with a stricter
// reminder prompt. Returns the parsed AnalysisResult.
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
	// Retry once with stricter prompt.
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
		Temperature:    0.3,
		MaxTokens:      c.MaxTokens,
		ResponseFormat: &responseFormat{Type: "json_object"},
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

	body, err := io.ReadAll(resp.Body)
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
