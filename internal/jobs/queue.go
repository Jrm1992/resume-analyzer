package jobs

import (
	"context"
	"errors"
	"sync"
)

var ErrQueueFull = errors.New("jobs: queue full")

type Handler func(ctx context.Context, j *Job)

type Queue struct {
	ch      chan *Job
	workers int
	wg      sync.WaitGroup
}

func NewQueue(workers, capacity int) *Queue {
	return &Queue{
		ch:      make(chan *Job, capacity),
		workers: workers,
	}
}

func (q *Queue) Enqueue(j *Job) error {
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

// Wait blocks until all workers have returned. Callers should cancel ctx first.
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
