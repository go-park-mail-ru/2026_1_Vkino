package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user/usecase"
	usecasemocks "github.com/go-park-mail-ru/2026_1_VKino/internal/app/user/usecase/mocks"
	"go.uber.org/mock/gomock"
)

func TestNewAuthMiddleware(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	u := usecase.NewAuthUsecase(
		usecasemocks.NewMockUserRepo(ctrl),
		usecasemocks.NewMockSessionRepo(ctrl),
		usecase.Config{},
	)

	got := NewAuthMiddleware(u)
	if got == nil {
		t.Fatal("expected non-nil middleware")
	}

	if got.usecase != u {
		t.Fatal("expected usecase to be set")
	}
}

func TestChain(t *testing.T) {
	t.Parallel()

	var order []string

	first := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "first-before")
			next.ServeHTTP(w, r)
			order = append(order, "first-after")
		})
	}

	second := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "second-before")
			next.ServeHTTP(w, r)
			order = append(order, "second-after")
		})
	}

	handler := Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
		w.WriteHeader(http.StatusNoContent)
	}), first, second)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rr.Code)
	}

	want := strings.Join([]string{"first-before", "second-before", "handler", "second-after", "first-after"}, ",")
	if got := strings.Join(order, ","); got != want {
		t.Fatalf("expected order %q, got %q", want, got)
	}
}

func TestCorsMiddleware(t *testing.T) {
	t.Parallel()

	cfg := CORSConfig{
		AllowedOrigins:   []string{"https://vkino.example"},
		AllowedMethods:   []string{http.MethodGet, http.MethodPost},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           600,
	}

	t.Run("preflight allowed origin", func(t *testing.T) {
		var nextCalled bool

		handler := CorsMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
		}))

		req := httptest.NewRequest(http.MethodOptions, "/", nil)
		req.Header.Set("Origin", "https://vkino.example")

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
		}

		if nextCalled {
			t.Fatal("expected next handler not to be called for preflight request")
		}

		if rr.Header().Get("Access-Control-Allow-Origin") != "https://vkino.example" {
			t.Fatalf("unexpected allow origin header %q", rr.Header().Get("Access-Control-Allow-Origin"))
		}
	})

	t.Run("regular request disallowed origin", func(t *testing.T) {
		var nextCalled bool

		handler := CorsMiddleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusAccepted)
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Origin", "https://other.example")

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if !nextCalled {
			t.Fatal("expected next handler to be called")
		}

		if rr.Code != http.StatusAccepted {
			t.Fatalf("expected status %d, got %d", http.StatusAccepted, rr.Code)
		}

		if rr.Header().Get("Access-Control-Allow-Origin") != "" {
			t.Fatalf("expected empty allow origin header, got %q", rr.Header().Get("Access-Control-Allow-Origin"))
		}
	})
}

func TestIsOriginAllowed(t *testing.T) {
	t.Parallel()

	if !isOriginAllowed("https://vkino.example", []string{"https://vkino.example"}) {
		t.Fatal("expected origin to be allowed")
	}

	if isOriginAllowed("https://other.example", []string{"https://vkino.example"}) {
		t.Fatal("expected origin to be rejected")
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	t.Parallel()

	t.Run("panic", func(t *testing.T) {
		handler := RecoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("boom")
		}))

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
		}

		if !strings.Contains(rr.Body.String(), "500 - Internal Server Error") {
			t.Fatalf("expected panic body, got %q", rr.Body.String())
		}
	})

	t.Run("no panic", func(t *testing.T) {
		handler := RecoveryMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
		}))

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

		if rr.Code != http.StatusCreated {
			t.Fatalf("expected status %d, got %d", http.StatusCreated, rr.Code)
		}
	})
}
