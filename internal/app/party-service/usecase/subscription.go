package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/party-service/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/common/capabilityerr"
)

const (
	roomMonthlyLimitExceededCode = "ROOM_MONTHLY_LIMIT_EXCEEDED"
	roomMembersLimitExceededCode = "ROOM_MEMBERS_LIMIT_EXCEEDED"
	watchPartyRoomsFeatureCode   = "watch_party_rooms"
	watchPartyUsersFeatureCode   = "watch_party_members"
)

func (s *service) ensureRoomCreationAllowed(ctx context.Context, userID int64) error {
	if s.subscriptionReader == nil {
		return nil
	}

	state, err := s.subscriptionReader.GetSubscriptionState(ctx, userID)
	if err != nil {
		return err
	}

	limit := state.GetCapabilities().MonthlyRoomLimit
	if limit == nil {
		return nil
	}

	if state.GetUsage().GetRoomsCreatedThisMonth() < *limit {
		return nil
	}

	return newLimitExceeded(
		roomMonthlyLimitExceededCode,
		watchPartyRoomsFeatureCode,
		*limit,
		state.GetUsage().GetRoomsCreatedThisMonth(),
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
	if activeUsers < state.GetCapabilities().GetMaxRoomMembers() {
		return nil
	}

	return newLimitExceeded(
		roomMembersLimitExceededCode,
		watchPartyUsersFeatureCode,
		state.GetCapabilities().GetMaxRoomMembers(),
		activeUsers,
		"Лимит активных участников комнаты по подписке владельца исчерпан.",
	)
}

func newLimitExceeded(
	code string,
	feature string,
	limit int32,
	used int32,
	message string,
) error {
	remaining := int32(0)
	limitValue := limit
	usedValue := used

	return capabilityerr.New(capabilityerr.Detail{
		Code:      code,
		Feature:   feature,
		Message:   message,
		Limit:     &limitValue,
		Used:      &usedValue,
		Remaining: &remaining,
	})
}
