package gofeat_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/w0rng/gofeat"
)

func TestNew_Validation(t *testing.T) {
	tests := []struct {
		name    string
		config  gofeat.Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: gofeat.Config{
				Features: []gofeat.Feature{
					{Name: "count", Aggregate: gofeat.Count},
				},
			},
			wantErr: false,
		},
		{
			name:    "no features",
			config:  gofeat.Config{},
			wantErr: true,
		},
		{
			name: "empty feature name",
			config: gofeat.Config{
				Features: []gofeat.Feature{
					{Name: "", Aggregate: gofeat.Count},
				},
			},
			wantErr: true,
		},
		{
			name: "nil aggregate",
			config: gofeat.Config{
				Features: []gofeat.Feature{
					{Name: "count"},
				},
			},
			wantErr: true,
		},
		{
			name: "nil window defaults to Lifetime",
			config: gofeat.Config{
				Features: []gofeat.Feature{
					{Name: "count", Aggregate: gofeat.Count, Window: nil},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := gofeat.New(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStore_PushGet(t *testing.T) {
	store, err := gofeat.New(gofeat.Config{
		Features: []gofeat.Feature{
			{Name: "count", Aggregate: gofeat.Count},
			{Name: "sum", Aggregate: gofeat.Sum("amount")},
		},
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	now := time.Now().UTC()

	err = store.Push(ctx, "user1",
		gofeat.Event{Timestamp: now.Add(-1 * time.Second), Data: map[string]any{"amount": 200.0}},
		gofeat.Event{Timestamp: now, Data: map[string]any{"amount": 100.0}},
	)
	if err != nil {
		t.Fatalf("Push failed: %v", err)
	}

	result, err := store.Get(ctx, "user1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	count, err := result.Int("count")
	if err != nil {
		t.Fatalf("Int failed: %v", err)
	}
	if count != 2 {
		t.Errorf("count: got %d, want 2", count)
	}

	sum, err := result.Float("sum")
	if err != nil {
		t.Fatalf("Float failed: %v", err)
	}
	if sum != 300.0 {
		t.Errorf("sum: got %f, want 300.0", sum)
	}
}

func TestStore_GetAt_PointInTime(t *testing.T) {
	store, err := gofeat.New(gofeat.Config{
		Features: []gofeat.Feature{
			{Name: "count", Aggregate: gofeat.Count, Window: gofeat.Sliding(1 * time.Hour)},
		},
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	// Push events at different times
	err = store.Push(ctx, "user1",
		gofeat.Event{Timestamp: now.Add(-2 * time.Hour), Data: nil}, // too old
		gofeat.Event{Timestamp: now.Add(-30 * time.Minute), Data: nil},
		gofeat.Event{Timestamp: now, Data: nil},
		gofeat.Event{Timestamp: now.Add(1 * time.Hour), Data: nil}, // in the future
	)
	if err != nil {
		t.Fatalf("Push failed: %v", err)
	}

	// Query at "now" should exclude future events
	result, err := store.GetAt(ctx, "user1", now)
	if err != nil {
		t.Fatalf("GetAt failed: %v", err)
	}

	count := result.IntOr("count", 0)
	if count != 2 {
		t.Errorf("count at now: got %d, want 2 (excludes future event)", count)
	}

	// Query at past time
	pastTime := now.Add(-1 * time.Hour)
	result, err = store.GetAt(ctx, "user1", pastTime)
	if err != nil {
		t.Fatalf("GetAt failed: %v", err)
	}

	count = result.IntOr("count", 0)
	if count != 1 {
		t.Errorf("count at past: got %d, want 1", count)
	}
}

func TestStore_TTL(t *testing.T) {
	store, err := gofeat.New(gofeat.Config{
		TTL: 1 * time.Hour,
		Features: []gofeat.Feature{
			{Name: "count", Aggregate: gofeat.Count},
		},
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	err = store.Push(ctx, "user1",
		gofeat.Event{Timestamp: now.Add(-2 * time.Hour), Data: nil},
		gofeat.Event{Timestamp: now.Add(-30 * time.Minute), Data: nil},
		gofeat.Event{Timestamp: now, Data: nil},
	)
	if err != nil {
		t.Fatalf("Push failed: %v", err)
	}

	// TTL should filter out events older than 1 hour
	result, err := store.GetAt(ctx, "user1", now)
	if err != nil {
		t.Fatalf("GetAt failed: %v", err)
	}

	count := result.IntOr("count", 0)
	if count != 2 {
		t.Errorf("count with TTL: got %d, want 2 (2h old event excluded)", count)
	}
}

func TestStore_Evict(t *testing.T) {
	store, err := gofeat.New(gofeat.Config{
		TTL: 1 * time.Hour,
		Features: []gofeat.Feature{
			{Name: "count", Aggregate: gofeat.Count},
		},
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	now := time.Now().UTC()

	// Push old events
	err = store.Push(ctx, "user1",
		gofeat.Event{Timestamp: now.Add(-2 * time.Hour), Data: nil},
		gofeat.Event{Timestamp: now, Data: nil},
	)
	if err != nil {
		t.Fatalf("Push failed: %v", err)
	}

	// Evict old events
	err = store.Evict(ctx)
	if err != nil {
		t.Fatalf("Evict failed: %v", err)
	}

	// Verify old events are gone
	stats, err := store.Stats(ctx)
	if err != nil {
		t.Fatalf("Stats failed: %v", err)
	}

	if stats.TotalEvents != 1 {
		t.Errorf("after evict: got %d events, want 1", stats.TotalEvents)
	}
}

func TestStore_Evict_NoTTL(t *testing.T) {
	store, err := gofeat.New(gofeat.Config{
		Features: []gofeat.Feature{
			{Name: "count", Aggregate: gofeat.Count},
		},
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	err = store.Evict(ctx)
	if err != nil {
		t.Errorf("Evict with no TTL should not error: %v", err)
	}
}

func TestStore_BatchGet(t *testing.T) {
	store, err := gofeat.New(gofeat.Config{
		Features: []gofeat.Feature{
			{Name: "count", Aggregate: gofeat.Count},
		},
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	now := time.Now().UTC()

	// Push events for multiple users
	for i := 1; i <= 3; i++ {
		entityID := "user" + string(rune('0'+i))
		events := make([]gofeat.Event, i)
		for j := 0; j < i; j++ {
			events[j] = gofeat.Event{Timestamp: now, Data: nil}
		}
		err = store.Push(ctx, entityID, events...)
		if err != nil {
			t.Fatalf("Push failed for %s: %v", entityID, err)
		}
	}

	// Batch get
	results, err := store.BatchGet(ctx, "user1", "user2", "user3")
	if err != nil {
		t.Fatalf("BatchGet failed: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("BatchGet: got %d results, want 3", len(results))
	}

	for i := 1; i <= 3; i++ {
		entityID := "user" + string(rune('0'+i))
		result, ok := results[entityID]
		if !ok {
			t.Errorf("BatchGet: missing result for %s", entityID)
			continue
		}

		count := result.IntOr("count", -1)
		if count != i {
			t.Errorf("%s count: got %d, want %d", entityID, count, i)
		}
	}
}

func TestStore_BatchGetAt(t *testing.T) {
	store, err := gofeat.New(gofeat.Config{
		Features: []gofeat.Feature{
			{Name: "count", Aggregate: gofeat.Count},
		},
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	// Push events before and after "now"
	err = store.Push(ctx, "user1",
		gofeat.Event{Timestamp: now.Add(-1 * time.Hour), Data: nil},
		gofeat.Event{Timestamp: now.Add(1 * time.Hour), Data: nil},
	)
	if err != nil {
		t.Fatalf("Push failed: %v", err)
	}

	results, err := store.BatchGetAt(ctx, now, "user1")
	if err != nil {
		t.Fatalf("BatchGetAt failed: %v", err)
	}

	count := results["user1"].IntOr("count", -1)
	if count != 1 {
		t.Errorf("count: got %d, want 1 (future event excluded)", count)
	}
}

func TestStore_ValidateUTC(t *testing.T) {
	store, err := gofeat.New(gofeat.Config{
		Features: []gofeat.Feature{
			{Name: "count", Aggregate: gofeat.Count},
		},
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Non-UTC timestamp should be rejected
	local := time.FixedZone("Local", 3600)
	event := gofeat.Event{
		Timestamp: time.Now().In(local),
		Data:      nil,
	}

	err = store.Push(ctx, "user1", event)
	if err == nil {
		t.Error("Push should reject non-UTC timestamp")
	}
}

// TestStore_ContextCancellation removed - we simplified the code by removing upfront ctx.Err() checks
// Context cancellation is still respected by the underlying storage operations if they check context

func TestStore_Stats(t *testing.T) {
	store, err := gofeat.New(gofeat.Config{
		Features: []gofeat.Feature{
			{Name: "count", Aggregate: gofeat.Count},
		},
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	now := time.Now().UTC()

	stats, err := store.Stats(ctx)
	if err != nil {
		t.Fatalf("Stats failed: %v", err)
	}

	if stats.Entities != 0 || stats.TotalEvents != 0 {
		t.Errorf("empty store: got %+v, want zeros", stats)
	}

	store.Push(ctx, "user1", gofeat.Event{Timestamp: now}, gofeat.Event{Timestamp: now})
	store.Push(ctx, "user2", gofeat.Event{Timestamp: now})

	stats, err = store.Stats(ctx)
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

func TestStore_CustomStorage(t *testing.T) {
	mockStorage := &mockStorage{events: make(map[string][]gofeat.Event)}

	store, err := gofeat.New(gofeat.Config{
		Storage: mockStorage,
		Features: []gofeat.Feature{
			{Name: "count", Aggregate: gofeat.Count},
		},
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	now := time.Now().UTC()

	err = store.Push(ctx, "user1", gofeat.Event{Timestamp: now})
	if err != nil {
		t.Fatalf("Push failed: %v", err)
	}

	if !mockStorage.pushCalled {
		t.Error("custom storage Push was not called")
	}

	_, err = store.Get(ctx, "user1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if !mockStorage.getCalled {
		t.Error("custom storage Get was not called")
	}
}

func TestStore_MultipleFeatures(t *testing.T) {
	store, err := gofeat.New(gofeat.Config{
		Features: []gofeat.Feature{
			{Name: "count", Aggregate: gofeat.Count},
			{Name: "sum", Aggregate: gofeat.Sum("amount")},
			{Name: "min", Aggregate: gofeat.Min("amount")},
			{Name: "max", Aggregate: gofeat.Max("amount")},
			{Name: "last_country", Aggregate: gofeat.Last("country")},
			{Name: "distinct_countries", Aggregate: gofeat.CountDistinct("country")},
		},
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	now := time.Now().UTC()

	err = store.Push(ctx, "user1",
		gofeat.Event{Timestamp: now.Add(-3 * time.Second), Data: map[string]any{"amount": 100.0, "country": "US"}},
		gofeat.Event{Timestamp: now.Add(-2 * time.Second), Data: map[string]any{"amount": 50.0, "country": "CA"}},
		gofeat.Event{Timestamp: now.Add(-1 * time.Second), Data: map[string]any{"amount": 200.0, "country": "US"}},
	)
	if err != nil {
		t.Fatalf("Push failed: %v", err)
	}

	result, err := store.Get(ctx, "user1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if count := result.IntOr("count", 0); count != 3 {
		t.Errorf("count: got %d, want 3", count)
	}

	if sum := result.FloatOr("sum", 0); sum != 350.0 {
		t.Errorf("sum: got %f, want 350.0", sum)
	}

	if minF := result.FloatOr("min", 0); minF != 50.0 {
		t.Errorf("min: got %f, want 50.0", minF)
	}

	if maxF := result.FloatOr("max", 0); maxF != 200.0 {
		t.Errorf("max: got %f, want 200.0", maxF)
	}

	if last := result.StringOr("last_country", ""); last != "US" {
		t.Errorf("last_country: got %s, want US", last)
	}

	if distinct := result.IntOr("distinct_countries", 0); distinct != 2 {
		t.Errorf("distinct_countries: got %d, want 2", distinct)
	}
}

// Mock storage for testing custom storage backends.
type mockStorage struct {
	events      map[string][]gofeat.Event
	pushCalled  bool
	getCalled   bool
	evictCalled bool
}

func (m *mockStorage) Push(_ context.Context, entityID string, events ...gofeat.Event) error {
	m.pushCalled = true
	m.events[entityID] = append(m.events[entityID], events...)
	return nil
}

func (m *mockStorage) Get(_ context.Context, entityID string, at time.Time) ([]gofeat.Event, error) {
	m.getCalled = true
	// Simple mock: return all events up to 'at'
	var result []gofeat.Event
	for _, e := range m.events[entityID] {
		if !e.Timestamp.After(at) {
			result = append(result, e)
		}
	}
	return result, nil
}

func (m *mockStorage) Evict(_ context.Context) error {
	m.evictCalled = true
	return nil
}

func (m *mockStorage) Stats(_ context.Context) (gofeat.StorageStats, error) {
	total := int64(0)
	for _, events := range m.events {
		total += int64(len(events))
	}
	return gofeat.StorageStats{
		Entities:    len(m.events),
		TotalEvents: total,
	}, nil
}

func (m *mockStorage) Close() error {
	return nil
}

// Error storage for testing error handling.
type errorStorage struct{}

func (e *errorStorage) Push(_ context.Context, _ string, _ ...gofeat.Event) error {
	return errors.New("storage error")
}

func (e *errorStorage) Get(_ context.Context, _ string, _ time.Time) ([]gofeat.Event, error) {
	return nil, errors.New("storage error")
}

func (e *errorStorage) Evict(_ context.Context) error {
	return errors.New("storage error")
}

func (e *errorStorage) Stats(_ context.Context) (gofeat.StorageStats, error) {
	return gofeat.StorageStats{}, errors.New("storage error")
}

func (e *errorStorage) Close() error {
	return errors.New("storage error")
}

func TestStore_StorageErrors(t *testing.T) {
	store, err := gofeat.New(gofeat.Config{
		TTL:     1 * time.Hour, // Need TTL for Evict to call storage
		Storage: &errorStorage{},
		Features: []gofeat.Feature{
			{Name: "count", Aggregate: gofeat.Count},
		},
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	ctx := context.Background()
	now := time.Now().UTC()

	err = store.Push(ctx, "user1", gofeat.Event{Timestamp: now})
	if err == nil {
		t.Error("Push should propagate storage error")
	}

	_, err = store.Get(ctx, "user1")
	if err == nil {
		t.Error("Get should propagate storage error")
	}

	_, err = store.GetAt(ctx, "user1", now)
	if err == nil {
		t.Error("GetAt should propagate storage error")
	}

	err = store.Evict(ctx)
	if err == nil {
		t.Error("Evict should propagate storage error")
	}

	_, err = store.Stats(ctx)
	if err == nil {
		t.Error("Stats should propagate storage error")
	}

	err = store.Close()
	if err == nil {
		t.Error("Close should propagate storage error")
	}
}
