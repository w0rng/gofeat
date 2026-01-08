# gofeat

Embedded feature store for Go. Real-time feature computation for fraud detection and ML pipelines without external dependencies.

## Why gofeat?

Existing solutions (Feast, Tecton, Redis + custom code) are overkill for many use cases:

- Require external infrastructure
- Complex setup and operations
- Add network latency

Most fraud detection tasks boil down to simple aggregations: "transactions in the last hour", "average amount in 24 hours", "distinct countries this week".

gofeat gives you this in a single library with zero dependencies.

## Features

- **In-memory storage** with pluggable backends (Redis, PostgreSQL, etc.)
- **Sliding windows** for time-based aggregations
- **Point-in-time correct** queries for ML training (no data leakage)
- **Extensible** - custom windows, aggregators, and storage
- **Thread-safe** with per-entity locking
- **Context support** for timeouts and cancellation

## Quick Start

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/w0rng/gofeat"
)

func main() {
    // Define features
    store, err := gofeat.New(gofeat.Config{
        TTL: 24 * time.Hour,
        Features: []gofeat.Feature{
            {Name: "tx_count_1h", Aggregate: gofeat.Count, Window: gofeat.Sliding(time.Hour)},
            {Name: "tx_sum_1h", Aggregate: gofeat.Sum, Field: "amount", Window: gofeat.Sliding(time.Hour)},
            {Name: "countries_1h", Aggregate: gofeat.CountDistinct, Field: "country", Window: gofeat.Sliding(time.Hour)},
            {Name: "last_country", Aggregate: gofeat.Last, Field: "country"},
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    defer store.Close()

    ctx := context.Background()

    // Push events
    err = store.Push(ctx, "user_123",
        gofeat.Event{
            Timestamp: time.Now().UTC(),
            Data: map[string]any{
                "amount":  100.50,
                "country": "US",
            },
        },
    )
    if err != nil {
        log.Fatal(err)
    }

    // Get features (real-time)
    result, err := store.Get(ctx, "user_123")
    if err != nil {
        log.Fatal(err)
    }

    txCount := result.IntOr("tx_count_1h", -1)
    txSum := result.FloatOr("tx_sum_1h", -1)
    countries := result.IntOr("countries_1h", -1)

    // Fraud detection logic
    if txCount > 10 && countries > 3 {
        log.Println("suspicious activity detected")
    }
}
```

## Aggregators

| Aggregator | Result | Use Case |
|------------|--------|----------|
| `Count` | int | Event frequency |
| `Sum` | float64 | Transaction volume |
| `Min` | float64 | Minimum amount |
| `Max` | float64 | Maximum amount |
| `Last` | any | Last country/device |
| `CountDistinct` | int | Unique countries/cards |

## Windows

```go
// Last hour
gofeat.Sliding(time.Hour)

// Last 24 hours
gofeat.Sliding(24 * time.Hour)

// All time (no window)
gofeat.Lifetime()

// Or nil for lifetime
{Name: "total_count", Aggregate: gofeat.Count}
```

## Point-in-Time Queries

For ML training, you need features computed at the time of each event, not current time. This prevents data leakage.

```go
// Real-time serving
result, _ := store.Get(ctx, "user_123")

// Historical query for ML training
result, _ := store.GetAt(ctx, "user_123", eventTimestamp)
```

## Batch Operations

Push multiple events at once:

```go
store.Push(ctx, "user_123",
    gofeat.Event{Timestamp: t1, Data: data1},
    gofeat.Event{Timestamp: t2, Data: data2},
    gofeat.Event{Timestamp: t3, Data: data3},
)
```

## Monitoring

```go
stats, _ := store.Stats(ctx)
log.Printf("entities: %d, events: %d", stats.Entities, stats.TotalEvents)
```

## Custom Storage

Implement the `Storage` interface for custom backends:

```go
type Storage interface {
    Push(ctx context.Context, entityID string, event Event) error
    Events(ctx context.Context, entityID string) ([]Event, error)
    Evict(ctx context.Context, before time.Time) error
    Close() error
}

// Example: Redis storage
store, _ := gofeat.New(gofeat.Config{
    Storage:  NewRedisStorage(redisClient),
    Features: features,
})
```

## Custom Aggregators

Implement the `Aggregator` interface:

```go
type Aggregator interface {
    Add(value any)
    Result() any
    Reset()
}

// Example: Median aggregator
type medianAgg struct {
    values []float64
}

func (a *medianAgg) Add(v any) {
    if f, ok := v.(float64); ok {
        a.values = append(a.values, f)
    }
}

func (a *medianAgg) Result() any {
    if len(a.values) == 0 {
        return 0.0
    }
    sort.Float64s(a.values)
    return a.values[len(a.values)/2]
}

func (a *medianAgg) Reset() {
    a.values = a.values[:0]
}

var Median gofeat.AggregatorFactory = func() gofeat.Aggregator {
    return &medianAgg{}
}

// Usage
{Name: "median_amount", Aggregate: Median, Field: "amount", Window: gofeat.Sliding(time.Hour)}
```

## Custom Windows

Implement the `Window` interface:

```go
type Window interface {
    Select(events []Event, t time.Time) []Event
}

// Example: Business hours only (9 AM - 5 PM)
type businessHoursWindow struct {
    duration time.Duration
}

func (w *businessHoursWindow) Select(events []Event, t time.Time) []Event {
    cutoff := t.Add(-w.duration)
    var result []Event
    for _, e := range events {
        hour := e.Timestamp.Hour()
        if e.Timestamp.After(cutoff) && hour >= 9 && hour < 17 {
            result = append(result, e)
        }
    }
    return result
}
```

## Design Decisions

### Event Time vs Processing Time

gofeat uses **event time** (timestamp from the event) rather than processing time (when event was received). This ensures correct feature computation for historical queries but means results may change if late events arrive.

### Lazy Eviction

Expired events are cleaned up during reads, not in background. This simplifies implementation and avoids concurrency issues. For high-throughput scenarios with many inactive entities, consider periodic cleanup.

### Data Types

Events with missing fields or wrong types are silently skipped (with a warning log). This handles sparse data gracefully without crashing.

## Limitations

- **In-memory by default** - data doesn't survive restarts (use custom storage for persistence)
- **No deduplication** - duplicate events are counted twice (handle at the application level)
- **UTC required** - all timestamps must be UTC

## License

MIT

## Contributing

Issues and PRs welcome.
