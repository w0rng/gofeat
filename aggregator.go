package gofeat

// Aggregator computes a value from a sequence of inputs.
type Aggregator interface {
	Add(data map[string]any)
	Result() any
}

// AggregatorFactory creates new Aggregator instances.
type AggregatorFactory = func() Aggregator

// Count counts the number of events.
func Count() Aggregator { return &countAgg{} }

type countAgg struct{ n int }

func (a *countAgg) Add(map[string]any) { a.n++ }
func (a *countAgg) Result() any        { return a.n }

// Sum computes the sum of float64 values.
func Sum(field string) AggregatorFactory {
	return func() Aggregator {
		return &sumAgg{field: field}
	}
}

type sumAgg struct {
	sum   float64
	field string
}

func (a *sumAgg) Add(data map[string]any) {
	v, ok := data[a.field]
	if !ok {
		return
	}
	if f, okF := toFloat64(v); okF {
		a.sum += f
	}
}
func (a *sumAgg) Result() any { return a.sum }

// Min computes the minimum float64 value.
func Min(field string) AggregatorFactory {
	return func() Aggregator {
		return &minAgg{field: field}
	}
}

type minAgg struct {
	min   float64
	valid bool
	field string
}

func (a *minAgg) Add(data map[string]any) {
	v, ok := data[a.field]
	if !ok {
		return
	}
	f, ok := toFloat64(v)
	if !ok {
		return
	}
	if !a.valid || f < a.min {
		a.min = f
		a.valid = true
	}
}

func (a *minAgg) Result() any {
	if !a.valid {
		return 0.0
	}
	return a.min
}

// Max computes the maximum float64 value.
func Max(field string) AggregatorFactory {
	return func() Aggregator {
		return &maxAgg{field: field}
	}
}

type maxAgg struct {
	max   float64
	valid bool
	field string
}

func (a *maxAgg) Add(data map[string]any) {
	v, ok := data[a.field]
	if !ok {
		return
	}
	f, ok := toFloat64(v)
	if !ok {
		return
	}
	if !a.valid || f > a.max {
		a.max = f
		a.valid = true
	}
}

func (a *maxAgg) Result() any {
	if !a.valid {
		return 0.0
	}
	return a.max
}

// Last returns the last non-nil value.
func Last(field string) AggregatorFactory {
	return func() Aggregator {
		return &lastAgg{field: field}
	}
}

type lastAgg struct {
	last  any
	field string
}

func (a *lastAgg) Add(data map[string]any) {
	v, ok := data[a.field]
	if !ok {
		return
	}
	a.last = v
}

func (a *lastAgg) Result() any { return a.last }

// CountDistinct counts unique values.
func CountDistinct(field string) AggregatorFactory {
	return func() Aggregator {
		return &countDistinctAgg{field: field, seen: make(map[any]struct{})}
	}
}

type countDistinctAgg struct {
	seen  map[any]struct{}
	field string
}

func (a *countDistinctAgg) Add(data map[string]any) {
	v, ok := data[a.field]
	if !ok {
		return
	}

	a.seen[v] = struct{}{}
}
func (a *countDistinctAgg) Result() any { return len(a.seen) }

func toFloat64(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case int32:
		return float64(n), true
	}
	return 0, false
}
