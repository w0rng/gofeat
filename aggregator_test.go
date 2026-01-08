package gofeat_test

import (
	"testing"

	"github.com/w0rng/gofeat"
)

func TestCountAggregator(t *testing.T) {
	agg := gofeat.Count()

	result := agg.Result()
	if result != 0 {
		t.Errorf("initial count: got %v, want 0", result)
	}

	agg.Add(nil)
	agg.Add(nil)
	agg.Add(nil)

	result = agg.Result()
	if result != 3 {
		t.Errorf("count after 3 adds: got %v, want 3", result)
	}
}

func TestSumAggregator(t *testing.T) {
	agg := gofeat.Sum("some")()

	result := agg.Result()
	if result != 0.0 {
		t.Errorf("initial sum: got %v, want 0.0", result)
	}

	agg.Add(map[string]any{"some": 100.0})
	agg.Add(map[string]any{"some": 50.5})
	agg.Add(map[string]any{"some": 25.25})

	result = agg.Result()
	expected := 175.75
	if result != expected {
		t.Errorf("sum: got %v, want %v", result, expected)
	}
}

func TestSumAggregator_TypeConversions(t *testing.T) {
	agg := gofeat.Sum("some")()

	// Test different numeric types
	agg.Add(map[string]any{"some": float64(100.0)})
	agg.Add(map[string]any{"some": float32(50.0)})
	agg.Add(map[string]any{"some": int(25)})
	agg.Add(map[string]any{"some": int64(10)})
	agg.Add(map[string]any{"some": int32(5)})

	result := agg.Result().(float64)
	expected := 190.0
	if result != expected {
		t.Errorf("sum with mixed types: got %v, want %v", result, expected)
	}
}

func TestSumAggregator_InvalidTypes(t *testing.T) {
	agg := gofeat.Sum("value")()

	agg.Add(map[string]any{"value": 100.0})
	agg.Add(map[string]any{"value": "invalid"}) // should be ignored
	agg.Add(map[string]any{"other": 999.0})     // should be ignored (wrong field)
	agg.Add(map[string]any{"value": 50.0})

	result := agg.Result().(float64)
	if result != 150.0 {
		t.Errorf("sum ignoring invalid types: got %v, want 150.0", result)
	}
}

func TestMinAggregator(t *testing.T) {
	agg := gofeat.Min("value")()

	// Initial state: no values
	result := agg.Result()
	if result != 0.0 {
		t.Errorf("initial min: got %v, want 0.0", result)
	}

	agg.Add(map[string]any{"value": 100.0})
	agg.Add(map[string]any{"value": 50.0})
	agg.Add(map[string]any{"value": 75.0})
	agg.Add(map[string]any{"value": 25.0})

	result = agg.Result()
	if result != 25.0 {
		t.Errorf("min: got %v, want 25.0", result)
	}
}

func TestMinAggregator_NegativeValues(t *testing.T) {
	agg := gofeat.Min("value")()

	agg.Add(map[string]any{"value": 100.0})
	agg.Add(map[string]any{"value": -50.0})
	agg.Add(map[string]any{"value": 0.0})

	result := agg.Result().(float64)
	if result != -50.0 {
		t.Errorf("min with negative: got %v, want -50.0", result)
	}
}

func TestMinAggregator_SingleValue(t *testing.T) {
	agg := gofeat.Min("value")()
	agg.Add(map[string]any{"value": 42.0})

	result := agg.Result().(float64)
	if result != 42.0 {
		t.Errorf("min with single value: got %v, want 42.0", result)
	}
}

func TestMaxAggregator(t *testing.T) {
	agg := gofeat.Max("value")()

	result := agg.Result()
	if result != 0.0 {
		t.Errorf("initial max: got %v, want 0.0", result)
	}

	agg.Add(map[string]any{"value": 100.0})
	agg.Add(map[string]any{"value": 200.0})
	agg.Add(map[string]any{"value": 150.0})
	agg.Add(map[string]any{"value": 75.0})

	result = agg.Result()
	if result != 200.0 {
		t.Errorf("max: got %v, want 200.0", result)
	}
}

