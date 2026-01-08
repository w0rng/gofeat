package gofeat

// Aggregator computes a value from a sequence of inputs.
type Aggregator interface {
	Add(value any)
	Result() any
	Reset()
}

// AggregatorFactory creates new Aggregator instances.
type AggregatorFactory = func() Aggregator

// Count counts the number of events.
func Count() Aggregator { return &countAgg{} }

type countAgg struct{ n int }

func (a *countAgg) Add(any)     { a.n++ }
func (a *countAgg) Result() any { return a.n }
func (a *countAgg) Reset()      { a.n = 0 }

// Sum computes the sum of float64 values.
func Sum() Aggregator { return &sumAgg{} }

type sumAgg struct{ sum float64 }

func (a *sumAgg) Add(v any) {
	if f, ok := toFloat64(v); ok {
		a.sum += f
	}
}
func (a *sumAgg) Result() any { return a.sum }
func (a *sumAgg) Reset()      { a.sum = 0 }

// Min computes the minimum float64 value.
func Min() Aggregator { return &minAgg{} }

type minAgg struct {
	min   float64
	valid bool
}

func (a *minAgg) Add(v any) {
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
func (a *minAgg) Reset() { a.min, a.valid = 0, false }

// Max computes the maximum float64 value.
func Max() Aggregator { return &maxAgg{} }

type maxAgg struct {
	max   float64
	valid bool
}

func (a *maxAgg) Add(v any) {
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
func (a *maxAgg) Reset() { a.max, a.valid = 0, false }

// Last returns the last non-nil value.
func Last() Aggregator { return &lastAgg{} }

type lastAgg struct{ last any }

func (a *lastAgg) Add(v any)   { a.last = v }
func (a *lastAgg) Result() any { return a.last }
func (a *lastAgg) Reset()      { a.last = nil }

// CountDistinct counts unique values.
func CountDistinct() Aggregator {
	return &countDistinctAgg{seen: make(map[any]struct{})}
}

type countDistinctAgg struct{ seen map[any]struct{} }

func (a *countDistinctAgg) Add(v any) {
	if v != nil {
		a.seen[v] = struct{}{}
	}
}
func (a *countDistinctAgg) Result() any { return len(a.seen) }
func (a *countDistinctAgg) Reset()      { clear(a.seen) }

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
