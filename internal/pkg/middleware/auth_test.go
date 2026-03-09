package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/usecase"
)

type mockUsecase struct {
	signInFn               func(email, password string) (domain.TokenPair, error)
	signUpFn               func(email, password string) (domain.TokenPair, error)
	refreshFn              func(email string) (domain.TokenPair, error)
	validateRefreshTokenFn func(token string) (string, error)
	validateAccessTokenFn  func(token string) (string, error)
	getConfigFn            func() usecase.Config
}

func (m *mockUsecase) SignIn(email, password string) (domain.TokenPair, error) {
	if m.signInFn == nil {
		panic("unexpected call: SignIn")
	}
	return m.signInFn(email, password)
}

func (m *mockUsecase) SignUp(email, password string) (domain.TokenPair, error) {
	if m.signUpFn == nil {
		panic("unexpected call: SignUp")
	}
	return m.signUpFn(email, password)
}

func (m *mockUsecase) Refresh(email string) (domain.TokenPair, error) {
	if m.refreshFn == nil {
		panic("unexpected call: Refresh")
	}
	return m.refreshFn(email)
}

func (m *mockUsecase) ValidateRefreshToken(token string) (string, error) {
	if m.validateRefreshTokenFn == nil {
		panic("unexpected call: ValidateRefreshToken")
	}
	return m.validateRefreshTokenFn(token)
}

func (m *mockUsecase) ValidateAccessToken(token string) (string, error) {
	if m.validateAccessTokenFn == nil {
		panic("unexpected call: ValidateAccessToken")
	}
	return m.validateAccessTokenFn(token)
}

func (m *mockUsecase) GetConfig() usecase.Config {
	if m.getConfigFn == nil {
		return usecase.Config{
			JWTSecret:         "test-secret",
			AccessTokenTTL:    time.Hour,
			RefreshTokenTTL:   24 * time.Hour,
			RefreshCookieName: "refresh_token",
			CookieSecure:      true,
		}
	}
	return m.getConfigFn()
}

func assertJSONContainsStringValue(t *testing.T, rr *httptest.ResponseRecorder, want string) {
	t.Helper()

	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v; body=%s", err, rr.Body.String())
	}

	for _, value := range body {
		if s, ok := value.(string); ok && s == want {
			return
		}
	}

	t.Fatalf("expected body to contain %q, got %v", want, body)
}

func TestAuthMiddleware_Middleware(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		authHeader         string
		validateEmail      string
		validateErr        error
		wantStatus         int
		wantBodyValue      string
		wantNextCalled     bool
		wantValidateCalled bool
		wantToken          string
		wantContextEmail   string
	}{
		{
			name:               "missing authorization header",
			wantStatus:         http.StatusUnauthorized,
			wantBodyValue:      "unauthorized",
			wantNextCalled:     false,
			wantValidateCalled: false,
		},
		{
			name:               "invalid authorization prefix",
			authHeader:         "Basic token",
			wantStatus:         http.StatusUnauthorized,
			wantBodyValue:      "unauthorized",
			wantNextCalled:     false,
			wantValidateCalled: false,
		},
		{
			name:               "empty bearer token",
			authHeader:         "Bearer   ",
			wantStatus:         http.StatusUnauthorized,
			wantBodyValue:      "unauthorized",
			wantNextCalled:     false,
			wantValidateCalled: false,
		},
		{
			name:               "validate access token error",
			authHeader:         "Bearer bad-token",
			validateErr:        errors.New("invalid token"),
			wantStatus:         http.StatusUnauthorized,
			wantBodyValue:      "unauthorized",
			wantNextCalled:     false,
			wantValidateCalled: true,
			wantToken:          "bad-token",
		},
		{
			name:               "success",
			authHeader:         "Bearer good-token",
			validateEmail:      "user@example.com",
			wantStatus:         http.StatusOK,
			wantNextCalled:     true,
			wantValidateCalled: true,
			wantToken:          "good-token",
			wantContextEmail:   "user@example.com",
		},
		{
			name:               "success with trimmed header",
			authHeader:         "   Bearer good-token   ",
			validateEmail:      "user@example.com",
			wantStatus:         http.StatusOK,
			wantNextCalled:     true,
			wantValidateCalled: true,
			wantToken:          "good-token",
			wantContextEmail:   "user@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var validateCalled bool
			var gotToken string
			var nextCalled bool
			var nextEmail string

			mock := &mockUsecase{
				validateAccessTokenFn: func(token string) (string, error) {
					validateCalled = true
					gotToken = token
					return tt.validateEmail, tt.validateErr
				},
			}

			m := &AuthMiddleware{usecase: mock}

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true

				email, ok := UserEmailFromContext(r.Context())
				if !ok {
					t.Fatal("expected email in context")
				}
				nextEmail = email

				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status":"ok"}`))
			})

			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			rr := httptest.NewRecorder()
			m.Middleware(next).ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rr.Code)
			}

			if validateCalled != tt.wantValidateCalled {
				t.Fatalf("expected ValidateAccessToken called=%v, got %v", tt.wantValidateCalled, validateCalled)
			}

			if tt.wantValidateCalled && gotToken != tt.wantToken {
				t.Fatalf("expected token %q, got %q", tt.wantToken, gotToken)
			}

			if nextCalled != tt.wantNextCalled {
				t.Fatalf("expected next called=%v, got %v", tt.wantNextCalled, nextCalled)
			}

			if tt.wantBodyValue != "" {
				assertJSONContainsStringValue(t, rr, tt.wantBodyValue)
				return
			}

			if nextEmail != tt.wantContextEmail {
				t.Fatalf("expected context email %q, got %q", tt.wantContextEmail, nextEmail)
			}

			if !strings.Contains(rr.Body.String(), `"status":"ok"`) {
				t.Fatalf("expected next handler response body, got %s", rr.Body.String())
			}
		})
	}
}

func TestUserEmailFromContext(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		ctx       context.Context
		wantEmail string
		wantOK    bool
	}{
		{
			name:      "email exists",
			ctx:       context.WithValue(context.Background(), UserEmailKey, "user@example.com"),
			wantEmail: "user@example.com",
			wantOK:    true,
		},
		{
			name:   "email missing",
			ctx:    context.Background(),
			wantOK: false,
		},
		{
			name:   "wrong type in context",
			ctx:    context.WithValue(context.Background(), UserEmailKey, 123),
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			email, ok := UserEmailFromContext(tt.ctx)

			if ok != tt.wantOK {
				t.Fatalf("expected ok=%v, got %v", tt.wantOK, ok)
			}
			if email != tt.wantEmail {
				t.Fatalf("expected email %q, got %q", tt.wantEmail, email)
			}
		})
	}
}
