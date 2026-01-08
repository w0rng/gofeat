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

	agg.Reset()
	result = agg.Result()
	if result != 0 {
		t.Errorf("count after reset: got %v, want 0", result)
	}
}

func TestSumAggregator(t *testing.T) {
	agg := gofeat.Sum()

	result := agg.Result()
	if result != 0.0 {
		t.Errorf("initial sum: got %v, want 0.0", result)
	}

	agg.Add(100.0)
	agg.Add(50.5)
	agg.Add(25.25)

	result = agg.Result()
	expected := 175.75
	if result != expected {
		t.Errorf("sum: got %v, want %v", result, expected)
	}

	agg.Reset()
	result = agg.Result()
	if result != 0.0 {
		t.Errorf("sum after reset: got %v, want 0.0", result)
	}
}

func TestSumAggregator_TypeConversions(t *testing.T) {
	agg := gofeat.Sum()

	// Test different numeric types
	agg.Add(float64(100.0))
	agg.Add(float32(50.0))
	agg.Add(int(25))
	agg.Add(int64(10))
	agg.Add(int32(5))

	result := agg.Result().(float64)
	expected := 190.0
	if result != expected {
		t.Errorf("sum with mixed types: got %v, want %v", result, expected)
	}
}

func TestSumAggregator_InvalidTypes(t *testing.T) {
	agg := gofeat.Sum()

	agg.Add(100.0)
	agg.Add("invalid") // should be ignored
	agg.Add(nil)       // should be ignored
	agg.Add(50.0)

	result := agg.Result().(float64)
	if result != 150.0 {
		t.Errorf("sum ignoring invalid types: got %v, want 150.0", result)
	}
}

func TestMinAggregator(t *testing.T) {
	agg := gofeat.Min()

	// Initial state: no values
	result := agg.Result()
	if result != 0.0 {
		t.Errorf("initial min: got %v, want 0.0", result)
	}

	agg.Add(100.0)
	agg.Add(50.0)
	agg.Add(75.0)
	agg.Add(25.0)

	result = agg.Result()
	if result != 25.0 {
		t.Errorf("min: got %v, want 25.0", result)
	}

	agg.Reset()
	result = agg.Result()
	if result != 0.0 {
		t.Errorf("min after reset: got %v, want 0.0", result)
	}
}

func TestMinAggregator_NegativeValues(t *testing.T) {
	agg := gofeat.Min()

	agg.Add(100.0)
	agg.Add(-50.0)
	agg.Add(0.0)

	result := agg.Result().(float64)
	if result != -50.0 {
		t.Errorf("min with negative: got %v, want -50.0", result)
	}
}

func TestMinAggregator_SingleValue(t *testing.T) {
	agg := gofeat.Min()
	agg.Add(42.0)

	result := agg.Result().(float64)
	if result != 42.0 {
		t.Errorf("min with single value: got %v, want 42.0", result)
	}
}

func TestMaxAggregator(t *testing.T) {
	agg := gofeat.Max()

	result := agg.Result()
	if result != 0.0 {
		t.Errorf("initial max: got %v, want 0.0", result)
	}

	agg.Add(100.0)
	agg.Add(200.0)
	agg.Add(150.0)
	agg.Add(75.0)

	result = agg.Result()
	if result != 200.0 {
		t.Errorf("max: got %v, want 200.0", result)
	}

	agg.Reset()
	result = agg.Result()
	if result != 0.0 {
		t.Errorf("max after reset: got %v, want 0.0", result)
	}
}

func TestMaxAggregator_NegativeValues(t *testing.T) {
	agg := gofeat.Max()

	agg.Add(-100.0)
	agg.Add(-50.0)
	agg.Add(-200.0)

	result := agg.Result().(float64)
	if result != -50.0 {
		t.Errorf("max with negatives: got %v, want -50.0", result)
	}
}

func TestLastAggregator(t *testing.T) {
	agg := gofeat.Last()

	result := agg.Result()
	if result != nil {
		t.Errorf("initial last: got %v, want nil", result)
	}

	agg.Add("first")
	agg.Add("second")
	agg.Add("third")

	result = agg.Result()
	if result != "third" {
		t.Errorf("last: got %v, want third", result)
	}

	agg.Reset()
	result = agg.Result()
	if result != nil {
		t.Errorf("last after reset: got %v, want nil", result)
	}
}

