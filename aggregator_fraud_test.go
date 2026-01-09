package gofeat_test

import (
	"math"
	"testing"
	"time"

	"github.com/w0rng/gofeat"
)

func TestVelocity(t *testing.T) {
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name   string
		events []gofeat.Event
		window time.Duration
		want   float64
	}{
		{
			name: "5 events over 5 minutes",
			events: []gofeat.Event{
				{Timestamp: now},
				{Timestamp: now.Add(1 * time.Minute)},
				{Timestamp: now.Add(2 * time.Minute)},
				{Timestamp: now.Add(3 * time.Minute)},
				{Timestamp: now.Add(5 * time.Minute)},
			},
			window: time.Hour,
			want:   1.0, // 5 events / 5 minutes = 1 event/min
		},
		{
			name: "10 events in same second",
			events: []gofeat.Event{
				{Timestamp: now},
				{Timestamp: now},
				{Timestamp: now},
				{Timestamp: now},
				{Timestamp: now},
				{Timestamp: now},
				{Timestamp: now},
				{Timestamp: now},
				{Timestamp: now},
				{Timestamp: now},
			},
			window: time.Hour,
			want:   10.0 / 60.0, // 10 events per hour window = 10/60 per minute
		},
		{
			name: "2 events 1 minute apart",
			events: []gofeat.Event{
				{Timestamp: now},
				{Timestamp: now.Add(1 * time.Minute)},
			},
			window: time.Hour,
			want:   2.0, // 2 events / 1 minute = 2 events/min
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agg := gofeat.Velocity(tt.window)()
			for _, e := range tt.events {
				agg.Add(e)
			}
			result := agg.Result().(float64)
			if math.Abs(result-tt.want) > 0.01 {
				t.Errorf("velocity: got %.2f, want %.2f", result, tt.want)
			}
		})
	}
}

func TestEntropy(t *testing.T) {
	tests := []struct {
		name   string
		events []gofeat.Event
		want   float64
		field  string
	}{
		{
			name: "all same value",
			events: []gofeat.Event{
				{Data: map[string]any{"device": "dev1"}},
				{Data: map[string]any{"device": "dev1"}},
				{Data: map[string]any{"device": "dev1"}},
			},
			field: "device",
			want:  0.0, // H = 0 for single value
		},
		{
			name: "two values equally distributed",
			events: []gofeat.Event{
				{Data: map[string]any{"device": "dev1"}},
				{Data: map[string]any{"device": "dev2"}},
			},
			field: "device",
			want:  1.0, // H = 1 for two equally probable values
		},
		{
			name: "four values equally distributed",
			events: []gofeat.Event{
				{Data: map[string]any{"country": "US"}},
				{Data: map[string]any{"country": "CA"}},
				{Data: map[string]any{"country": "MX"}},
				{Data: map[string]any{"country": "BR"}},
			},
			field: "country",
			want:  2.0, // H = 2 for four equally probable values
		},
		{
			name: "skewed distribution",
			events: []gofeat.Event{
				{Data: map[string]any{"device": "dev1"}},
				{Data: map[string]any{"device": "dev1"}},
				{Data: map[string]any{"device": "dev1"}},
				{Data: map[string]any{"device": "dev2"}},
			},
			field: "device",
			want:  0.81, // H ≈ 0.81 for 75%/25% distribution
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agg := gofeat.Entropy(tt.field)()
			for _, e := range tt.events {
				agg.Add(e)
			}
			result := agg.Result().(float64)
			if math.Abs(result-tt.want) > 0.01 {
				t.Errorf("entropy: got %.2f, want %.2f", result, tt.want)
			}
		})
	}
}

func TestUniqueRatio(t *testing.T) {
	tests := []struct {
		name   string
		events []gofeat.Event
		want   float64
	}{
		{
			name: "all unique values",
			events: []gofeat.Event{
				{Data: map[string]any{"card": "1111"}},
				{Data: map[string]any{"card": "2222"}},
				{Data: map[string]any{"card": "3333"}},
			},
			want: 1.0, // 3 unique / 3 total = 1.0
		},
		{
			name: "all same value",
			events: []gofeat.Event{
				{Data: map[string]any{"card": "1111"}},
				{Data: map[string]any{"card": "1111"}},
				{Data: map[string]any{"card": "1111"}},
			},
			want: 0.33, // 1 unique / 3 total ≈ 0.33
		},
		{
			name: "half unique",
			events: []gofeat.Event{
				{Data: map[string]any{"email": "a@x.com"}},
				{Data: map[string]any{"email": "b@x.com"}},
				{Data: map[string]any{"email": "a@x.com"}},
				{Data: map[string]any{"email": "b@x.com"}},
			},
			want: 0.5, // 2 unique / 4 total = 0.5
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := "card"
			if tt.name == "half unique" {
				field = "email"
			}
			agg := gofeat.UniqueRatio(field)()
			for _, e := range tt.events {
				agg.Add(e)
			}
			result := agg.Result().(float64)
			if math.Abs(result-tt.want) > 0.01 {
				t.Errorf("unique ratio: got %.2f, want %.2f", result, tt.want)
			}
		})
	}
}

