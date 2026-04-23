package jobs

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
)

var (
	ErrQueueFull   = errors.New("jobs: queue full")
	ErrQueueClosed = errors.New("jobs: queue closed")
)

type Handler func(ctx context.Context, j *Job)

type Queue struct {
	ch        chan *Job
	workers   int
	wg        sync.WaitGroup
	closed    atomic.Bool
	closeOnce sync.Once
}

func NewQueue(workers, capacity int) *Queue {
	return &Queue{
		ch:      make(chan *Job, capacity),
		workers: workers,
	}
}

func (q *Queue) Enqueue(j *Job) error {
	if q.closed.Load() {
		return ErrQueueClosed
	}
	select {
	case q.ch <- j:
		return nil
	default:
		return ErrQueueFull
	}
}

func (q *Queue) Start(ctx context.Context, handler Handler) {
	for i := 0; i < q.workers; i++ {
		q.wg.Add(1)
		go q.run(ctx, handler)
	}
}

// Close stops accepting new jobs and closes the channel so in-flight
// buffered jobs drain before workers exit. Safe to call multiple times.
func (q *Queue) Close() {
	q.closeOnce.Do(func() {
		q.closed.Store(true)
		close(q.ch)
	})
}

// Wait blocks until all workers have returned. Callers should cancel ctx
// (or call Close) first.
func (q *Queue) Wait() { q.wg.Wait() }

func (q *Queue) run(ctx context.Context, handler Handler) {
	defer q.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case j, ok := <-q.ch:
			if !ok {
				return
			}
			handler(ctx, j)
		}
	}
}
