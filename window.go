package gofeat

import (
	"sort"
	"time"
)

// Window filters events within a time range.
type Window interface {
	Select(events []Event, t time.Time) []Event
}

type slidingWindow struct {
	duration time.Duration
}

// Sliding returns a window that selects events within the last duration.
func Sliding(d time.Duration) Window {
	return &slidingWindow{duration: d}
}

func (w *slidingWindow) Select(events []Event, t time.Time) []Event {
	cutoff := t.Add(-w.duration)
	for i, e := range events {
		if !e.Timestamp.Before(cutoff) {
			return events[i:]
		}
	}
	return nil
}

type lifetimeWindow struct{}

// Lifetime returns a window that selects all events.
func Lifetime() Window {
	return &lifetimeWindow{}
}

func (w *lifetimeWindow) Select(events []Event, t time.Time) []Event {
	// Events are sorted ascending, find first event > t
	idx := sort.Search(len(events), func(i int) bool {
		return events[i].Timestamp.After(t)
	})
	return events[:idx]
}
