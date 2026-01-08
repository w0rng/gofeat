package gofeat

import "time"

type Config struct {
	TTL      time.Duration
	Features []Feature
	Storage  Storage // optional, defaults to in-memory
}

// Feature defines a single feature computation.
type Feature struct {
	Name      string
	Aggregate AggregatorFactory
	Window    Window // nil for Lifetime
}
