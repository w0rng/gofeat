package gofeat

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type Store struct {
	storage  Storage
	features []Feature
}

func New(cfg Config) (*Store, error) {
	if len(cfg.Features) == 0 {
		return nil, errors.New("gofeat: at least one feature required")
	}
	for i, f := range cfg.Features {
		if f.Name == "" {
			return nil, errors.New("gofeat: feature name required")
		}
		if f.Aggregate == nil {
			return nil, errors.New("gofeat: feature aggregate required")
		}
		if f.Window == nil {
			cfg.Features[i].Window = Lifetime()
		}
	}

	storage := cfg.Storage
	if storage == nil {
		storage = NewMemoryStorage(cfg.TTL)
	}

	return &Store{
		storage:  storage,
		features: cfg.Features,
	}, nil
}

func (s *Store) Push(ctx context.Context, entityID string, events ...Event) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	for i, e := range events {
		if err := s.validateEvent(e); err != nil {
			return fmt.Errorf("invalid event %d: %w", i, err)
		}
	}

	err := s.storage.Push(ctx, entityID, events...)
	if err != nil {
		return fmt.Errorf("failed to push events: %w", err)
	}

	return nil
}

func (s *Store) Get(ctx context.Context, entityID string) (Result, error) {
	return s.GetAt(ctx, entityID, time.Now().UTC())
}

func (s *Store) GetAt(ctx context.Context, entityID string, at time.Time) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}

	events, err := s.storage.Get(ctx, entityID, at)
	if err != nil {
		return Result{}, fmt.Errorf("failed to fetch events: %w", err)
	}

	values := make(map[string]any, len(s.features))
	for _, f := range s.features {
		if err := ctx.Err(); err != nil {
			return Result{}, err
		}

		selected := f.Window.Select(events, at)
		agg := f.Aggregate()
		for _, e := range selected {
			agg.Add(e.Data)
		}
		values[f.Name] = agg.Result()
	}

	return newResult(values), nil
}

func (s *Store) BatchGet(ctx context.Context, entityIDs ...string) (map[string]Result, error) {
	return s.BatchGetAt(ctx, time.Now().UTC(), entityIDs...)
}

func (s *Store) BatchGetAt(ctx context.Context, at time.Time, entityIDs ...string) (map[string]Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	results := make(map[string]Result, len(entityIDs))
	for _, entityID := range entityIDs {
		result, err := s.GetAt(ctx, entityID, at)
		if err != nil {
			return nil, err
		}
		results[entityID] = result
	}

	return results, nil
}

func (s *Store) Evict(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	err := s.storage.Evict(ctx)
	if err != nil {
		return fmt.Errorf("failed to evict events: %w", err)
	}

	return nil
}

func (s *Store) Stats(ctx context.Context) (StorageStats, error) {
	stat, err := s.storage.Stats(ctx)
	if err != nil {
		return StorageStats{}, fmt.Errorf("failed to stats: %w", err)
	}

	return stat, nil
}

func (s *Store) Close() error {
	err := s.storage.Close()
	if err != nil {
		return fmt.Errorf("error closing storage: %w", err)
	}

	return nil
}

func (s *Store) validateEvent(e Event) error {
	if e.Timestamp.Location() != time.UTC {
		return errors.New("timestamp must be in UTC")
	}
	return nil
}
