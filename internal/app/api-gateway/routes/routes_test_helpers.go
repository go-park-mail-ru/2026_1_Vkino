package routes

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	authv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/auth/v1"
	moviev1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/movie/v1"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/httpserver"
	"github.com/stretchr/testify/require"
)

type testConfig struct {
	authTimeout       time.Duration
	userTimeout       time.Duration
	movieTimeout      time.Duration
	refreshCookieName string
	cookieSecure      bool
}

func (c testConfig) AuthRequestTimeout() time.Duration  { return c.authTimeout }
func (c testConfig) UserRequestTimeout() time.Duration  { return c.userTimeout }
func (c testConfig) MovieRequestTimeout() time.Duration { return c.movieTimeout }
func (c testConfig) RefreshCookieName() string          { return c.refreshCookieName }
func (c testConfig) CookieSecure() bool                 { return c.cookieSecure }

func newAuthHandler(t *testing.T, cfg Config, client authv1.AuthServiceClient) http.Handler {
	t.Helper()

	server := httpserver.New(Auth(cfg, client)...)

	return server.Handler()
}

func newMovieHandler(t *testing.T, client moviev1.MovieServiceClient) http.Handler {
	t.Helper()

	server := httpserver.New(Movie(testConfig{}, client)...)

	return server.Handler()
}

func newUserHandler(t *testing.T, cfg Config, client UserClient) http.Handler {
	t.Helper()

	server := httpserver.New(User(cfg, client)...)

	return server.Handler()
}

func doRequest(handler http.Handler, method, path string, body io.Reader) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, body)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	return rr
}

func requireJSONError(t *testing.T, rr *httptest.ResponseRecorder, statusCode int, message string) {
	t.Helper()

	require.Equal(t, statusCode, rr.Code)
	require.JSONEq(t, fmt.Sprintf(`{"Error":%q}`, message), strings.TrimSpace(rr.Body.String()))
}
