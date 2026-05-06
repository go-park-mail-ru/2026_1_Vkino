package errmap

import "testing"

type testResult struct {
	Value string
}

func TestResolve(t *testing.T) {
	t.Parallel()

	rules := map[int]testResult{
		2: {Value: "two"},
		1: {Value: "one"},
	}

	m := New([]int{1, 2}, rules, func(subject int, key int) bool {
		return subject == key
	})

	res, ok := m.Resolve(2)
	if !ok || res.Value != "two" {
		t.Fatalf("expected match for 2")
	}
}

func TestResolveMissingRule(t *testing.T) {
	t.Parallel()

	m := New([]int{1}, map[int]testResult{}, func(subject int, key int) bool {
		return subject == key
	})

	_, ok := m.Resolve(1)
	if ok {
		t.Fatal("expected missing rule")
	}
}
