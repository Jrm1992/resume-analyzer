package jobs

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestQueue_ProcessesJobs(t *testing.T) {
	q := NewQueue(2, 4) // 2 workers, capacity 4
	var processed int32
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	q.Start(ctx, func(ctx context.Context, j *Job) {
		atomic.AddInt32(&processed, 1)
	})

	s := NewStore()
	for i := 0; i < 4; i++ {
		j := s.Create("r", "j", "")
		if err := q.Enqueue(j); err != nil {
			t.Fatalf("enqueue: %v", err)
		}
	}
	// Wait for drain.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if atomic.LoadInt32(&processed) == 4 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("processed %d/4", atomic.LoadInt32(&processed))
}

func TestQueue_EnqueueFullReturnsError(t *testing.T) {
	q := NewQueue(0, 1) // 0 workers — nothing drains; capacity 1
	if err := q.Enqueue(&Job{ID: "a"}); err != nil {
		t.Fatalf("first enqueue: %v", err)
	}
	err := q.Enqueue(&Job{ID: "b"})
	if !errors.Is(err, ErrQueueFull) {
		t.Errorf("err = %v, want ErrQueueFull", err)
	}
}

func TestQueue_StopOnContextCancel(t *testing.T) {
	q := NewQueue(1, 2)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	q.Start(ctx, func(ctx context.Context, j *Job) {
		<-ctx.Done()
		close(done)
	})
	s := NewStore()
	j := s.Create("r", "j", "")
	_ = q.Enqueue(j)
	time.Sleep(10 * time.Millisecond) // Give worker time to dequeue the job
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("worker did not exit after ctx cancel")
	}
	q.Wait()
}
