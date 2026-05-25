package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/domain"
)

func (u *UserUsecase) BuySubscriptionWithVKinoCoins(
	ctx context.Context,
	userID int64,
	tariffID int64,
) (domain.VKinoCoinsSubscriptionPurchase, error) {
	if err := validateVKinoCoinsSubscriptionPurchaseInput(userID, tariffID); err != nil {
		return domain.VKinoCoinsSubscriptionPurchase{}, err
	}

	purchase, err := u.userRepo.BuySubscriptionWithVKinoCoins(ctx, userID, tariffID)
	if err != nil {
		return domain.VKinoCoinsSubscriptionPurchase{}, mapVKinoCoinsSubscriptionPurchaseError(err)
	}

	return purchase, nil
}

func validateVKinoCoinsSubscriptionPurchaseInput(userID int64, tariffID int64) error {
	if userID <= 0 {
		return domain.ErrInvalidToken
	}

	if tariffID <= 0 {
		return domain.ErrSubscriptionTariffNotFound
	}

	return nil
}

func mapVKinoCoinsSubscriptionPurchaseError(err error) error {
	switch {
	case errors.Is(err, domain.ErrSubscriptionTariffNotFound):
		return domain.ErrSubscriptionTariffNotFound
	case errors.Is(err, domain.ErrTariffNotAvailableForCoins):
		return domain.ErrTariffNotAvailableForCoins
	case errors.Is(err, domain.ErrInsufficientVKinoCoins):
		return domain.ErrInsufficientVKinoCoins
	case errors.Is(err, domain.ErrUserNotFound):
		return domain.ErrInvalidToken
	default:
		return fmt.Errorf("%w: buy subscription with vkino coins: %w", domain.ErrInternal, err)
	}
}
