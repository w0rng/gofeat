# gofeat [![Go Version](https://img.shields.io/github/go-mod/go-version/w0rng/gofeat)](https://go.dev/) [![CI](https://github.com/w0rng/gofeat/actions/workflows/ci.yaml/badge.svg)](https://github.com/w0rng/gofeat/actions) [![Go Report Card](https://goreportcard.com/badge/github.com/w0rng/gofeat)](https://goreportcard.com/report/github.com/w0rng/gofeat) [![Coverage Status](https://coveralls.io/repos/github/w0rng/gofeat/badge.svg?branch=main)](https://coveralls.io/github/w0rng/gofeat?branch=main)

Embedded feature store for Go. Real-time feature computation for fraud detection and ML pipelines without external dependencies.

## Why gofeat?

**Problem**: You need to detect fraud in real-time, but...

- **Feast** requires Redis + Python + infrastructure team
- **Tecton** is $$$$$ enterprise solution  
- **Custom Redis + code** is brittle and hard to maintain

**Reality**: 90% of fraud detection is simple aggregations:
- "How many transactions in the last hour?"
- "What's the velocity: events per minute?"
- "How many unique countries this week?"

**gofeat gives you this in 10 lines of Go**:

```go
store, _ := gofeat.New(gofeat.Config{
    Features: []gofeat.Feature{
        {Name: "tx_velocity", Aggregate: gofeat.Velocity(time.Hour), Window: gofeat.Sliding(5 * time.Minute)},
        {Name: "unique_cards", Aggregate: gofeat.UniqueRatio("card"), Window: gofeat.Sliding(5 * time.Minute)},
    },
})
```

### What You Get

✅ **Zero dependencies** - pure Go stdlib, no Redis/Kafka/Docker  
✅ **Fast** - 1μs latency, 2M events/sec throughput (benchmarks below)  
✅ **Point-in-time correct** - for ML training without data leakage  
✅ **Fraud-specific aggregators** - velocity, entropy, unique ratio built-in  
✅ **Production-ready** - thread-safe, tested, 90% coverage

## Quick Start

### Basic Fraud Detection

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/w0rng/gofeat"
)

