package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie-service/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/common/capabilityerr"
	userv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/user/v1"
)

const (
	subscriptionFeatureForbiddenCode       = "SUBSCRIPTION_FEATURE_FORBIDDEN"
	paidContentFeatureCode                 = "paid_content"
	smartContinueFeatureCode               = "smart_continue"
	requiredPaidContentLevel         int32 = 2
	requiredSmartContinueLevel       int32 = 2
)

func (u *MovieUsecase) subscriptionState(
	ctx context.Context,
	userID int64,
) (*userv1.GetSubscriptionCapabilitiesResponse, error) {
	if userID <= 0 || u.subscriptionReader == nil {
		return defaultSubscriptionState(), nil
	}

	return u.subscriptionReader.GetSubscriptionState(ctx, userID)
}

func (u *MovieUsecase) viewerSubscriptionState(
	ctx context.Context,
) (*userv1.GetSubscriptionCapabilitiesResponse, error) {
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

	if state.GetCapabilities().GetCanWatchPaidContent() {
		return nil
	}

	return newFeatureForbidden(
		subscriptionFeatureForbiddenCode,
		paidContentFeatureCode,
		state.GetSubscription().GetLevel(),
		requiredPaidContentLevel,
		"Платный контент недоступен на вашем уровне подписки.",
	)
}

func (u *MovieUsecase) ensureSmartContinueAllowed(ctx context.Context, userID int64) error {
	state, err := u.subscriptionState(ctx, userID)
	if err != nil {
		return err
	}

	if state.GetCapabilities().GetCanUseSmartContinue() {
		return nil
	}

	return newFeatureForbidden(
		subscriptionFeatureForbiddenCode,
		smartContinueFeatureCode,
		state.GetSubscription().GetLevel(),
		requiredSmartContinueLevel,
		"Умное продолжение просмотра недоступно на вашем уровне подписки.",
	)
}

func newFeatureForbidden(
	code string,
	feature string,
	currentLevel int32,
	requiredLevel int32,
	message string,
) error {
	currentLevelValue := currentLevel
	requiredLevelValue := requiredLevel

	return capabilityerr.New(capabilityerr.Detail{
		Code:          code,
		Feature:       feature,
		Message:       message,
		CurrentLevel:  &currentLevelValue,
		RequiredLevel: &requiredLevelValue,
	})
}
