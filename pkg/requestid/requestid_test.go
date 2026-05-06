package requestid

import (
	"context"
	"testing"
)

func TestGenerate(t *testing.T) {
	t.Parallel()

	id := Generate()
	if id == "" {
		t.Fatal("expected non-empty id")
	}

	if len(id) != 32 {
		t.Fatalf("expected 32-char hex id, got %d", len(id))
	}
}

func TestNormalizeEmpty(t *testing.T) {
	t.Parallel()

	id := Normalize("   ")
	if id == "" {
		t.Fatal("expected generated id for empty input")
	}
}

func TestContextWithID(t *testing.T) {
	t.Parallel()

	ctx := ContextWithID(nil, "custom")
	got, ok := FromContext(ctx)
	if !ok {
		t.Fatal("expected id in context")
	}
	if got != "custom" {
		t.Fatalf("expected custom id, got %q", got)
	}
}

func TestFromContextMissing(t *testing.T) {
	t.Parallel()

	_, ok := FromContext(context.Background())
	if ok {
		t.Fatal("expected missing id")
	}
}
