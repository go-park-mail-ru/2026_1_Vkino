package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_VKino/services/auth-service/internal/domain"
)

func (u *AuthUsecase) Refresh(ctx context.Context, email string) (domain.TokenPair, error) {
	user, err := u.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return domain.TokenPair{}, domain.ErrNoSession
	}

	if _, err = u.sessionRepo.GetSession(ctx, user.ID); err != nil {
		return domain.TokenPair{}, domain.ErrNoSession
	}

	return u.tokenPairGenerate(ctx, user)
}
