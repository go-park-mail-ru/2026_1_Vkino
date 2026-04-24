package usecase

import (
	"context"

	domain2 "github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth-service/domain"
)

func (u *AuthUsecase) Refresh(ctx context.Context, email string) (domain2.TokenPair, error) {
	user, err := u.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return domain2.TokenPair{}, domain2.ErrNoSession
	}

	if _, err = u.sessionRepo.GetSession(ctx, user.ID); err != nil {
		return domain2.TokenPair{}, domain2.ErrNoSession
	}

	return u.tokenPairGenerate(ctx, user)
}
