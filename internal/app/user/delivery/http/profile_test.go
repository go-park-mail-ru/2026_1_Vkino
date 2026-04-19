package http

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	stdhttp "net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user/usecase"
	usecasemocks "github.com/go-park-mail-ru/2026_1_VKino/internal/app/user/usecase/mocks"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/pkg/middleware"
	"go.uber.org/mock/gomock"
)

func multipartRequest(t *testing.T, birthdate string, filename string, data []byte) (*stdhttp.Request, string) {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	if birthdate != "" {
		if err := writer.WriteField("birthdate", birthdate); err != nil {
			t.Fatalf("write birthdate: %v", err)
		}
	}

	if filename != "" {
		part, err := writer.CreateFormFile("avatar", filename)
		if err != nil {
			t.Fatalf("create file part: %v", err)
		}

		if _, err = part.Write(data); err != nil {
			t.Fatalf("write file part: %v", err)
		}
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	req := httptest.NewRequest(stdhttp.MethodPut, "/user/profile", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req, writer.FormDataContentType()
}

func multipartRequestWithoutAvatarContentType(
	t *testing.T,
	birthdate string,
	filename string,
	data []byte,
) *stdhttp.Request {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	if birthdate != "" {
		if err := writer.WriteField("birthdate", birthdate); err != nil {
			t.Fatalf("write birthdate: %v", err)
		}
	}

	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="avatar"; filename="%s"`, filename))

	part, err := writer.CreatePart(header)
	if err != nil {
		t.Fatalf("create file part: %v", err)
	}

	if _, err = part.Write(data); err != nil {
		t.Fatalf("write file part: %v", err)
	}

	if err = writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	req := httptest.NewRequest(stdhttp.MethodPut, "/user/profile", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req
}

func TestHandler_UpdateProfile(t *testing.T) {
	t.Parallel()

	t.Run("unauthorized", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mu := usecasemocks.NewMockUsecase(ctrl)
		h := NewHandler(mu)

		req, _ := multipartRequest(t, "", "", nil)
		rr := httptest.NewRecorder()

		h.UpdateProfile(rr, req)

		if rr.Code != stdhttp.StatusUnauthorized {
			t.Fatalf("expected status %d, got %d", stdhttp.StatusUnauthorized, rr.Code)
		}

		assertJSONContainsStringValue(t, rr, "unauthorized")
	})

	t.Run("invalid multipart", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mu := usecasemocks.NewMockUsecase(ctrl)
		h := NewHandler(mu)

		req := httptest.NewRequest(stdhttp.MethodPut, "/user/profile", bytes.NewBufferString("broken"))
		req.Header.Set("Content-Type", "multipart/form-data")
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthCtxKey, usecase.AuthContext{UserId: 1}))

		rr := httptest.NewRecorder()
		h.UpdateProfile(rr, req)

		if rr.Code != stdhttp.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", stdhttp.StatusBadRequest, rr.Code)
		}

		assertJSONContainsStringValue(t, rr, "invalid avatar")
	})

	t.Run("empty avatar", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mu := usecasemocks.NewMockUsecase(ctrl)
		h := NewHandler(mu)

		req, _ := multipartRequest(t, "", "avatar.png", nil)
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthCtxKey, usecase.AuthContext{UserId: 1}))

		rr := httptest.NewRecorder()
		h.UpdateProfile(rr, req)

		if rr.Code != stdhttp.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", stdhttp.StatusBadRequest, rr.Code)
		}

		assertJSONContainsStringValue(t, rr, "invalid avatar")
	})

	t.Run("usecase error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mu := usecasemocks.NewMockUsecase(ctrl)
		mu.EXPECT().
			UpdateProfile(gomock.Any(), int64(1), "2001-09-12", gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, _ int64, birthdate string, body io.Reader, size int64, contentType string) (domain.ProfileResponse, error) {
				if birthdate != "2001-09-12" || body != nil || size != 0 || contentType != "" {
					t.Fatalf("unexpected forwarded args: birthdate=%q body=%v size=%d contentType=%q", birthdate, body, size, contentType)
				}

				return domain.ProfileResponse{}, domain.ErrInvalidBirthdate
			})

		h := NewHandler(mu)
		req, _ := multipartRequest(t, "2001-09-12", "", nil)
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthCtxKey, usecase.AuthContext{UserId: 1}))

		rr := httptest.NewRecorder()
		h.UpdateProfile(rr, req)

		if rr.Code != stdhttp.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", stdhttp.StatusBadRequest, rr.Code)
		}

		assertJSONContainsStringValue(t, rr, "invalid birthdate")
	})

	t.Run("success with avatar", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mu := usecasemocks.NewMockUsecase(ctrl)
		mu.EXPECT().
			UpdateProfile(gomock.Any(), int64(1), "2001-09-12", gomock.Any(), int64(3), gomock.Any()).
			DoAndReturn(func(_ context.Context, _ int64, birthdate string, body io.Reader, size int64, contentType string) (domain.ProfileResponse, error) {
				if birthdate != "2001-09-12" || body == nil || size != 3 {
					t.Fatalf("unexpected forwarded args: birthdate=%q body=%v size=%d contentType=%q", birthdate, body, size, contentType)
				}

				return domain.ProfileResponse{Email: "user@example.com"}, nil
			})

		h := NewHandler(mu)
		req, _ := multipartRequest(t, "2001-09-12", "avatar.png", []byte("img"))
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthCtxKey, usecase.AuthContext{UserId: 1}))

		rr := httptest.NewRecorder()
		h.UpdateProfile(rr, req)

		if rr.Code != stdhttp.StatusOK {
			t.Fatalf("expected status %d, got %d", stdhttp.StatusOK, rr.Code)
		}

		assertJSONContainsStringValue(t, rr, "user@example.com")
	})

	t.Run("missing avatar content type does not fallback to filename", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mu := usecasemocks.NewMockUsecase(ctrl)
		mu.EXPECT().
			UpdateProfile(gomock.Any(), int64(1), "2001-09-12", gomock.Any(), int64(3), "").
			Return(domain.ProfileResponse{Email: "user@example.com"}, nil)

		h := NewHandler(mu)
		req := multipartRequestWithoutAvatarContentType(t, "2001-09-12", "avatar.png", []byte("img"))
		req = req.WithContext(context.WithValue(req.Context(), middleware.AuthCtxKey, usecase.AuthContext{UserId: 1}))

		rr := httptest.NewRecorder()
		h.UpdateProfile(rr, req)

		if rr.Code != stdhttp.StatusOK {
			t.Fatalf("expected status %d, got %d", stdhttp.StatusOK, rr.Code)
		}

		assertJSONContainsStringValue(t, rr, "user@example.com")
	})
}

func TestHandler_ChangePassword(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		withAuth   bool
		body       string
		setupMocks func(mu *usecasemocks.MockUsecase)
		wantStatus int
		wantString string
	}{
		{
			name:       "unauthorized",
			body:       mustJSON(t, domain.ChangePasswordRequest{OldPassword: "oldpass1", NewPassword: "newpass1"}),
			wantStatus: stdhttp.StatusUnauthorized,
			wantString: "unauthorized",
		},
		{
			name:       "invalid json",
			withAuth:   true,
			body:       `{"old_password":"oldpass1"`,
			wantStatus: stdhttp.StatusInternalServerError,
			wantString: "internal server error",
		},
		{
			name:     "usecase error",
			withAuth: true,
			body:     mustJSON(t, domain.ChangePasswordRequest{OldPassword: "oldpass1", NewPassword: "newpass1"}),
			setupMocks: func(mu *usecasemocks.MockUsecase) {
				mu.EXPECT().
					ChangePassword(gomock.Any(), int64(1), "oldpass1", "newpass1").
					Return(domain.ErrInvalidCredentials)
			},
			wantStatus: stdhttp.StatusUnauthorized,
			wantString: "invalid credentials",
		},
		{
			name:     "success",
			withAuth: true,
			body:     mustJSON(t, domain.ChangePasswordRequest{OldPassword: "oldpass1", NewPassword: "newpass1"}),
			setupMocks: func(mu *usecasemocks.MockUsecase) {
				mu.EXPECT().
					ChangePassword(gomock.Any(), int64(1), "oldpass1", "newpass1").
					Return(nil)
			},
			wantStatus: stdhttp.StatusOK,
			wantString: "password updated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mu := usecasemocks.NewMockUsecase(ctrl)
			if tt.setupMocks != nil {
				tt.setupMocks(mu)
			}

			h := NewHandler(mu)
			req := httptest.NewRequest(stdhttp.MethodPost, "/user/change-password", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			if tt.withAuth {
				req = req.WithContext(context.WithValue(req.Context(), middleware.AuthCtxKey, usecase.AuthContext{UserId: 1}))
			}

			rr := httptest.NewRecorder()
			h.ChangePassword(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rr.Code)
			}

			assertJSONContainsStringValue(t, rr, tt.wantString)
		})
	}
}
