package worker

import (
	"context"
	"log"
	"time"

	"github.com/Igordro1d/job_scheduler/internal/store"
)

type Reaper struct {
	store    store.Store
	interval time.Duration
	timeout  time.Duration
}

func NewReaper(s store.Store, interval, timeout time.Duration) *Reaper {
	return &Reaper{
		store:    s,
		interval: interval,
		timeout:  timeout,
	}
}

func (r *Reaper) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(r.interval):
		}

		reclaimed, err := r.store.ReclaimStale(ctx, r.timeout)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("reaper: reclaim error: %v", err)
			continue
		}

		if reclaimed > 0 {
			log.Printf("reaper: reclaimed %d stale job(s)", reclaimed)
		}
	}
}
