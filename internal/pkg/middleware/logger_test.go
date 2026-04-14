package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/pkg/logger"
)

func TestLoggerMiddleware_UsesIncomingRequestID(t *testing.T) {
	t.Parallel()

	baseLogger, err := logger.New(logger.Config{Format: "json"})
	if err != nil {
		t.Fatalf("new logger: %v", err)
	}

	var output bytes.Buffer
	baseLogger.SetOutput(&output)

	var ctxLogger *logger.Logger

	middleware := LoggerMiddleware(baseLogger)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxLogger = logger.FromContext(r.Context())
		w.WriteHeader(http.StatusCreated)
	}))

	req := httptest.NewRequest(http.MethodPost, "/user/sign-in", nil)
	req.Header.Set(requestIDHeader, "req-42")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if ctxLogger == nil {
		t.Fatal("expected logger in request context")
	}

	if got := rr.Header().Get(requestIDHeader); got != "req-42" {
		t.Fatalf("expected request id header req-42, got %q", got)
	}

	payload := decodeLogLine(t, output.String())

	if payload["request_id"] != "req-42" {
		t.Fatalf("expected request_id req-42, got %v", payload["request_id"])
	}

	if payload["method"] != http.MethodPost {
		t.Fatalf("expected method %s, got %v", http.MethodPost, payload["method"])
	}

	if payload["path"] != "/user/sign-in" {
		t.Fatalf("expected path /user/sign-in, got %v", payload["path"])
	}

	if payload["status"] != float64(http.StatusCreated) {
		t.Fatalf("expected status %d, got %v", http.StatusCreated, payload["status"])
	}

	if payload["msg"] != "request handled" {
		t.Fatalf("expected request handled message, got %v", payload["msg"])
	}

	if payload["duration"] == "" {
		t.Fatal("expected duration field")
	}
}

func TestLoggerMiddleware_GeneratesRequestIDWhenMissing(t *testing.T) {
	t.Parallel()

	baseLogger, err := logger.New(logger.Config{Format: "json"})
	if err != nil {
		t.Fatalf("new logger: %v", err)
	}

	var output bytes.Buffer
	baseLogger.SetOutput(&output)

	handler := LoggerMiddleware(baseLogger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := w.(http.Flusher); !ok {
			t.Fatal("expected wrapped response writer to preserve http.Flusher")
		}

		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/movie/selection/all", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	requestID := rr.Header().Get(requestIDHeader)
	if requestID == "" {
		t.Fatal("expected generated request id header")
	}

	payload := decodeLogLine(t, output.String())
	if payload["request_id"] != requestID {
		t.Fatalf("expected logged request_id %q, got %v", requestID, payload["request_id"])
	}
}

func TestRecoveryMiddleware_UsesRequestLogger(t *testing.T) {
	t.Parallel()

	baseLogger, err := logger.New(logger.Config{Format: "json"})
	if err != nil {
		t.Fatalf("new logger: %v", err)
	}

	var output bytes.Buffer
	baseLogger.SetOutput(&output)

	handler := LoggerMiddleware(baseLogger)(RecoveryMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	})))

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}

	requestID := rr.Header().Get(requestIDHeader)
	if requestID == "" {
		t.Fatal("expected request id header")
	}

	var panicLogged bool
	for _, line := range strings.Split(strings.TrimSpace(output.String()), "\n") {
		payload := decodeLogLine(t, line)
		if payload["msg"] != "panic recovered" {
			continue
		}

		panicLogged = true
		if payload["request_id"] != requestID {
			t.Fatalf("expected panic log request_id %q, got %v", requestID, payload["request_id"])
		}
	}

	if !panicLogged {
		t.Fatal("expected panic log entry")
	}
}

func decodeLogLine(t *testing.T, line string) map[string]any {
	t.Helper()

	var payload map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(line)), &payload); err != nil {
		t.Fatalf("decode log line: %v", err)
	}

	return payload
}
