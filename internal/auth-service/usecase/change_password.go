package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/auth-service/domain"
)

func (u *AuthUsecase) ChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error {
	if !domain.ValidatePassword(newPassword) {
		return domain.ErrInvalidCredentials
	}

	user, err := u.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return domain.ErrInvalidToken
	}

	err = u.passwordService.Compare(user.Password, oldPassword)
	if err != nil {
		return domain.ErrPasswordMismatch
	}

	newPasswordHash, err := u.passwordService.Hash(newPassword)
	if err != nil {
		return domain.ErrInternal
	}

	err = u.userRepo.UpdatePassword(ctx, userID, newPasswordHash)
	if err != nil {
		return domain.ErrInternal
	}

	return nil
}
