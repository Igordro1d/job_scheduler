package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/Igordro1d/job_scheduler/internal/executor"
	"github.com/Igordro1d/job_scheduler/internal/job"
	"github.com/Igordro1d/job_scheduler/internal/store"
	"github.com/Igordro1d/job_scheduler/internal/worker"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	connString := "postgres://scheduler:scheduler@localhost:5432/scheduler"

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		log.Fatalf("failed to create pool: %v", err)
	}
	defer pool.Close()

	err = pool.Ping(ctx)
	if err != nil {
		log.Fatalf("ping failed: %v", err)
	}

	st := store.NewPostgres(pool)

	registry := executor.NewRegistry()
	registry.Register("print", func(ctx context.Context, j *job.Job) error {
		log.Printf("running job %s with payload %s", j.ID, j.Payload)
		return nil
	})

	workers := worker.NewPool(st, registry, 4, time.Second)

	log.Println("scheduler started")
	workers.Run(ctx)
	log.Println("scheduler stopped")
}
