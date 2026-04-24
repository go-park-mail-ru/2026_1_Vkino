package usecase

import (
	"context"

	domain2 "github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth-service/domain"
)

func (u *AuthUsecase) SignIn(ctx context.Context, email, password string) (domain2.TokenPair, error) {
	user, err := u.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return domain2.TokenPair{}, domain2.ErrInvalidCredentials
	}

	err = u.passwordService.Compare(user.Password, password)
	if err != nil {
		return domain2.TokenPair{}, domain2.ErrInvalidCredentials
	}

	return u.tokenPairGenerate(ctx, user)
}
