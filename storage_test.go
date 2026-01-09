package gofeat_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/w0rng/gofeat"
)

func TestMemoryStorage_Push_Single(t *testing.T) {
	s := gofeat.NewMemoryStorage(0) // no TTL
	ctx := context.Background()

	event := gofeat.Event{
		Timestamp: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		Data:      map[string]any{"amount": 100.0},
	}

	err := s.Push(ctx, "user1", event)
	if err != nil {
		t.Fatalf("Push failed: %v", err)
	}

	events, err := s.Get(ctx, "user1", time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	if !events[0].Timestamp.Equal(event.Timestamp) {
		t.Errorf("timestamp mismatch: got %v, want %v", events[0].Timestamp, event.Timestamp)
	}
}

func TestMemoryStorage_Push_Batch(t *testing.T) {
	s := gofeat.NewMemoryStorage(0) // no TTL
	ctx := context.Background()

	events := []gofeat.Event{
		{Timestamp: time.Date(2024, 1, 1, 12, 30, 0, 0, time.UTC), Data: map[string]any{"id": 2}},
		{Timestamp: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC), Data: map[string]any{"id": 1}},
		{Timestamp: time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC), Data: map[string]any{"id": 3}},
	}

	err := s.Push(ctx, "user1", events...)
	if err != nil {
		t.Fatalf("Push failed: %v", err)
	}

	got, err := s.Get(ctx, "user1", time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if len(got) != 3 {
		t.Fatalf("expected 3 events, got %d", len(got))
	}

	// Events should be sorted by timestamp
	if got[0].Data["id"] != 1 {
		t.Errorf("first event id: got %v, want 1", got[0].Data["id"])
	}
	if got[1].Data["id"] != 2 {
		t.Errorf("second event id: got %v, want 2", got[1].Data["id"])
	}
	if got[2].Data["id"] != 3 {
		t.Errorf("third event id: got %v, want 3", got[2].Data["id"])
	}
}

func TestMemoryStorage_Push_MaintainsSortOrder(t *testing.T) {
	s := gofeat.NewMemoryStorage(0) // no TTL
	ctx := context.Background()

	// Push first event
	err := s.Push(ctx, "user1", gofeat.Event{
		Timestamp: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		Data:      map[string]any{"id": 1},
	})
	if err != nil {
		t.Fatalf("Push failed: %v", err)
	}

	// Push event in the middle
	err = s.Push(ctx, "user1", gofeat.Event{
		Timestamp: time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
		Data:      map[string]any{"id": 0},
	})
	if err != nil {
		t.Fatalf("Push failed: %v", err)
	}

	// Push event at the end
	err = s.Push(ctx, "user1", gofeat.Event{
		Timestamp: time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC),
		Data:      map[string]any{"id": 2},
	})
	if err != nil {
		t.Fatalf("Push failed: %v", err)
	}

	events, err := s.Get(ctx, "user1", time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}

	for i := range len(events) {
		if events[i].Data["id"] != i {
			t.Errorf("event %d: got id %v, want %d", i, events[i].Data["id"], i)
		}
	}
}

func TestMemoryStorage_Get_NonExistent(t *testing.T) {
	s := gofeat.NewMemoryStorage(0) // no TTL
	ctx := context.Background()

	events, err := s.Get(ctx, "nonexistent", time.Now().UTC())
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if events != nil {
		t.Errorf("expected nil for non-existent entity, got %v", events)
	}
}

func TestMemoryStorage_Evict(t *testing.T) {
	s := gofeat.NewMemoryStorage(1 * time.Hour) // 1 hour TTL
	ctx := context.Background()

	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	events := []gofeat.Event{
		{Timestamp: now.Add(-2 * time.Hour), Data: map[string]any{"id": 1}},
		{Timestamp: now.Add(-1 * time.Hour), Data: map[string]any{"id": 2}},
		{Timestamp: now.Add(-30 * time.Minute), Data: map[string]any{"id": 3}},
		{Timestamp: now, Data: map[string]any{"id": 4}},
	}

	err := s.Push(ctx, "user1", events...)
	if err != nil {
		t.Fatalf("Push failed: %v", err)
	}

	// Evict events older than 1 hour from now
	err = s.Evict(ctx)
	if err != nil {
		t.Fatalf("Evict failed: %v", err)
	}

	remaining, err := s.Get(ctx, "user1", time.Now().UTC())
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	// All events should be evicted since they're all in the past
	if len(remaining) != 0 {
		t.Fatalf("expected 0 events after eviction (all are old), got %d", len(remaining))
	}
}

