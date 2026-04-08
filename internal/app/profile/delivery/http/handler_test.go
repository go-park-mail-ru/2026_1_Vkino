package http

import (
	"context"
	"encoding/json"
	stdhttp "net/http"
	"net/http/httptest"
	"testing"

	authusecase "github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/usecase"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/profile/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/pkg/middleware"
)

type usecaseStub struct {
	resp   domain.ProfileResponse
	err    error
	gotID  int64
	called bool
}

func (s *usecaseStub) GetProfile(_ context.Context, userID int64) (domain.ProfileResponse, error) {
	s.called = true
	s.gotID = userID

	if s.err != nil {
		return domain.ProfileResponse{}, s.err
	}

	return s.resp, nil
}

func decodeBody[T any](t *testing.T, rr *httptest.ResponseRecorder) T {
	t.Helper()

	var v T
	if err := json.Unmarshal(rr.Body.Bytes(), &v); err != nil {
		t.Fatalf("decode body: %v; body=%s", err, rr.Body.String())
	}

	return v
}

func assertJSONContainsStringValue(t *testing.T, rr *httptest.ResponseRecorder, want string) {
	t.Helper()

	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode generic body: %v; body=%s", err, rr.Body.String())
	}

	for _, value := range body {
		if s, ok := value.(string); ok && s == want {
			return
		}
	}

	t.Fatalf("expected body to contain string value %q, got %v", want, body)
}

func TestHandler_GetProfile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		ctxUserID      int64
		usecaseResp    domain.ProfileResponse
		usecaseErr     error
		wantStatus     int
		wantStringBody string
		wantCalled     bool
	}{
		{
			name:           "unauthorized",
			wantStatus:     stdhttp.StatusUnauthorized,
			wantStringBody: "unauthorized",
		},
		{
			name:           "user not found",
			ctxUserID:      7,
			usecaseErr:     domain.ErrUserNotFound,
			wantStatus:     stdhttp.StatusNotFound,
			wantStringBody: "user not found",
			wantCalled:     true,
		},
		{
			name:      "success",
			ctxUserID: 7,
			usecaseResp: domain.ProfileResponse{
				Email: "user@example.com",
			},
			wantStatus:     stdhttp.StatusOK,
			wantStringBody: "user@example.com",
			wantCalled:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			stub := &usecaseStub{
				resp: tt.usecaseResp,
				err:  tt.usecaseErr,
			}

			h := NewHandler(stub)

			req := httptest.NewRequest(stdhttp.MethodGet, "/auth/me", nil)
			if tt.ctxUserID != 0 {
				ctx := context.WithValue(req.Context(), middleware.AuthCtxKey, authusecase.AuthContext{
					UserId: tt.ctxUserID,
					Email:  "from-token@example.com",
				})
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()
			h.GetProfile(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rr.Code)
			}

			if !tt.wantCalled {
				if stub.called {
					t.Fatal("expected usecase not to be called")
				}

				assertJSONContainsStringValue(t, rr, tt.wantStringBody)

				return
			}

			if !stub.called {
				t.Fatal("expected usecase to be called")
			}

			if stub.gotID != tt.ctxUserID {
				t.Fatalf("expected user id %d, got %d", tt.ctxUserID, stub.gotID)
			}

			if tt.usecaseErr != nil {
				assertJSONContainsStringValue(t, rr, tt.wantStringBody)

				return
			}

			got := decodeBody[domain.ProfileResponse](t, rr)
			if got != tt.usecaseResp {
				t.Fatalf("expected response %+v, got %+v", tt.usecaseResp, got)
			}
		})
	}
}
