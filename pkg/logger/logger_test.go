package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestNew(t *testing.T) {
	t.Parallel()

	log, err := New(Config{Level: "debug", Format: "json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if log.entry == nil {
		t.Fatal("expected non-nil log entry")
	}

	if log.entry.Logger.Level != logrus.DebugLevel {
		t.Fatalf("expected debug level, got %s", log.entry.Logger.Level)
	}

	if _, ok := log.entry.Logger.Formatter.(*logrus.JSONFormatter); !ok {
		t.Fatalf("expected json formatter, got %T", log.entry.Logger.Formatter)
	}
}

func TestNew_Defaults(t *testing.T) {
	t.Parallel()

	log, err := New(Config{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if log.entry.Logger.Level != logrus.InfoLevel {
		t.Fatalf("expected info level, got %s", log.entry.Logger.Level)
	}

	if _, ok := log.entry.Logger.Formatter.(*logrus.TextFormatter); !ok {
		t.Fatalf("expected text formatter, got %T", log.entry.Logger.Formatter)
	}
}

func TestNew_InvalidConfig(t *testing.T) {
	t.Parallel()

	if _, err := New(Config{Level: "nope"}); err == nil {
		t.Fatal("expected invalid level error")
	}

	if _, err := New(Config{Format: "xml"}); err == nil {
		t.Fatal("expected invalid format error")
	}
}

func TestNew_OutputPathWritesToFile(t *testing.T) {
	t.Parallel()

	logFile := filepath.Join(t.TempDir(), "logs", "vkinoapi.log")

	log, err := New(Config{
		Format:     "json",
		OutputPath: logFile,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	log.Info("written to file")

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}

	if len(bytes.TrimSpace(data)) == 0 {
		t.Fatal("expected non-empty log file")
	}
}

func TestContextWithLoggerAndFromContext(t *testing.T) {
	t.Parallel()

	log, err := New(Config{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx := ContextWithLogger(context.Background(), log)

	got := FromContext(ctx)
	if got != log {
		t.Fatal("expected same logger from context")
	}
}

func TestFromContext_DefaultLogger(t *testing.T) {
	t.Parallel()

	got := FromContext(context.Background())
	if got == nil || got.entry == nil {
		t.Fatal("expected default logger")
	}
}

func TestWithFieldWritesField(t *testing.T) {
	t.Parallel()

	log, err := New(Config{Format: "json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var output bytes.Buffer
	log.SetOutput(&output)

	log.WithField("request_id", "req-1").Info("hello")

	var payload map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(output.Bytes()), &payload); err != nil {
		t.Fatalf("unmarshal log payload: %v", err)
	}

	if payload["request_id"] != "req-1" {
		t.Fatalf("expected request_id req-1, got %v", payload["request_id"])
	}

	if payload["msg"] != "hello" {
		t.Fatalf("expected message hello, got %v", payload["msg"])
	}
}

func TestAddFieldPropagatesToExistingChildLogger(t *testing.T) {
	t.Parallel()

	log, err := New(Config{Format: "json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var output bytes.Buffer
	log.SetOutput(&output)

	requestLogger := log.WithField("request_id", "req-1")
	childLogger := requestLogger.WithField("scope", "handler")

	requestLogger.AddField("user_id", int64(42))
	requestLogger.AddField("email", "user@example.com")

	childLogger.Info("handled")

	var payload map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(output.Bytes()), &payload); err != nil {
		t.Fatalf("unmarshal log payload: %v", err)
	}

	if payload["request_id"] != "req-1" {
		t.Fatalf("expected request_id req-1, got %v", payload["request_id"])
	}

	if payload["scope"] != "handler" {
		t.Fatalf("expected scope handler, got %v", payload["scope"])
	}

	if payload["user_id"] != float64(42) {
		t.Fatalf("expected user_id 42, got %v", payload["user_id"])
	}

	if payload["email"] != "user@example.com" {
		t.Fatalf("expected email user@example.com, got %v", payload["email"])
	}
}

func TestFatal(t *testing.T) {
	t.Parallel()

	log, err := New(Config{Format: "json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var output bytes.Buffer
	log.SetOutput(&output)

	var exitCode int

	log.entry.Logger.ExitFunc = func(code int) {
		exitCode = code
	}

	log.Fatal("fatal message")

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}

	var payload map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(output.Bytes()), &payload); err != nil {
		t.Fatalf("unmarshal log payload: %v", err)
	}

	if payload["msg"] != "fatal message" {
		t.Fatalf("expected message fatal message, got %v", payload["msg"])
	}

	if payload["level"] != "fatal" {
		t.Fatalf("expected level fatal, got %v", payload["level"])
	}
}
