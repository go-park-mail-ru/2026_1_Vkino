package usecase

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_VKino/services/auth-service/internal/domain"
)

func (u *AuthUsecase) tokenPairGenerate(ctx context.Context, user *domain.User) (domain.TokenPair, error) {
	accessToken, err := u.jwtService.GenerateToken(user.Email, user.ID, u.cfg.AccessTokenTTL)
	if err != nil {
		return domain.TokenPair{}, fmt.Errorf("access token generate error: %w", err)
	}

	refreshToken, err := u.jwtService.GenerateToken(user.Email, user.ID, u.cfg.RefreshTokenTTL)
	if err != nil {
		return domain.TokenPair{}, fmt.Errorf("refresh token generate error: %w", err)
	}

	expiresAt := u.clockService.Now().Add(u.cfg.RefreshTokenTTL)

	err = u.sessionRepo.SaveSession(ctx, user.ID, refreshToken, expiresAt)
	if err != nil {
		return domain.TokenPair{}, fmt.Errorf("save session: %w", err)
	}

	return domain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}