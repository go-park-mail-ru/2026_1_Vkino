package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user/usecase"
	usecasemocks "github.com/go-park-mail-ru/2026_1_VKino/internal/app/user/usecase/mocks"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/logger"
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

	testUserID := int64(1)

	tests := []struct {
		name               string
		authHeader         string
		validateAuth       usecase.AuthContext
		validateErr        error
		wantStatus         int
		wantBodyValue      string
		wantNextCalled     bool
		wantValidateCalled bool
		wantToken          string
		wantContextAuth    *usecase.AuthContext
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
			validateErr:        fmt.Errorf("invalid token"),
			wantStatus:         http.StatusUnauthorized,
			wantBodyValue:      "unauthorized",
			wantNextCalled:     false,
			wantValidateCalled: true,
			wantToken:          "bad-token",
		},
		{
			name:               "success",
			authHeader:         "Bearer good-token",
			validateAuth:       usecase.AuthContext{UserId: testUserID, Email: "user@example.com"},
			wantStatus:         http.StatusOK,
			wantNextCalled:     true,
			wantValidateCalled: true,
			wantToken:          "good-token",
			wantContextAuth:    &usecase.AuthContext{UserId: testUserID, Email: "user@example.com"},
		},
		{
			name:               "success with trimmed header",
			authHeader:         "   Bearer good-token   ",
			validateAuth:       usecase.AuthContext{UserId: testUserID, Email: "user@example.com"},
			wantStatus:         http.StatusOK,
			wantNextCalled:     true,
			wantValidateCalled: true,
			wantToken:          "good-token",
			wantContextAuth:    &usecase.AuthContext{UserId: testUserID, Email: "user@example.com"},
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
					Return(tt.validateAuth, tt.validateErr)
			}

			var nextCalled bool

			var nextAuth usecase.AuthContext

			m := &AuthMiddleware{usecase: mu}

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true

				auth, err := AuthFromContext(r.Context())
				if err != nil {
					t.Fatalf("expected auth in context, got error: %v", err)
				}

				nextAuth = auth

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

			if tt.wantContextAuth != nil {
				if nextAuth.UserId != tt.wantContextAuth.UserId {
					t.Fatalf("expected context user_id %v, got %v", tt.wantContextAuth.UserId, nextAuth.UserId)
				}

				if nextAuth.Email != tt.wantContextAuth.Email {
					t.Fatalf("expected context email %q, got %q", tt.wantContextAuth.Email, nextAuth.Email)
				}
			}

			if !strings.Contains(rr.Body.String(), `"status":"ok"`) {
				t.Fatalf("expected next handler response body, got %s", rr.Body.String())
			}
		})
	}
}

func TestAuthFromContext(t *testing.T) {
	t.Parallel()

	testUserID := int64(1)

	tests := []struct {
		name     string
		ctx      context.Context
		wantAuth usecase.AuthContext
		wantErr  bool
	}{
		{
			name: "auth exists",
			ctx: context.WithValue(context.Background(), AuthCtxKey,
				usecase.AuthContext{UserId: testUserID, Email: "user@example.com"}),
			wantAuth: usecase.AuthContext{UserId: testUserID, Email: "user@example.com"},
			wantErr:  false,
		},
		{
			name:    "auth missing",
			ctx:     context.Background(),
			wantErr: true,
		},
		{
			name:    "wrong type in context",
			ctx:     context.WithValue(context.Background(), AuthCtxKey, 123),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			auth, err := AuthFromContext(tt.ctx)

			if tt.wantErr {
				if !errors.Is(err, ErrMidlware) {
					t.Fatalf("expected ErrMidlware, got %v", err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if auth.UserId != tt.wantAuth.UserId {
				t.Fatalf("expected user_id %v, got %v", tt.wantAuth.UserId, auth.UserId)
			}

			if auth.Email != tt.wantAuth.Email {
				t.Fatalf("expected email %q, got %q", tt.wantAuth.Email, auth.Email)
			}
		})
	}
}

func TestAuthMiddleware_EnrichesRequestLogger(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	baseLogger, err := logger.New(logger.Config{Format: "json"})
	if err != nil {
		t.Fatalf("new logger: %v", err)
	}

	var output bytes.Buffer
	baseLogger.SetOutput(&output)

	mu := usecasemocks.NewMockUsecase(ctrl)
	mu.EXPECT().
		ValidateAccessToken("good-token").
		Return(usecase.AuthContext{
			UserId: 7,
			Email:  "user@example.com",
		}, nil)

	authMiddleware := &AuthMiddleware{usecase: mu}
	handler := LoggerMiddleware(baseLogger)(authMiddleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.FromContext(r.Context()).Info("inside handler")
		w.WriteHeader(http.StatusOK)
	})))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer good-token")
	req.Header.Set(requestIDHeader, "req-7")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 log lines, got %d", len(lines))
	}

	for _, line := range lines {
		payload := decodeLogLine(t, line)

		if payload["request_id"] != "req-7" {
			t.Fatalf("expected request_id req-7, got %v", payload["request_id"])
		}

		if payload["user_id"] != float64(7) {
			t.Fatalf("expected user_id 7, got %v", payload["user_id"])
		}

		if payload["email"] != "user@example.com" {
			t.Fatalf("expected email user@example.com, got %v", payload["email"])
		}
	}
}
