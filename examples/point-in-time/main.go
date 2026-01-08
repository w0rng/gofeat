package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/w0rng/gofeat"
)

// Demonstrates point-in-time correct feature computation for ML training
// This prevents data leakage by computing features as they were at event time
func main() {
	store, err := gofeat.New(gofeat.Config{
		TTL: 7 * 24 * time.Hour,
		Features: []gofeat.Feature{
			{Name: "tx_count_1h", Aggregate: gofeat.Count, Window: gofeat.Sliding(time.Hour)},
			{Name: "tx_sum_1h", Aggregate: gofeat.Sum("amount"), Window: gofeat.Sliding(time.Hour)},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	ctx := context.Background()
	userID := "user_789"

	// Historical transactions with labels (for ML training)
	transactions := []struct {
		event   gofeat.Event
		isFraud bool
	}{
		{gofeat.Event{Timestamp: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC), Data: map[string]any{"amount": 50.0}}, false},
		{gofeat.Event{Timestamp: time.Date(2024, 1, 1, 10, 30, 0, 0, time.UTC), Data: map[string]any{"amount": 75.0}}, false},
		{gofeat.Event{Timestamp: time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC), Data: map[string]any{"amount": 100.0}}, false},
		{gofeat.Event{Timestamp: time.Date(2024, 1, 1, 11, 15, 0, 0, time.UTC), Data: map[string]any{"amount": 500.0}}, true},
		{gofeat.Event{Timestamp: time.Date(2024, 1, 1, 11, 20, 0, 0, time.UTC), Data: map[string]any{"amount": 600.0}}, true},
	}

	// Push all events first
	for _, tx := range transactions {
		if err := store.Push(ctx, userID, tx.event); err != nil {
			log.Fatal(err)
		}
	}

	// Generate training dataset with point-in-time correct features
	fmt.Println("=== ML Training Dataset (Point-in-Time Correct) ===")
	fmt.Println()
	fmt.Printf("%-20s | %-8s | %-10s | %-8s | %s\n", "Timestamp", "Amount", "tx_count", "tx_sum", "Label")
	fmt.Println("---------------------|----------|------------|----------|-------")

	for _, tx := range transactions {
		// GetAt computes features as they were BEFORE this transaction
		// This is what you'd have known at decision time
		result, err := store.GetAt(ctx, userID, tx.event.Timestamp.Add(-time.Millisecond))
		if err != nil {
			log.Fatal(err)
		}

		label := "legit"
		if tx.isFraud {
			label = "FRAUD"
		}

		fmt.Printf("%-20s | $%-7.0f | %-10d | $%-7.0f | %s\n",
			tx.event.Timestamp.Format("2006-01-02 15:04:05"),
			tx.event.Data["amount"],
			result.IntOr("tx_count_1h", -1),
			result.FloatOr("tx_sum_1h", -1),
			label,
		)
	}

	fmt.Println()
	fmt.Println("Notice how tx_count and tx_sum increase for each transaction,")
	fmt.Println("but each row only sees events that happened BEFORE it.")
	fmt.Println("This prevents data leakage in ML training.")

	// Compare with incorrect approach (using current time)
	fmt.Println()
	fmt.Println("=== WRONG: Using Get() instead of GetAt() ===")
	fmt.Println("This would leak future information into all training examples!")
	fmt.Println()

	result, _ := store.Get(ctx, userID)
	fmt.Printf("All rows would show: tx_count=%d, tx_sum=$%.0f\n",
		result.IntOr("tx_count_1h", -1),
		result.FloatOr("tx_sum_1h", -1),
	)
}
