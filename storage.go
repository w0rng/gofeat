package gofeat

import (
	"context"
	"fmt"
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
	entities sync.Map // string -> *entityStore
	ttl      time.Duration
}

type entityStore struct {
	mu     sync.RWMutex
	events []Event
}

func NewMemoryStorage(ttl time.Duration) Storage {
	return &memoryStorage{
		ttl: ttl,
	}
}

func (s *memoryStorage) Push(ctx context.Context, entityID string, events ...Event) error {
	v, _ := s.entities.LoadOrStore(entityID, &entityStore{})
	es, ok := v.(*entityStore)
	if !ok {
		return fmt.Errorf("unknown storage type: %v", v)
	}

	es.mu.Lock()
	defer es.mu.Unlock()

	// Optimize for single event: binary insert O(log n) instead of full sort O(n log n)
	if len(events) == 1 {
		e := events[0]
		idx := sort.Search(len(es.events), func(i int) bool {
			return es.events[i].Timestamp.After(e.Timestamp)
		})
		es.events = slices.Insert(es.events, idx, e)
	} else {
		// Batch: append all, then sort once
		es.events = append(es.events, events...)
		sort.Slice(es.events, func(i, j int) bool {
			return es.events[i].Timestamp.Before(es.events[j].Timestamp)
		})
	}

	return nil
}

func (s *memoryStorage) Get(ctx context.Context, entityID string, at time.Time) ([]Event, error) {
	v, ok := s.entities.Load(entityID)
	if !ok {
		return nil, nil
	}

	es, ok := v.(*entityStore)
	if !ok {
		return nil, fmt.Errorf("unknown storage format: %v", v)
	}
	es.mu.RLock()
	defer es.mu.RUnlock()

	// Apply TTL and point-in-time cutoff
	cutoff := time.Time{}
	if s.ttl > 0 {
		cutoff = at.Add(-s.ttl)
	}

	// Filter events: timestamp <= at AND timestamp > cutoff
	// Pre-allocate to avoid multiple allocations
	filtered := make([]Event, 0, len(es.events))
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
	if s.ttl == 0 {
		return nil
	}

	before := time.Now().UTC().Add(-s.ttl)

	s.entities.Range(func(key, value any) bool {
		es, ok := value.(*entityStore)
		if !ok {
			return true
		}
		es.mu.Lock()
		idx := sort.Search(len(es.events), func(i int) bool {
			return !es.events[i].Timestamp.Before(before)
		})
		if idx > 0 {
			es.events = es.events[idx:]
		}
		es.mu.Unlock()
		return true
	})

	return nil
}

func (s *memoryStorage) Stats(ctx context.Context) (StorageStats, error) {
	var entities int
	var total int64

	s.entities.Range(func(key, value any) bool {
		entities++
		es, ok := value.(*entityStore)
		if !ok {
			return true
		}
		es.mu.RLock()
		total += int64(len(es.events))
		es.mu.RUnlock()
		return true
	})

	return StorageStats{
		Entities:    entities,
		TotalEvents: total,
	}, nil
}

func (s *memoryStorage) Close() error {
	return nil
}
