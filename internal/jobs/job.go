package jobs

import (
	"time"

	"github.com/jose/resume-analyzer/internal/llm"
)

type Status string

const (
	StatusQueued  Status = "queued"
	StatusRunning Status = "running"
	StatusDone    Status = "done"
	StatusFailed  Status = "failed"
)

type Job struct {
	ID        string
	Status    Status
	CreatedAt time.Time
	UpdatedAt time.Time
	Resume    string
	JD        string
	Language  string // "" (auto), "pt", "en", "es"
	Result    *llm.AnalysisResult
	Err       string
}
