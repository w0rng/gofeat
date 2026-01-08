package gofeat

import "fmt"

type Result struct {
	values map[string]any
}

func newResult(values map[string]any) Result {
	return Result{values: values}
}

func (r Result) Int(name string) (int, error) {
	v, ok := r.values[name]
	if !ok {
		return 0, fmt.Errorf("feature %q not found", name)
	}
	i, ok := v.(int)
	if !ok {
		return 0, fmt.Errorf("feature %q: expected int, got %T", name, v)
	}
	return i, nil
}

func (r Result) IntOr(name string, defaultValue int) int {
	v, err := r.Int(name)
	if err != nil {
		return defaultValue
	}
	return v
}

func (r Result) Float(name string) (float64, error) {
	v, ok := r.values[name]
	if !ok {
		return 0, fmt.Errorf("feature %q not found", name)
	}
	f, ok := v.(float64)
	if !ok {
		return 0, fmt.Errorf("feature %q: expected float64, got %T", name, v)
	}
	return f, nil
}

func (r Result) FloatOr(name string, defaultValue float64) float64 {
	v, err := r.Float(name)
	if err != nil {
		return defaultValue
	}
	return v
}

func (r Result) String(name string) (string, error) {
	v, ok := r.values[name]
	if !ok {
		return "", fmt.Errorf("feature %q not found", name)
	}
	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("feature %q: expected string, got %T", name, v)
	}
	return s, nil
}

func (r Result) StringOr(name, defaultValue string) string {
	v, err := r.String(name)
	if err != nil {
		return defaultValue
	}
	return v
}

func (r Result) Any(name string) (any, bool) {
	v, ok := r.values[name]
	return v, ok
}

func (r Result) All() map[string]any {
	return r.values
}
