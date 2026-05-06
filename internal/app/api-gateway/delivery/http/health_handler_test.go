package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHealthHandler(t *testing.T) {
	t.Parallel()

	h := NewHealthHandler()

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()

	h.Health(rr, req)

	res := rr.Result()
	defer func() {
		_ = res.Body.Close()
	}()

	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Equal(t, "ok", rr.Body.String())
}
