package usecase

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_VKino/services/auth-service/internal/domain"
)

func (u *AuthUsecase) SignUp(ctx context.Context, email, password string) (domain.TokenPair, error) {
	if !domain.Validate(email, password) {
		return domain.TokenPair{}, domain.ErrInvalidCredentials
	}

	_, err := u.userRepo.GetUserByEmail(ctx, email)
	if err == nil {
		return domain.TokenPair{}, domain.ErrUserAlreadyExists
	}

	passwordHash, err := u.passwordService.Hash(password)
	if err != nil {
		return domain.TokenPair{}, fmt.Errorf("password hash error: %w", err)
	}

	user, err := u.userRepo.CreateUser(ctx, email, passwordHash)
	if err != nil {
		return domain.TokenPair{}, err
	}

	return u.tokenPairGenerate(ctx, user)
}