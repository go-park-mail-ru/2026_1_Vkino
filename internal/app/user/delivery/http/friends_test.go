package http

import (
	"context"
	stdhttp "net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user/usecase"
	usecasemocks "github.com/go-park-mail-ru/2026_1_VKino/internal/app/user/usecase/mocks"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/pkg/middleware"
	"go.uber.org/mock/gomock"
)

func TestHandler_SearchUsersByEmail(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		withAuth   bool
		query      string
		setupMocks func(mu *usecasemocks.MockUsecase)
		wantStatus int
		wantString string
	}{
		{
			name:       "unauthorized",
			query:      "example",
			wantStatus: stdhttp.StatusUnauthorized,
			wantString: "unauthorized",
		},
		{
			name:     "invalid query",
			withAuth: true,
			setupMocks: func(mu *usecasemocks.MockUsecase) {
				mu.EXPECT().
					SearchUsersByEmail(gomock.Any(), int64(1), "").
					Return(nil, domain.ErrInvalidSearchQuery)
			},
			wantStatus: stdhttp.StatusBadRequest,
			wantString: "invalid email query",
		},
		{
			name:     "success",
			withAuth: true,
			query:    "example",
			setupMocks: func(mu *usecasemocks.MockUsecase) {
				mu.EXPECT().
					SearchUsersByEmail(gomock.Any(), int64(1), "example").
					Return([]domain.UserSearchResult{
						{ID: 2, Email: "friend@example.com", IsFriend: true},
					}, nil)
			},
			wantStatus: stdhttp.StatusOK,
			wantString: "friend@example.com",
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

			req := httptest.NewRequest(stdhttp.MethodGet, "/user/search?email="+tt.query, nil)
			if tt.withAuth {
				req = req.WithContext(context.WithValue(req.Context(), middleware.AuthCtxKey, usecase.AuthContext{UserId: 1}))
			}

			rr := httptest.NewRecorder()
			h.SearchUsersByEmail(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rr.Code)
			}

			if tt.name == "success" {
				got := decodeBody[[]domain.UserSearchResult](t, rr)
				if len(got) != 1 || got[0].Email != tt.wantString {
					t.Fatalf("unexpected response: %#v", got)
				}

				return
			}

			assertJSONContainsStringValue(t, rr, tt.wantString)
		})
	}
}

func TestHandler_AddFriend(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		withAuth   bool
		pathID     string
		setupMocks func(mu *usecasemocks.MockUsecase)
		wantStatus int
		wantString string
	}{
		{
			name:       "unauthorized",
			pathID:     "2",
			wantStatus: stdhttp.StatusUnauthorized,
			wantString: "unauthorized",
		},
		{
			name:       "invalid path id",
			withAuth:   true,
			pathID:     "bad",
			wantStatus: stdhttp.StatusBadRequest,
			wantString: "invalid friend id",
		},
		{
			name:     "usecase error",
			withAuth: true,
			pathID:   "2",
			setupMocks: func(mu *usecasemocks.MockUsecase) {
				mu.EXPECT().
					AddFriend(gomock.Any(), int64(1), int64(2)).
					Return(domain.FriendResponse{}, domain.ErrAlreadyFriends)
			},
			wantStatus: stdhttp.StatusConflict,
			wantString: "users are already friends",
		},
		{
			name:     "success",
			withAuth: true,
			pathID:   "2",
			setupMocks: func(mu *usecasemocks.MockUsecase) {
				mu.EXPECT().
					AddFriend(gomock.Any(), int64(1), int64(2)).
					Return(domain.FriendResponse{ID: 2, Email: "friend@example.com"}, nil)
			},
			wantStatus: stdhttp.StatusCreated,
			wantString: "friend@example.com",
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
			req := httptest.NewRequest(stdhttp.MethodPost, "/user/friends/"+tt.pathID, nil)
			req.SetPathValue("id", tt.pathID)
			if tt.withAuth {
				req = req.WithContext(context.WithValue(req.Context(), middleware.AuthCtxKey, usecase.AuthContext{UserId: 1}))
			}

			rr := httptest.NewRecorder()
			h.AddFriend(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rr.Code)
			}

			assertJSONContainsStringValue(t, rr, tt.wantString)
		})
	}
}

func TestHandler_DeleteFriend(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		withAuth   bool
		pathID     string
		setupMocks func(mu *usecasemocks.MockUsecase)
		wantStatus int
		wantString string
	}{
		{
			name:       "unauthorized",
			pathID:     "2",
			wantStatus: stdhttp.StatusUnauthorized,
			wantString: "unauthorized",
		},
		{
			name:       "invalid path id",
			withAuth:   true,
			pathID:     "bad",
			wantStatus: stdhttp.StatusBadRequest,
			wantString: "invalid friend id",
		},
		{
			name:     "usecase error",
			withAuth: true,
			pathID:   "2",
			setupMocks: func(mu *usecasemocks.MockUsecase) {
				mu.EXPECT().
					DeleteFriend(gomock.Any(), int64(1), int64(2)).
					Return(domain.ErrFriendNotFound)
			},
			wantStatus: stdhttp.StatusNotFound,
			wantString: "friend not found",
		},
		{
			name:     "success",
			withAuth: true,
			pathID:   "2",
			setupMocks: func(mu *usecasemocks.MockUsecase) {
				mu.EXPECT().
					DeleteFriend(gomock.Any(), int64(1), int64(2)).
					Return(nil)
			},
			wantStatus: stdhttp.StatusOK,
			wantString: "friend deleted",
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
			req := httptest.NewRequest(stdhttp.MethodDelete, "/user/friends/"+tt.pathID, nil)
			req.SetPathValue("id", tt.pathID)
			if tt.withAuth {
				req = req.WithContext(context.WithValue(req.Context(), middleware.AuthCtxKey, usecase.AuthContext{UserId: 1}))
			}

			rr := httptest.NewRecorder()
			h.DeleteFriend(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rr.Code)
			}

			assertJSONContainsStringValue(t, rr, tt.wantString)
		})
	}
}
