package jobs

import (
	"context"
	"testing"
	"time"
)

func TestJanitor_DeletesExpiredJobs(t *testing.T) {
	s := NewStore()
	j1 := s.Create("r", "j", "", "")
	// Age j1.
	s.Update(j1.ID, func(j *Job) { j.CreatedAt = time.Now().Add(-10 * time.Minute) })
	_ = s.Create("r2", "j2", "", "")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	j := NewJanitor(s, 5*time.Minute, 20*time.Millisecond)
	j.Start(ctx)

	time.Sleep(60 * time.Millisecond)
	if _, ok := s.Get(j1.ID); ok {
		t.Error("expected j1 to be cleaned")
	}
}
