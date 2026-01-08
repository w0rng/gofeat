package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pkg/errors"
	"github.com/w0rng/gofeat"
)

type PostgresStorage struct {
	db  *sql.DB
	ttl time.Duration
}

func New(db *sql.DB, ttl time.Duration) *PostgresStorage {
	return &PostgresStorage{db: db, ttl: ttl}
}

// Push inserts events into PostgreSQL.
func (s *PostgresStorage) Push(ctx context.Context, entityID string, events ...gofeat.Event) error {
	const query = `
		INSERT INTO events (entity, ts, data)
		VALUES ($1, $2, $3)
	`

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "begin tx")
	}
	defer tx.Rollback()

	for _, e := range events {
		dataJSON, err := json.Marshal(e.Data)
		if err != nil {
			return errors.Wrap(err, "marshal event data")
		}

		if _, err := tx.ExecContext(ctx, query, entityID, e.Timestamp, dataJSON); err != nil {
			return errors.Wrap(err, "insert event")
		}
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "commit tx")
	}

	return nil
}

// Get returns events for an entity filtered by TTL relative to the given time.
func (s *PostgresStorage) Get(ctx context.Context, entityID string, at time.Time) ([]gofeat.Event, error) {
	var query string
	var args []interface{}

	if s.ttl > 0 {
		cutoff := at.Add(-s.ttl)
		query = `
			SELECT ts, data
			FROM events
			WHERE entity = $1 AND ts > $2 AND ts <= $3
			ORDER BY ts ASC
		`
		args = []interface{}{entityID, cutoff, at}
	} else {
		query = `
			SELECT ts, data
			FROM events
			WHERE entity = $1 AND ts <= $2
			ORDER BY ts ASC
		`
		args = []interface{}{entityID, at}
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "query events")
	}
	defer rows.Close()

	var events []gofeat.Event
	for rows.Next() {
		var e gofeat.Event

		var dataJSON []byte
		if err := rows.Scan(&e.Timestamp, &dataJSON); err != nil {
			return nil, errors.Wrap(err, "scan event")
		}

		if err := json.Unmarshal(dataJSON, &e.Data); err != nil {
			return nil, errors.Wrap(err, "unmarshal event data")
		}

		events = append(events, e)
	}

	return events, nil
}

// Evict removes events older than TTL from current time.
func (s *PostgresStorage) Evict(ctx context.Context) error {
	if s.ttl == 0 {
		return nil
	}

	before := time.Now().UTC().Add(-s.ttl)
	const query = `DELETE FROM events WHERE ts < $1`
	if _, err := s.db.ExecContext(ctx, query, before); err != nil {
		return errors.Wrap(err, "evict events")
	}
	return nil
}

// Stats returns simple storage statistics.
func (s *PostgresStorage) Stats(ctx context.Context) (gofeat.StorageStats, error) {
	const query = `SELECT COUNT(*) FROM events`
	var count int64
	if err := s.db.QueryRowContext(ctx, query).Scan(&count); err != nil {
		return gofeat.StorageStats{}, errors.Wrap(err, "stats")
	}
	return gofeat.StorageStats{TotalEvents: count}, nil
}

// Close closes DB connection.
func (s *PostgresStorage) Close() error {
	return s.db.Close()
}
