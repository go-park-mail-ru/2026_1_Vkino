package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/domain"
)

func (u *UserUsecase) ActivateSubscription(
	ctx context.Context,
	userID, tariffID, paymentID int64,
) (domain.SubscriptionInfo, error) {
	if userID <= 0 || tariffID <= 0 || paymentID <= 0 {
		return domain.SubscriptionInfo{}, domain.ErrInvalidToken
	}

	if _, err := u.userRepo.GetUserByID(ctx, userID); err != nil {
		return domain.SubscriptionInfo{}, err
	}

	tariff, err := u.userRepo.GetSubscriptionTariffByID(ctx, tariffID)
	if err != nil {
		return domain.SubscriptionInfo{}, err
	}

	current, err := u.userRepo.GetActiveSubscription(ctx, userID)
	if err != nil {
		return domain.SubscriptionInfo{}, fmt.Errorf("%w: %w", domain.ErrInternal, err)
	}

	now := u.clockService.Now()
	startsAt := now

	if current.ID != 0 && current.ActiveUntil != nil && current.Level <= tariff.Level {
		expiresAt, parseErr := time.Parse(time.RFC3339, *current.ActiveUntil)
		if parseErr == nil && expiresAt.After(now) {
			startsAt = expiresAt
		}
	}

	expiresAt := startsAt.AddDate(0, 0, int(tariff.DurationDays))

	if err = u.userRepo.DeactivateUserSubscriptions(ctx, userID); err != nil {
		return domain.SubscriptionInfo{}, fmt.Errorf("%w: %w", domain.ErrInternal, err)
	}

	if err = u.userRepo.CreateUserSubscription(ctx, userID, tariffID, now, expiresAt); err != nil {
		return domain.SubscriptionInfo{}, fmt.Errorf("%w: %w", domain.ErrInternal, err)
	}

	activeUntil := expiresAt.Format(time.RFC3339)

	return domain.SubscriptionInfo{
		ID:          tariff.ID,
		Code:        tariff.Code,
		Name:        tariff.Title,
		Level:       tariff.Level,
		ActiveUntil: &activeUntil,
	}, nil
}
