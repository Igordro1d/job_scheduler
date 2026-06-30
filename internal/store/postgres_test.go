package store

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/Igordro1d/job_scheduler/internal/job"
	"github.com/jackc/pgx/v5/pgxpool"
)

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		url = "postgres://scheduler:scheduler@localhost:5432/scheduler"
	}

	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		t.Fatalf("create pool: %v", err)
	}

	err = pool.Ping(context.Background())
	if err != nil {
		pool.Close()
		t.Skipf("no database available: %v", err)
	}

	_, err = pool.Exec(context.Background(), "TRUNCATE jobs")
	if err != nil {
		pool.Close()
		t.Fatalf("truncate: %v", err)
	}

	return pool
}

func TestEnqueueAndGet(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	defer pool.Close()

	st := NewPostgres(pool)

	created, err := st.Enqueue(ctx, job.EnqueueParams{
		Type:     "send_email",
		Payload:  json.RawMessage(`{"to":"a@b.com"}`),
		Priority: 5,
	})
	if err != nil {
		t.Fatalf("enqueue: %v", err)
	}

	if created.ID == "" {
		t.Error("expected generated id, got empty")
	}
	if created.Status != job.StatusPending {
		t.Errorf("expected status pending, got %q", created.Status)
	}
	if created.MaxRetries != 3 {
		t.Errorf("expected default max_retries 3, got %d", created.MaxRetries)
	}

	got, err := st.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	if got.ID != created.ID {
		t.Errorf("id mismatch: got %q want %q", got.ID, created.ID)
	}
	if got.Type != "send_email" {
		t.Errorf("type mismatch: got %q", got.Type)
	}
	if got.Priority != 5 {
		t.Errorf("priority mismatch: got %d", got.Priority)
	}
}
