package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/w0rng/gofeat"
)

// Simple fraud detection based on velocity and diversity checks
func main() {
	store, err := gofeat.New(gofeat.Config{
		TTL: 24 * time.Hour,
		Features: []gofeat.Feature{
			// Velocity features
			{Name: "tx_count_1h", Aggregate: gofeat.Count, Window: gofeat.Sliding(time.Hour)},
			{Name: "tx_count_24h", Aggregate: gofeat.Count, Window: gofeat.Sliding(24 * time.Hour)},
			{Name: "tx_sum_1h", Aggregate: gofeat.Sum("amount"), Window: gofeat.Sliding(time.Hour)},

			// Diversity features
			{Name: "countries_1h", Aggregate: gofeat.CountDistinct("country"), Window: gofeat.Sliding(time.Hour)},
			{Name: "devices_24h", Aggregate: gofeat.CountDistinct("device_id"), Window: gofeat.Sliding(24 * time.Hour)},

			// Amount features
			{Name: "max_amount_24h", Aggregate: gofeat.Max("amount"), Window: gofeat.Sliding(24 * time.Hour)},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	ctx := context.Background()

	// Simulate transaction history for a user
	userID := "user_456"
	now := time.Now().UTC()

	history := []gofeat.Event{
		{Timestamp: now.Add(-23 * time.Hour), Data: map[string]any{"amount": 50.0, "country": "US", "device_id": "dev_1"}},
		{Timestamp: now.Add(-22 * time.Hour), Data: map[string]any{"amount": 75.0, "country": "US", "device_id": "dev_1"}},
		{Timestamp: now.Add(-5 * time.Hour), Data: map[string]any{"amount": 100.0, "country": "US", "device_id": "dev_1"}},
	}

	if err := store.Push(ctx, userID, history...); err != nil {
		log.Fatal(err)
	}

	// New transaction comes in - check if it's suspicious
	newTx := gofeat.Event{
		Timestamp: now,
		Data: map[string]any{
			"amount":    500.0,
			"country":   "RU",
			"device_id": "dev_2",
		},
	}

	// Push and evaluate
	if err := store.Push(ctx, userID, newTx); err != nil {
		log.Fatal(err)
	}

	result, err := store.Get(ctx, userID)
	if err != nil {
		log.Fatal(err)
	}

	// Extract features
	txCount1h := result.IntOr("tx_count_1h", -1)
	txCount24h := result.IntOr("tx_count_24h", -1)
	txSum1h := result.FloatOr("tx_sum_1h", -1)
	countries1h := result.IntOr("countries_1h", -1)
	devices24h := result.IntOr("devices_24h", -1)
	maxAmount24h := result.FloatOr("max_amount_24h", -1)

	fmt.Println("=== Fraud Detection Analysis ===")
	fmt.Printf("Transactions (1h):    %d\n", txCount1h)
	fmt.Printf("Transactions (24h):   %d\n", txCount24h)
	fmt.Printf("Sum (1h):             $%.2f\n", txSum1h)
	fmt.Printf("Countries (1h):       %d\n", countries1h)
	fmt.Printf("Devices (24h):        %d\n", devices24h)
	fmt.Printf("Max amount (24h):     $%.2f\n", maxAmount24h)
	fmt.Println()

	// Simple rule-based fraud detection
	riskScore := 0
	reasons := []string{}

	if countries1h > 1 {
		riskScore += 30
		reasons = append(reasons, fmt.Sprintf("multiple countries in 1h (%d)", countries1h))
	}

	if devices24h > 1 {
		riskScore += 20
		reasons = append(reasons, fmt.Sprintf("multiple devices in 24h (%d)", devices24h))
	}

	if maxAmount24h > 400 {
		riskScore += 25
		reasons = append(reasons, fmt.Sprintf("high transaction amount ($%.2f)", maxAmount24h))
	}

	if txSum1h > 1000 {
		riskScore += 25
		reasons = append(reasons, fmt.Sprintf("high velocity ($%.2f in 1h)", txSum1h))
	}

	fmt.Printf("Risk Score: %d/100\n", riskScore)
	if len(reasons) > 0 {
		fmt.Println("Risk Factors:")
		for _, r := range reasons {
			fmt.Printf("  - %s\n", r)
		}
	}

	if riskScore >= 50 {
		fmt.Println("\n⚠️  TRANSACTION FLAGGED FOR REVIEW")
	} else {
		fmt.Println("\n✅ Transaction approved")
	}
}
