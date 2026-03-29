package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type CustomClaims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

type Config struct {
	JWTSecret         string        `mapstructure:"jwt_secret"`
	AccessTokenTTL    time.Duration `mapstructure:"access_token_ttl"`
	RefreshTokenTTL   time.Duration `mapstructure:"refresh_token_ttl"`
	RefreshCookieName string        `mapstructure:"refresh_cookie_name"`
	CookieSecure      bool          `mapstructure:"cookie_secure"`
}

type AuthUsecase struct {
	userRepo    repository.UserRepo
	sessionRepo repository.SessionRepo
	cfg         Config
}

type AuthContext struct {
	UserId uuid.UUID
	Email  string
}

func NewAuthUsecase(userRepo repository.UserRepo, sessionRepo repository.SessionRepo, cfg Config) *AuthUsecase {
	return &AuthUsecase{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		cfg:         cfg,
	}
}

func (u *AuthUsecase) SignIn(ctx context.Context, email, password string) (domain.TokenPair, error) {
	user, err := u.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return domain.TokenPair{}, domain.ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return domain.TokenPair{}, domain.ErrInvalidCredentials
	}

	return u.tokenPairGenerate(ctx, user)
}

func (u *AuthUsecase) SignUp(ctx context.Context, email, password string) (domain.TokenPair, error) {
	if !domain.Validate(email, password) {
		return domain.TokenPair{}, domain.ErrInvalidCredentials
	}

	_, err := u.userRepo.GetUserByEmail(ctx, email)
	if err == nil {
		return domain.TokenPair{}, domain.ErrUserAlreadyExists
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return domain.TokenPair{}, fmt.Errorf("bcrypt generate error: %w", err)
	}

	user, err := u.userRepo.CreateUser(ctx, email, string(passwordHash))
	if err != nil {
		return domain.TokenPair{}, err
	}

	return u.tokenPairGenerate(ctx, user)
}

func (u *AuthUsecase) Refresh(ctx context.Context, email string) (domain.TokenPair, error) {
	user, err := u.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return domain.TokenPair{}, domain.ErrNoSession
	}

	if _, err := u.sessionRepo.GetSession(ctx, user.ID); err != nil {
		return domain.TokenPair{}, domain.ErrNoSession
	}

	return u.tokenPairGenerate(ctx, user)
}

func (u *AuthUsecase) ValidateRefreshToken(ctx context.Context, tokenString string) (string, error) {
	claims, err := u.parseToken(tokenString)
	if err != nil {
		return "", err
	}

	if claims.Subject == "" {
		return "", domain.ErrInvalidToken
	}

	user, err := u.userRepo.GetUserByEmail(ctx, claims.Subject)
	if err != nil {
		return "", domain.ErrNoSession
	}

	storedRefreshToken, err := u.sessionRepo.GetSession(ctx, user.ID)
	if err != nil {
		return "", domain.ErrNoSession
	}

	if storedRefreshToken != tokenString {
		return "", domain.ErrInvalidToken
	}

	return claims.Subject, nil
}

func (u *AuthUsecase) LogOut(ctx context.Context, email string) error {
	user, err := u.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil
	}

	err = u.sessionRepo.DeleteSession(ctx, user.ID)
	if err != nil {
		if errors.Is(err, domain.ErrNoSession) {
			return nil
		}
		return err
	}
	return nil
}

func (u *AuthUsecase) ValidateAccessToken(tokenString string) (AuthContext, error) {
	claims, err := u.parseToken(tokenString)
	if err != nil {
		return AuthContext{}, err
	}

	if claims.Subject == "" {
		return AuthContext{}, domain.ErrInvalidToken
	}

	return AuthContext{
		UserId: claims.UserID,
		Email:  claims.Subject,
	}, nil
}

func (u *AuthUsecase) GetConfig() Config {
	return u.cfg
}

func (u *AuthUsecase) tokenGenerate(userEmail string, userID uuid.UUID, tokenTTL time.Duration) (string, error) {
	claims := CustomClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userEmail,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	stringToken, err := token.SignedString([]byte(u.cfg.JWTSecret))
	if err != nil {
		return "", err
	}

	return stringToken, nil
}

func (u *AuthUsecase) tokenPairGenerate(ctx context.Context, user *domain.User) (domain.TokenPair, error) {
	accessToken, err := u.tokenGenerate(user.Email, user.ID, u.cfg.AccessTokenTTL)
	if err != nil {
		return domain.TokenPair{}, fmt.Errorf("token generate error: %w", err)
	}

	refreshToken, err := u.tokenGenerate(user.Email, user.ID, u.cfg.RefreshTokenTTL)
	if err != nil {
		return domain.TokenPair{}, fmt.Errorf("token generate error: %w", err)
	}

	expiresAt := time.Now().Add(u.cfg.RefreshTokenTTL)

	err = u.sessionRepo.SaveSession(ctx, user.ID, refreshToken, expiresAt)
	if err != nil {
		return domain.TokenPair{}, fmt.Errorf("save session: %w", err)
	}

	return domain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (u *AuthUsecase) parseToken(tokenString string) (*CustomClaims, error) {
	claims := &CustomClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if token.Method == nil || token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, domain.ErrInvalidToken
		}
		return []byte(u.cfg.JWTSecret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, domain.ErrInvalidToken
	}

	return claims, nil
}
