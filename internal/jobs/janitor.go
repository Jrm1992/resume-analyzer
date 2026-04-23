package jobs

import (
	"context"
	"log/slog"
	"time"
)

type Janitor struct {
	store    *Store
	ttl      time.Duration
	interval time.Duration
}

func NewJanitor(store *Store, ttl, interval time.Duration) *Janitor {
	return &Janitor{store: store, ttl: ttl, interval: interval}
}

func (j *Janitor) Start(ctx context.Context) {
	go func() {
		t := time.NewTicker(j.interval)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				n := j.store.DeleteOlderThan(j.ttl)
				if n > 0 {
					slog.Info("janitor: cleaned expired jobs", "count", n)
				}
			}
		}
	}()
}
