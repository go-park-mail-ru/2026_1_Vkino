package httpserver

import (
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

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	s.mux.ServeHTTP(w, req)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusOK, w.Code)
}
