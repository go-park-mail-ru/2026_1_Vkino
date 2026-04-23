package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/auth-service/domain"
)

func (u *AuthUsecase) SignIn(ctx context.Context, email, password string) (domain.TokenPair, error) {
	user, err := u.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return domain.TokenPair{}, domain.ErrInvalidCredentials
	}

	err = u.passwordService.Compare(user.Password, password)
	if err != nil {
		return domain.TokenPair{}, domain.ErrInvalidCredentials
	}

	return u.tokenPairGenerate(ctx, user)
}
