package gofeat_test

import (
	"context"
	"testing"
	"time"

	"github.com/w0rng/gofeat"
)

func BenchmarkStore_Push_Single(b *testing.B) {
	store, _ := gofeat.New(gofeat.Config{
		Features: []gofeat.Feature{
			{Name: "count", Aggregate: gofeat.Count},
		},
	})
	defer store.Close()

	ctx := context.Background()
	now := time.Now().UTC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.Push(ctx, "user1", gofeat.Event{
			Timestamp: now.Add(time.Duration(i) * time.Second),
			Data:      map[string]any{"amount": 100.0},
		})
	}
}

func BenchmarkStore_Push_Batch10(b *testing.B) {
	store, _ := gofeat.New(gofeat.Config{
		Features: []gofeat.Feature{
			{Name: "count", Aggregate: gofeat.Count},
		},
	})
	defer store.Close()

	ctx := context.Background()
	now := time.Now().UTC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		events := make([]gofeat.Event, 10)
		for j := 0; j < 10; j++ {
			events[j] = gofeat.Event{
				Timestamp: now.Add(time.Duration(i*10+j) * time.Second),
				Data:      map[string]any{"amount": 100.0},
			}
		}
		store.Push(ctx, "user1", events...)
	}
}

func BenchmarkStore_Get(b *testing.B) {
	store, _ := gofeat.New(gofeat.Config{
		Features: []gofeat.Feature{
			{Name: "count", Aggregate: gofeat.Count},
			{Name: "sum", Aggregate: gofeat.Sum, Field: "amount"},
		},
	})
	defer store.Close()

	ctx := context.Background()
	now := time.Now().UTC()

	// Prepopulate with 100 events
	events := make([]gofeat.Event, 100)
	for i := 0; i < 100; i++ {
		events[i] = gofeat.Event{
			Timestamp: now.Add(time.Duration(i) * time.Second),
			Data:      map[string]any{"amount": 100.0},
		}
	}
	store.Push(ctx, "user1", events...)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.Get(ctx, "user1")
	}
}

func BenchmarkStore_GetAt(b *testing.B) {
	store, _ := gofeat.New(gofeat.Config{
		TTL: 1 * time.Hour,
		Features: []gofeat.Feature{
			{Name: "count", Aggregate: gofeat.Count, Window: gofeat.Sliding(30 * time.Minute)},
			{Name: "sum", Aggregate: gofeat.Sum, Field: "amount", Window: gofeat.Sliding(30 * time.Minute)},
		},
	})
	defer store.Close()

	ctx := context.Background()
	now := time.Now().UTC()

	// Prepopulate
	events := make([]gofeat.Event, 1000)
	for i := 0; i < 1000; i++ {
		events[i] = gofeat.Event{
			Timestamp: now.Add(time.Duration(i-500) * time.Second),
			Data:      map[string]any{"amount": 100.0},
		}
	}
	store.Push(ctx, "user1", events...)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.GetAt(ctx, "user1", now)
	}
}

func BenchmarkStore_BatchGet(b *testing.B) {
	store, _ := gofeat.New(gofeat.Config{
		Features: []gofeat.Feature{
			{Name: "count", Aggregate: gofeat.Count},
		},
	})
	defer store.Close()

	ctx := context.Background()
	now := time.Now().UTC()

	// Prepopulate 10 entities with 100 events each
	for i := 0; i < 10; i++ {
		events := make([]gofeat.Event, 100)
		for j := 0; j < 100; j++ {
			events[j] = gofeat.Event{
				Timestamp: now.Add(time.Duration(j) * time.Second),
				Data:      map[string]any{"amount": 100.0},
			}
		}
		store.Push(ctx, "user"+string(rune('0'+i)), events...)
	}

	entityIDs := make([]string, 10)
	for i := 0; i < 10; i++ {
		entityIDs[i] = "user" + string(rune('0'+i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.BatchGet(ctx, entityIDs...)
	}
}

func BenchmarkStore_MultipleFeatures(b *testing.B) {
	store, _ := gofeat.New(gofeat.Config{
		Features: []gofeat.Feature{
			{Name: "count", Aggregate: gofeat.Count},
			{Name: "sum", Aggregate: gofeat.Sum, Field: "amount"},
			{Name: "min", Aggregate: gofeat.Min, Field: "amount"},
			{Name: "max", Aggregate: gofeat.Max, Field: "amount"},
			{Name: "last_country", Aggregate: gofeat.Last, Field: "country"},
			{Name: "distinct_countries", Aggregate: gofeat.CountDistinct, Field: "country"},
		},
	})
	defer store.Close()

	ctx := context.Background()
	now := time.Now().UTC()

	// Prepopulate
	events := make([]gofeat.Event, 100)
	for i := 0; i < 100; i++ {
		events[i] = gofeat.Event{
			Timestamp: now.Add(time.Duration(i) * time.Second),
			Data: map[string]any{
				"amount":  float64(i * 10),
				"country": []string{"US", "CA", "MX"}[i%3],
			},
		}
	}
	store.Push(ctx, "user1", events...)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.Get(ctx, "user1")
	}
}

func BenchmarkStorage_Push(b *testing.B) {
	s := gofeat.NewMemoryStorage()
	ctx := context.Background()
	now := time.Now().UTC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Push(ctx, "user1", gofeat.Event{
			Timestamp: now.Add(time.Duration(i) * time.Second),
			Data:      map[string]any{"amount": 100.0},
		})
	}
}

