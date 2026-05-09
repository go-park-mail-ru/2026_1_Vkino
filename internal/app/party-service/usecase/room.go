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

	return s.partyRepo.GetOverview(ctx, userID)
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

	if room.Visibility == "private" && !isRoomMember(room.Members, userID) {
		return domain.RoomResponse{}, domain.ErrAccessDenied
	}

	return domain.RoomResponse{Room: *room}, nil
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

	if req.RoomID <= 0 && req.InviteLink == "" {
		return domain.RoomResponse{}, domain.ErrInvalidInviteLink
	}

	if s.partyRepo == nil {
		return domain.RoomResponse{}, domain.ErrInternal
	}

	roomID := req.RoomID
	if roomID <= 0 {
		inviteCode := normalizeInviteLink(req.InviteLink)
		if inviteCode == "" {
			return domain.RoomResponse{}, domain.ErrInvalidInviteLink
		}

		invite, err := s.partyRepo.GetInvite(ctx, inviteCode)
		if err != nil {
			return domain.RoomResponse{}, err
		}

		roomID = invite.RoomID
	}

	room, err := s.partyRepo.AddMember(ctx, roomID, userID)
	if err != nil {
		return domain.RoomResponse{}, err
	}

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
