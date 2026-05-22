package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie-service/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/subscription"
)

func (u *MovieUsecase) subscriptionState(ctx context.Context, userID int64) (subscription.State, error) {
	if userID <= 0 || u.subscriptionReader == nil {
		return subscription.DefaultState(), nil
	}

	return u.subscriptionReader.GetSubscriptionState(ctx, userID)
}

func (u *MovieUsecase) viewerSubscriptionState(ctx context.Context) (subscription.State, error) {
	return u.subscriptionState(ctx, viewerIDFromContext(ctx))
}

func (u *MovieUsecase) ensurePaidContentAllowed(ctx context.Context, episode *domain.Episode) error {
	if episode == nil || (!episode.IsPaid && !episode.MovieIsPaid) {
		return nil
	}

	state, err := u.viewerSubscriptionState(ctx)
	if err != nil {
		return err
	}

	if state.Capabilities.CanWatchPaidContent {
		return nil
	}

	return subscription.NewFeatureForbidden(
		subscription.CodePaidContentForbidden,
		subscription.FeaturePaidContent,
		state.Subscription.Level,
		2,
		"Платный контент недоступен на вашем уровне подписки.",
	)
}

func (u *MovieUsecase) ensureSmartContinueAllowed(ctx context.Context, userID int64) error {
	state, err := u.subscriptionState(ctx, userID)
	if err != nil {
		return err
	}

	if state.Capabilities.CanUseSmartContinue {
		return nil
	}

	return subscription.NewFeatureForbidden(
		subscription.CodeSmartContinueForbidden,
		subscription.FeatureSmartContinue,
		state.Subscription.Level,
		2,
		"Умное продолжение просмотра недоступно на вашем уровне подписки.",
	)
}
