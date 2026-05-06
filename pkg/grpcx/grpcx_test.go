package grpcx

import (
	"context"
	"testing"
	"time"
)

func TestDialEmptyAddress(t *testing.T) {
	t.Parallel()

	if _, err := Dial(context.Background(), ClientConfig{}); err == nil {
		t.Fatal("expected error for empty address")
	}
}

func TestWithTimeoutDefault(t *testing.T) {
	t.Parallel()

	ctx, cancel := WithTimeout(context.Background(), 0)
	defer cancel()

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("expected deadline")
	}

	if time.Until(deadline) <= 0 {
		t.Fatal("expected future deadline")
	}
}

func TestListen(t *testing.T) {
	t.Parallel()

	lis, err := Listen(0)
	if err != nil {
		t.Fatalf("Listen error: %v", err)
	}

	_ = lis.Close()
}
