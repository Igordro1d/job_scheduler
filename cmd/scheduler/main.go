package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Igordro1d/job_scheduler/internal/api"
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
	reaper := worker.NewReaper(st, 10*time.Second, 30*time.Second)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: api.NewServer(st).Routes(),
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		reaper.Run(ctx)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("http server error: %v", err)
		}
	}()

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(shutdownCtx)
	}()

	log.Println("scheduler started on :8080")
	workers.Run(ctx)
	wg.Wait()
	log.Println("scheduler stopped")
}
