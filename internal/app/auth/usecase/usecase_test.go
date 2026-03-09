package usecase

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type mockUserRepo struct {
	getUserByEmailFn func(email string) (*domain.User, error)
	createUserFn     func(email, password string) (*domain.User, error)
}

func (m *mockUserRepo) GetUserByEmail(email string) (*domain.User, error) {
	if m.getUserByEmailFn == nil {
		panic("unexpected call: GetUserByEmail")
	}
	return m.getUserByEmailFn(email)
}

func (m *mockUserRepo) GetUserByID(id uuid.UUID) (*domain.User, error) {
	panic("unexpected call: GetUserByID")
}

func (m *mockUserRepo) CreateUser(email string, password string) (*domain.User, error) {
	if m.createUserFn == nil {
		panic("unexpected call: CreateUser")
	}
	return m.createUserFn(email, password)
}

func (m *mockUserRepo) UpdateUser(login string, password string) (*domain.User, error) {
	panic("unexpected call: UpdateUser")
}

func (m *mockUserRepo) GetAllUsers() ([]*domain.User, error) {
	panic("unexpected call: GetAllUsers")
}

func (m *mockUserRepo) DeleteUser(login string) error {
	panic("unexpected call: DeleteUser")
}

type mockSessionRepo struct {
	getSessionFn  func(email string) (*domain.TokenPair, error)
	saveSessionFn func(email string, tokenPair domain.TokenPair) error
}

func (m *mockSessionRepo) SaveSession(email string, tokenPair domain.TokenPair) error {
	if m.saveSessionFn == nil {
		panic("unexpected call: SaveSession")
	}
	return m.saveSessionFn(email, tokenPair)
}

func (m *mockSessionRepo) GetSession(email string) (*domain.TokenPair, error) {
	if m.getSessionFn == nil {
		panic("unexpected call: GetSession")
	}
	return m.getSessionFn(email)
}

func (m *mockSessionRepo) DeleteSession(email string) error {
	panic("unexpected call: DeleteSession")
}

var _ repository.UserRepo = (*mockUserRepo)(nil)
var _ repository.SessionRepo = (*mockSessionRepo)(nil)

