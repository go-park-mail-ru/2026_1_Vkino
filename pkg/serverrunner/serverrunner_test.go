package serverrunner

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

var errBoom = errors.New("boom")

func TestRunHTTPRequiresContext(t *testing.T) {
	t.Parallel()

	var runnerCtx context.Context

	if err := RunHTTP(runnerCtx, nil, "svc", func() error { return nil },
		func(context.Context) error { return nil }); !errors.Is(err, errRunnerContextRequired) {
		t.Fatalf("expected errRunnerContextRequired, got %v", err)
	}
}

func TestRunHTTPContextCancel(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	runCh := make(chan struct{})

	if err := RunHTTP(ctx, nil, "svc", func() error {
		<-runCh

		return http.ErrServerClosed
	}, func(context.Context) error {
		close(runCh)

		return nil
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunGRPCContextCancel(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	runCh := make(chan struct{})

	if err := RunGRPC(ctx, nil, "svc", func() error {
		<-runCh

		return nil
	}, func() {
		close(runCh)
	}, func() {}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunGRPCServeError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	err := RunGRPC(ctx, nil, "svc", func() error {
		return errBoom
	}, func() {}, func() {})
	if err == nil {
		t.Fatal("expected error")
	}
}
