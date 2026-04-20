package usecase_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user/usecase/mocks"
	"go.uber.org/mock/gomock"
)

func TestAuthUsecase_SearchUsersByEmail(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		query           string
		setupMocks      func(userRepo *mocks.MockUserRepo)
		want            []domain.UserSearchResult
		wantErrIs       error
		wantErrContains string
	}{
		{
			name:  "invalid token",
			query: "user@example.com",
			setupMocks: func(userRepo *mocks.MockUserRepo) {
				userRepo.EXPECT().
					GetUserByID(gomock.Any(), int64(7)).
					Return(nil, errors.New("not found"))
			},
			wantErrIs: domain.ErrInvalidToken,
		},
		{
			name:  "invalid query",
			query: "   ",
			setupMocks: func(userRepo *mocks.MockUserRepo) {
				userRepo.EXPECT().
					GetUserByID(gomock.Any(), int64(7)).
					Return(&domain.User{ID: 7}, nil)
			},
			wantErrIs: domain.ErrInvalidSearchQuery,
		},
		{
			name:  "repository error",
			query: "example",
			setupMocks: func(userRepo *mocks.MockUserRepo) {
				userRepo.EXPECT().
					GetUserByID(gomock.Any(), int64(7)).
					Return(&domain.User{ID: 7}, nil)
				userRepo.EXPECT().
					SearchUsersByEmail(gomock.Any(), int64(7), "example").
					Return(nil, errors.New("db failed"))
			},
			wantErrIs:       domain.ErrInternal,
			wantErrContains: "search users by email",
		},
		{
			name:  "success",
			query: "  example  ",
			setupMocks: func(userRepo *mocks.MockUserRepo) {
				userRepo.EXPECT().
					GetUserByID(gomock.Any(), int64(7)).
					Return(&domain.User{ID: 7}, nil)
				userRepo.EXPECT().
					SearchUsersByEmail(gomock.Any(), int64(7), "example").
					Return([]domain.UserSearchResult{
						{ID: 2, Email: "friend@example.com", IsFriend: true},
						{ID: 3, Email: "new@example.com", IsFriend: false},
					}, nil)
			},
			want: []domain.UserSearchResult{
				{ID: 2, Email: "friend@example.com", IsFriend: true},
				{ID: 3, Email: "new@example.com", IsFriend: false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userRepo := mocks.NewMockUserRepo(ctrl)
			sessionRepo := mocks.NewMockSessionRepo(ctrl)
			if tt.setupMocks != nil {
				tt.setupMocks(userRepo)
			}

			u := newTestUsecase(userRepo, sessionRepo)
			got, err := u.SearchUsersByEmail(context.Background(), 7, tt.query)

			if tt.wantErrIs != nil {
				if !errors.Is(err, tt.wantErrIs) {
					t.Fatalf("expected error %v, got %v", tt.wantErrIs, err)
				}

				if tt.wantErrContains != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErrContains)) {
					t.Fatalf("expected error containing %q, got %v", tt.wantErrContains, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(got) != len(tt.want) {
				t.Fatalf("expected %d results, got %d", len(tt.want), len(got))
			}

			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("expected result %#v, got %#v", tt.want[i], got[i])
				}
			}
		})
	}
}

func TestAuthUsecase_AddFriend(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		friendID        int64
		setupMocks      func(userRepo *mocks.MockUserRepo)
		want            domain.FriendResponse
		wantErrIs       error
		wantErrContains string
	}{
		{
			name:     "invalid token",
			friendID: 2,
			setupMocks: func(userRepo *mocks.MockUserRepo) {
				userRepo.EXPECT().
					GetUserByID(gomock.Any(), int64(7)).
					Return(nil, errors.New("not found"))
			},
			wantErrIs: domain.ErrInvalidToken,
		},
		{
			name:     "self friendship",
			friendID: 7,
			setupMocks: func(userRepo *mocks.MockUserRepo) {
				userRepo.EXPECT().
					GetUserByID(gomock.Any(), int64(7)).
					Return(&domain.User{ID: 7}, nil)
			},
			wantErrIs: domain.ErrSelfFriendship,
		},
		{
			name:     "friend user not found",
			friendID: 2,
			setupMocks: func(userRepo *mocks.MockUserRepo) {
				userRepo.EXPECT().
					GetUserByID(gomock.Any(), int64(7)).
					Return(&domain.User{ID: 7}, nil)
				userRepo.EXPECT().
					GetUserByID(gomock.Any(), int64(2)).
					Return(nil, errors.New("not found"))
			},
			wantErrIs: domain.ErrUserNotFound,
		},
		{
			name:     "already friends",
			friendID: 2,
			setupMocks: func(userRepo *mocks.MockUserRepo) {
				userRepo.EXPECT().
					GetUserByID(gomock.Any(), int64(7)).
					Return(&domain.User{ID: 7}, nil)
				userRepo.EXPECT().
					GetUserByID(gomock.Any(), int64(2)).
					Return(&domain.User{ID: 2, Email: "friend@example.com"}, nil)
				userRepo.EXPECT().
					AddFriend(gomock.Any(), int64(7), int64(2)).
					Return(domain.ErrAlreadyFriends)
			},
			wantErrIs: domain.ErrAlreadyFriends,
		},
		{
			name:     "repository error",
			friendID: 2,
			setupMocks: func(userRepo *mocks.MockUserRepo) {
				userRepo.EXPECT().
					GetUserByID(gomock.Any(), int64(7)).
					Return(&domain.User{ID: 7}, nil)
				userRepo.EXPECT().
					GetUserByID(gomock.Any(), int64(2)).
					Return(&domain.User{ID: 2, Email: "friend@example.com"}, nil)
				userRepo.EXPECT().
					AddFriend(gomock.Any(), int64(7), int64(2)).
					Return(errors.New("db failed"))
			},
			wantErrIs:       domain.ErrInternal,
			wantErrContains: "add friend",
		},
		{
			name:     "success",
			friendID: 2,
			setupMocks: func(userRepo *mocks.MockUserRepo) {
				userRepo.EXPECT().
					GetUserByID(gomock.Any(), int64(7)).
					Return(&domain.User{ID: 7}, nil)
				userRepo.EXPECT().
					GetUserByID(gomock.Any(), int64(2)).
					Return(&domain.User{ID: 2, Email: "friend@example.com"}, nil)
				userRepo.EXPECT().
					AddFriend(gomock.Any(), int64(7), int64(2)).
					Return(nil)
			},
			want: domain.FriendResponse{ID: 2, Email: "friend@example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userRepo := mocks.NewMockUserRepo(ctrl)
			sessionRepo := mocks.NewMockSessionRepo(ctrl)
			if tt.setupMocks != nil {
				tt.setupMocks(userRepo)
			}

			u := newTestUsecase(userRepo, sessionRepo)
			got, err := u.AddFriend(context.Background(), 7, tt.friendID)

			if tt.wantErrIs != nil {
				if !errors.Is(err, tt.wantErrIs) {
					t.Fatalf("expected error %v, got %v", tt.wantErrIs, err)
				}

				if tt.wantErrContains != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErrContains)) {
					t.Fatalf("expected error containing %q, got %v", tt.wantErrContains, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tt.want {
				t.Fatalf("expected %#v, got %#v", tt.want, got)
			}
		})
	}
}

