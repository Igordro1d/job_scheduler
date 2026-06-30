package store

import (
	"context"

	"github.com/Igordro1d/job_scheduler/internal/job"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const jobColumns = `id, type, payload, priority, status, depends_on,
	max_retries, retry_count, idempotency_key, locked_by, locked_at,
	run_after, created_at, updated_at`

type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(pool *pgxpool.Pool) *Postgres {
	return &Postgres{pool: pool}
}

func (p *Postgres) Enqueue(ctx context.Context, params job.EnqueueParams) (*job.Job, error) {
	query := `INSERT INTO jobs (type, payload, priority, depends_on, idempotency_key)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING` + jobColumns

	row := p.pool.QueryRow(ctx, query,
		params.Type, params.Payload, params.Priority, params.DependsOn, params.IdempotencyKey)

	return scanJob(row)
}

func (p *Postgres) GetByID(ctx context.Context, id string) (*job.Job, error) {
	query := `SELECT ` + jobColumns + ` FROM jobs WHERE id = $1`

	row := p.pool.QueryRow(ctx, query, id)

	return scanJob(row)
}

func scanJob(row pgx.Row) (*job.Job, error) {
	var j job.Job
	err := row.Scan(
		&j.ID, &j.Type, &j.Payload, &j.Priority, &j.Status, &j.DependsOn,
		&j.MaxRetries, &j.RetryCount, &j.IdempotencyKey, &j.LockedBy, &j.LockedAt,
		&j.RunAfter, &j.CreatedAt, &j.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &j, nil
}
