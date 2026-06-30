package worker

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Igordro1d/job_scheduler/internal/executor"
	"github.com/Igordro1d/job_scheduler/internal/job"
	"github.com/Igordro1d/job_scheduler/internal/store"
)

type Pool struct {
	store        store.Store
	registry     *executor.Registry
	numWorkers   int
	pollInterval time.Duration
}

func NewPool(s store.Store, r *executor.Registry, numWorkers int, pollInterval time.Duration) *Pool {
	return &Pool{
		store:        s,
		registry:     r,
		numWorkers:   numWorkers,
		pollInterval: pollInterval,
	}
}

func (p *Pool) Run(ctx context.Context) {
	var wg sync.WaitGroup

	for i := 0; i < p.numWorkers; i++ {
		wg.Add(1)
		workerID := fmt.Sprintf("worker-%d", i)
		go func() {
			defer wg.Done()
			p.work(ctx, workerID)
		}()
	}

	wg.Wait()
}

func (p *Pool) work(ctx context.Context, workerID string) {
	for {
		if ctx.Err() != nil {
			return
		}

		claimed, err := p.store.Claim(ctx, workerID)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("%s: claim error: %v", workerID, err)
			p.wait(ctx)
			continue
		}

		if claimed == nil {
			p.wait(ctx)
			continue
		}

		p.process(workerID, claimed)
	}
}

func (p *Pool) process(workerID string, j *job.Job) {
	ctx := context.Background()

	handler, ok := p.registry.Get(j.Type)
	if !ok {
		log.Printf("%s: no handler for type %q, marking failed", workerID, j.Type)
		p.markFailed(ctx, workerID, j)
		return
	}

	err := handler(ctx, j)
	if err != nil {
		log.Printf("%s: job %s (%s) failed: %v", workerID, j.ID, j.Type, err)
		p.markFailed(ctx, workerID, j)
		return
	}

	err = p.store.MarkCompleted(ctx, j.ID)
	if err != nil {
		log.Printf("%s: mark completed error for %s: %v", workerID, j.ID, err)
	}
}

func (p *Pool) markFailed(ctx context.Context, workerID string, j *job.Job) {
	err := p.store.MarkFailed(ctx, j.ID)
	if err != nil {
		log.Printf("%s: mark failed error for %s: %v", workerID, j.ID, err)
	}
}

func (p *Pool) wait(ctx context.Context) {
	select {
	case <-ctx.Done():
	case <-time.After(p.pollInterval):
	}
}
