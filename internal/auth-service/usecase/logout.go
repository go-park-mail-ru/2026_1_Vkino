package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/auth-service/domain"
)

func (u *AuthUsecase) LogOut(ctx context.Context, email string) error {
	user, err := u.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("logout failed: %w", err)
	}

	err = u.sessionRepo.DeleteSession(ctx, user.ID)
	if err != nil {
		if errors.Is(err, domain.ErrNoSession) {
			return nil
		}

		return err
	}

	return nil
}
