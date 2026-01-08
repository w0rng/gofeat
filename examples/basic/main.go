package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/w0rng/gofeat"
)

func main() {
	store, err := gofeat.New(gofeat.Config{
		TTL: 24 * time.Hour,
		Features: []gofeat.Feature{
			{Name: "tx_count_1h", Aggregate: gofeat.Count, Window: gofeat.Sliding(time.Hour)},
			{Name: "tx_sum_1h", Aggregate: gofeat.Sum("amount"), Window: gofeat.Sliding(time.Hour)},
			{Name: "tx_max_1h", Aggregate: gofeat.Max("amount"), Window: gofeat.Sliding(time.Hour)},
			{Name: "last_country", Aggregate: gofeat.Last("country")},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	ctx := context.Background()

	// Push some events
	events := []gofeat.Event{
		{Timestamp: time.Now().UTC().Add(-30 * time.Minute), Data: map[string]any{"amount": 100.0, "country": "US"}},
		{Timestamp: time.Now().UTC().Add(-20 * time.Minute), Data: map[string]any{"amount": 250.0, "country": "CA"}},
		{Timestamp: time.Now().UTC().Add(-10 * time.Minute), Data: map[string]any{"amount": 75.0, "country": "US"}},
	}

	if err := store.Push(ctx, "user_123", events...); err != nil {
		log.Fatal(err)
	}

	// Get features
	result, err := store.Get(ctx, "user_123")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Transaction count (1h): %d\n", result.IntOr("tx_count_1h", -1))
	fmt.Printf("Transaction sum (1h):   %.2f\n", result.FloatOr("tx_sum_1h", -1))
	fmt.Printf("Transaction max (1h):   %.2f\n", result.FloatOr("tx_max_1h", -1))
	fmt.Printf("Last country:           %s\n", result.StringOr("last_country", ""))

	// Stats
	stats, _ := store.Stats(ctx)
	fmt.Printf("\nStore stats: %d entities, %d events\n", stats.Entities, stats.TotalEvents)
}