func TestTimeSinceFirst(t *testing.T) {
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name   string
		events []gofeat.Event
		want   time.Duration
	}{
		{
			name: "single event",
			events: []gofeat.Event{
				{Timestamp: now},
			},
			want: 0,
		},
		{
			name: "events 1 hour apart",
			events: []gofeat.Event{
				{Timestamp: now},
				{Timestamp: now.Add(1 * time.Hour)},
			},
			want: 1 * time.Hour,
		},
		{
			name: "events out of order",
			events: []gofeat.Event{
				{Timestamp: now.Add(1 * time.Hour)},
				{Timestamp: now}, // first event
				{Timestamp: now.Add(2 * time.Hour)},
			},
			want: 2 * time.Hour, // from earliest to latest
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agg := gofeat.TimeSinceFirst()()
			for _, e := range tt.events {
				agg.Add(e)
			}
			result := agg.Result().(time.Duration)
			if result != tt.want {
				t.Errorf("time since first: got %v, want %v", result, tt.want)
			}
		})
	}
}

func TestPercentile(t *testing.T) {
	events := []gofeat.Event{
		{Data: map[string]any{"amount": 10.0}},
		{Data: map[string]any{"amount": 20.0}},
		{Data: map[string]any{"amount": 30.0}},
		{Data: map[string]any{"amount": 40.0}},
		{Data: map[string]any{"amount": 50.0}},
		{Data: map[string]any{"amount": 60.0}},
		{Data: map[string]any{"amount": 70.0}},
		{Data: map[string]any{"amount": 80.0}},
		{Data: map[string]any{"amount": 90.0}},
		{Data: map[string]any{"amount": 100.0}},
	}

	tests := []struct {
		name string
		p    float64
		want float64
	}{
		{"p25", 0.25, 30.0},
		{"p50", 0.5, 50.0},
		{"p75", 0.75, 70.0},
		{"p95", 0.95, 90.0},
		{"p99", 0.99, 90.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agg := gofeat.Percentile("amount", tt.p)()
			for _, e := range events {
				agg.Add(e)
			}
			result := agg.Result().(float64)
			if math.Abs(result-tt.want) > 0.01 {
				t.Errorf("percentile %.2f: got %.2f, want %.2f", tt.p, result, tt.want)
			}
		})
	}
}

func TestStandardDeviation(t *testing.T) {
	tests := []struct {
		name   string
		events []gofeat.Event
		want   float64
	}{
		{
			name: "all same values",
			events: []gofeat.Event{
				{Data: map[string]any{"amount": 100.0}},
				{Data: map[string]any{"amount": 100.0}},
				{Data: map[string]any{"amount": 100.0}},
			},
			want: 0.0,
		},
		{
			name: "simple sequence",
			events: []gofeat.Event{
				{Data: map[string]any{"amount": 10.0}},
				{Data: map[string]any{"amount": 20.0}},
				{Data: map[string]any{"amount": 30.0}},
			},
			want: 8.16, // std dev ≈ 8.16
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agg := gofeat.StandardDeviation("amount")()
			for _, e := range tt.events {
				agg.Add(e)
			}
			result := agg.Result().(float64)
			if math.Abs(result-tt.want) > 0.5 {
				t.Errorf("std dev: got %.2f, want %.2f", result, tt.want)
			}
		})
	}
}

func TestMean(t *testing.T) {
	tests := []struct {
		name   string
		events []gofeat.Event
		want   float64
	}{
		{
			name: "simple average",
			events: []gofeat.Event{
				{Data: map[string]any{"amount": 100.0}},
				{Data: map[string]any{"amount": 200.0}},
				{Data: map[string]any{"amount": 300.0}},
			},
			want: 200.0,
		},
		{
			name: "single value",
			events: []gofeat.Event{
				{Data: map[string]any{"amount": 42.0}},
			},
			want: 42.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agg := gofeat.Mean("amount")()
			for _, e := range tt.events {
				agg.Add(e)
			}
			result := agg.Result().(float64)
			if math.Abs(result-tt.want) > 0.01 {
				t.Errorf("mean: got %.2f, want %.2f", result, tt.want)
			}
		})
	}
}

