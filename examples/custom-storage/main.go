package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/w0rng/gofeat"

	pgstore "github.com/w0rng/gofeat/examples/custom-storage/postgres"
)

func main() {
	ctx := context.Background()

	container, err := postgres.Run(
		ctx,
		"postgres:16",
		postgres.WithDatabase("gofeat"),
		postgres.WithUsername("user"),
		postgres.WithPassword("password"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer container.Terminate(ctx)

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := migrate(db); err != nil {
		log.Fatal(err)
	}

	store, err := gofeat.New(
		gofeat.Config{
			Features: []gofeat.Feature{
				{Name: "sum", Aggregate: gofeat.Sum("value"), Window: gofeat.Sliding(time.Hour)},
			},
			Storage: pgstore.New(db, time.Hour),
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	now := time.Now().UTC()

	err = store.Push(ctx, "user-1",
		gofeat.Event{now, map[string]any{"value": 1}},
		gofeat.Event{now, map[string]any{"value": 2}},
		gofeat.Event{now, map[string]any{"value": 3}},
	)
	if err != nil {
		log.Fatal(err)
	}

	result, err := store.Get(
		ctx,
		"user-1",
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("sum:", result.FloatOr("sum", -1))
}

func migrate(db *sql.DB) error {
	const query = `
		CREATE TABLE IF NOT EXISTS events (
			entity TEXT NOT NULL,
			ts TIMESTAMPTZ NOT NULL,
			data JSONB NOT NULL
		);

		CREATE INDEX IF NOT EXISTS idx_events_entity_ts ON events(entity, ts);`

	_, err := db.Exec(query)
	return err
}
