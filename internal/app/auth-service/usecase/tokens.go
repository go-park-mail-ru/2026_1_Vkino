package usecase

import (
	"context"
	"fmt"

	domain2 "github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth-service/domain"
)

func (u *AuthUsecase) tokenPairGenerate(ctx context.Context, user *domain2.User) (domain2.TokenPair, error) {
	accessToken, err := u.jwtService.GenerateToken(user.Email, user.ID, u.cfg.AccessTokenTTL)
	if err != nil {
		return domain2.TokenPair{}, fmt.Errorf("access token generate error: %w", err)
	}

	refreshToken, err := u.jwtService.GenerateToken(user.Email, user.ID, u.cfg.RefreshTokenTTL)
	if err != nil {
		return domain2.TokenPair{}, fmt.Errorf("refresh token generate error: %w", err)
	}

	expiresAt := u.clockService.Now().Add(u.cfg.RefreshTokenTTL)

	err = u.sessionRepo.SaveSession(ctx, user.ID, refreshToken, expiresAt)
	if err != nil {
		return domain2.TokenPair{}, fmt.Errorf("save session: %w", err)
	}

	return domain2.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