// Real fraud detection scenarios

func TestFraudScenario_CardTesting(t *testing.T) {
	// Card testing attack: 15 small transactions in 5 minutes
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	events := make([]gofeat.Event, 0, 15)
	for i := range 15 {
		events = append(events, gofeat.Event{
			Timestamp: now.Add(time.Duration(i) * 20 * time.Second), // Every 20 seconds
			Data: map[string]any{
				"amount": 2.0 + float64(i)*0.5,          // Small amounts $2-9
				"card":   "card_" + string(rune('0'+i)), // Different cards
			},
		})
	}

	// Test velocity
	velocityAgg := gofeat.Velocity(time.Hour)()
	for _, e := range events {
		velocityAgg.Add(e)
	}
	velocity := velocityAgg.Result().(float64)

	// 15 transactions over ~5 minutes = 3 tx/min
	if velocity < 2.5 || velocity > 3.5 {
		t.Errorf("velocity: got %.2f, want ~3.0 tx/min", velocity)
	}

	// Test unique ratio for cards
	ratioAgg := gofeat.UniqueRatio("card")()
	for _, e := range events {
		ratioAgg.Add(e)
	}
	ratio := ratioAgg.Result().(float64)

	// All different cards = 1.0 ratio (very suspicious)
	if ratio < 0.95 {
		t.Errorf("unique ratio: got %.2f, want ~1.0 (all unique cards)", ratio)
	}

	// Test mean amount
	meanAgg := gofeat.Mean("amount")()
	for _, e := range events {
		meanAgg.Add(e)
	}
	mean := meanAgg.Result().(float64)

	// Small average amount
	if mean > 10.0 {
		t.Errorf("mean amount: got %.2f, want <$10 (card testing amounts)", mean)
	}

	// VERDICT: velocity > 3 AND unique_ratio > 0.9 AND mean < 10 = CARD TESTING
	isCardTesting := velocity > 3.0 && ratio > 0.9 && mean < 10.0
	if !isCardTesting {
		t.Error("failed to detect card testing pattern")
	}
}

func TestFraudScenario_LocationDiversity(t *testing.T) {
	// Impossible travel: transactions from 5 different countries in 1 hour
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	countries := []string{"US", "UK", "CN", "BR", "AU"}
	events := make([]gofeat.Event, 0, len(countries))

	for i, country := range countries {
		events = append(events, gofeat.Event{
			Timestamp: now.Add(time.Duration(i) * 10 * time.Minute),
			Data: map[string]any{
				"country": country,
				"amount":  100.0,
			},
		})
	}

	// Test entropy
	entropyAgg := gofeat.Entropy("country")()
	for _, e := range events {
		entropyAgg.Add(e)
	}
	entropy := entropyAgg.Result().(float64)

	// 5 equally distributed countries = high entropy
	if entropy < 2.0 {
		t.Errorf("entropy: got %.2f, want >2.0 (high diversity)", entropy)
	}

	// Test distinct count
	distinctAgg := gofeat.DistinctCount("country")()
	for _, e := range events {
		distinctAgg.Add(e)
	}
	distinct := distinctAgg.Result().(int)

	if distinct != 5 {
		t.Errorf("distinct countries: got %d, want 5", distinct)
	}

	// VERDICT: distinct > 3 AND entropy > 2.0 = IMPOSSIBLE TRAVEL
	isImpossibleTravel := distinct > 3 && entropy > 2.0
	if !isImpossibleTravel {
		t.Error("failed to detect impossible travel pattern")
	}
}

func TestFraudScenario_NewAccountAbuse(t *testing.T) {
	// New account immediately uses promo code
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	events := []gofeat.Event{
		{Timestamp: now, Data: map[string]any{"amount": 50.0}}, // Account created + first tx
	}

	// Test account age
	ageAgg := gofeat.TimeSinceFirst()()
	for _, e := range events {
		ageAgg.Add(e)
	}
	age := ageAgg.Result().(time.Duration)

	// Brand new account
	if age != 0 {
		t.Errorf("account age: got %v, want 0 (new account)", age)
	}

	// VERDICT: age < 1 hour AND promo_used = true = PROMO ABUSE RISK
	isPromoAbuseRisk := age < time.Hour
	if !isPromoAbuseRisk {
		t.Error("failed to detect new account pattern")
	}
}