func testConfig() Config {
	return Config{
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

func newTestUsecase(userRepo *mockUserRepo, sessionRepo *mockSessionRepo) *AuthUsecase {
	return NewAuthUsecase(userRepo, sessionRepo, testConfig())
}

func TestAuthUsecase_SignIn(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		email             string
		password          string
		userRepoUser      *domain.User
		userRepoErr       error
		saveSessionErr    error
		wantErrIs         error
		wantErrContains   string
		wantSaveCalled    bool
		wantSavedEmail    string
		wantValidateToken bool
	}{
		{
			name:        "user not found",
			email:       "user@example.com",
			password:    "qwerty",
			userRepoErr: errors.New("not found"),
			wantErrIs:   domain.ErrInvalidCredentials,
		},
		{
			name:     "wrong password",
			email:    "user@example.com",
			password: "wrong-password",
			userRepoUser: &domain.User{
				Email:    "user@example.com",
				Password: hashPassword(t, "correct-password"),
			},
			wantErrIs: domain.ErrInvalidCredentials,
		},
		{
			name:     "save session error",
			email:    "user@example.com",
			password: "qwerty",
			userRepoUser: &domain.User{
				Email:    "user@example.com",
				Password: hashPassword(t, "qwerty"),
			},
			saveSessionErr:  errors.New("session save failed"),
			wantErrContains: "failed to save session",
			wantSaveCalled:  true,
			wantSavedEmail:  "user@example.com",
		},
		{
			name:     "success",
			email:    "user@example.com",
			password: "qwerty",
			userRepoUser: &domain.User{
				Email:    "user@example.com",
				Password: hashPassword(t, "qwerty"),
			},
			wantSaveCalled:    true,
			wantSavedEmail:    "user@example.com",
			wantValidateToken: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var saveCalled bool
			var savedEmail string
			var savedPair domain.TokenPair

			userRepo := &mockUserRepo{
				getUserByEmailFn: func(email string) (*domain.User, error) {
					if email != tt.email {
						t.Fatalf("expected email %q, got %q", tt.email, email)
					}
					return tt.userRepoUser, tt.userRepoErr
				},
			}

			sessionRepo := &mockSessionRepo{
				saveSessionFn: func(email string, tokenPair domain.TokenPair) error {
					saveCalled = true
					savedEmail = email
					savedPair = tokenPair
					return tt.saveSessionErr
				},
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
				if saveCalled != tt.wantSaveCalled {
					t.Fatalf("expected save called=%v, got %v", tt.wantSaveCalled, saveCalled)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !saveCalled {
				t.Fatal("expected SaveSession to be called")
			}
			if savedEmail != tt.wantSavedEmail {
				t.Fatalf("expected saved email %q, got %q", tt.wantSavedEmail, savedEmail)
			}
			if got != savedPair {
				t.Fatal("expected returned token pair to equal saved pair")
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

	tests := []struct {
		name            string
		email           string
		password        string
		existingUser    *domain.User
		getUserErr      error
		createUserResp  *domain.User
		createUserErr   error
		saveSessionErr  error
		wantErrIs       error
		wantErrContains string
		wantCreateCall  bool
		wantSaveCall    bool
	}{
		{
			name:     "user already exists",
			email:    "user@example.com",
			password: "qwerty",
			existingUser: &domain.User{
				Email: "user@example.com",
			},
			wantErrIs: domain.ErrUserAlreadyExists,
		},
		{
			name:           "create user error",
			email:          "user@example.com",
			password:       "qwerty",
			getUserErr:     errors.New("not found"),
			createUserErr:  errors.New("create failed"),
			wantCreateCall: true,
		},
		{
			name:       "save session error",
			email:      "user@example.com",
			password:   "qwerty",
			getUserErr: errors.New("not found"),
			createUserResp: &domain.User{
				Email: "user@example.com",
			},
			saveSessionErr:  errors.New("save session failed"),
			wantCreateCall:  true,
			wantSaveCall:    true,
			wantErrContains: "failed to save session",
		},
		{
			name:       "success",
			email:      "user@example.com",
			password:   "qwerty",
			getUserErr: errors.New("not found"),
			createUserResp: &domain.User{
				Email: "user@example.com",
			},
			wantCreateCall: true,
			wantSaveCall:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var createCalled bool
			var createEmail string
			var createPassword string
			var saveCalled bool

			userRepo := &mockUserRepo{
				getUserByEmailFn: func(email string) (*domain.User, error) {
					if email != tt.email {
						t.Fatalf("expected email %q, got %q", tt.email, email)
					}
					return tt.existingUser, tt.getUserErr
				},
				createUserFn: func(email, password string) (*domain.User, error) {
					createCalled = true
					createEmail = email
					createPassword = password
					return tt.createUserResp, tt.createUserErr
				},
			}

			sessionRepo := &mockSessionRepo{
				saveSessionFn: func(email string, tokenPair domain.TokenPair) error {
					saveCalled = true
					return tt.saveSessionErr
				},
			}

			u := newTestUsecase(userRepo, sessionRepo)

			got, err := u.SignUp(tt.email, tt.password)

			if tt.wantErrIs != nil {
				if !errors.Is(err, tt.wantErrIs) {
					t.Fatalf("expected error %v, got %v", tt.wantErrIs, err)
				}
				if createCalled {
					t.Fatal("expected CreateUser not to be called")
				}
				return
			}

			if tt.createUserErr != nil {
				if !errors.Is(err, tt.createUserErr) {
					t.Fatalf("expected error %v, got %v", tt.createUserErr, err)
				}
				if createCalled != tt.wantCreateCall {
					t.Fatalf("expected create called=%v, got %v", tt.wantCreateCall, createCalled)
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
				if createCalled != tt.wantCreateCall {
					t.Fatalf("expected create called=%v, got %v", tt.wantCreateCall, createCalled)
				}
				if saveCalled != tt.wantSaveCall {
					t.Fatalf("expected save called=%v, got %v", tt.wantSaveCall, saveCalled)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !createCalled {
				t.Fatal("expected CreateUser to be called")
			}
			if createEmail != tt.email {
				t.Fatalf("expected create email %q, got %q", tt.email, createEmail)
			}
			if createPassword == tt.password {
				t.Fatal("expected password to be hashed before CreateUser")
			}
			if err := bcrypt.CompareHashAndPassword([]byte(createPassword), []byte(tt.password)); err != nil {
				t.Fatalf("expected valid bcrypt hash, got error: %v", err)
			}
			if !saveCalled {
				t.Fatal("expected SaveSession to be called")
			}
			if got.AccessToken == "" || got.RefreshToken == "" {
				t.Fatal("expected non-empty tokens")
			}
		})
	}
}

func TestAuthUsecase_Refresh(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		email           string
		sessionErr      error
		user            *domain.User
		userErr         error
		saveSessionErr  error
		wantErrIs       error
		wantErrContains string
		wantSaveCall    bool
	}{
		{
			name:       "no session",
			email:      "user@example.com",
			sessionErr: errors.New("not found"),
			wantErrIs:  domain.ErrNoSession,
		},
		{
			name:      "user not found",
			email:     "user@example.com",
			userErr:   errors.New("not found"),
			wantErrIs: domain.ErrNoSession,
		},
		{
			name:  "save session error",
			email: "user@example.com",
			user: &domain.User{
				Email: "user@example.com",
			},
			saveSessionErr:  errors.New("save session failed"),
			wantErrContains: "failed to save session",
			wantSaveCall:    true,
		},
		{
			name:  "success",
			email: "user@example.com",
			user: &domain.User{
				Email: "user@example.com",
			},
			wantSaveCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var saveCalled bool
			var savedEmail string
			var savedPair domain.TokenPair

			userRepo := &mockUserRepo{
				getUserByEmailFn: func(email string) (*domain.User, error) {
					if email != tt.email {
						t.Fatalf("expected email %q, got %q", tt.email, email)
					}
					return tt.user, tt.userErr
				},
			}

			sessionRepo := &mockSessionRepo{
				getSessionFn: func(email string) (*domain.TokenPair, error) {
					if email != tt.email {
						t.Fatalf("expected session email %q, got %q", tt.email, email)
					}
					if tt.sessionErr != nil {
						return nil, tt.sessionErr
					}
					return &domain.TokenPair{RefreshToken: "old-refresh"}, nil
				},
				saveSessionFn: func(email string, tokenPair domain.TokenPair) error {
					saveCalled = true
					savedEmail = email
					savedPair = tokenPair
					return tt.saveSessionErr
				},
			}

			u := newTestUsecase(userRepo, sessionRepo)

			got, err := u.Refresh(tt.email)

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
				if saveCalled != tt.wantSaveCall {
					t.Fatalf("expected save called=%v, got %v", tt.wantSaveCall, saveCalled)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !saveCalled {
				t.Fatal("expected SaveSession to be called")
			}
			if savedEmail != tt.email {
				t.Fatalf("expected saved email %q, got %q", tt.email, savedEmail)
			}
			if got != savedPair {
				t.Fatal("expected returned token pair to equal saved pair")
			}
			if got.AccessToken == "" || got.RefreshToken == "" {
				t.Fatal("expected non-empty tokens")
			}
		})
	}
}

func TestAuthUsecase_ValidateRefreshToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		makeToken     func(t *testing.T, u *AuthUsecase) string
		sessionPair   *domain.TokenPair
		sessionErr    error
		wantEmail     string
		wantErrIs     error
		wantGetCalled bool
	}{
		{
			name: "invalid token string",
			makeToken: func(t *testing.T, u *AuthUsecase) string {
				t.Helper()
				return "not-a-jwt"
			},
			wantErrIs: domain.ErrInvalidToken,
		},
		{
			name: "empty subject",
			makeToken: func(t *testing.T, u *AuthUsecase) string {
				t.Helper()
				token, err := u.tokenGenerate("", time.Hour)
				if err != nil {
					t.Fatalf("generate token: %v", err)
				}
				return token
			},
			wantErrIs: domain.ErrInvalidToken,
		},
		{
			name: "no session",
			makeToken: func(t *testing.T, u *AuthUsecase) string {
				t.Helper()
				token, err := u.tokenGenerate("user@example.com", time.Hour)
				if err != nil {
					t.Fatalf("generate token: %v", err)
				}
				return token
			},
			sessionErr:    errors.New("not found"),
			wantErrIs:     domain.ErrNoSession,
			wantGetCalled: true,
		},
		{
			name: "refresh token mismatch",
			makeToken: func(t *testing.T, u *AuthUsecase) string {
				t.Helper()
				token, err := u.tokenGenerate("user@example.com", time.Hour)
				if err != nil {
					t.Fatalf("generate token: %v", err)
				}
				return token
			},
			sessionPair:   &domain.TokenPair{RefreshToken: "another-token"},
			wantErrIs:     domain.ErrInvalidToken,
			wantGetCalled: true,
		},
		{
			name: "success",
			makeToken: func(t *testing.T, u *AuthUsecase) string {
				t.Helper()
				token, err := u.tokenGenerate("user@example.com", time.Hour)
				if err != nil {
					t.Fatalf("generate token: %v", err)
				}
				return token
			},
			wantEmail:     "user@example.com",
			wantGetCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var getCalled bool
			var gotEmail string

			sessionRepo := &mockSessionRepo{}
			userRepo := &mockUserRepo{}
			u := newTestUsecase(userRepo, sessionRepo)

			token := tt.makeToken(t, u)

			sessionRepo.getSessionFn = func(email string) (*domain.TokenPair, error) {
				getCalled = true
				gotEmail = email

				if tt.sessionErr != nil {
					return nil, tt.sessionErr
				}
				if tt.wantEmail != "" {
					return &domain.TokenPair{RefreshToken: token}, nil
				}
				return tt.sessionPair, nil
			}

			email, err := u.ValidateRefreshToken(token)

			if tt.wantErrIs != nil {
				if !errors.Is(err, tt.wantErrIs) {
					t.Fatalf("expected error %v, got %v", tt.wantErrIs, err)
				}
				if getCalled != tt.wantGetCalled {
					t.Fatalf("expected get session called=%v, got %v", tt.wantGetCalled, getCalled)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if email != tt.wantEmail {
				t.Fatalf("expected email %q, got %q", tt.wantEmail, email)
			}
			if getCalled != tt.wantGetCalled {
				t.Fatalf("expected get session called=%v, got %v", tt.wantGetCalled, getCalled)
			}
			if gotEmail != tt.wantEmail {
				t.Fatalf("expected session email %q, got %q", tt.wantEmail, gotEmail)
			}
		})
	}
}

func TestAuthUsecase_ValidateAccessToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		makeToken func(t *testing.T, u *AuthUsecase) string
		wantEmail string
		wantErrIs error
	}{
		{
			name: "invalid token string",
			makeToken: func(t *testing.T, u *AuthUsecase) string {
				t.Helper()
				return "bad-token"
			},
			wantErrIs: domain.ErrInvalidToken,
		},
		{
			name: "empty subject",
			makeToken: func(t *testing.T, u *AuthUsecase) string {
				t.Helper()
				token, err := u.tokenGenerate("", time.Hour)
				if err != nil {
					t.Fatalf("generate token: %v", err)
				}
				return token
			},
			wantErrIs: domain.ErrInvalidToken,
		},
		{
			name: "success",
			makeToken: func(t *testing.T, u *AuthUsecase) string {
				t.Helper()
				token, err := u.tokenGenerate("user@example.com", time.Hour)
				if err != nil {
					t.Fatalf("generate token: %v", err)
				}
				return token
			},
			wantEmail: "user@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := newTestUsecase(&mockUserRepo{}, &mockSessionRepo{})

			token := tt.makeToken(t, u)
			email, err := u.ValidateAccessToken(token)

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

func TestAuthUsecase_parseToken(t *testing.T) {
	t.Parallel()

	u := newTestUsecase(&mockUserRepo{}, &mockSessionRepo{})

	t.Run("wrong signing method", func(t *testing.T) {
		token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.RegisteredClaims{
			Subject:   "user@example.com",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		})

		tokenString, err := token.SignedString([]byte(u.cfg.JWTSecret))
		if err != nil {
			t.Fatalf("sign token: %v", err)
		}

		_, err = u.parseToken(tokenString)
		if !errors.Is(err, domain.ErrInvalidToken) {
			t.Fatalf("expected error %v, got %v", domain.ErrInvalidToken, err)
		}
	})

	t.Run("expired token", func(t *testing.T) {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
			Subject:   "user@example.com",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		})

		tokenString, err := token.SignedString([]byte(u.cfg.JWTSecret))
		if err != nil {
			t.Fatalf("sign token: %v", err)
		}

		_, err = u.parseToken(tokenString)
		if !errors.Is(err, domain.ErrInvalidToken) {
			t.Fatalf("expected error %v, got %v", domain.ErrInvalidToken, err)
		}
	})

	t.Run("success", func(t *testing.T) {
		tokenString, err := u.tokenGenerate("user@example.com", time.Hour)
		if err != nil {
			t.Fatalf("generate token: %v", err)
		}

		claims, err := u.parseToken(tokenString)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if claims.Subject != "user@example.com" {
			t.Fatalf("expected subject %q, got %q", "user@example.com", claims.Subject)
		}
	})
}

func TestAuthUsecase_GetConfig(t *testing.T) {
	t.Parallel()

	cfg := testConfig()
	u := NewAuthUsecase(&mockUserRepo{}, &mockSessionRepo{}, cfg)

	got := u.GetConfig()

	if got != cfg {
		t.Fatalf("expected config %+v, got %+v", cfg, got)
	}
}
