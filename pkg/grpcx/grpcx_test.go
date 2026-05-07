package grpcx

import (
	"context"
	"errors"
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
		if errors.Is(err, ErrListenPermissionDenied) {
			t.Skipf("listen is not allowed in this environment: %v", err)
		}

		t.Fatalf("Listen error: %v", err)
	}

	_ = lis.Close()
}
