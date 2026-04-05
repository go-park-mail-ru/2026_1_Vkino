package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/domain"
)

//go:generate mockgen -source=./usecase.go -destination=./mocks/usecase_mock.go -package=mocks
type Usecase interface {
	SignIn(ctx context.Context, email, password string) (domain.TokenPair, error)
	SignUp(ctx context.Context, email, password string) (domain.TokenPair, error)
	Refresh(ctx context.Context, email string) (domain.TokenPair, error)
	ValidateRefreshToken(ctx context.Context, tokenString string) (string, error)
	LogOut(ctx context.Context, email string) error

	ValidateAccessToken(tokenString string) (AuthContext, error)
	GetConfig() Config
}
