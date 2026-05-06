package serverrunner

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

func TestRunHTTPRequiresContext(t *testing.T) {
	t.Parallel()

	if err := RunHTTP(nil, nil, "svc", func() error { return nil }, func(context.Context) error { return nil }); err == nil {
		t.Fatal("expected error for nil context")
	}
}

func TestRunHTTPContextCancel(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	runCh := make(chan struct{})

	err := RunHTTP(ctx, nil, "svc", func() error {
		<-runCh
		return http.ErrServerClosed
	}, func(context.Context) error {
		close(runCh)
		return nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunGRPCContextCancel(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	runCh := make(chan struct{})

	err := RunGRPC(ctx, nil, "svc", func() error {
		<-runCh
		return nil
	}, func() {
		close(runCh)
	}, func() {})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunGRPCServeError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	expected := errors.New("boom")

	err := RunGRPC(ctx, nil, "svc", func() error {
		return expected
	}, func() {}, func() {})

	if err == nil {
		t.Fatal("expected error")
	}
}
