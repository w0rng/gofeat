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

	// Get returns events for an entity filtered by TTL relative to the given time.
	// Returns events where: timestamp <= at AND timestamp > at - TTL.
	Get(ctx context.Context, entityID string, at time.Time) ([]Event, error)

	// Evict removes events older than TTL from current time.
	Evict(ctx context.Context) error

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
	ttl      time.Duration
}

type entityStore struct {
	mu     sync.RWMutex
	events []Event
}

func NewMemoryStorage(ttl time.Duration) Storage {
	return &memoryStorage{
		entities: make(map[string]*entityStore),
		ttl:      ttl,
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

func (s *memoryStorage) Get(ctx context.Context, entityID string, at time.Time) ([]Event, error) {
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

	// Apply TTL and point-in-time cutoff
	cutoff := time.Time{}
	if s.ttl > 0 {
		cutoff = at.Add(-s.ttl)
	}

	// Filter events: timestamp <= at AND timestamp > cutoff
	var filtered []Event
	for _, e := range es.events {
		if e.Timestamp.After(at) {
			continue
		}
		if !cutoff.IsZero() && !e.Timestamp.After(cutoff) {
			continue
		}
		filtered = append(filtered, e)
	}

	return filtered, nil
}

func (s *memoryStorage) Evict(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if s.ttl == 0 {
		return nil
	}

	before := time.Now().UTC().Add(-s.ttl)

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
