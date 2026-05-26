package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/domain"
)

func (u *UserUsecase) GrantVKinoCoins(
	ctx context.Context,
	userID int64,
	coinsAmount int32,
	operationType string,
	description string,
	referenceKey string,
) (domain.VKinoCoinsGrant, error) {
	if userID <= 0 {
		return domain.VKinoCoinsGrant{}, domain.ErrInvalidToken
	}

	if coinsAmount <= 0 || operationType == "" || description == "" || referenceKey == "" {
		return domain.VKinoCoinsGrant{}, domain.ErrInvalidCoinsAmount
	}

	coinsGranted, balance, err := u.userRepo.GrantVKinoCoins(
		ctx,
		userID,
		coinsAmount,
		operationType,
		description,
		referenceKey,
	)
	if err != nil {
		switch err {
		case domain.ErrUserNotFound:
			return domain.VKinoCoinsGrant{}, domain.ErrInvalidToken
		case domain.ErrInvalidCoinsAmount:
			return domain.VKinoCoinsGrant{}, domain.ErrInvalidCoinsAmount
		default:
			return domain.VKinoCoinsGrant{}, domain.ErrInternal
		}
	}

	return domain.VKinoCoinsGrant{
		CoinsGranted:      coinsGranted,
		VKinoCoinsBalance: balance,
	}, nil
}
