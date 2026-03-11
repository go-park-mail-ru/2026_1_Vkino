package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	usecasemocks "github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/usecase/mocks"
	"go.uber.org/mock/gomock"
)

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
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mu := usecasemocks.NewMockUsecase(ctrl)
			if tt.wantValidateCalled {
				mu.EXPECT().
					ValidateAccessToken(tt.wantToken).
					Return(tt.validateEmail, tt.validateErr)
			}

			var nextCalled bool
			var nextEmail string

			m := &AuthMiddleware{usecase: mu}

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true

				email, err := UserEmailFromContext(r.Context())
				if err != nil {
					t.Fatalf("expected email in context, got error: %v", err)
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
		wantErr   bool
	}{
		{
			name:      "email exists",
			ctx:       context.WithValue(context.Background(), UserEmailKey, "user@example.com"),
			wantEmail: "user@example.com",
			wantErr:   false,
		},
		{
			name:    "email missing",
			ctx:     context.Background(),
			wantErr: true,
		},
		{
			name:    "wrong type in context",
			ctx:     context.WithValue(context.Background(), UserEmailKey, 123),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			email, err := UserEmailFromContext(tt.ctx)

			if tt.wantErr {
				if !errors.Is(err, ErrMidlware) {
					t.Fatalf("expected ErrMidlware, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if email != tt.wantEmail {
				t.Fatalf("expected email %q, got %q", tt.wantEmail, email)
			}
		})
	}
}
