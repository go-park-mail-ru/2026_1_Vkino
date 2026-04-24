package usecase

import (
	"context"
	"fmt"

	domain2 "github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth-service/domain"
)

func (u *AuthUsecase) SignUp(ctx context.Context, email, password string) (domain2.TokenPair, error) {
	if !domain2.Validate(email, password) {
		return domain2.TokenPair{}, domain2.ErrInvalidCredentials
	}

	_, err := u.userRepo.GetUserByEmail(ctx, email)
	if err == nil {
		return domain2.TokenPair{}, domain2.ErrUserAlreadyExists
	}

	passwordHash, err := u.passwordService.Hash(password)
	if err != nil {
		return domain2.TokenPair{}, fmt.Errorf("password hash error: %w", err)
	}

	user, err := u.userRepo.CreateUser(ctx, email, passwordHash)
	if err != nil {
		return domain2.TokenPair{}, err
	}

	return u.tokenPairGenerate(ctx, user)
}
