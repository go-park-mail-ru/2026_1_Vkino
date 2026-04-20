package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_VKino/services/auth-service/internal/domain"
)

func (u *AuthUsecase) ValidateRefreshToken(ctx context.Context, tokenString string) (string, error) {
	authCtx, err := u.jwtService.ParseToken(tokenString)
	if err != nil {
		return "", err
	}

	if authCtx.Email == "" {
		return "", domain.ErrInvalidToken
	}

	user, err := u.userRepo.GetUserByEmail(ctx, authCtx.Email)
	if err != nil {
		return "", domain.ErrNoSession
	}

	storedRefreshToken, err := u.sessionRepo.GetSession(ctx, user.ID)
	if err != nil {
		return "", domain.ErrNoSession
	}

	if storedRefreshToken != tokenString {
		return "", domain.ErrInvalidToken
	}

	return authCtx.Email, nil
}

func (u *AuthUsecase) ValidateAccessToken(tokenString string) (AuthContext, error) {
	authCtx, err := u.jwtService.ParseToken(tokenString)
	if err != nil {
		return AuthContext{}, err
	}

	if authCtx.Email == "" {
		return AuthContext{}, domain.ErrInvalidToken
	}

	return AuthContext{
		UserID: authCtx.UserID,
		Email:  authCtx.Email,
	}, nil
}