func TestLastAggregator_DifferentTypes(t *testing.T) {
	agg := gofeat.Last()

	agg.Add(42)
	result := agg.Result()
	if result != 42 {
		t.Errorf("last int: got %v, want 42", result)
	}

	agg.Add("string")
	result = agg.Result()
	if result != "string" {
		t.Errorf("last string: got %v, want string", result)
	}

	agg.Add(3.14)
	result = agg.Result()
	if result != 3.14 {
		t.Errorf("last float: got %v, want 3.14", result)
	}
}

func TestCountDistinctAggregator(t *testing.T) {
	agg := gofeat.CountDistinct()

	result := agg.Result()
	if result != 0 {
		t.Errorf("initial count distinct: got %v, want 0", result)
	}

	agg.Add("US")
	agg.Add("CA")
	agg.Add("US") // duplicate
	agg.Add("MX")
	agg.Add("CA") // duplicate

	result = agg.Result()
	if result != 3 {
		t.Errorf("count distinct: got %v, want 3", result)
	}

	agg.Reset()
	result = agg.Result()
	if result != 0 {
		t.Errorf("count distinct after reset: got %v, want 0", result)
	}
}

func TestCountDistinctAggregator_Numbers(t *testing.T) {
	agg := gofeat.CountDistinct()

	agg.Add(1)
	agg.Add(2)
	agg.Add(1)
	agg.Add(3)
	agg.Add(2)

	result := agg.Result().(int)
	if result != 3 {
		t.Errorf("count distinct numbers: got %v, want 3", result)
	}
}

func TestCountDistinctAggregator_NilValues(t *testing.T) {
	agg := gofeat.CountDistinct()

	agg.Add("US")
	agg.Add(nil) // should be ignored
	agg.Add("CA")
	agg.Add(nil) // should be ignored

	result := agg.Result().(int)
	if result != 2 {
		t.Errorf("count distinct with nil: got %v, want 2", result)
	}
}

func TestCountDistinctAggregator_MixedTypes(t *testing.T) {
	agg := gofeat.CountDistinct()

	agg.Add("string")
	agg.Add(42)
	agg.Add("string") // duplicate
	agg.Add(42.0)     // different type than int(42)
	agg.Add(42)       // duplicate

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

func TestAggregators_MultipleResets(t *testing.T) {
	tests := []struct {
		name    string
		factory gofeat.AggregatorFactory
		addFunc func(gofeat.Aggregator)
		want    any
	}{
		{
			name:    "Count",
			factory: gofeat.Count,
			addFunc: func(a gofeat.Aggregator) { a.Add(nil); a.Add(nil) },
			want:    2,
		},
		{
			name:    "Sum",
			factory: gofeat.Sum,
			addFunc: func(a gofeat.Aggregator) { a.Add(100.0); a.Add(50.0) },
			want:    150.0,
		},
		{
			name:    "Min",
			factory: gofeat.Min,
			addFunc: func(a gofeat.Aggregator) { a.Add(100.0); a.Add(50.0) },
			want:    50.0,
		},
		{
			name:    "Max",
			factory: gofeat.Max,
			addFunc: func(a gofeat.Aggregator) { a.Add(100.0); a.Add(200.0) },
			want:    200.0,
		},
		{
			name:    "Last",
			factory: gofeat.Last,
			addFunc: func(a gofeat.Aggregator) { a.Add("value") },
			want:    "value",
		},
		{
			name:    "CountDistinct",
			factory: gofeat.CountDistinct,
			addFunc: func(a gofeat.Aggregator) { a.Add("a"); a.Add("b"); a.Add("a") },
			want:    2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agg := tt.factory()

			// Add, reset, add again
			tt.addFunc(agg)
			agg.Reset()
			tt.addFunc(agg)

			result := agg.Result()
			if result != tt.want {
				t.Errorf("after reset and re-add: got %v, want %v", result, tt.want)
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
			addFunc: func(a gofeat.Aggregator) { a.Add(nil) },
		},
		{
			name:    "Sum",
			factory: gofeat.Sum,
			addFunc: func(a gofeat.Aggregator) { a.Add(100.0) },
		},
		{
			name:    "Min",
			factory: gofeat.Min,
			addFunc: func(a gofeat.Aggregator) { a.Add(100.0) },
		},
		{
			name:    "Max",
			factory: gofeat.Max,
			addFunc: func(a gofeat.Aggregator) { a.Add(100.0) },
		},
		{
			name:    "Last",
			factory: gofeat.Last,
			addFunc: func(a gofeat.Aggregator) { a.Add("value") },
		},
		{
			name:    "CountDistinct",
			factory: gofeat.CountDistinct,
			addFunc: func(a gofeat.Aggregator) { a.Add("value") },
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