func TestAuthUsecase_DeleteFriend(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		friendID        int64
		setupMocks      func(userRepo *mocks.MockUserRepo)
		wantErrIs       error
		wantErrContains string
	}{
		{
			name:     "invalid token",
			friendID: 2,
			setupMocks: func(userRepo *mocks.MockUserRepo) {
				userRepo.EXPECT().
					GetUserByID(gomock.Any(), int64(7)).
					Return(nil, errors.New("not found"))
			},
			wantErrIs: domain.ErrInvalidToken,
		},
		{
			name:     "self friendship",
			friendID: 7,
			setupMocks: func(userRepo *mocks.MockUserRepo) {
				userRepo.EXPECT().
					GetUserByID(gomock.Any(), int64(7)).
					Return(&domain.User{ID: 7}, nil)
			},
			wantErrIs: domain.ErrSelfFriendship,
		},
		{
			name:     "friend user not found",
			friendID: 2,
			setupMocks: func(userRepo *mocks.MockUserRepo) {
				userRepo.EXPECT().
					GetUserByID(gomock.Any(), int64(7)).
					Return(&domain.User{ID: 7}, nil)
				userRepo.EXPECT().
					GetUserByID(gomock.Any(), int64(2)).
					Return(nil, errors.New("not found"))
			},
			wantErrIs: domain.ErrUserNotFound,
		},
		{
			name:     "friendship not found",
			friendID: 2,
			setupMocks: func(userRepo *mocks.MockUserRepo) {
				userRepo.EXPECT().
					GetUserByID(gomock.Any(), int64(7)).
					Return(&domain.User{ID: 7}, nil)
				userRepo.EXPECT().
					GetUserByID(gomock.Any(), int64(2)).
					Return(&domain.User{ID: 2}, nil)
				userRepo.EXPECT().
					DeleteFriend(gomock.Any(), int64(7), int64(2)).
					Return(domain.ErrFriendNotFound)
			},
			wantErrIs: domain.ErrFriendNotFound,
		},
		{
			name:     "repository error",
			friendID: 2,
			setupMocks: func(userRepo *mocks.MockUserRepo) {
				userRepo.EXPECT().
					GetUserByID(gomock.Any(), int64(7)).
					Return(&domain.User{ID: 7}, nil)
				userRepo.EXPECT().
					GetUserByID(gomock.Any(), int64(2)).
					Return(&domain.User{ID: 2}, nil)
				userRepo.EXPECT().
					DeleteFriend(gomock.Any(), int64(7), int64(2)).
					Return(errors.New("db failed"))
			},
			wantErrIs:       domain.ErrInternal,
			wantErrContains: "delete friend",
		},
		{
			name:     "success",
			friendID: 2,
			setupMocks: func(userRepo *mocks.MockUserRepo) {
				userRepo.EXPECT().
					GetUserByID(gomock.Any(), int64(7)).
					Return(&domain.User{ID: 7}, nil)
				userRepo.EXPECT().
					GetUserByID(gomock.Any(), int64(2)).
					Return(&domain.User{ID: 2}, nil)
				userRepo.EXPECT().
					DeleteFriend(gomock.Any(), int64(7), int64(2)).
					Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userRepo := mocks.NewMockUserRepo(ctrl)
			sessionRepo := mocks.NewMockSessionRepo(ctrl)
			if tt.setupMocks != nil {
				tt.setupMocks(userRepo)
			}

			u := newTestUsecase(userRepo, sessionRepo)
			err := u.DeleteFriend(context.Background(), 7, tt.friendID)

			if tt.wantErrIs != nil {
				if !errors.Is(err, tt.wantErrIs) {
					t.Fatalf("expected error %v, got %v", tt.wantErrIs, err)
				}

				if tt.wantErrContains != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErrContains)) {
					t.Fatalf("expected error containing %q, got %v", tt.wantErrContains, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
