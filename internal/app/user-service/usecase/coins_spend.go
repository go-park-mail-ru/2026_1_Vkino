package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/domain"
)

func (u *UserUsecase) SpendVKinoCoins(
	ctx context.Context,
	userID int64,
	coinsAmount int32,
	operationType string,
	description string,
) (domain.VKinoCoinsSpend, error) {
	if err := validateVKinoCoinsSpendInput(userID, coinsAmount, operationType, description); err != nil {
		return domain.VKinoCoinsSpend{}, err
	}

	balance, err := u.userRepo.SpendVKinoCoins(ctx, userID, coinsAmount, operationType, description)
	if err != nil {
		return domain.VKinoCoinsSpend{}, mapVKinoCoinsSpendError(err)
	}

	return domain.VKinoCoinsSpend{
		CoinsSpent:        coinsAmount,
		VKinoCoinsBalance: balance,
	}, nil
}

func validateVKinoCoinsSpendInput(
	userID int64,
	coinsAmount int32,
	operationType string,
	description string,
) error {
	if userID <= 0 {
		return domain.ErrInvalidToken
	}

	if coinsAmount <= 0 {
		return domain.ErrInvalidCoinsAmount
	}

	if strings.TrimSpace(operationType) == "" || strings.TrimSpace(description) == "" {
		return domain.ErrInvalidCoinsAmount
	}

	return nil
}

func mapVKinoCoinsSpendError(err error) error {
	switch {
	case errors.Is(err, domain.ErrInsufficientVKinoCoins):
		return domain.ErrInsufficientVKinoCoins
	case errors.Is(err, domain.ErrUserNotFound):
		return domain.ErrInvalidToken
	case errors.Is(err, domain.ErrInvalidCoinsAmount):
		return domain.ErrInvalidCoinsAmount
	default:
		return fmt.Errorf("%w: spend vkino coins: %w", domain.ErrInternal, err)
	}
}
