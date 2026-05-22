package usecase

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/subscription"
)

func (u *UserUsecase) GetSubscriptionCapabilities(ctx context.Context, userID int64) (subscription.State, error) {
	if userID <= 0 {
		return subscription.State{}, domain.ErrInvalidToken
	}

	state, err := u.resolveSubscriptionState(ctx, userID)
	if err != nil {
		return subscription.State{}, fmt.Errorf("%w: %w", domain.ErrInternal, err)
	}

	return state, nil
}

func (u *UserUsecase) resolveSubscriptionState(ctx context.Context, userID int64) (subscription.State, error) {
	info, err := u.resolveSubscriptionInfo(ctx, userID)
	if err != nil {
		return subscription.State{}, err
	}

	if info.ID == 0 {
		defaultState := subscription.DefaultState()
		defaultState.Subscription = info

		return defaultState, nil
	}

	options, err := u.userRepo.GetSubscriptionTariffOptions(ctx, info.ID)
	if err != nil {
		return subscription.State{}, err
	}

	capabilities, err := subscription.ParseCapabilities(options)
	if err != nil {
		return subscription.State{}, err
	}

	usage, err := u.resolveSubscriptionUsage(ctx, userID, capabilities)
	if err != nil {
		return subscription.State{}, err
	}

	return subscription.State{
		Subscription: info,
		Capabilities: capabilities,
		Usage:        usage,
	}, nil
}

func (u *UserUsecase) resolveSubscriptionInfo(ctx context.Context, userID int64) (subscription.Info, error) {
	info, err := u.userRepo.GetActiveSubscription(ctx, userID)
	if err != nil {
		return subscription.Info{}, err
	}

	if info.ID != 0 {
		return info, nil
	}

	info, err = u.userRepo.GetSubscriptionTariffByCode(ctx, subscription.FreeTariffCode)
	if err != nil {
		return subscription.Info{}, err
	}

	if info.ID == 0 {
		return subscription.DefaultState().Subscription, nil
	}

	return info, nil
}

func (u *UserUsecase) resolveSubscriptionUsage(
	ctx context.Context,
	userID int64,
	capabilities subscription.Capabilities,
) (subscription.Usage, error) {
	coinsReceivedToday, err := u.userRepo.GetCoinsReceivedToday(ctx, userID)
	if err != nil {
		return subscription.Usage{}, err
	}

	roomsCreatedThisMonth, err := u.userRepo.GetRoomsCreatedThisMonth(ctx, userID)
	if err != nil {
		return subscription.Usage{}, err
	}

	return subscription.Usage{
		CoinsReceivedToday:      coinsReceivedToday,
		CoinsRemainingToday:     capabilities.CoinsRemaining(coinsReceivedToday),
		RoomsCreatedThisMonth:   roomsCreatedThisMonth,
		RoomsRemainingThisMonth: capabilities.RoomsRemaining(roomsCreatedThisMonth),
	}, nil
}
