package usecase

import (
	"context"
	"errors"

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
	if err := validateVKinoCoinsGrantInput(userID, coinsAmount, operationType, description, referenceKey); err != nil {
		return domain.VKinoCoinsGrant{}, err
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
		return domain.VKinoCoinsGrant{}, mapVKinoCoinsGrantError(err)
	}

	return domain.VKinoCoinsGrant{
		CoinsGranted:      coinsGranted,
		VKinoCoinsBalance: balance,
	}, nil
}

func validateVKinoCoinsGrantInput(
	userID int64,
	coinsAmount int32,
	operationType string,
	description string,
	referenceKey string,
) error {
	if userID <= 0 {
		return domain.ErrInvalidToken
	}

	if coinsAmount <= 0 || operationType == "" || description == "" || referenceKey == "" {
		return domain.ErrInvalidCoinsAmount
	}

	return nil
}

func mapVKinoCoinsGrantError(err error) error {
	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		return domain.ErrInvalidToken
	case errors.Is(err, domain.ErrInvalidCoinsAmount):
		return domain.ErrInvalidCoinsAmount
	default:
		return domain.ErrInternal
	}
}