func BenchmarkStorage_Get(b *testing.B) {
	s := gofeat.NewMemoryStorage()
	ctx := context.Background()
	now := time.Now().UTC()

	// Prepopulate
	events := make([]gofeat.Event, 1000)
	for i := 0; i < 1000; i++ {
		events[i] = gofeat.Event{
			Timestamp: now.Add(time.Duration(i) * time.Second),
			Data:      map[string]any{"amount": 100.0},
		}
	}
	s.Push(ctx, "user1", events...)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Get(ctx, "user1")
	}
}

func BenchmarkWindow_Sliding(b *testing.B) {
	now := time.Now().UTC()
	events := make([]gofeat.Event, 1000)
	for i := 0; i < 1000; i++ {
		events[i] = gofeat.Event{
			Timestamp: now.Add(time.Duration(i-500) * time.Minute),
			Data:      map[string]any{},
		}
	}

	window := gofeat.Sliding(1 * time.Hour)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		window.Select(events, now)
	}
}

func BenchmarkWindow_Lifetime(b *testing.B) {
	now := time.Now().UTC()
	events := make([]gofeat.Event, 1000)
	for i := 0; i < 1000; i++ {
		events[i] = gofeat.Event{
			Timestamp: now.Add(time.Duration(i-500) * time.Minute),
			Data:      map[string]any{},
		}
	}

	window := gofeat.Lifetime()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		window.Select(events, now)
	}
}

// Parallel benchmarks.
func BenchmarkStore_PushParallel(b *testing.B) {
	store, _ := gofeat.New(gofeat.Config{
		Features: []gofeat.Feature{
			{Name: "count", Aggregate: gofeat.Count},
		},
	})
	defer store.Close()

	ctx := context.Background()
	now := time.Now().UTC()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			store.Push(ctx, "user1", gofeat.Event{
				Timestamp: now.Add(time.Duration(i) * time.Second),
				Data:      map[string]any{"amount": 100.0},
			})
			i++
		}
	})
}

func BenchmarkStore_GetParallel(b *testing.B) {
	store, _ := gofeat.New(gofeat.Config{
		Features: []gofeat.Feature{
			{Name: "count", Aggregate: gofeat.Count},
		},
	})
	defer store.Close()

	ctx := context.Background()
	now := time.Now().UTC()

	// Prepopulate
	events := make([]gofeat.Event, 100)
	for i := 0; i < 100; i++ {
		events[i] = gofeat.Event{
			Timestamp: now.Add(time.Duration(i) * time.Second),
			Data:      map[string]any{"amount": 100.0},
		}
	}
	store.Push(ctx, "user1", events...)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			store.Get(ctx, "user1")
		}
	})
}

func BenchmarkStore_MixedReadWrite(b *testing.B) {
	store, _ := gofeat.New(gofeat.Config{
		Features: []gofeat.Feature{
			{Name: "count", Aggregate: gofeat.Count},
		},
	})
	defer store.Close()

	ctx := context.Background()
	now := time.Now().UTC()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i%10 == 0 {
				// 10% writes
				store.Push(ctx, "user1", gofeat.Event{
					Timestamp: now.Add(time.Duration(i) * time.Second),
					Data:      map[string]any{},
				})
			} else {
				// 90% reads
				store.Get(ctx, "user1")
			}
			i++
		}
	})
}
