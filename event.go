package gofeat

import "time"

// Event represents a single event with timestamp and arbitrary data.
type Event struct {
	Timestamp time.Time
	Data      map[string]any
}
