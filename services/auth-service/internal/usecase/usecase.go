package usecase

import (
	"context"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/services/auth-service/internal/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/services/auth-service/internal/repository"
	clocksvc "github.com/go-park-mail-ru/2026_1_VKino/services/auth-service/internal/service/clock"
	jwtsvc "github.com/go-park-mail-ru/2026_1_VKino/services/auth-service/internal/service/jwt"
	passwordsvc "github.com/go-park-mail-ru/2026_1_VKino/services/auth-service/internal/service/password"
)

type Config struct {
	JWTSecret       string        `mapstructure:"jwt_secret"`
	AccessTokenTTL  time.Duration `mapstructure:"access_token_ttl"`
	RefreshTokenTTL time.Duration `mapstructure:"refresh_token_ttl"`
	Issuer          string        `mapstructure:"issuer"`
}

type AuthContext = domain.AuthContext

type Usecase interface {
	SignIn(ctx context.Context, email, password string) (domain.TokenPair, error)
	SignUp(ctx context.Context, email, password string) (domain.TokenPair, error)
	Refresh(ctx context.Context, email string) (domain.TokenPair, error)
	ValidateRefreshToken(ctx context.Context, tokenString string) (string, error)
	LogOut(ctx context.Context, email string) error
	ValidateAccessToken(tokenString string) (AuthContext, error)
	ChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error
	GetConfig() Config
}

type AuthUsecase struct {
	userRepo        repository.UserRepo
	sessionRepo     repository.SessionRepo
	jwtService      jwtsvc.Service
	passwordService passwordsvc.Service
	clockService    clocksvc.Service
	cfg             Config
}
