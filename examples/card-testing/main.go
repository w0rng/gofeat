package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/w0rng/gofeat"
)

// Card testing detection: fraudsters test stolen cards with small transactions
// Pattern: 10-20 transactions of $1-5 within 5 minutes, each with different card
func main() {
	store, err := gofeat.New(gofeat.Config{
		TTL: 24 * time.Hour,
		Features: []gofeat.Feature{
			// Velocity features
			{Name: "tx_velocity", Aggregate: gofeat.Velocity(time.Hour), Window: gofeat.Sliding(5 * time.Minute)},
			{Name: "tx_count_5min", Aggregate: gofeat.Count, Window: gofeat.Sliding(5 * time.Minute)},

			// Card diversity features
			{Name: "unique_cards_ratio", Aggregate: gofeat.UniqueRatio("card_last4"), Window: gofeat.Sliding(5 * time.Minute)},
			{Name: "distinct_cards", Aggregate: gofeat.DistinctCount("card_last4"), Window: gofeat.Sliding(5 * time.Minute)},

			// Amount features
			{Name: "avg_amount_5min", Aggregate: gofeat.Mean("amount"), Window: gofeat.Sliding(5 * time.Minute)},
			{Name: "max_amount_5min", Aggregate: gofeat.Max("amount"), Window: gofeat.Sliding(5 * time.Minute)},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	ctx := context.Background()

	fmt.Println("=== Card Testing Detection Demo ===")
	fmt.Println()

	// Scenario 1: Normal user behavior
	fmt.Println("Scenario 1: Normal User")
	fmt.Println("------------------------")
	testNormalUser(ctx, store)
	fmt.Println()

	// Scenario 2: Card testing attack
	fmt.Println("Scenario 2: Card Testing Attack")
	fmt.Println("--------------------------------")
	testCardTestingAttack(ctx, store)
}

func testNormalUser(ctx context.Context, store *gofeat.Store) {
	userID := "user_normal"
	now := time.Now().UTC()

	// Normal user: 3 transactions in 5 minutes, same card
	events := []gofeat.Event{
		{
			Timestamp: now.Add(-4 * time.Minute),
			Data: map[string]any{
				"amount":     45.99,
				"card_last4": "1234",
			},
		},
		{
			Timestamp: now.Add(-2 * time.Minute),
			Data: map[string]any{
				"amount":     12.50,
				"card_last4": "1234",
			},
		},
		{
			Timestamp: now,
			Data: map[string]any{
				"amount":     89.00,
				"card_last4": "1234",
			},
		},
	}

	if err := store.Push(ctx, userID, events...); err != nil {
		log.Fatal(err)
	}

	result, err := store.Get(ctx, userID)
	if err != nil {
		log.Fatal(err)
	}

	printFeatures(result)
	verdict := analyzeRisk(result)
	fmt.Printf("\n%s\n", verdict)
}

func testCardTestingAttack(ctx context.Context, store *gofeat.Store) {
	userID := "user_fraudster"
	now := time.Now().UTC()

	// Card testing: 15 small transactions in 5 minutes, all different cards
	var events []gofeat.Event
	for i := range 15 {
		events = append(events, gofeat.Event{
			Timestamp: now.Add(-5*time.Minute + time.Duration(i)*20*time.Second),
			Data: map[string]any{
				"amount":     1.0 + rand.Float64()*4.0, // $1-5
				"card_last4": fmt.Sprintf("%04d", 1000+i),
			},
		})
	}

	if err := store.Push(ctx, userID, events...); err != nil {
		log.Fatal(err)
	}

	result, err := store.Get(ctx, userID)
	if err != nil {
		log.Fatal(err)
	}

	printFeatures(result)
	verdict := analyzeRisk(result)
	fmt.Printf("\n%s\n", verdict)
}

func printFeatures(result gofeat.Result) {
	velocity := result.FloatOr("tx_velocity", 0)
	count := result.IntOr("tx_count_5min", 0)
	uniqueRatio := result.FloatOr("unique_cards_ratio", 0)
	distinctCards := result.IntOr("distinct_cards", 0)
	avgAmount := result.FloatOr("avg_amount_5min", 0)
	maxAmount := result.FloatOr("max_amount_5min", 0)

	fmt.Printf("Transactions (5 min):     %d\n", count)
	fmt.Printf("Velocity:                 %.1f tx/min\n", velocity)
	fmt.Printf("Distinct cards:           %d\n", distinctCards)
	fmt.Printf("Unique cards ratio:       %.2f\n", uniqueRatio)
	fmt.Printf("Average amount:           $%.2f\n", avgAmount)
	fmt.Printf("Max amount:               $%.2f\n", maxAmount)
}

func analyzeRisk(result gofeat.Result) string {
	velocity := result.FloatOr("tx_velocity", 0)
	count := result.IntOr("tx_count_5min", 0)
	uniqueRatio := result.FloatOr("unique_cards_ratio", 0)
	avgAmount := result.FloatOr("avg_amount_5min", 0)

	riskScore := 0
	reasons := []string{}

	// Rule 1: High velocity (>2 tx/min)
	if velocity > 2.0 {
		riskScore += 30
		reasons = append(reasons, fmt.Sprintf("high velocity (%.1f tx/min)", velocity))
	}

	// Rule 2: Many transactions in short time
	if count >= 10 {
		riskScore += 25
		reasons = append(reasons, fmt.Sprintf("high transaction count (%d in 5 min)", count))
	}

	// Rule 3: High card diversity (>0.8 means almost all unique)
	if uniqueRatio > 0.8 {
		riskScore += 30
		reasons = append(reasons, fmt.Sprintf("high card diversity (%.0f%% unique)", uniqueRatio*100))
	}

	// Rule 4: Small amounts (avg < $10)
	if avgAmount > 0 && avgAmount < 10.0 {
		riskScore += 15
		reasons = append(reasons, fmt.Sprintf("small amounts (avg $%.2f)", avgAmount))
	}

	// Build verdict
	verdict := fmt.Sprintf("Risk Score: %d/100\n", riskScore)
	if len(reasons) > 0 {
		verdict += "Risk Factors:\n"
		for _, r := range reasons {
			verdict += fmt.Sprintf("  • %s\n", r)
		}
	}

	if riskScore >= 60 {
		verdict += "\n🚨 CARD TESTING DETECTED - BLOCK IMMEDIATELY"
	} else if riskScore >= 40 {
		verdict += "\n⚠️  Suspicious activity - Flag for review"
	} else {
		verdict += "\n✅ Normal activity"
	}

	return verdict
}
