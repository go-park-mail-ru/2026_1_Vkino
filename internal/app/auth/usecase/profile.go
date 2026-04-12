package usecase

import (
	"context"
	"errors"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/domain"
)

func (u *AuthUsecase) GetProfile(ctx context.Context, userID int64) (domain.Response, error) {
	user, err := u.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return domain.Response{}, domain.ErrUserNotFound
		}

		return domain.Response{}, err
	}

	return domain.Response{
		Email: user.Email,
	}, nil
}
