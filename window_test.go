package gofeat_test

import (
	"testing"
	"time"

	"github.com/w0rng/gofeat"
)

func TestSlidingWindow(t *testing.T) {
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	events := []gofeat.Event{
		{Timestamp: now.Add(-2 * time.Hour), Data: map[string]any{"id": 1}},
		{Timestamp: now.Add(-90 * time.Minute), Data: map[string]any{"id": 2}},
		{Timestamp: now.Add(-30 * time.Minute), Data: map[string]any{"id": 3}},
		{Timestamp: now, Data: map[string]any{"id": 4}},
	}

	window := gofeat.Sliding(1 * time.Hour)
	selected := window.Select(events, now)

	// Should select events within last hour: id 3 and 4
	if len(selected) != 2 {
		t.Fatalf("expected 2 events, got %d", len(selected))
	}

	if selected[0].Data["id"] != 3 {
		t.Errorf("first selected event: got id %v, want 3", selected[0].Data["id"])
	}

	if selected[1].Data["id"] != 4 {
		t.Errorf("second selected event: got id %v, want 4", selected[1].Data["id"])
	}
}

func TestSlidingWindow_AllInWindow(t *testing.T) {
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	events := []gofeat.Event{
		{Timestamp: now.Add(-30 * time.Minute), Data: map[string]any{"id": 1}},
		{Timestamp: now.Add(-15 * time.Minute), Data: map[string]any{"id": 2}},
		{Timestamp: now, Data: map[string]any{"id": 3}},
	}

	window := gofeat.Sliding(1 * time.Hour)
	selected := window.Select(events, now)

	if len(selected) != 3 {
		t.Fatalf("expected 3 events, got %d", len(selected))
	}
}

func TestSlidingWindow_NoneInWindow(t *testing.T) {
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	events := []gofeat.Event{
		{Timestamp: now.Add(-2 * time.Hour), Data: map[string]any{"id": 1}},
		{Timestamp: now.Add(-90 * time.Minute), Data: map[string]any{"id": 2}},
	}

	window := gofeat.Sliding(1 * time.Hour)
	selected := window.Select(events, now)

	if selected != nil {
		t.Errorf("expected nil for empty selection, got %d events", len(selected))
	}
}

func TestSlidingWindow_EmptyEvents(t *testing.T) {
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	window := gofeat.Sliding(1 * time.Hour)

	selected := window.Select([]gofeat.Event{}, now)
	if len(selected) != 0 {
		t.Error("expected empty slice for empty events")
	}

	selected = window.Select(nil, now)
	if selected != nil {
		t.Error("expected nil for nil events")
	}
}

func TestSlidingWindow_ExactBoundary(t *testing.T) {
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	events := []gofeat.Event{
		{Timestamp: now.Add(-1*time.Hour - 1*time.Second), Data: map[string]any{"id": 1}}, // just outside
		{Timestamp: now.Add(-1 * time.Hour), Data: map[string]any{"id": 2}},               // exactly on boundary
		{Timestamp: now.Add(-30 * time.Minute), Data: map[string]any{"id": 3}},
	}

	window := gofeat.Sliding(1 * time.Hour)
	selected := window.Select(events, now)

	// Event at exact cutoff should be included
	if len(selected) != 2 {
		t.Fatalf("expected 2 events (including boundary), got %d", len(selected))
	}

	if selected[0].Data["id"] != 2 {
		t.Errorf("first event: got id %v, want 2 (boundary event)", selected[0].Data["id"])
	}
}

func TestSlidingWindow_DifferentDurations(t *testing.T) {
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	events := []gofeat.Event{
		{Timestamp: now.Add(-25 * time.Hour), Data: map[string]any{"id": 1}},
		{Timestamp: now.Add(-12 * time.Hour), Data: map[string]any{"id": 2}},
		{Timestamp: now.Add(-30 * time.Minute), Data: map[string]any{"id": 3}},
		{Timestamp: now, Data: map[string]any{"id": 4}},
	}

	tests := []struct {
		duration      time.Duration
		expectedCount int
		expectedFirst int
	}{
		{duration: 1 * time.Hour, expectedCount: 2, expectedFirst: 3},
		{duration: 24 * time.Hour, expectedCount: 3, expectedFirst: 2},
		{duration: 48 * time.Hour, expectedCount: 4, expectedFirst: 1},
	}

	for _, tt := range tests {
		window := gofeat.Sliding(tt.duration)
		selected := window.Select(events, now)

		if len(selected) != tt.expectedCount {
			t.Errorf("duration %v: got %d events, want %d", tt.duration, len(selected), tt.expectedCount)
		}

		if len(selected) > 0 && selected[0].Data["id"] != tt.expectedFirst {
			t.Errorf("duration %v: first event id %v, want %d", tt.duration, selected[0].Data["id"], tt.expectedFirst)
		}
	}
}

