package store

import (
	"context"
	"time"

	"github.com/Igordro1d/job_scheduler/internal/job"
)

type Store interface {
	Enqueue(ctx context.Context, params job.EnqueueParams) (*job.Job, error)
	GetByID(ctx context.Context, id string) (*job.Job, error)
	ListDeadLetter(ctx context.Context) ([]*job.Job, error)
	ListRecent(ctx context.Context, limit int) ([]*job.Job, error)
	Claim(ctx context.Context, workerID string) (*job.Job, error)
	MarkCompleted(ctx context.Context, id string) error
	MarkFailed(ctx context.Context, id string) error
	ReclaimStale(ctx context.Context, timeout time.Duration) (int64, error)
}
