package gofeat

import (
	"context"
	"slices"
	"sort"
	"sync"
	"time"
)

// Storage is the interface for event storage backends.
type Storage interface {
	// Push adds events for an entity.
	Push(ctx context.Context, entityID string, events ...Event) error

	// Get returns all events for an entity.
	Get(ctx context.Context, entityID string) ([]Event, error)

	// Evict removes events older than the given time.
	Evict(ctx context.Context, before time.Time) error

	// Stats returns storage statistics.
	Stats(ctx context.Context) (StorageStats, error)

	// Close closes the storage.
	Close() error
}

type StorageStats struct {
	Entities    int
	TotalEvents int64
}

// memoryStorage is an in-memory implementation of Storage.
type memoryStorage struct {
	mu       sync.RWMutex
	entities map[string]*entityStore
}

type entityStore struct {
	mu     sync.RWMutex
	events []Event
}

func NewMemoryStorage() Storage {
	return &memoryStorage{
		entities: make(map[string]*entityStore),
	}
}

func (s *memoryStorage) getOrCreate(entityID string) *entityStore {
	s.mu.RLock()
	es, ok := s.entities[entityID]
	s.mu.RUnlock()
	if ok {
		return es
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if es, ok = s.entities[entityID]; ok {
		return es
	}

	es = &entityStore{}
	s.entities[entityID] = es
	return es
}

func (s *memoryStorage) Push(ctx context.Context, entityID string, events ...Event) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	es := s.getOrCreate(entityID)

	es.mu.Lock()
	defer es.mu.Unlock()

	if len(events) == 1 {
		e := events[0]
		idx := sort.Search(len(es.events), func(i int) bool {
			return es.events[i].Timestamp.After(e.Timestamp)
		})
		es.events = slices.Insert(es.events, idx, e)
	} else {
		// Batch: append all, then sort once instead of N insertions
		es.events = append(es.events, events...)
		sort.Slice(es.events, func(i, j int) bool {
			return es.events[i].Timestamp.Before(es.events[j].Timestamp)
		})
	}

	return nil
}

func (s *memoryStorage) Get(ctx context.Context, entityID string) ([]Event, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	s.mu.RLock()
	es, ok := s.entities[entityID]
	s.mu.RUnlock()

	if !ok {
		return nil, nil
	}

	es.mu.RLock()
	defer es.mu.RUnlock()

	// No copy - caller must not modify the slice
	return es.events, nil
}

func (s *memoryStorage) Evict(ctx context.Context, before time.Time) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.RLock()
	entities := make([]*entityStore, 0, len(s.entities))
	for _, es := range s.entities {
		entities = append(entities, es)
	}
	s.mu.RUnlock()

	for _, es := range entities {
		if err := ctx.Err(); err != nil {
			return err
		}

		es.mu.Lock()
		idx := sort.Search(len(es.events), func(i int) bool {
			return !es.events[i].Timestamp.Before(before)
		})
		if idx > 0 {
			es.events = es.events[idx:]
		}
		es.mu.Unlock()
	}

	return nil
}

func (s *memoryStorage) Stats(ctx context.Context) (StorageStats, error) {
	if err := ctx.Err(); err != nil {
		return StorageStats{}, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	var total int64
	for _, es := range s.entities {
		es.mu.RLock()
		total += int64(len(es.events))
		es.mu.RUnlock()
	}

	return StorageStats{
		Entities:    len(s.entities),
		TotalEvents: total,
	}, nil
}

func (s *memoryStorage) Close() error {
	return nil
}
