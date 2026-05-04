package httpserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPort(t *testing.T) {
	s := New(Port(8080))
	assert.Equal(t, ":8080", s.server.Addr)
}

func TestTimeout(t *testing.T) {
	timeouts := TimeoutsConfig{
		ReadHeader: 1 * time.Second,
		Read:       2 * time.Second,
		Write:      3 * time.Second,
		Idle:       4 * time.Second,
	}

	s := New(Timeout(timeouts))

	assert.Equal(t, 1*time.Second, s.server.ReadHeaderTimeout)
	assert.Equal(t, 2*time.Second, s.server.ReadTimeout)
	assert.Equal(t, 3*time.Second, s.server.WriteTimeout)
	assert.Equal(t, 4*time.Second, s.server.IdleTimeout)
}

func TestWithRoute(t *testing.T) {
	handlerCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true

		w.WriteHeader(http.StatusOK)
	})

	s := New(WithRoute("/test", handler))

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	s.mux.ServeHTTP(w, req)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestWithMiddlewareRoute(t *testing.T) {
	middlewareCalled := false
	handlerCalled := false

	mw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			middlewareCalled = true

			next.ServeHTTP(w, r)
		})
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true

		w.WriteHeader(http.StatusAccepted)
	})

	s := New(WithMiddlewareRoute("/test", handler, mw))

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	s.mux.ServeHTTP(w, req)

	assert.True(t, middlewareCalled)
	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusAccepted, w.Code)
}

func TestWithMiddleware(t *testing.T) {
	var order []string

	mw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "mw")

			next.ServeHTTP(w, r)
		})
	}

	s := New(
		WithMiddleware(mw),
		WithRoute("/test", func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "handler")

			w.WriteHeader(http.StatusNoContent)
		}),
	)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	s.server.Handler.ServeHTTP(w, req)

	assert.Equal(t, []string{"mw", "handler"}, order)
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestRun(t *testing.T) {
	s := New()
	s.server.Addr = "127.0.0.1"

	err := s.Run()

	assert.Error(t, err)
}
