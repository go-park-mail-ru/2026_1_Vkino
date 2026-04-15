package usecase_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user/usecase"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user/usecase/mocks"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthUsecase_GetProfile(t *testing.T) {
	t.Parallel()

	t.Run("invalid token", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		userRepo := mocks.NewMockUserRepo(ctrl)
		sessionRepo := mocks.NewMockSessionRepo(ctrl)
		userRepo.EXPECT().GetUserByID(gomock.Any(), int64(7)).Return(nil, errors.New("not found"))

		u := newTestUsecase(userRepo, sessionRepo)
		_, err := u.GetProfile(context.Background(), 7)
		if !errors.Is(err, domain.ErrInvalidToken) {
			t.Fatalf("expected ErrInvalidToken, got %v", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		userRepo := mocks.NewMockUserRepo(ctrl)
		sessionRepo := mocks.NewMockSessionRepo(ctrl)
		store := mocks.NewMockFileStorage(ctrl)

		birthdate, _ := time.Parse("2006-01-02", "2001-09-12")
		avatar := "avatars/7.png"

		userRepo.EXPECT().
			GetUserByID(gomock.Any(), int64(7)).
			Return(&domain.User{
				ID:            7,
				Email:         "user@example.com",
				Birthdate:     &birthdate,
				AvatarFileKey: &avatar,
			}, nil)
		store.EXPECT().
			PresignGetObject(gomock.Any(), avatar, time.Duration(0)).
			Return("https://cdn.example/avatars/7.png", nil)

		u := usecase.NewAuthUsecaseWithStorage(userRepo, sessionRepo, store, testConfig())
		got, err := u.GetProfile(context.Background(), 7)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got.Email != "user@example.com" || got.AvatarURL == "" || got.Birthdate == nil || *got.Birthdate != "2001-09-12" {
			t.Fatalf("unexpected profile: %#v", got)
		}
	})
}

func TestAuthUsecase_LogOut(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		setupMocks      func(userRepo *mocks.MockUserRepo, sessionRepo *mocks.MockSessionRepo)
		wantErrIs       error
		wantErrContains string
	}{
		{
			name: "get user error",
			setupMocks: func(userRepo *mocks.MockUserRepo, sessionRepo *mocks.MockSessionRepo) {
				userRepo.EXPECT().
					GetUserByEmail(gomock.Any(), "user@example.com").
					Return(nil, errors.New("db failed"))
			},
			wantErrContains: "logOut failed",
		},
		{
			name: "no session is ignored",
			setupMocks: func(userRepo *mocks.MockUserRepo, sessionRepo *mocks.MockSessionRepo) {
				userRepo.EXPECT().
					GetUserByEmail(gomock.Any(), "user@example.com").
					Return(&domain.User{ID: 7, Email: "user@example.com"}, nil)
				sessionRepo.EXPECT().
					DeleteSession(gomock.Any(), int64(7)).
					Return(domain.ErrNoSession)
			},
		},
		{
			name: "delete session error",
			setupMocks: func(userRepo *mocks.MockUserRepo, sessionRepo *mocks.MockSessionRepo) {
				userRepo.EXPECT().
					GetUserByEmail(gomock.Any(), "user@example.com").
					Return(&domain.User{ID: 7, Email: "user@example.com"}, nil)
				sessionRepo.EXPECT().
					DeleteSession(gomock.Any(), int64(7)).
					Return(errors.New("delete failed"))
			},
			wantErrContains: "delete failed",
		},
		{
			name: "success",
			setupMocks: func(userRepo *mocks.MockUserRepo, sessionRepo *mocks.MockSessionRepo) {
				userRepo.EXPECT().
					GetUserByEmail(gomock.Any(), "user@example.com").
					Return(&domain.User{ID: 7, Email: "user@example.com"}, nil)
				sessionRepo.EXPECT().
					DeleteSession(gomock.Any(), int64(7)).
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
			tt.setupMocks(userRepo, sessionRepo)

			u := newTestUsecase(userRepo, sessionRepo)
			err := u.LogOut(context.Background(), "user@example.com")

			if tt.wantErrIs != nil {
				if !errors.Is(err, tt.wantErrIs) {
					t.Fatalf("expected error %v, got %v", tt.wantErrIs, err)
				}

				return
			}

			if tt.wantErrContains != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErrContains) {
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

func TestAuthUsecase_ChangePassword(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		oldPass    string
		newPass    string
		setupMocks func(userRepo *mocks.MockUserRepo, sessionRepo *mocks.MockSessionRepo)
		wantErrIs  error
	}{
		{
			name:      "invalid new password",
			oldPass:   "qwerty1",
			newPass:   "short",
			wantErrIs: domain.ErrInvalidCredentials,
		},
		{
			name:    "user not found",
			oldPass: "qwerty1",
			newPass: "newpass1",
			setupMocks: func(userRepo *mocks.MockUserRepo, sessionRepo *mocks.MockSessionRepo) {
				userRepo.EXPECT().GetUserByID(gomock.Any(), int64(7)).Return(nil, errors.New("not found"))
			},
			wantErrIs: domain.ErrInvalidToken,
		},
		{
			name:    "old password mismatch",
			oldPass: "wrong1",
			newPass: "newpass1",
			setupMocks: func(userRepo *mocks.MockUserRepo, sessionRepo *mocks.MockSessionRepo) {
				userRepo.EXPECT().
					GetUserByID(gomock.Any(), int64(7)).
					Return(&domain.User{ID: 7, Password: hashPassword(t, "oldpass1")}, nil)
			},
			wantErrIs: domain.ErrPasswordMismatch,
		},
		{
			name:    "update error",
			oldPass: "oldpass1",
			newPass: "newpass1",
			setupMocks: func(userRepo *mocks.MockUserRepo, sessionRepo *mocks.MockSessionRepo) {
				userRepo.EXPECT().
					GetUserByID(gomock.Any(), int64(7)).
					Return(&domain.User{ID: 7, Password: hashPassword(t, "oldpass1")}, nil)
				userRepo.EXPECT().
					UpdatePassword(gomock.Any(), int64(7), gomock.Any()).
					Return(domain.ErrInternal)
			},
			wantErrIs: domain.ErrInternal,
		},
		{
			name:    "success",
			oldPass: "oldpass1",
			newPass: "newpass1",
			setupMocks: func(userRepo *mocks.MockUserRepo, sessionRepo *mocks.MockSessionRepo) {
				userRepo.EXPECT().
					GetUserByID(gomock.Any(), int64(7)).
					Return(&domain.User{ID: 7, Password: hashPassword(t, "oldpass1")}, nil)
				userRepo.EXPECT().
					UpdatePassword(gomock.Any(), int64(7), gomock.Any()).
					DoAndReturn(func(_ context.Context, _ int64, passwordHash string) error {
						if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte("newpass1")); err != nil {
							t.Fatalf("expected new password to be hashed: %v", err)
						}

						return nil
					})
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
				tt.setupMocks(userRepo, sessionRepo)
			}

			u := newTestUsecase(userRepo, sessionRepo)
			err := u.ChangePassword(context.Background(), 7, tt.oldPass, tt.newPass)

			if tt.wantErrIs != nil {
				if !errors.Is(err, tt.wantErrIs) {
					t.Fatalf("expected error %v, got %v", tt.wantErrIs, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
