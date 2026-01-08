package main

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/w0rng/gofeat"
)

// Custom aggregator: Median
type medianAgg struct {
	values []float64
	field  string
}

func (a *medianAgg) Add(data map[string]any) {
	v, ok := data[a.field]
	if !ok {
		return
	}
	switch n := v.(type) {
	case float64:
		a.values = append(a.values, n)
	case int:
		a.values = append(a.values, float64(n))
	}
}

func (a *medianAgg) Result() any {
	if len(a.values) == 0 {
		return 0.0
	}
	sorted := make([]float64, len(a.values))
	copy(sorted, a.values)
	sort.Float64s(sorted)

	n := len(sorted)
	if n%2 == 0 {
		return (sorted[n/2-1] + sorted[n/2]) / 2
	}
	return sorted[n/2]
}

// Median is a custom aggregator factory
func Median(field string) gofeat.AggregatorFactory {
	return func() gofeat.Aggregator {
		return &medianAgg{field: field}
	}
}

// Custom aggregator: Percentile
type percentileAgg struct {
	values     []float64
	percentile float64
	field      string
}

func Percentile(field string, p float64) gofeat.AggregatorFactory {
	return func() gofeat.Aggregator {
		return &percentileAgg{field: field, percentile: p}
	}
}

func (a *percentileAgg) Add(data map[string]any) {
	v, ok := data[a.field]
	if !ok {
		return
	}
	switch n := v.(type) {
	case float64:
		a.values = append(a.values, n)
	case int:
		a.values = append(a.values, float64(n))
	}
}

func (a *percentileAgg) Result() any {
	if len(a.values) == 0 {
		return 0.0
	}
	sorted := make([]float64, len(a.values))
	copy(sorted, a.values)
	sort.Float64s(sorted)

	idx := int(float64(len(sorted)-1) * a.percentile / 100)
	return sorted[idx]
}

func main() {
	store, err := gofeat.New(gofeat.Config{
		TTL: 24 * time.Hour,
		Features: []gofeat.Feature{
			// Built-in aggregators
			{Name: "count", Aggregate: gofeat.Count, Window: gofeat.Sliding(time.Hour)},
			{Name: "sum", Aggregate: gofeat.Sum("amount"), Window: gofeat.Sliding(time.Hour)},
			{Name: "min", Aggregate: gofeat.Min("amount"), Window: gofeat.Sliding(time.Hour)},
			{Name: "max", Aggregate: gofeat.Max("amount"), Window: gofeat.Sliding(time.Hour)},

			// Custom aggregators
			{Name: "median", Aggregate: Median("amount"), Window: gofeat.Sliding(time.Hour)},
			{Name: "p90", Aggregate: Percentile("amount", 90), Window: gofeat.Sliding(time.Hour)},
			{Name: "p95", Aggregate: Percentile("amount", 95), Window: gofeat.Sliding(time.Hour)},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	ctx := context.Background()

	// Generate some transaction data
	amounts := []float64{10, 20, 30, 50, 75, 100, 150, 200, 500, 1000}
	now := time.Now().UTC()

	for i, amount := range amounts {
		err := store.Push(ctx, "user_1", gofeat.Event{
			Timestamp: now.Add(-time.Duration(len(amounts)-i) * time.Minute),
			Data:      map[string]any{"amount": amount},
		})
		if err != nil {
			log.Fatal(err)
		}
	}

	result, err := store.Get(ctx, "user_1")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== Custom Aggregators Demo ===")
	fmt.Println()
	fmt.Printf("Input amounts: %v\n", amounts)
	fmt.Println()
	fmt.Println("Built-in aggregators:")
	fmt.Printf("  Count:  %d\n", result.IntOr("count", -1))
	fmt.Printf("  Sum:    $%.2f\n", result.FloatOr("sum", -1))
	fmt.Printf("  Min:    $%.2f\n", result.FloatOr("min", -1))
	fmt.Printf("  Max:    $%.2f\n", result.FloatOr("max", -1))
	fmt.Println()
	fmt.Println("Custom aggregators:")
	fmt.Printf("  Median: $%.2f\n", result.FloatOr("median", -1))
	fmt.Printf("  P90:    $%.2f\n", result.FloatOr("p90", -1))
	fmt.Printf("  P95:    $%.2f\n", result.FloatOr("p95", -1))
}
