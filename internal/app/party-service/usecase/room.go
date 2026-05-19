package usecase

import (
	"context"
	"strings"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/party-service/domain"
)

func (s *service) GetOverview(ctx context.Context, userID int64) (domain.OverviewResponse, error) {
	if userID <= 0 {
		return domain.OverviewResponse{}, domain.ErrInvalidUserID
	}

	if s.partyRepo == nil {
		return domain.OverviewResponse{}, domain.ErrInternal
	}

	overview, err := s.partyRepo.GetOverview(ctx, userID)
	if err != nil {
		return domain.OverviewResponse{}, err
	}

	maskOverviewInviteLinks(&overview, userID)

	return overview, nil
}

func (s *service) GetRoom(ctx context.Context, userID, roomID int64) (domain.RoomResponse, error) {
	if userID <= 0 {
		return domain.RoomResponse{}, domain.ErrInvalidUserID
	}

	if roomID <= 0 {
		return domain.RoomResponse{}, domain.ErrInvalidRoomID
	}

	if s.partyRepo == nil {
		return domain.RoomResponse{}, domain.ErrInternal
	}

	room, err := s.partyRepo.GetRoomByID(ctx, roomID)
	if err != nil {
		return domain.RoomResponse{}, err
	}

	if !isRoomMember(room.Members, userID) {
		return domain.RoomResponse{}, domain.ErrAccessDenied
	}

	if activated, err := s.activatePendingMemberIfNeeded(ctx, roomID, userID, room.Members); err != nil {
		return domain.RoomResponse{}, err
	} else if activated {
		room, err = s.partyRepo.GetRoomByID(ctx, roomID)
		if err != nil {
			return domain.RoomResponse{}, err
		}
	}

	maskRoomInviteLink(room, userID)

	return domain.RoomResponse{Room: *room}, nil
}

func (s *service) GetRoomInvite(ctx context.Context, userID, roomID int64) (domain.RoomInviteResponse, error) {
	if userID <= 0 {
		return domain.RoomInviteResponse{}, domain.ErrInvalidUserID
	}

	if roomID <= 0 {
		return domain.RoomInviteResponse{}, domain.ErrInvalidRoomID
	}

	if s.partyRepo == nil {
		return domain.RoomInviteResponse{}, domain.ErrInternal
	}

	room, err := s.partyRepo.GetRoomByID(ctx, roomID)
	if err != nil {
		return domain.RoomInviteResponse{}, err
	}

	if room.HostUserID != userID {
		return domain.RoomInviteResponse{}, domain.ErrAccessDenied
	}

	return domain.RoomInviteResponse{
		RoomID:     room.ID,
		InviteLink: room.InviteLink,
	}, nil
}

func (s *service) InviteFriendToRoom(
	ctx context.Context,
	userID int64,
	req domain.InviteFriendToRoomRequest,
) (domain.InviteFriendToRoomResponse, error) {
	if userID <= 0 {
		return domain.InviteFriendToRoomResponse{}, domain.ErrInvalidUserID
	}

	if req.RoomID <= 0 {
		return domain.InviteFriendToRoomResponse{}, domain.ErrInvalidRoomID
	}

	if req.InvitedUserID <= 0 || req.InvitedUserID == userID {
		return domain.InviteFriendToRoomResponse{}, domain.ErrInvalidUserID
	}

	if s.partyRepo == nil {
		return domain.InviteFriendToRoomResponse{}, domain.ErrInternal
	}

	room, err := s.partyRepo.GetRoomByID(ctx, req.RoomID)
	if err != nil {
		return domain.InviteFriendToRoomResponse{}, err
	}

	if room.HostUserID != userID {
		return domain.InviteFriendToRoomResponse{}, domain.ErrAccessDenied
	}

	if err = s.partyRepo.InviteMember(ctx, req.RoomID, req.InvitedUserID); err != nil {
		return domain.InviteFriendToRoomResponse{}, err
	}

	return domain.InviteFriendToRoomResponse{
		RoomID:        req.RoomID,
		InvitedUserID: req.InvitedUserID,
		Status:        "pending",
	}, nil
}

