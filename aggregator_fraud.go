package gofeat

import (
	"math"
	"sort"
	"time"
)

// Velocity computes events per minute within a time window.
// Use this to detect velocity abuse (e.g., too many transactions per minute).
func Velocity(window time.Duration) AggregatorFactory {
	return func() Aggregator {
		return &velocityAgg{
			window:    window,
			firstTime: time.Time{},
			lastTime:  time.Time{},
		}
	}
}

type velocityAgg struct {
	count     int
	window    time.Duration
	firstTime time.Time
	lastTime  time.Time
}

func (a *velocityAgg) Add(e Event) {
	a.count++
	if a.firstTime.IsZero() || e.Timestamp.Before(a.firstTime) {
		a.firstTime = e.Timestamp
	}
	if a.lastTime.IsZero() || e.Timestamp.After(a.lastTime) {
		a.lastTime = e.Timestamp
	}
}

func (a *velocityAgg) Result() any {
	if a.count == 0 || a.window == 0 {
		return 0.0
	}

	// If we have only one event or all events at same time
	duration := a.lastTime.Sub(a.firstTime)
	if duration == 0 {
		// All events at same time - return count per window
		return float64(a.count) / a.window.Minutes()
	}

	// Calculate events per minute based on actual duration
	minutes := duration.Minutes()
	if minutes == 0 {
		return float64(a.count)
	}
	return float64(a.count) / minutes
}

// Entropy computes Shannon entropy for a field's values.
// High entropy = many different values (suspicious for device_id, IP, etc).
func Entropy(field string) AggregatorFactory {
	return func() Aggregator {
		return &entropyAgg{
			field:  field,
			counts: make(map[any]int),
		}
	}
}

type entropyAgg struct {
	field  string
	counts map[any]int
	total  int
}

func (a *entropyAgg) Add(e Event) {
	v, ok := e.Data[a.field]
	if !ok {
		return
	}
	a.counts[v]++
	a.total++
}

func (a *entropyAgg) Result() any {
	if a.total == 0 {
		return 0.0
	}

	var entropy float64
	for _, count := range a.counts {
		if count > 0 {
			p := float64(count) / float64(a.total)
			entropy -= p * math.Log2(p)
		}
	}
	return entropy
}

// UniqueRatio computes the ratio of unique values to total events.
// Returns float64 from 0.0 to 1.0.
// 1.0 = all values unique (suspicious for cards, emails, etc)
// 0.0 = all values same.
func UniqueRatio(field string) AggregatorFactory {
	return func() Aggregator {
		return &uniqueRatioAgg{
			field: field,
			seen:  make(map[any]struct{}),
		}
	}
}

type uniqueRatioAgg struct {
	field string
	seen  map[any]struct{}
	total int
}

func (a *uniqueRatioAgg) Add(e Event) {
	v, ok := e.Data[a.field]
	if !ok {
		return
	}
	a.seen[v] = struct{}{}
	a.total++
}

func (a *uniqueRatioAgg) Result() any {
	if a.total == 0 {
		return 0.0
	}
	return float64(len(a.seen)) / float64(a.total)
}

// TimeSinceFirst returns duration since the first event.
// Useful for account age, time since first transaction, etc.
func TimeSinceFirst() AggregatorFactory {
	return func() Aggregator {
		return &timeSinceFirstAgg{
			firstTime: time.Time{},
			lastTime:  time.Time{},
		}
	}
}

type timeSinceFirstAgg struct {
	firstTime time.Time
	lastTime  time.Time
}

func (a *timeSinceFirstAgg) Add(e Event) {
	if a.firstTime.IsZero() || e.Timestamp.Before(a.firstTime) {
		a.firstTime = e.Timestamp
	}
	if a.lastTime.IsZero() || e.Timestamp.After(a.lastTime) {
		a.lastTime = e.Timestamp
	}
}

func (a *timeSinceFirstAgg) Result() any {
	if a.firstTime.IsZero() {
		return time.Duration(0)
	}
	return a.lastTime.Sub(a.firstTime)
}

// Percentile computes the percentile value for a numeric field.
// p should be between 0.0 and 1.0 (e.g., 0.95 for p95, 0.99 for p99).
// Use this for outlier detection.
func Percentile(field string, p float64) AggregatorFactory {
	return func() Aggregator {
		return &percentileAgg{
			field: field,
			p:     p,
		}
	}
}

type percentileAgg struct {
	field  string
	p      float64
	values []float64
}

func (a *percentileAgg) Add(e Event) {
	v, ok := e.Data[a.field]
	if !ok {
		return
	}
	if f, okF := toFloat64(v); okF {
		a.values = append(a.values, f)
	}
}

func (a *percentileAgg) Result() any {
	if len(a.values) == 0 {
		return 0.0
	}

	sorted := make([]float64, len(a.values))
	copy(sorted, a.values)
	sort.Float64s(sorted)

	index := int(float64(len(sorted)-1) * a.p)
	if index < 0 {
		index = 0
	}
	if index >= len(sorted) {
		index = len(sorted) - 1
	}

	return sorted[index]
}

// StandardDeviation computes the standard deviation for a numeric field.
// Use this for anomaly detection (e.g., Z-score calculation).
func StandardDeviation(field string) AggregatorFactory {
	return func() Aggregator {
		return &stdDevAgg{field: field}
	}
}

type stdDevAgg struct {
	field  string
	values []float64
}

func (a *stdDevAgg) Add(e Event) {
	v, ok := e.Data[a.field]
	if !ok {
		return
	}
	if f, okF := toFloat64(v); okF {
		a.values = append(a.values, f)
	}
}

func (a *stdDevAgg) Result() any {
	if len(a.values) == 0 {
		return 0.0
	}

	// Calculate mean
	var sum float64
	for _, v := range a.values {
		sum += v
	}
	mean := sum / float64(len(a.values))

	// Calculate variance
	var variance float64
	for _, v := range a.values {
		diff := v - mean
		variance += diff * diff
	}
	variance /= float64(len(a.values))

	// Standard deviation is square root of variance
	return math.Sqrt(variance)
}

// Mean computes the average value for a numeric field.
func Mean(field string) AggregatorFactory {
	return func() Aggregator {
		return &meanAgg{field: field}
	}
}

type meanAgg struct {
	field string
	sum   float64
	count int
}

func (a *meanAgg) Add(e Event) {
	v, ok := e.Data[a.field]
	if !ok {
		return
	}
	if f, okF := toFloat64(v); okF {
		a.sum += f
		a.count++
	}
}

func (a *meanAgg) Result() any {
	if a.count == 0 {
		return 0.0
	}
	return a.sum / float64(a.count)
}
