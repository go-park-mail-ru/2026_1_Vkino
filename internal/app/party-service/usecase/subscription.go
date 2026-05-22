package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/party-service/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/subscription"
)

func (s *service) ensureRoomCreationAllowed(ctx context.Context, userID int64) error {
	if s.subscriptionReader == nil {
		return nil
	}

	state, err := s.subscriptionReader.GetSubscriptionState(ctx, userID)
	if err != nil {
		return err
	}

	limit := state.Capabilities.MonthlyRoomLimit
	if limit == nil {
		return nil
	}

	if state.Usage.RoomsCreatedThisMonth < *limit {
		return nil
	}

	return subscription.NewLimitExceeded(
		subscription.CodeRoomMonthlyLimitExceeded,
		subscription.FeatureWatchPartyRooms,
		*limit,
		state.Usage.RoomsCreatedThisMonth,
		"Лимит комнат совместного просмотра за месяц исчерпан.",
	)
}

func (s *service) ensureRoomMemberCapacity(ctx context.Context, room *domain.Room, userID int64) error {
	if room == nil || s.subscriptionReader == nil || s.eventBroker == nil {
		return nil
	}

	if s.eventBroker.IsUserActive(room.ID, userID) {
		return nil
	}

	state, err := s.subscriptionReader.GetSubscriptionState(ctx, room.HostUserID)
	if err != nil {
		return err
	}

	activeUsers := s.eventBroker.ActiveUsers(room.ID)
	if activeUsers < state.Capabilities.MaxRoomMembers {
		return nil
	}

	return subscription.NewLimitExceeded(
		subscription.CodeRoomMembersLimitExceeded,
		subscription.FeatureWatchPartyUsers,
		state.Capabilities.MaxRoomMembers,
		activeUsers,
		"Лимит активных участников комнаты по подписке владельца исчерпан.",
	)
}