func TestMaxAggregator_NegativeValues(t *testing.T) {
	agg := gofeat.Max("value")()

	agg.Add(map[string]any{"value": -100.0})
	agg.Add(map[string]any{"value": -50.0})
	agg.Add(map[string]any{"value": -200.0})

	result := agg.Result().(float64)
	if result != -50.0 {
		t.Errorf("max with negatives: got %v, want -50.0", result)
	}
}

func TestLastAggregator(t *testing.T) {
	agg := gofeat.Last("value")()

	result := agg.Result()
	if result != nil {
		t.Errorf("initial last: got %v, want nil", result)
	}

	agg.Add(map[string]any{"value": "first"})
	agg.Add(map[string]any{"value": "second"})
	agg.Add(map[string]any{"value": "third"})

	result = agg.Result()
	if result != "third" {
		t.Errorf("last: got %v, want third", result)
	}
}

func TestLastAggregator_DifferentTypes(t *testing.T) {
	agg := gofeat.Last("value")()

	agg.Add(map[string]any{"value": 42})
	result := agg.Result()
	if result != 42 {
		t.Errorf("last int: got %v, want 42", result)
	}

	agg.Add(map[string]any{"value": "string"})
	result = agg.Result()
	if result != "string" {
		t.Errorf("last string: got %v, want string", result)
	}

	agg.Add(map[string]any{"value": 3.14})
	result = agg.Result()
	if result != 3.14 {
		t.Errorf("last float: got %v, want 3.14", result)
	}
}

func TestCountDistinctAggregator(t *testing.T) {
	agg := gofeat.CountDistinct("country")()

	result := agg.Result()
	if result != 0 {
		t.Errorf("initial count distinct: got %v, want 0", result)
	}

	agg.Add(map[string]any{"country": "US"})
	agg.Add(map[string]any{"country": "CA"})
	agg.Add(map[string]any{"country": "US"}) // duplicate
	agg.Add(map[string]any{"country": "MX"})
	agg.Add(map[string]any{"country": "CA"}) // duplicate

	result = agg.Result()
	if result != 3 {
		t.Errorf("count distinct: got %v, want 3", result)
	}
}

func TestCountDistinctAggregator_Numbers(t *testing.T) {
	agg := gofeat.CountDistinct("value")()

	agg.Add(map[string]any{"value": 1})
	agg.Add(map[string]any{"value": 2})
	agg.Add(map[string]any{"value": 1})
	agg.Add(map[string]any{"value": 3})
	agg.Add(map[string]any{"value": 2})

	result := agg.Result().(int)
	if result != 3 {
		t.Errorf("count distinct numbers: got %v, want 3", result)
	}
}

func TestCountDistinctAggregator_MissingField(t *testing.T) {
	agg := gofeat.CountDistinct("country")()

	agg.Add(map[string]any{"country": "US"})
	agg.Add(map[string]any{"other": "value"}) // should be ignored (missing field)
	agg.Add(map[string]any{"country": "CA"})
	agg.Add(map[string]any{}) // should be ignored (missing field)

	result := agg.Result().(int)
	if result != 2 {
		t.Errorf("count distinct with missing field: got %v, want 2", result)
	}
}

func TestCountDistinctAggregator_MixedTypes(t *testing.T) {
	agg := gofeat.CountDistinct("value")()

	agg.Add(map[string]any{"value": "string"})
	agg.Add(map[string]any{"value": 42})
	agg.Add(map[string]any{"value": "string"}) // duplicate
	agg.Add(map[string]any{"value": 42.0})     // different type than int(42)
	agg.Add(map[string]any{"value": 42})       // duplicate

	result := agg.Result().(int)
	// "string", 42, 42.0 = 3 distinct values
	if result != 3 {
		t.Errorf("count distinct mixed types: got %v, want 3", result)
	}
}

