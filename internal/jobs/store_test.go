package jobs

import (
	"sync"
	"testing"
	"time"
)

func TestStore_CreateAndGet(t *testing.T) {
	s := NewStore()
	j := s.Create("resume text", "jd text", "")
	if j.ID == "" {
		t.Fatal("empty ID")
	}
	if j.Status != StatusQueued {
		t.Errorf("status = %q", j.Status)
	}
	got, ok := s.Get(j.ID)
	if !ok {
		t.Fatal("not found")
	}
	if got.Resume != "resume text" {
		t.Errorf("resume = %q", got.Resume)
	}
}

func TestStore_Update(t *testing.T) {
	s := NewStore()
	j := s.Create("r", "j", "")
	s.Update(j.ID, func(j *Job) {
		j.Status = StatusRunning
	})
	got, _ := s.Get(j.ID)
	if got.Status != StatusRunning {
		t.Errorf("status = %q", got.Status)
	}
	if got.UpdatedAt.Before(got.CreatedAt) {
		t.Error("UpdatedAt not advanced")
	}
}

func TestStore_GetMissing(t *testing.T) {
	s := NewStore()
	if _, ok := s.Get("nope"); ok {
		t.Error("expected miss")
	}
}

func TestStore_DeleteOlderThan(t *testing.T) {
	s := NewStore()
	j1 := s.Create("a", "b", "")
	j2 := s.Create("c", "d", "")
	// age j1 artificially
	s.Update(j1.ID, func(j *Job) { j.CreatedAt = time.Now().Add(-2 * time.Hour) })

	n := s.DeleteOlderThan(time.Hour)
	if n != 1 {
		t.Errorf("deleted %d, want 1", n)
	}
	if _, ok := s.Get(j1.ID); ok {
		t.Error("j1 should be deleted")
	}
	if _, ok := s.Get(j2.ID); !ok {
		t.Error("j2 should remain")
	}
}

func TestStore_ConcurrentAccess(t *testing.T) {
	s := NewStore()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			j := s.Create("r", "j", "")
			s.Update(j.ID, func(j *Job) { j.Status = StatusRunning })
			_, _ = s.Get(j.ID)
		}()
	}
	wg.Wait()
}
