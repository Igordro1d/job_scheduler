package store

import (
	"context"
	"errors"
	"time"

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
	if params.Payload == nil {
		return nil, errors.New("payload is required")
	}

	query := `INSERT INTO jobs (type, payload, priority, depends_on, idempotency_key)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING ` + jobColumns

	row := p.pool.QueryRow(ctx, query,
		params.Type, params.Payload, params.Priority, params.DependsOn, params.IdempotencyKey)

	return scanJob(row)
}

func (p *Postgres) GetByID(ctx context.Context, id string) (*job.Job, error) {
	query := `SELECT ` + jobColumns + ` FROM jobs WHERE id = $1`

	row := p.pool.QueryRow(ctx, query, id)

	return scanJob(row)
}

func (p *Postgres) Claim(ctx context.Context, workerID string) (*job.Job, error) {
	query := `UPDATE jobs
		SET status = 'in_progress', locked_by = $1, locked_at = now(), updated_at = now()
		WHERE id = (
			SELECT j.id FROM jobs j
			WHERE j.status = 'pending'
			  AND (j.run_after IS NULL OR j.run_after <= now())
			  AND (j.depends_on IS NULL OR EXISTS (
				SELECT 1 FROM jobs parent
				WHERE parent.id = j.depends_on AND parent.status = 'completed'))
			ORDER BY j.priority DESC, j.created_at
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		)
		RETURNING ` + jobColumns

	row := p.pool.QueryRow(ctx, query, workerID)

	claimed, err := scanJob(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return claimed, nil
}

func (p *Postgres) MarkCompleted(ctx context.Context, id string) error {
	query := `UPDATE jobs
		SET status = 'completed', locked_by = NULL, locked_at = NULL, updated_at = now()
		WHERE id = $1`

	_, err := p.pool.Exec(ctx, query, id)
	return err
}

func (p *Postgres) MarkFailed(ctx context.Context, id string) error {
	query := `UPDATE jobs
		SET status = 'failed', locked_by = NULL, locked_at = NULL, updated_at = now()
		WHERE id = $1`

	_, err := p.pool.Exec(ctx, query, id)
	return err
}

func (p *Postgres) ReclaimStale(ctx context.Context, timeout time.Duration) (int64, error) {
	query := `UPDATE jobs
		SET status = 'pending', locked_by = NULL, locked_at = NULL, updated_at = now()
		WHERE status = 'in_progress'
		  AND locked_at < now() - make_interval(secs => $1)`

	tag, err := p.pool.Exec(ctx, query, timeout.Seconds())
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
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
