package usecase

import "github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/domain"

type Usecase interface {
	SignIn(email, password string) (domain.TokenPair, error)
	SignUp(email, password string) (domain.TokenPair, error)
	Refresh(email string) (domain.TokenPair, error)
	ValidateRefreshToken(tokenString string) (string, error)
	ValidateAccessToken(tokenString string) (string, error)
	LogOut(email string) error
	GetConfig() Config
}