func TestMemoryStorage_Evict_NoTTL(t *testing.T) {
	s := gofeat.NewMemoryStorage(0) // no TTL
	ctx := context.Background()

	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	err := s.Push(ctx, "user1",
		gofeat.Event{Timestamp: now.Add(-2 * time.Hour), Data: map[string]any{}},
		gofeat.Event{Timestamp: now.Add(-1 * time.Hour), Data: map[string]any{}},
	)
	if err != nil {
		t.Fatalf("Push failed: %v", err)
	}

	// Evict should do nothing when TTL is 0
	err = s.Evict(ctx)
	if err != nil {
		t.Fatalf("Evict failed: %v", err)
	}

	events, err := s.Get(ctx, "user1", now)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if len(events) != 2 {
		t.Errorf("expected 2 events (evict should not remove anything when TTL=0), got %d", len(events))
	}
}

func TestMemoryStorage_Get_WithTTL(t *testing.T) {
	s := gofeat.NewMemoryStorage(1 * time.Hour) // 1 hour TTL
	ctx := context.Background()

	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	events := []gofeat.Event{
		{Timestamp: now.Add(-2 * time.Hour), Data: map[string]any{"id": 1}},
		{Timestamp: now.Add(-30 * time.Minute), Data: map[string]any{"id": 2}},
		{Timestamp: now, Data: map[string]any{"id": 3}},
	}

	err := s.Push(ctx, "user1", events...)
	if err != nil {
		t.Fatalf("Push failed: %v", err)
	}

	// Get at 'now' should only return events within 1 hour TTL
	result, err := s.Get(ctx, "user1", now)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	// Should have events 2 and 3 (within 1 hour of 'now')
	if len(result) != 2 {
		t.Fatalf("expected 2 events within TTL, got %d", len(result))
	}

	if result[0].Data["id"] != 2 {
		t.Errorf("first event: got id %v, want 2", result[0].Data["id"])
	}
	if result[1].Data["id"] != 3 {
		t.Errorf("second event: got id %v, want 3", result[1].Data["id"])
	}
}

func TestMemoryStorage_Stats(t *testing.T) {
	s := gofeat.NewMemoryStorage(0) // no TTL
	ctx := context.Background()

	stats, err := s.Stats(ctx)
	if err != nil {
		t.Fatalf("Stats failed: %v", err)
	}

	if stats.Entities != 0 || stats.TotalEvents != 0 {
		t.Errorf("empty storage: got %+v, want zeros", stats)
	}

	// Add events for multiple entities
	now := time.Now().UTC()
	s.Push(ctx, "user1", gofeat.Event{Timestamp: now}, gofeat.Event{Timestamp: now})
	s.Push(ctx, "user2", gofeat.Event{Timestamp: now})

	stats, err = s.Stats(ctx)
	if err != nil {
		t.Fatalf("Stats failed: %v", err)
	}

	if stats.Entities != 2 {
		t.Errorf("entities: got %d, want 2", stats.Entities)
	}
	if stats.TotalEvents != 3 {
		t.Errorf("total events: got %d, want 3", stats.TotalEvents)
	}
}

func TestMemoryStorage_Concurrency(t *testing.T) {
	s := gofeat.NewMemoryStorage(0) // no TTL
	ctx := context.Background()

	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	// Concurrent writes to different entities
	var wg sync.WaitGroup
	for i := range 100 {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			entityID := "user" + string(rune('0'+id%10))
			for j := range 10 {
				event := gofeat.Event{
					Timestamp: now.Add(time.Duration(j) * time.Second),
					Data:      map[string]any{"id": j},
				}
				s.Push(ctx, entityID, event)
			}
		}(i)
	}

	wg.Wait()

	stats, err := s.Stats(ctx)
	if err != nil {
		t.Fatalf("Stats failed: %v", err)
	}

	if stats.Entities != 10 {
		t.Errorf("entities: got %d, want 10", stats.Entities)
	}
	if stats.TotalEvents != 1000 {
		t.Errorf("total events: got %d, want 1000", stats.TotalEvents)
	}

	// Verify events are sorted for each entity
	queryTime := now.Add(10 * time.Second)
	for i := range 10 {
		entityID := "user" + string(rune('0'+i))
		events, err := s.Get(ctx, entityID, queryTime)
		if err != nil {
			t.Fatalf("Get failed for %s: %v", entityID, err)
		}

		for j := 1; j < len(events); j++ {
			if events[j].Timestamp.Before(events[j-1].Timestamp) {
				t.Errorf("entity %s: events not sorted at index %d", entityID, j)
			}
		}
	}
}

func TestMemoryStorage_Close(t *testing.T) {
	s := gofeat.NewMemoryStorage(0) // no TTL
	err := s.Close()
	if err != nil {
		t.Errorf("Close returned error: %v", err)
	}
}