func main() {
    store, err := gofeat.New(gofeat.Config{
        TTL: 24 * time.Hour,
        Features: []gofeat.Feature{
            // Velocity: transactions per minute
            {Name: "tx_velocity", Aggregate: gofeat.Velocity(time.Hour), Window: gofeat.Sliding(5 * time.Minute)},
            
            // Count: total transactions
            {Name: "tx_count_5min", Aggregate: gofeat.Count, Window: gofeat.Sliding(5 * time.Minute)},
            
            // Unique ratio: 1.0 = all different cards (suspicious)
            {Name: "card_diversity", Aggregate: gofeat.UniqueRatio("card"), Window: gofeat.Sliding(5 * time.Minute)},
            
            // Average amount
            {Name: "avg_amount", Aggregate: gofeat.Mean("amount"), Window: gofeat.Sliding(5 * time.Minute)},
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    defer store.Close()

    ctx := context.Background()

    // Push transaction event
    err = store.Push(ctx, "user_123",
        gofeat.Event{
            Timestamp: time.Now().UTC(),
            Data: map[string]any{
                "amount": 100.50,
                "card":   "1234",
            },
        },
    )
    if err != nil {
        log.Fatal(err)
    }

    // Get features for real-time fraud scoring
    result, err := store.Get(ctx, "user_123")
    if err != nil {
        log.Fatal(err)
    }

    velocity := result.FloatOr("tx_velocity", 0)
    cardDiversity := result.FloatOr("card_diversity", 0)
    avgAmount := result.FloatOr("avg_amount", 0)

    // Simple fraud detection logic
    if velocity > 3.0 && cardDiversity > 0.8 && avgAmount < 10 {
        log.Println("🚨 CARD TESTING DETECTED")
    }
}
```

## Fraud-Specific Aggregators

gofeat includes aggregators designed specifically for fraud detection:

| Aggregator | Returns | Use Case |
|------------|---------|----------|
| `Velocity(window)` | float64 | Events per minute - detect velocity abuse |
| `Entropy(field)` | float64 | Shannon entropy - detect diversity attacks |
| `UniqueRatio(field)` | float64 | Unique/total ratio - detect card testing |
| `TimeSinceFirst()` | Duration | Account age - flag new accounts |
| `Percentile(field, p)` | float64 | P95/P99 - detect outliers |
| `StandardDeviation(field)` | float64 | Std dev - calculate Z-scores |
| `Mean(field)` | float64 | Average value |

### Basic Aggregators

| Aggregator | Returns | Use Case |
|------------|---------|----------|
| `Count` | int | Event frequency |
| `Sum(field)` | float64 | Transaction volume |
| `Min(field)` | float64 | Minimum amount |
| `Max(field)` | float64 | Maximum amount |
| `Last(field)` | any | Last country/device |
| `DistinctCount(field)` | int | Unique countries/cards |

## Windows

```go
// Last 5 minutes
gofeat.Sliding(5 * time.Minute)

// Last 24 hours
gofeat.Sliding(24 * time.Hour)

// All time (no window)
gofeat.Lifetime()
```

## Point-in-Time Queries

For ML training, you need features computed at the time of each event, not current time. This prevents data leakage.

```go
// Real-time serving (uses current time)
result, _ := store.Get(ctx, "user_123")

// Historical query for ML training (uses event time)
result, _ := store.GetAt(ctx, "user_123", eventTimestamp)
```

See [examples/point-in-time](examples/point-in-time) for a complete ML training pipeline.

## Batch Operations

Push multiple events at once:

```go
store.Push(ctx, "user_123",
    gofeat.Event{Timestamp: t1, Data: data1},
    gofeat.Event{Timestamp: t2, Data: data2},
    gofeat.Event{Timestamp: t3, Data: data3},
)

// Batch get for multiple entities
results, _ := store.BatchGet(ctx, "user_1", "user_2", "user_3")
```

## Monitoring

```go
stats, _ := store.Stats(ctx)
log.Printf("entities: %d, events: %d", stats.Entities, stats.TotalEvents)
```

## Performance

Benchmarked on AMD Ryzen 5 5600 (6-core):

| Operation | Latency | Throughput | Memory |
|-----------|---------|------------|--------|
| `Push` (single) | 530 ns | ~2M events/sec | 603 B/op |
| `Get` | 1.1 μs | ~900K ops/sec | 3.8 KB/op |
| `GetAt` (point-in-time) | 17 μs | ~60K ops/sec | 33 KB/op |
| `Aggregation` (Count) | 30 ns | ~33M ops/sec | 48 B/op |

**Parallel performance**:
- `Push` (concurrent): 838 ns/op
- `Get` (concurrent): 956 ns/op

Run benchmarks: `go test -bench=. -benchmem`

### vs Feast/Tecton

| | gofeat | Feast | Tecton |
|---|--------|-------|--------|
| **Latency** | 1 μs | ~10-50 ms | ~5-20 ms |
| **Dependencies** | None | Redis/DynamoDB | Kafka/Cloud |
| **Setup time** | 30 seconds | Hours/Days | Days/Weeks |
| **Cost** | Free | Infrastructure | $$$$ |
| **Point-in-time** | ✅ Native | ⚠️ Complex | ✅ Native |

gofeat is **10-50x faster** for single-service use cases (up to 100K events/sec).

Use Feast/Tecton when you need:
- Multi-service feature sharing
- Petabyte-scale feature storage
- Enterprise support

## Custom Storage

Implement the `Storage` interface for custom backends:

```go
type Storage interface {
    Push(ctx context.Context, entityID string, events ...Event) error
    Get(ctx context.Context, entityID string, at time.Time) ([]Event, error)
    Evict(ctx context.Context) error
    Stats(ctx context.Context) (StorageStats, error)
    Close() error
}

// Example: Redis storage
store, _ := gofeat.New(gofeat.Config{
    Storage:  NewRedisStorage(redisClient, 24*time.Hour),
    Features: features,
})
```

See [examples/custom-storage](examples/custom-storage) for PostgreSQL implementation.

**Note**: Storage implementations are responsible for:
- Applying TTL filtering in the `Get` method
- Evicting old events in the `Evict` method based on their internal TTL
- Keeping events sorted by timestamp per entity

## Custom Aggregators

Implement the `Aggregator` interface:

```go
type Aggregator interface {
    Add(e Event)
    Result() any
}

// Example: Median aggregator
type medianAgg struct {
    values []float64
    field  string
}

func (a *medianAgg) Add(e Event) {
    v, ok := e.Data[a.field]
    if !ok {
        return
    }
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

func Median(field string) gofeat.AggregatorFactory {
    return func() gofeat.Aggregator {
        return &medianAgg{field: field}
    }
}

// Usage
{Name: "median_amount", Aggregate: Median("amount"), Window: gofeat.Sliding(time.Hour)}
```

See [examples/custom-aggregator](examples/custom-aggregator) for a complete example.

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

## Examples

- [basic](examples/basic) - Simple transaction counting
- [fraud-detection](examples/fraud-detection) - Multi-feature fraud scoring
- [card-testing](examples/card-testing) - Detect card testing attacks with velocity + diversity
- [point-in-time](examples/point-in-time) - ML training without data leakage
- [custom-storage](examples/custom-storage) - PostgreSQL storage backend
- [custom-aggregator](examples/custom-aggregator) - Custom median aggregator

## Design Decisions

### Event Time vs Processing Time

gofeat uses **event time** (timestamp from the event) rather than processing time (when event was received). This ensures correct feature computation for historical queries but means results may change if late events arrive.

### Lazy Eviction

Expired events are cleaned up during reads, not in background. This simplifies implementation and avoids concurrency issues. For high-throughput scenarios with many inactive entities, consider periodic cleanup with `store.Evict(ctx)`.

### Data Types

Events with missing fields or wrong types are silently skipped. This handles sparse data gracefully without crashing your pipeline.

## Limitations

- **In-memory by default** - data doesn't survive restarts (use custom storage for persistence)
- **No deduplication** - duplicate events are counted twice (handle at the application level)
- **UTC required** - all timestamps must be UTC
- **Single-service** - designed for 10K-100K events/sec, not distributed petabyte-scale

For distributed, petabyte-scale feature stores, use Feast or Tecton.

## Related Projects

- [Feast](https://feast.dev/) - Full-featured feature store with infrastructure requirements
- [Tecton](https://tecton.ai/) - Enterprise feature platform
- [Feathr](https://github.com/feathr-ai/feathr) - Feature store from LinkedIn
