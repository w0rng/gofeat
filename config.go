package gofeat

import "time"

type Config struct {
	Features []Feature
	Storage  Storage       // optional, defaults to in-memory with no TTL
	TTL      time.Duration // Used only if Storage is not provided
}

// Feature defines a single feature computation.
type Feature struct {
	Name      string
	Aggregate AggregatorFactory
	Window    Window // nil for Lifetime
}
