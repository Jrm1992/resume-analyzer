package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	nethttp "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jose/resume-analyzer/internal/config"
	apphttp "github.com/jose/resume-analyzer/internal/http"
	"github.com/jose/resume-analyzer/internal/jobs"
	"github.com/jose/resume-analyzer/internal/llm"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run() error {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	tpl, err := apphttp.LoadTemplates()
	if err != nil {
		return fmt.Errorf("templates: %w", err)
	}

	store := jobs.NewStore()
	queue := jobs.NewQueue(cfg.Workers, cfg.QueueCapacity)
	janitor := jobs.NewJanitor(store, cfg.JobTTL, 5*time.Minute)

	llmClient := &llm.Client{
		BaseURL:   cfg.LLMBaseURL,
		APIKey:    cfg.LLMAPIKey,
		Model:     cfg.LLMModel,
		MaxTokens: cfg.LLMMaxTokens,
		Timeout:   cfg.LLMTimeout,
		HTTP:      &nethttp.Client{Timeout: cfg.LLMTimeout + 5*time.Second},
	}

	srv := &apphttp.Server{
		Config:    cfg,
		Templates: tpl,
		Store:     store,
		Queue:     queue,
		Analyzer:  llmClient,
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	queue.Start(ctx, srv.JobHandler)
	janitor.Start(ctx)

	httpSrv := &nethttp.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           srv.Router(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("listening", "addr", httpSrv.Addr, "model", cfg.LLMModel, "base_url", cfg.LLMBaseURL)
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, nethttp.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		slog.Info("shutdown signal received")
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("http: %w", err)
		}
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.LLMTimeout+5*time.Second)
	defer shutdownCancel()
	_ = httpSrv.Shutdown(shutdownCtx)

	cancel() // signal workers to stop
	queue.Wait()
	slog.Info("shutdown complete")
	return nil
}
