package usecase_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user/usecase"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user/usecase/mocks"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func testConfig() usecase.Config {
	return usecase.Config{
		JWTSecret:         "super-secret-key",
		AccessTokenTTL:    time.Hour,
		RefreshTokenTTL:   24 * time.Hour,
		RefreshCookieName: "refresh_token",
		CookieSecure:      true,
	}
}

func hashPassword(t *testing.T, password string) string {
	t.Helper()

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	return string(hash)
}

func newTestUsecase(userRepo *mocks.MockUserRepo, sessionRepo *mocks.MockSessionRepo) *usecase.AuthUsecase {
	return usecase.NewAuthUsecase(userRepo, sessionRepo, testConfig())
}

func makeToken(t *testing.T, secret, subject string, userID int64, ttl time.Duration, method jwt.SigningMethod) string {
	t.Helper()

	claims := usecase.CustomClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(method, claims)

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	return tokenString
}

func TestAuthUsecase_SignIn(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name              string
		email             string
		password          string
		setupMocks        func(userRepo *mocks.MockUserRepo, sessionRepo *mocks.MockSessionRepo)
		wantErrIs         error
		wantErrContains   string
		wantValidateToken bool
	}{
		{
			name:     "user not found",
			email:    "user@example.com",
			password: "qwerty",
			setupMocks: func(userRepo *mocks.MockUserRepo, sessionRepo *mocks.MockSessionRepo) {
				userRepo.EXPECT().
					GetUserByEmail(gomock.Any(), "user@example.com").
					Return(nil, fmt.Errorf("not found"))
			},
			wantErrIs: domain.ErrInvalidCredentials,
		},
		{
			name:     "wrong password",
			email:    "user@example.com",
			password: "wrong-password",
			setupMocks: func(userRepo *mocks.MockUserRepo, sessionRepo *mocks.MockSessionRepo) {
				userRepo.EXPECT().
					GetUserByEmail(gomock.Any(), "user@example.com").
					Return(&domain.User{
						Email:    "user@example.com",
						Password: hashPassword(t, "correct-password"),
					}, nil)
			},
			wantErrIs: domain.ErrInvalidCredentials,
		},
		{
			name:     "save session error",
			email:    "user@example.com",
			password: "qwerty",
			setupMocks: func(userRepo *mocks.MockUserRepo, sessionRepo *mocks.MockSessionRepo) {
				userRepo.EXPECT().
					GetUserByEmail(gomock.Any(), "user@example.com").
					Return(&domain.User{
						Email:    "user@example.com",
						Password: hashPassword(t, "qwerty"),
					}, nil)
				sessionRepo.EXPECT().
					SaveSession(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(fmt.Errorf("session save failed"))
			},
			wantErrContains: "save session",
		},
		{
			name:     "success",
			email:    "user@example.com",
			password: "qwerty",
			setupMocks: func(userRepo *mocks.MockUserRepo, sessionRepo *mocks.MockSessionRepo) {
				userRepo.EXPECT().
					GetUserByEmail(gomock.Any(), "user@example.com").
					Return(&domain.User{
						Email:    "user@example.com",
						Password: hashPassword(t, "qwerty"),
						ID:       int64(1),
					}, nil)
				sessionRepo.EXPECT().
					SaveSession(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, _ int64, _ string, _ time.Time) error {
						return nil
					})
			},
			wantValidateToken: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := mocks.NewMockUserRepo(ctrl)
			sessionRepo := mocks.NewMockSessionRepo(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(userRepo, sessionRepo)
			}

			u := newTestUsecase(userRepo, sessionRepo)

			got, err := u.SignIn(context.Background(), tt.email, tt.password)

			if tt.wantErrIs != nil {
				if !errors.Is(err, tt.wantErrIs) {
					t.Fatalf("expected error %v, got %v", tt.wantErrIs, err)
				}
				return
			}

			if tt.wantErrContains != "" {
				if err == nil {
					t.Fatal("expected non-nil error, got nil")
				}

				if !strings.Contains(err.Error(), tt.wantErrContains) {
					t.Fatalf("expected error to contain %q, got %q", tt.wantErrContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.AccessToken == "" || got.RefreshToken == "" {
				t.Fatal("expected non-empty tokens")
			}

			if tt.wantValidateToken {
				auth, err := u.ValidateAccessToken(got.AccessToken)
				if err != nil {
					t.Fatalf("validate access token: %v", err)
				}

				if auth.Email != tt.email {
					t.Fatalf("expected token subject %q, got %q", tt.email, auth.Email)
				}
			}
		})
	}
}

func TestAuthUsecase_SignUp(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name            string
		email           string
		password        string
		setupMocks      func(userRepo *mocks.MockUserRepo, sessionRepo *mocks.MockSessionRepo)
		wantErrIs       error
		wantErrContains string
	}{
		{
			name:     "user already exists",
			email:    "user@example.com",
			password: "qwerty1",
			setupMocks: func(userRepo *mocks.MockUserRepo, sessionRepo *mocks.MockSessionRepo) {
				userRepo.EXPECT().
					GetUserByEmail(gomock.Any(), "user@example.com").
					Return(&domain.User{Email: "user@example.com"}, nil)
			},
			wantErrIs: domain.ErrUserAlreadyExists,
		},
		{
			name:     "create user error",
			email:    "user@example.com",
			password: "qwerty1",
			setupMocks: func(userRepo *mocks.MockUserRepo, sessionRepo *mocks.MockSessionRepo) {
				userRepo.EXPECT().
					GetUserByEmail(gomock.Any(), "user@example.com").
					Return(nil, fmt.Errorf("not found"))
				userRepo.EXPECT().
					CreateUser(gomock.Any(), "user@example.com", gomock.Any()).
					Return(nil, fmt.Errorf("create failed"))
			},
			wantErrContains: "create failed",
		},
		{
			name:     "save session error",
			email:    "user@example.com",
			password: "qwerty1",
			setupMocks: func(userRepo *mocks.MockUserRepo, sessionRepo *mocks.MockSessionRepo) {
				userRepo.EXPECT().
					GetUserByEmail(gomock.Any(), "user@example.com").
					Return(nil, fmt.Errorf("not found"))
				userRepo.EXPECT().
					CreateUser(gomock.Any(), "user@example.com", gomock.Any()).
					DoAndReturn(func(_ context.Context, email, password string) (*domain.User, error) {
						if err := bcrypt.CompareHashAndPassword([]byte(password), []byte("qwerty1")); err != nil {
							t.Errorf("password not properly hashed: %v", err)
						}
						return &domain.User{Email: email, ID: int64(1)}, nil
					})
				sessionRepo.EXPECT().
					SaveSession(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(fmt.Errorf("save session failed"))
			},
			wantErrContains: "save session",
		},
		{
			name:     "success",
			email:    "user@example.com",
			password: "qwerty1",
			setupMocks: func(userRepo *mocks.MockUserRepo, sessionRepo *mocks.MockSessionRepo) {
				userRepo.EXPECT().
					GetUserByEmail(gomock.Any(), "user@example.com").
					Return(nil, fmt.Errorf("not found"))
				userRepo.EXPECT().
					CreateUser(gomock.Any(), "user@example.com", gomock.Any()).
					DoAndReturn(func(_ context.Context, email, password string) (*domain.User, error) {
						if err := bcrypt.CompareHashAndPassword([]byte(password), []byte("qwerty1")); err != nil {
							t.Errorf("password not properly hashed: %v", err)
						}
						return &domain.User{Email: email, ID: int64(1)}, nil
					})
				sessionRepo.EXPECT().
					SaveSession(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, _ int64, _ string, _ time.Time) error {
						return nil
					})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := mocks.NewMockUserRepo(ctrl)
			sessionRepo := mocks.NewMockSessionRepo(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(userRepo, sessionRepo)
			}

			u := newTestUsecase(userRepo, sessionRepo)

			got, err := u.SignUp(context.Background(), tt.email, tt.password)

			if tt.wantErrIs != nil {
				if !errors.Is(err, tt.wantErrIs) {
					t.Fatalf("expected error %v, got %v", tt.wantErrIs, err)
				}
				return
			}

			if tt.wantErrContains != "" {
				if err == nil {
					t.Fatal("expected non-nil error, got nil")
				}

				if !strings.Contains(err.Error(), tt.wantErrContains) {
					t.Fatalf("expected error to contain %q, got %q", tt.wantErrContains, err.Error())
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.AccessToken == "" || got.RefreshToken == "" {
				t.Fatal("expected non-empty tokens")
			}
		})
	}
}

func TestAuthUsecase_Refresh(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name            string
		email           string
		setupMocks      func(userRepo *mocks.MockUserRepo, sessionRepo *mocks.MockSessionRepo)
		wantErrIs       error
		wantErrContains string
	}{
		{
			name:  "no session",
			email: "user@example.com",
			setupMocks: func(userRepo *mocks.MockUserRepo, sessionRepo *mocks.MockSessionRepo) {
				userRepo.EXPECT().
					GetUserByEmail(gomock.Any(), "user@example.com").
					Return(&domain.User{Email: "user@example.com", ID: int64(1)}, nil)
				sessionRepo.EXPECT().
					GetSession(gomock.Any(), int64(1)).
					Return("", fmt.Errorf("not found"))
			},
			wantErrIs: domain.ErrNoSession,
		},
		{
			name:  "user not found",
			email: "user@example.com",
			setupMocks: func(userRepo *mocks.MockUserRepo, sessionRepo *mocks.MockSessionRepo) {
				userRepo.EXPECT().
					GetUserByEmail(gomock.Any(), "user@example.com").
					Return(nil, fmt.Errorf("not found"))
			},
			wantErrIs: domain.ErrNoSession,
		},
		{
			name:  "save session error",
			email: "user@example.com",
			setupMocks: func(userRepo *mocks.MockUserRepo, sessionRepo *mocks.MockSessionRepo) {
				userRepo.EXPECT().
					GetUserByEmail(gomock.Any(), "user@example.com").
					Return(&domain.User{Email: "user@example.com", ID: int64(1)}, nil)
				sessionRepo.EXPECT().
					GetSession(gomock.Any(), int64(1)).
					Return("old-refresh", nil)
				sessionRepo.EXPECT().
					SaveSession(gomock.Any(), int64(1), gomock.Any(), gomock.Any()).
					Return(fmt.Errorf("save session failed"))
			},
			wantErrContains: "save session",
		},
		{
			name:  "success",
			email: "user@example.com",
			setupMocks: func(userRepo *mocks.MockUserRepo, sessionRepo *mocks.MockSessionRepo) {
				userRepo.EXPECT().
					GetUserByEmail(gomock.Any(), "user@example.com").
					Return(&domain.User{Email: "user@example.com", ID: int64(1)}, nil)
				sessionRepo.EXPECT().
					GetSession(gomock.Any(), int64(1)).
					Return("old-refresh", nil)
				sessionRepo.EXPECT().
					SaveSession(gomock.Any(), int64(1), gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, _ int64, _ string, _ time.Time) error {
						return nil
					})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := mocks.NewMockUserRepo(ctrl)
			sessionRepo := mocks.NewMockSessionRepo(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(userRepo, sessionRepo)
			}

			u := newTestUsecase(userRepo, sessionRepo)

			got, err := u.Refresh(context.Background(), tt.email)

			if tt.wantErrIs != nil {
				if !errors.Is(err, tt.wantErrIs) {
					t.Fatalf("expected error %v, got %v", tt.wantErrIs, err)
				}
				return
			}

			if tt.wantErrContains != "" {
				if err == nil {
					t.Fatal("expected non-nil error, got nil")
				}

				if !strings.Contains(err.Error(), tt.wantErrContains) {
					t.Fatalf("expected error to contain %q, got %q", tt.wantErrContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.AccessToken == "" || got.RefreshToken == "" {
				t.Fatal("expected non-empty tokens")
			}
		})
	}
}

func TestAuthUsecase_ValidateRefreshToken(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(func() {
		ctrl.Finish()
	})

	testUserID := int64(1)

	tests := []struct {
		name       string
		makeToken  func(t *testing.T, u *usecase.AuthUsecase) string
		setupMocks func(userRepo *mocks.MockUserRepo, sessionRepo *mocks.MockSessionRepo, token string)
		wantEmail  string
		wantErrIs  error
	}{
		{
			name: "invalid token string",
			makeToken: func(t *testing.T, u *usecase.AuthUsecase) string {
				t.Helper()
				return "not-a-jwt"
			},
			wantErrIs: domain.ErrInvalidToken,
		},
		{
			name: "empty subject",
			makeToken: func(t *testing.T, u *usecase.AuthUsecase) string {
				t.Helper()
				return makeToken(t, u.GetConfig().JWTSecret, "", testUserID, time.Hour, jwt.SigningMethodHS256)
			},
			wantErrIs: domain.ErrInvalidToken,
		},
		{
			name: "no session",
			makeToken: func(t *testing.T, u *usecase.AuthUsecase) string {
				t.Helper()
				return makeToken(t, u.GetConfig().JWTSecret, "user@example.com", testUserID, time.Hour, jwt.SigningMethodHS256)
			},
			setupMocks: func(userRepo *mocks.MockUserRepo, sessionRepo *mocks.MockSessionRepo, token string) {
				userRepo.EXPECT().
					GetUserByEmail(gomock.Any(), "user@example.com").
					Return(&domain.User{Email: "user@example.com", ID: testUserID}, nil)
				sessionRepo.EXPECT().
					GetSession(gomock.Any(), testUserID).
					Return("", fmt.Errorf("not found"))
			},
			wantErrIs: domain.ErrNoSession,
		},
		{
			name: "refresh token mismatch",
			makeToken: func(t *testing.T, u *usecase.AuthUsecase) string {
				t.Helper()
				return makeToken(t, u.GetConfig().JWTSecret, "user@example.com", testUserID, time.Hour, jwt.SigningMethodHS256)
			},
			setupMocks: func(userRepo *mocks.MockUserRepo, sessionRepo *mocks.MockSessionRepo, token string) {
				userRepo.EXPECT().
					GetUserByEmail(gomock.Any(), "user@example.com").
					Return(&domain.User{Email: "user@example.com", ID: testUserID}, nil)
				sessionRepo.EXPECT().
					GetSession(gomock.Any(), testUserID).
					Return("another-token", nil)
			},
			wantErrIs: domain.ErrInvalidToken,
		},
		{
			name: "success",
			makeToken: func(t *testing.T, u *usecase.AuthUsecase) string {
				t.Helper()
				return makeToken(t, u.GetConfig().JWTSecret, "user@example.com", testUserID, time.Hour, jwt.SigningMethodHS256)
			},
			setupMocks: func(userRepo *mocks.MockUserRepo, sessionRepo *mocks.MockSessionRepo, token string) {
				userRepo.EXPECT().
					GetUserByEmail(gomock.Any(), "user@example.com").
					Return(&domain.User{Email: "user@example.com", ID: testUserID}, nil)
				sessionRepo.EXPECT().
					GetSession(gomock.Any(), testUserID).
					Return(token, nil)
			},
			wantEmail: "user@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			userRepo := mocks.NewMockUserRepo(ctrl)
			sessionRepo := mocks.NewMockSessionRepo(ctrl)
			u := newTestUsecase(userRepo, sessionRepo)

			token := tt.makeToken(t, u)

			if tt.setupMocks != nil {
				tt.setupMocks(userRepo, sessionRepo, token)
			}

			email, err := u.ValidateRefreshToken(context.Background(), token)

			if tt.wantErrIs != nil {
				if !errors.Is(err, tt.wantErrIs) {
					t.Fatalf("expected error %v, got %v", tt.wantErrIs, err)
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

func TestAuthUsecase_UpdateProfile_BirthdateOnly(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := mocks.NewMockUserRepo(ctrl)
	sessionRepo := mocks.NewMockSessionRepo(ctrl)
	u := newTestUsecase(userRepo, sessionRepo)

	birthdate, _ := time.Parse("2006-01-02", "2001-09-12")
	userRepo.EXPECT().
		GetUserByID(gomock.Any(), int64(7)).
		Return(&domain.User{ID: 7, Email: "user@example.com"}, nil)
	userRepo.EXPECT().
		UpdateBirthdate(gomock.Any(), int64(7), gomock.Any()).
		Return(&domain.User{ID: 7, Email: "user@example.com", Birthdate: &birthdate}, nil)

	resp, err := u.UpdateProfile(context.Background(), 7, "2001-09-12", nil, 0, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Email != "user@example.com" {
		t.Fatalf("expected email user@example.com, got %q", resp.Email)
	}

	if resp.Birthdate == nil || *resp.Birthdate != "2001-09-12" {
		t.Fatalf("expected birthdate 2001-09-12, got %v", resp.Birthdate)
	}
}

func TestAuthUsecase_UpdateProfile_BirthdateAndAvatar(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := mocks.NewMockUserRepo(ctrl)
	sessionRepo := mocks.NewMockSessionRepo(ctrl)
	store := mocks.NewMockFileStorage(ctrl)
	u := usecase.NewAuthUsecaseWithStorage(userRepo, sessionRepo, store, testConfig())

	birthdate, _ := time.Parse("2006-01-02", "2002-10-13")
	oldAvatar := "users/7/avatar/old.jpg"
	newAvatar := ""

	userRepo.EXPECT().
		GetUserByID(gomock.Any(), int64(7)).
		Return(&domain.User{ID: 7, Email: "user@example.com", AvatarFileKey: &oldAvatar}, nil)

	userRepo.EXPECT().
		UpdateBirthdate(gomock.Any(), int64(7), gomock.Any()).
		Return(&domain.User{ID: 7, Email: "user@example.com", Birthdate: &birthdate, AvatarFileKey: &oldAvatar}, nil)

	putCall := store.EXPECT().
		PutObject(gomock.Any(), gomock.Any(), gomock.Any(), int64(3), "image/png").
		Return(nil)

	updateCall := userRepo.EXPECT().
		UpdateAvatarFileKey(gomock.Any(), int64(7), gomock.Any()).
		DoAndReturn(func(_ context.Context, _ int64, avatarFileKey *string) (*domain.User, error) {
			newAvatar = *avatarFileKey
			return &domain.User{ID: 7, Email: "user@example.com", Birthdate: &birthdate, AvatarFileKey: avatarFileKey}, nil
		})

	deleteCall := store.EXPECT().
		DeleteObject(gomock.Any(), oldAvatar).
		Return(nil)

	presignCall := store.EXPECT().
		PresignGetObject(gomock.Any(), gomock.Any(), time.Duration(0)).
		DoAndReturn(func(_ context.Context, key string, _ time.Duration) (string, error) {
			if newAvatar == "" {
				t.Fatalf("expected new avatar key to be set")
			}
			return "https://example.com/" + key, nil
		})

	gomock.InOrder(putCall, updateCall, deleteCall, presignCall)

	resp, err := u.UpdateProfile(
		context.Background(),
		7,
		"2002-10-13",
		strings.NewReader("img"),
		3,
		"image/png",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.AvatarURL == "" {
		t.Fatal("expected non-empty avatar url")
	}
}

func TestAuthUsecase_ValidateAccessToken(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(func() {
		ctrl.Finish()
	})

	testUserID := int64(1)

	tests := []struct {
		name      string
		makeToken func(t *testing.T, u *usecase.AuthUsecase) string
		wantAuth  usecase.AuthContext
		wantErrIs error
	}{
		{
			name: "invalid token string",
			makeToken: func(t *testing.T, u *usecase.AuthUsecase) string {
				t.Helper()
				return "bad-token"
			},
			wantErrIs: domain.ErrInvalidToken,
		},
		{
			name: "empty subject",
			makeToken: func(t *testing.T, u *usecase.AuthUsecase) string {
				t.Helper()
				return makeToken(t, u.GetConfig().JWTSecret, "", testUserID, time.Hour, jwt.SigningMethodHS256)
			},
			wantErrIs: domain.ErrInvalidToken,
		},
		{
			name: "success",
			makeToken: func(t *testing.T, u *usecase.AuthUsecase) string {
				t.Helper()
				return makeToken(t, u.GetConfig().JWTSecret, "user@example.com", testUserID, time.Hour, jwt.SigningMethodHS256)
			},
			wantAuth: usecase.AuthContext{
				UserId: testUserID,
				Email:  "user@example.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			u := newTestUsecase(mocks.NewMockUserRepo(ctrl), mocks.NewMockSessionRepo(ctrl))

			token := tt.makeToken(t, u)

			auth, err := u.ValidateAccessToken(token)

			if tt.wantErrIs != nil {
				if !errors.Is(err, tt.wantErrIs) {
					t.Fatalf("expected error %v, got %v", tt.wantErrIs, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if auth.Email != tt.wantAuth.Email {
				t.Fatalf("expected email %q, got %q", tt.wantAuth.Email, auth.Email)
			}

			if auth.UserId != tt.wantAuth.UserId {
				t.Fatalf("expected user_id %v, got %v", tt.wantAuth.UserId, auth.UserId)
			}
		})
	}
}

func TestAuthUsecase_ValidateAccessToken_ParseScenarios(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	t.Cleanup(func() {
		ctrl.Finish()
	})

	u := newTestUsecase(mocks.NewMockUserRepo(ctrl), mocks.NewMockSessionRepo(ctrl))
	testUserID := int64(1)

	t.Run("wrong signing method", func(t *testing.T) {
		tokenString := makeToken(t, u.GetConfig().JWTSecret, "user@example.com", testUserID, time.Hour,
			jwt.SigningMethodHS512)

		_, err := u.ValidateAccessToken(tokenString)
		if !errors.Is(err, domain.ErrInvalidToken) {
			t.Fatalf("expected error %v, got %v", domain.ErrInvalidToken, err)
		}
	})

	t.Run("expired token", func(t *testing.T) {
		t.Parallel()
		tokenString := makeToken(t, u.GetConfig().JWTSecret, "user@example.com", testUserID, -time.Hour,
			jwt.SigningMethodHS256)

		_, err := u.ValidateAccessToken(tokenString)
		if !errors.Is(err, domain.ErrInvalidToken) {
			t.Fatalf("expected error %v, got %v", domain.ErrInvalidToken, err)
		}
	})

	t.Run("success", func(t *testing.T) {
		tokenString := makeToken(t, u.GetConfig().JWTSecret, "user@example.com", testUserID, time.Hour,
			jwt.SigningMethodHS256)

		auth, err := u.ValidateAccessToken(tokenString)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if auth.Email != "user@example.com" {
			t.Fatalf("expected email %q, got %q", "user@example.com", auth.Email)
		}

		if auth.UserId != testUserID {
			t.Fatalf("expected user_id %v, got %v", testUserID, auth.UserId)
		}
	})
}

func TestAuthUsecase_GetConfig(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := testConfig()
	u := usecase.NewAuthUsecase(mocks.NewMockUserRepo(ctrl), mocks.NewMockSessionRepo(ctrl), cfg)

	got := u.GetConfig()

	if got != cfg {
		t.Fatalf("expected config %+v, got %+v", cfg, got)
	}
}
