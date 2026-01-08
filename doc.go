// Package gofeat provides an embedded feature store for real-time feature computation.
//
// It is designed for fraud detection and ML pipelines where you need
// to compute aggregations over time windows without external dependencies.
//
// Basic usage:
//
//	store, err := gofeat.New(gofeat.Config{
//	    TTL: 24 * time.Hour,
//	    Features: []gofeat.Feature{
//	        {Name: "tx_count_1h", Aggregate: gofeat.Count, Window: gofeat.Sliding(time.Hour)},
//	        {Name: "tx_sum_1h", Aggregate: gofeat.Sum, Field: "amount", Window: gofeat.Sliding(time.Hour)},
//	    },
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer store.Close()
//
//	// Push events
//	store.Push(ctx, "user_123", gofeat.Event{
//	    Timestamp: time.Now().UTC(),
//	    Data: map[string]any{"amount": 100.0},
//	})
//
//	// Get features
//	result, _ := store.Get(ctx, "user_123")
//	count := result.IntOr("tx_count_1h", -1)
package gofeat
