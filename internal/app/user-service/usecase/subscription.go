package usecase

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/domain"
)

func (u *UserUsecase) GetSubscriptionCapabilities(
	ctx context.Context,
	userID int64,
) (domain.SubscriptionState, error) {
	if userID <= 0 {
		return domain.SubscriptionState{}, domain.ErrInvalidToken
	}

	state, err := u.resolveSubscriptionState(ctx, userID)
	if err != nil {
		return domain.SubscriptionState{}, fmt.Errorf("%w: %w", domain.ErrInternal, err)
	}

	return state, nil
}

func (u *UserUsecase) resolveSubscriptionState(
	ctx context.Context,
	userID int64,
) (domain.SubscriptionState, error) {
	info, err := u.resolveSubscriptionInfo(ctx, userID)
	if err != nil {
		return domain.SubscriptionState{}, err
	}

	if info.ID == 0 {
		defaultState := domain.DefaultSubscriptionState()
		defaultState.Subscription = info

		return defaultState, nil
	}

	options, err := u.userRepo.GetSubscriptionTariffOptions(ctx, info.ID)
	if err != nil {
		return domain.SubscriptionState{}, err
	}

	capabilities, err := domain.ParseSubscriptionCapabilities(options)
	if err != nil {
		return domain.SubscriptionState{}, err
	}

	usage, err := u.resolveSubscriptionUsage(ctx, userID, capabilities)
	if err != nil {
		return domain.SubscriptionState{}, err
	}

	return domain.SubscriptionState{
		Subscription: info,
		Capabilities: capabilities,
		Usage:        usage,
	}, nil
}

func (u *UserUsecase) resolveSubscriptionInfo(
	ctx context.Context,
	userID int64,
) (domain.SubscriptionInfo, error) {
	info, err := u.userRepo.GetActiveSubscription(ctx, userID)
	if err != nil {
		return domain.SubscriptionInfo{}, err
	}

	if info.ID != 0 {
		return info, nil
	}

	info, err = u.userRepo.GetSubscriptionTariffByCode(ctx, domain.SubscriptionFreeTariffCode)
	if err != nil {
		return domain.SubscriptionInfo{}, err
	}

	if info.ID == 0 {
		return domain.DefaultSubscriptionState().Subscription, nil
	}

	return info, nil
}

func (u *UserUsecase) resolveSubscriptionUsage(
	ctx context.Context,
	userID int64,
	capabilities domain.SubscriptionCapabilities,
) (domain.SubscriptionUsage, error) {
	coinsReceivedToday, err := u.userRepo.GetCoinsReceivedToday(ctx, userID)
	if err != nil {
		return domain.SubscriptionUsage{}, err
	}

	roomsCreatedThisMonth, err := u.userRepo.GetRoomsCreatedThisMonth(ctx, userID)
	if err != nil {
		return domain.SubscriptionUsage{}, err
	}

	return domain.SubscriptionUsage{
		CoinsReceivedToday:      coinsReceivedToday,
		CoinsRemainingToday:     capabilities.CoinsRemaining(coinsReceivedToday),
		RoomsCreatedThisMonth:   roomsCreatedThisMonth,
		RoomsRemainingThisMonth: capabilities.RoomsRemaining(roomsCreatedThisMonth),
	}, nil
}
