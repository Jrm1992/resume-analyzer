package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port          int
	LLMBaseURL    string
	LLMAPIKey     string
	LLMModel      string
	LLMMaxTokens  int
	MaxPDFBytes   int64
	LLMTimeout    time.Duration
	Workers       int
	QueueCapacity int
	JobTTL        time.Duration
}

func Load() (*Config, error) {
	port, err := getInt("PORT", 8080)
	if err != nil {
		return nil, err
	}
	maxMB, err := getInt("MAX_PDF_MB", 10)
	if err != nil {
		return nil, err
	}
	timeoutSec, err := getInt("LLM_TIMEOUT_SEC", 120)
	if err != nil {
		return nil, err
	}
	workers, err := getInt("WORKERS", 2)
	if err != nil {
		return nil, err
	}
	queueCap, err := getInt("QUEUE_CAPACITY", 100)
	if err != nil {
		return nil, err
	}
	ttlMin, err := getInt("JOB_TTL_MIN", 60)
	if err != nil {
		return nil, err
	}
	maxTokens, err := getInt("LLM_MAX_TOKENS", 4000)
	if err != nil {
		return nil, err
	}

	apiKey := getStr("LLM_API_KEY", "")
	if apiKey == "" {
		return nil, errors.New("env LLM_API_KEY: required (set to your provider API key)")
	}

	baseURL := strings.TrimRight(getStr("LLM_BASE_URL", "https://api.openai.com/v1"), "/")

	return &Config{
		Port:          port,
		LLMBaseURL:    baseURL,
		LLMAPIKey:     apiKey,
		LLMModel:      getStr("LLM_MODEL", "gpt-4o-mini"),
		LLMMaxTokens:  maxTokens,
		MaxPDFBytes:   int64(maxMB) * 1024 * 1024,
		LLMTimeout:    time.Duration(timeoutSec) * time.Second,
		Workers:       workers,
		QueueCapacity: queueCap,
		JobTTL:        time.Duration(ttlMin) * time.Minute,
	}, nil
}

func getStr(key, def string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	return v
}

func getInt(key string, def int) (int, error) {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("env %s: %w", key, err)
	}
	if n <= 0 {
		return 0, fmt.Errorf("env %s: must be positive, got %d", key, n)
	}
	return n, nil
}