func TestAggregators_Factory(t *testing.T) {
	// Test that factories create independent instances
	agg1 := gofeat.Count()
	agg2 := gofeat.Count()

	agg1.Add(nil)
	agg1.Add(nil)

	result1 := agg1.Result()
	result2 := agg2.Result()

	if result1 != 2 {
		t.Errorf("agg1 count: got %v, want 2", result1)
	}
	if result2 != 0 {
		t.Errorf("agg2 count: got %v, want 0 (should be independent)", result2)
	}
}

func TestAggregators_Add(t *testing.T) {
	tests := []struct {
		name    string
		factory gofeat.AggregatorFactory
		addFunc func(gofeat.Aggregator)
		want    any
	}{
		{
			name:    "Count",
			factory: gofeat.Count,
			addFunc: func(a gofeat.Aggregator) { a.Add(map[string]any{}); a.Add(map[string]any{}) },
			want:    2,
		},
		{
			name:    "Sum",
			factory: gofeat.Sum("value"),
			addFunc: func(a gofeat.Aggregator) { a.Add(map[string]any{"value": 100.0}); a.Add(map[string]any{"value": 50.0}) },
			want:    150.0,
		},
		{
			name:    "Min",
			factory: gofeat.Min("value"),
			addFunc: func(a gofeat.Aggregator) { a.Add(map[string]any{"value": 100.0}); a.Add(map[string]any{"value": 50.0}) },
			want:    50.0,
		},
		{
			name:    "Max",
			factory: gofeat.Max("value"),
			addFunc: func(a gofeat.Aggregator) {
				a.Add(map[string]any{"value": 100.0})
				a.Add(map[string]any{"value": 200.0})
			},
			want: 200.0,
		},
		{
			name:    "Last",
			factory: gofeat.Last("value"),
			addFunc: func(a gofeat.Aggregator) { a.Add(map[string]any{"value": "value"}) },
			want:    "value",
		},
		{
			name:    "CountDistinct",
			factory: gofeat.CountDistinct("value"),
			addFunc: func(a gofeat.Aggregator) {
				a.Add(map[string]any{"value": "a"})
				a.Add(map[string]any{"value": "b"})
				a.Add(map[string]any{"value": "a"})
			},
			want: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agg := tt.factory()

			tt.addFunc(agg)

			result := agg.Result()
			if result != tt.want {
				t.Errorf("got %v, want %v", result, tt.want)
			}
		})
	}
}

func BenchmarkAggregators(b *testing.B) {
	benchmarks := []struct {
		name    string
		factory gofeat.AggregatorFactory
		addFunc func(gofeat.Aggregator)
	}{
		{
			name:    "Count",
			factory: gofeat.Count,
			addFunc: func(a gofeat.Aggregator) { a.Add(map[string]any{}) },
		},
		{
			name:    "Sum",
			factory: gofeat.Sum("value"),
			addFunc: func(a gofeat.Aggregator) { a.Add(map[string]any{"value": 100.0}) },
		},
		{
			name:    "Min",
			factory: gofeat.Min("value"),
			addFunc: func(a gofeat.Aggregator) { a.Add(map[string]any{"value": 100.0}) },
		},
		{
			name:    "Max",
			factory: gofeat.Max("value"),
			addFunc: func(a gofeat.Aggregator) { a.Add(map[string]any{"value": 100.0}) },
		},
		{
			name:    "Last",
			factory: gofeat.Last("value"),
			addFunc: func(a gofeat.Aggregator) { a.Add(map[string]any{"value": "value"}) },
		},
		{
			name:    "CountDistinct",
			factory: gofeat.CountDistinct("value"),
			addFunc: func(a gofeat.Aggregator) { a.Add(map[string]any{"value": "value"}) },
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			agg := bm.factory()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				bm.addFunc(agg)
			}
		})
	}
}