func TestLifetimeWindow(t *testing.T) {
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	events := []gofeat.Event{
		{Timestamp: now.Add(-2 * time.Hour), Data: map[string]any{"id": 1}},
		{Timestamp: now.Add(-1 * time.Hour), Data: map[string]any{"id": 2}},
		{Timestamp: now, Data: map[string]any{"id": 3}},
		{Timestamp: now.Add(1 * time.Hour), Data: map[string]any{"id": 4}}, // future
	}

	window := gofeat.Lifetime()
	selected := window.Select(events, now)

	// Should select all events up to (and including) now
	if len(selected) != 3 {
		t.Fatalf("expected 3 events, got %d", len(selected))
	}

	for i := 0; i < 3; i++ {
		if selected[i].Data["id"] != i+1 {
			t.Errorf("event %d: got id %v, want %d", i, selected[i].Data["id"], i+1)
		}
	}
}

func TestLifetimeWindow_AllEvents(t *testing.T) {
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	events := []gofeat.Event{
		{Timestamp: now.Add(-10 * time.Hour), Data: map[string]any{"id": 1}},
		{Timestamp: now.Add(-5 * time.Hour), Data: map[string]any{"id": 2}},
		{Timestamp: now.Add(-1 * time.Hour), Data: map[string]any{"id": 3}},
	}

	// Query at a time after all events
	window := gofeat.Lifetime()
	selected := window.Select(events, now)

	if len(selected) != 3 {
		t.Fatalf("expected all 3 events, got %d", len(selected))
	}
}

func TestLifetimeWindow_NoEvents(t *testing.T) {
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	events := []gofeat.Event{
		{Timestamp: now.Add(1 * time.Hour), Data: map[string]any{"id": 1}},
		{Timestamp: now.Add(2 * time.Hour), Data: map[string]any{"id": 2}},
	}

	// All events are in the future
	window := gofeat.Lifetime()
	selected := window.Select(events, now)

	if len(selected) != 0 {
		t.Errorf("expected empty slice when all events are future, got %d events", len(selected))
	}
}

func TestLifetimeWindow_EmptyEvents(t *testing.T) {
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	window := gofeat.Lifetime()

	selected := window.Select([]gofeat.Event{}, now)
	if len(selected) != 0 {
		t.Error("expected empty slice for empty events")
	}

	selected = window.Select(nil, now)
	if selected != nil {
		t.Error("expected nil for nil events")
	}
}

func TestLifetimeWindow_ExactBoundary(t *testing.T) {
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	events := []gofeat.Event{
		{Timestamp: now.Add(-1 * time.Hour), Data: map[string]any{"id": 1}},
		{Timestamp: now, Data: map[string]any{"id": 2}}, // exactly at query time
		{Timestamp: now.Add(1 * time.Nanosecond), Data: map[string]any{"id": 3}},
	}

	window := gofeat.Lifetime()
	selected := window.Select(events, now)

	// Event at exact time should be included, future event excluded
	if len(selected) != 2 {
		t.Fatalf("expected 2 events, got %d", len(selected))
	}

	if selected[1].Data["id"] != 2 {
		t.Errorf("last selected event: got id %v, want 2", selected[1].Data["id"])
	}
}

func TestLifetimeWindow_HistoricalQuery(t *testing.T) {
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	events := []gofeat.Event{
		{Timestamp: now.Add(-3 * time.Hour), Data: map[string]any{"id": 1}},
		{Timestamp: now.Add(-2 * time.Hour), Data: map[string]any{"id": 2}},
		{Timestamp: now.Add(-1 * time.Hour), Data: map[string]any{"id": 3}},
		{Timestamp: now, Data: map[string]any{"id": 4}},
	}

	window := gofeat.Lifetime()

	// Query at 2 hours ago
	queryTime := now.Add(-2 * time.Hour)
	selected := window.Select(events, queryTime)

	// Should get events 1 and 2
	if len(selected) != 2 {
		t.Fatalf("expected 2 events for historical query, got %d", len(selected))
	}

	if selected[0].Data["id"] != 1 || selected[1].Data["id"] != 2 {
		t.Errorf("historical query: got ids %v, %v; want 1, 2",
			selected[0].Data["id"], selected[1].Data["id"])
	}
}

func TestWindow_Integration(t *testing.T) {
	// Test that windows work correctly when used together in features
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	events := []gofeat.Event{
		{Timestamp: now.Add(-3 * time.Hour), Data: map[string]any{"amount": 100.0}},
		{Timestamp: now.Add(-2 * time.Hour), Data: map[string]any{"amount": 200.0}},
		{Timestamp: now.Add(-30 * time.Minute), Data: map[string]any{"amount": 300.0}},
		{Timestamp: now, Data: map[string]any{"amount": 400.0}},
	}

	sliding1h := gofeat.Sliding(1 * time.Hour)
	sliding2h := gofeat.Sliding(2 * time.Hour)
	lifetime := gofeat.Lifetime()

	selected1h := sliding1h.Select(events, now)
	selected2h := sliding2h.Select(events, now)
	selectedLifetime := lifetime.Select(events, now)

	if len(selected1h) != 2 {
		t.Errorf("1h window: got %d events, want 2", len(selected1h))
	}

	// 2h window: events at -2h, -30m, now = 3 events (-2h is exactly on boundary)
	if len(selected2h) != 3 {
		t.Errorf("2h window: got %d events, want 3", len(selected2h))
	}

	if len(selectedLifetime) != 4 {
		t.Errorf("lifetime window: got %d events, want 4", len(selectedLifetime))
	}
}
