package store

import (
	"context"

	"github.com/Igordro1d/job_scheduler/internal/job"
)

type Store interface {
	Enqueue(ctx context.Context, params job.EnqueueParams) (*job.Job, error)
	GetByID(ctx context.Context, id string) (*job.Job, error)
	Claim(ctx context.Context, workerID string) (*job.Job, error)
	MarkCompleted(ctx context.Context, id string) error
	MarkFailed(ctx context.Context, id string) error
}