func (s *service) CreateRoom(
	ctx context.Context,
	userID int64,
	req domain.CreateRoomRequest,
) (domain.RoomResponse, error) {
	if userID <= 0 {
		return domain.RoomResponse{}, domain.ErrInvalidUserID
	}

	if req.Name == "" {
		return domain.RoomResponse{}, domain.ErrInvalidRoomName
	}

	if req.Visibility == "" {
		return domain.RoomResponse{}, domain.ErrInvalidVisibility
	}

	if s.partyRepo == nil {
		return domain.RoomResponse{}, domain.ErrInternal
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Visibility = strings.TrimSpace(strings.ToLower(req.Visibility))

	room, err := s.partyRepo.CreateRoom(ctx, userID, req)
	if err != nil {
		return domain.RoomResponse{}, err
	}

	return domain.RoomResponse{Room: *room}, nil
}

func (s *service) JoinRoom(
	ctx context.Context,
	userID int64,
	req domain.JoinRoomRequest,
) (domain.RoomResponse, error) {
	if userID <= 0 {
		return domain.RoomResponse{}, domain.ErrInvalidUserID
	}

	if req.InviteLink == "" {
		return domain.RoomResponse{}, domain.ErrInvalidInviteLink
	}

	if s.partyRepo == nil {
		return domain.RoomResponse{}, domain.ErrInternal
	}

	inviteCode := normalizeInviteLink(req.InviteLink)
	if inviteCode == "" {
		return domain.RoomResponse{}, domain.ErrInvalidInviteLink
	}

	invite, err := s.partyRepo.GetInvite(ctx, inviteCode)
	if err != nil {
		return domain.RoomResponse{}, err
	}

	room, err := s.partyRepo.AddMember(ctx, invite.RoomID, userID)
	if err != nil {
		return domain.RoomResponse{}, err
	}

	maskRoomInviteLink(room, userID)

	return domain.RoomResponse{Room: *room}, nil
}

func (s *service) DeleteRoom(ctx context.Context, userID, roomID int64) (domain.DeleteRoomResponse, error) {
	if userID <= 0 {
		return domain.DeleteRoomResponse{}, domain.ErrInvalidUserID
	}

	if roomID <= 0 {
		return domain.DeleteRoomResponse{}, domain.ErrInvalidRoomID
	}

	if s.partyRepo == nil {
		return domain.DeleteRoomResponse{}, domain.ErrInternal
	}

	room, err := s.partyRepo.GetRoomByID(ctx, roomID)
	if err != nil {
		return domain.DeleteRoomResponse{}, err
	}

	if room.HostUserID != userID {
		return domain.DeleteRoomResponse{}, domain.ErrAccessDenied
	}

	if err = s.partyRepo.DeleteRoom(ctx, roomID); err != nil {
		return domain.DeleteRoomResponse{}, err
	}

	return domain.DeleteRoomResponse{
		RoomID:  roomID,
		Success: true,
	}, nil
}

func (s *service) SubscribeRoom(
	ctx context.Context,
	userID int64,
	req domain.SubscribeRoomRequest,
) (<-chan domain.RoomEvent, func(), error) {
	if userID <= 0 {
		return nil, nil, domain.ErrInvalidUserID
	}

	if req.RoomID <= 0 {
		return nil, nil, domain.ErrInvalidRoomID
	}

	if s.eventBroker == nil {
		return nil, nil, domain.ErrNotImplemented
	}

	room, err := s.partyRepo.GetRoomByID(ctx, req.RoomID)
	if err != nil {
		return nil, nil, err
	}

	if !isRoomMember(room.Members, userID) {
		return nil, nil, domain.ErrAccessDenied
	}

	if _, err = s.activatePendingMemberIfNeeded(ctx, req.RoomID, userID, room.Members); err != nil {
		return nil, nil, err
	}

	return s.eventBroker.Subscribe(ctx, req.RoomID)
}

func isRoomMember(members []domain.RoomMember, userID int64) bool {
	for _, member := range members {
		if member.UserID == userID {
			return true
		}
	}

	return false
}

func findRoomMember(members []domain.RoomMember, userID int64) (domain.RoomMember, bool) {
	for _, member := range members {
		if member.UserID == userID {
			return member, true
		}
	}

	return domain.RoomMember{}, false
}

func (s *service) activatePendingMemberIfNeeded(
	ctx context.Context,
	roomID, userID int64,
	members []domain.RoomMember,
) (bool, error) {
	member, ok := findRoomMember(members, userID)
	if !ok || member.Role == "host" || member.Status != "pending" {
		return false, nil
	}

	if err := s.partyRepo.ActivateMember(ctx, roomID, userID); err != nil {
		return false, err
	}

	return true, nil
}

func normalizeInviteLink(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	value = strings.TrimRight(value, "/")
	if idx := strings.LastIndex(value, "/"); idx >= 0 {
		value = value[idx+1:]
	}

	return strings.TrimSpace(value)
}

func maskOverviewInviteLinks(overview *domain.OverviewResponse, userID int64) {
	maskRoomCardsInviteLinks(overview.ActiveRooms, userID)
	maskRoomCardsInviteLinks(overview.MyRooms, userID)
	maskRoomCardsInviteLinks(overview.FeaturedRooms, userID)
}

func maskRoomCardsInviteLinks(items []domain.RoomCard, userID int64) {
	for i := range items {
		if items[i].HostUserID != userID {
			items[i].InviteLink = ""
		}
	}
}

func maskRoomInviteLink(room *domain.Room, userID int64) {
	if room.HostUserID != userID {
		room.InviteLink = ""
	}
}
