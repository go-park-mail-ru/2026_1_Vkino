package usecase

import (
	"context"

	domain2 "github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth-service/domain"
)

func (u *AuthUsecase) ChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error {
	if !domain2.ValidatePassword(newPassword) {
		return domain2.ErrInvalidCredentials
	}

	user, err := u.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return domain2.ErrInvalidToken
	}

	err = u.passwordService.Compare(user.Password, oldPassword)
	if err != nil {
		return domain2.ErrPasswordMismatch
	}

	newPasswordHash, err := u.passwordService.Hash(newPassword)
	if err != nil {
		return domain2.ErrInternal
	}

	err = u.userRepo.UpdatePassword(ctx, userID, newPasswordHash)
	if err != nil {
		return domain2.ErrInternal
	}

	return nil
}
