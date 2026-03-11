package usecase_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/usecase"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/usecase/mocks"
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

func makeToken(t *testing.T, secret, subject string, ttl time.Duration, method jwt.SigningMethod) string {
	t.Helper()

	token := jwt.NewWithClaims(method, jwt.RegisteredClaims{
		Subject:   subject,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	})

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
					GetUserByEmail("user@example.com").
					Return(nil, errors.New("not found"))
			},
			wantErrIs: domain.ErrInvalidCredentials,
		},
		{
			name:     "wrong password",
			email:    "user@example.com",
			password: "wrong-password",
			setupMocks: func(userRepo *mocks.MockUserRepo, sessionRepo *mocks.MockSessionRepo) {
				userRepo.EXPECT().
					GetUserByEmail("user@example.com").
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
					GetUserByEmail("user@example.com").
					Return(&domain.User{
						Email:    "user@example.com",
						Password: hashPassword(t, "qwerty"),
					}, nil)
				sessionRepo.EXPECT().
					SaveSession("user@example.com", gomock.Any()).
					Return(errors.New("session save failed"))
			},
			wantErrContains: "save session",
		},
		{
			name:     "success",
			email:    "user@example.com",
			password: "qwerty",
			setupMocks: func(userRepo *mocks.MockUserRepo, sessionRepo *mocks.MockSessionRepo) {
				userRepo.EXPECT().
					GetUserByEmail("user@example.com").
					Return(&domain.User{
						Email:    "user@example.com",
						Password: hashPassword(t, "qwerty"),
					}, nil)
				sessionRepo.EXPECT().
					SaveSession("user@example.com", gomock.Any()).
					DoAndReturn(func(email string, tokenPair domain.TokenPair) error {
						if tokenPair.AccessToken == "" || tokenPair.RefreshToken == "" {
							t.Error("expected non-empty tokens")
						}
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

			got, err := u.SignIn(tt.email, tt.password)

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
				email, err := u.ValidateAccessToken(got.AccessToken)
				if err != nil {
					t.Fatalf("validate access token: %v", err)
				}
				if email != tt.email {
					t.Fatalf("expected token subject %q, got %q", tt.email, email)
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
			password: "qwerty",
			setupMocks: func(userRepo *mocks.MockUserRepo, sessionRepo *mocks.MockSessionRepo) {
				userRepo.EXPECT().
					GetUserByEmail("user@example.com").
					Return(&domain.User{Email: "user@example.com"}, nil)
			},
			wantErrIs: domain.ErrUserAlreadyExists,
		},
		{
			name:     "create user error",
			email:    "user@example.com",
			password: "qwerty",
			setupMocks: func(userRepo *mocks.MockUserRepo, sessionRepo *mocks.MockSessionRepo) {
				userRepo.EXPECT().
					GetUserByEmail("user@example.com").
					Return(nil, errors.New("not found"))
				userRepo.EXPECT().
					CreateUser("user@example.com", gomock.Any()).
					Return(nil, errors.New("create failed"))
			},
			wantErrContains: "create failed",
		},
		{
			name:     "save session error",
			email:    "user@example.com",
			password: "qwerty",
			setupMocks: func(userRepo *mocks.MockUserRepo, sessionRepo *mocks.MockSessionRepo) {
				userRepo.EXPECT().
					GetUserByEmail("user@example.com").
					Return(nil, errors.New("not found"))
				userRepo.EXPECT().
					CreateUser("user@example.com", gomock.Any()).
					DoAndReturn(func(email, password string) (*domain.User, error) {
						if err := bcrypt.CompareHashAndPassword([]byte(password), []byte("qwerty")); err != nil {
							t.Errorf("password not properly hashed: %v", err)
						}
						return &domain.User{Email: email}, nil
					})
				sessionRepo.EXPECT().
					SaveSession("user@example.com", gomock.Any()).
					Return(errors.New("save session failed"))
			},
			wantErrContains: "save session",
		},
		{
			name:     "success",
			email:    "user@example.com",
			password: "qwerty",
			setupMocks: func(userRepo *mocks.MockUserRepo, sessionRepo *mocks.MockSessionRepo) {
				userRepo.EXPECT().
					GetUserByEmail("user@example.com").
					Return(nil, errors.New("not found"))
				userRepo.EXPECT().
					CreateUser("user@example.com", gomock.Any()).
					DoAndReturn(func(email, password string) (*domain.User, error) {
						if err := bcrypt.CompareHashAndPassword([]byte(password), []byte("qwerty")); err != nil {
							t.Errorf("password not properly hashed: %v", err)
						}
						return &domain.User{Email: email}, nil
					})
				sessionRepo.EXPECT().
					SaveSession("user@example.com", gomock.Any()).
					DoAndReturn(func(email string, tokenPair domain.TokenPair) error {
						if tokenPair.AccessToken == "" || tokenPair.RefreshToken == "" {
							t.Error("expected non-empty tokens")
						}
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

			got, err := u.SignUp(tt.email, tt.password)

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
				sessionRepo.EXPECT().
					GetSession("user@example.com").
					Return(nil, errors.New("not found"))
			},
			wantErrIs: domain.ErrNoSession,
		},
		{
			name:  "user not found",
			email: "user@example.com",
			setupMocks: func(userRepo *mocks.MockUserRepo, sessionRepo *mocks.MockSessionRepo) {
				sessionRepo.EXPECT().
					GetSession("user@example.com").
					Return(&domain.TokenPair{RefreshToken: "old-refresh"}, nil)
				userRepo.EXPECT().
					GetUserByEmail("user@example.com").
					Return(nil, errors.New("not found"))
			},
			wantErrIs: domain.ErrNoSession,
		},
		{
			name:  "save session error",
			email: "user@example.com",
			setupMocks: func(userRepo *mocks.MockUserRepo, sessionRepo *mocks.MockSessionRepo) {
				sessionRepo.EXPECT().
					GetSession("user@example.com").
					Return(&domain.TokenPair{RefreshToken: "old-refresh"}, nil)
				userRepo.EXPECT().
					GetUserByEmail("user@example.com").
					Return(&domain.User{Email: "user@example.com"}, nil)
				sessionRepo.EXPECT().
					SaveSession("user@example.com", gomock.Any()).
					Return(errors.New("save session failed"))
			},
			wantErrContains: "save session",
		},
		{
			name:  "success",
			email: "user@example.com",
			setupMocks: func(userRepo *mocks.MockUserRepo, sessionRepo *mocks.MockSessionRepo) {
				sessionRepo.EXPECT().
					GetSession("user@example.com").
					Return(&domain.TokenPair{RefreshToken: "old-refresh"}, nil)
				userRepo.EXPECT().
					GetUserByEmail("user@example.com").
					Return(&domain.User{Email: "user@example.com"}, nil)
				sessionRepo.EXPECT().
					SaveSession("user@example.com", gomock.Any()).
					DoAndReturn(func(email string, tokenPair domain.TokenPair) error {
						if tokenPair.AccessToken == "" || tokenPair.RefreshToken == "" {
							t.Error("expected non-empty tokens")
						}
						return nil
					})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Подготовка (Setup)
			userRepo := mocks.NewMockUserRepo(ctrl)
			sessionRepo := mocks.NewMockSessionRepo(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(userRepo, sessionRepo)
			}

			u := newTestUsecase(userRepo, sessionRepo)

			// Действие (Action)
			got, err := u.Refresh(tt.email)

			// Проверка (Assert)
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
	defer ctrl.Finish()

	tests := []struct {
		name       string
		makeToken  func(t *testing.T, u *usecase.AuthUsecase) string
		setupMocks func(sessionRepo *mocks.MockSessionRepo, token string)
		wantEmail  string
		wantErrIs  error
	}{
		{
			name: "invalid token string",
			makeToken: func(t *testing.T, u *usecase.AuthUsecase) string {
				return "not-a-jwt"
			},
			wantErrIs: domain.ErrInvalidToken,
		},
		{
			name: "empty subject",
			makeToken: func(t *testing.T, u *usecase.AuthUsecase) string {
				return makeToken(t, u.GetConfig().JWTSecret, "", time.Hour, jwt.SigningMethodHS256)
			},
			wantErrIs: domain.ErrInvalidToken,
		},
		{
			name: "no session",
			makeToken: func(t *testing.T, u *usecase.AuthUsecase) string {
				return makeToken(t, u.GetConfig().JWTSecret, "user@example.com", time.Hour, jwt.SigningMethodHS256)
			},
			setupMocks: func(sessionRepo *mocks.MockSessionRepo, token string) {
				sessionRepo.EXPECT().
					GetSession("user@example.com").
					Return(nil, errors.New("not found"))
			},
			wantErrIs: domain.ErrNoSession,
		},
		{
			name: "refresh token mismatch",
			makeToken: func(t *testing.T, u *usecase.AuthUsecase) string {
				return makeToken(t, u.GetConfig().JWTSecret, "user@example.com", time.Hour, jwt.SigningMethodHS256)
			},
			setupMocks: func(sessionRepo *mocks.MockSessionRepo, token string) {
				sessionRepo.EXPECT().
					GetSession("user@example.com").
					Return(&domain.TokenPair{RefreshToken: "another-token"}, nil)
			},
			wantErrIs: domain.ErrInvalidToken,
		},
		{
			name: "success",
			makeToken: func(t *testing.T, u *usecase.AuthUsecase) string {
				return makeToken(t, u.GetConfig().JWTSecret, "user@example.com", time.Hour, jwt.SigningMethodHS256)
			},
			setupMocks: func(sessionRepo *mocks.MockSessionRepo, token string) {
				sessionRepo.EXPECT().
					GetSession("user@example.com").
					Return(&domain.TokenPair{RefreshToken: token}, nil)
			},
			wantEmail: "user@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Подготовка (Setup)
			userRepo := mocks.NewMockUserRepo(ctrl)
			sessionRepo := mocks.NewMockSessionRepo(ctrl)
			u := newTestUsecase(userRepo, sessionRepo)

			token := tt.makeToken(t, u)

			if tt.setupMocks != nil {
				tt.setupMocks(sessionRepo, token)
			}

			// Действие (Action)
			email, err := u.ValidateRefreshToken(token)

			// Проверка (Assert)
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

func TestAuthUsecase_ValidateAccessToken(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name      string
		makeToken func(t *testing.T, u *usecase.AuthUsecase) string
		wantEmail string
		wantErrIs error
	}{
		{
			name: "invalid token string",
			makeToken: func(t *testing.T, u *usecase.AuthUsecase) string {
				return "bad-token"
			},
			wantErrIs: domain.ErrInvalidToken,
		},
		{
			name: "empty subject",
			makeToken: func(t *testing.T, u *usecase.AuthUsecase) string {
				return makeToken(t, u.GetConfig().JWTSecret, "", time.Hour, jwt.SigningMethodHS256)
			},
			wantErrIs: domain.ErrInvalidToken,
		},
		{
			name: "success",
			makeToken: func(t *testing.T, u *usecase.AuthUsecase) string {
				return makeToken(t, u.GetConfig().JWTSecret, "user@example.com", time.Hour, jwt.SigningMethodHS256)
			},
			wantEmail: "user@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Подготовка (Setup)
			u := newTestUsecase(mocks.NewMockUserRepo(ctrl), mocks.NewMockSessionRepo(ctrl))

			token := tt.makeToken(t, u)

			// Действие (Action)
			email, err := u.ValidateAccessToken(token)

			// Проверка (Assert)
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

func TestAuthUsecase_ValidateAccessToken_ParseScenarios(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	u := newTestUsecase(mocks.NewMockUserRepo(ctrl), mocks.NewMockSessionRepo(ctrl))

	t.Run("wrong signing method", func(t *testing.T) {
		tokenString := makeToken(t, u.GetConfig().JWTSecret, "user@example.com", time.Hour, jwt.SigningMethodHS512)
		_, err := u.ValidateAccessToken(tokenString)
		if !errors.Is(err, domain.ErrInvalidToken) {
			t.Fatalf("expected error %v, got %v", domain.ErrInvalidToken, err)
		}
	})

	t.Run("expired token", func(t *testing.T) {
		tokenString := makeToken(t, u.GetConfig().JWTSecret, "user@example.com", -time.Hour, jwt.SigningMethodHS256)
		_, err := u.ValidateAccessToken(tokenString)
		if !errors.Is(err, domain.ErrInvalidToken) {
			t.Fatalf("expected error %v, got %v", domain.ErrInvalidToken, err)
		}
	})

	t.Run("success", func(t *testing.T) {
		tokenString := makeToken(t, u.GetConfig().JWTSecret, "user@example.com", time.Hour, jwt.SigningMethodHS256)
		email, err := u.ValidateAccessToken(tokenString)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if email != "user@example.com" {
			t.Fatalf("expected email %q, got %q", "user@example.com", email)
		}
	})
}

func TestAuthUsecase_GetConfig(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Подготовка (Setup)
	cfg := testConfig()
	u := usecase.NewAuthUsecase(mocks.NewMockUserRepo(ctrl), mocks.NewMockSessionRepo(ctrl), cfg)

	// Действие (Action)
	got := u.GetConfig()

	// Проверка (Assert)
	if got != cfg {
		t.Fatalf("expected config %+v, got %+v", cfg, got)
	}
}
