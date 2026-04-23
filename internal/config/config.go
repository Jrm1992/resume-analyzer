package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port          int
	OllamaURL     string
	OllamaModel   string
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
	return &Config{
		Port:          port,
		OllamaURL:     getStr("OLLAMA_URL", "http://localhost:11434"),
		OllamaModel:   getStr("OLLAMA_MODEL", "llama3.1:8b"),
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
