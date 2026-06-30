package main

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	connString := "postgres://scheduler:scheduler@localhost:5432/scheduler"

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		log.Fatalf("failed to create pool: %v", err)
	}

	defer pool.Close()

	err = pool.Ping(ctx)
	if err != nil {
		log.Fatalf("ping failed: %v", err)
	}
}
