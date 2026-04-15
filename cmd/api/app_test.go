package main

import (
	"strings"
	"testing"
)

func TestRun_ConfigError(t *testing.T) {
	t.Parallel()

	path := "/definitely/missing/config.yaml"
	err := Run(&path)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "unable to load config") {
		t.Fatalf("unexpected error: %v", err)
	}
}
