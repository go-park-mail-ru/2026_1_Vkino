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
	if err := validateGetRoomRequest(userID, roomID, s.partyRepo); err != nil {
		return domain.RoomResponse{}, err
	}

	room, err := s.partyRepo.GetRoomByID(ctx, roomID)
	if err != nil {
		return domain.RoomResponse{}, err
	}

	room, err = s.ensureActiveRoomMember(ctx, userID, roomID, room)
	if err != nil {
		return domain.RoomResponse{}, err
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
	if err := validateInviteFriendToRoomRequest(userID, req, s.partyRepo); err != nil {
		return domain.InviteFriendToRoomResponse{}, err
	}

	room, err := s.partyRepo.GetRoomByID(ctx, req.RoomID)
	if err != nil {
		return domain.InviteFriendToRoomResponse{}, err
	}

	if err = ensureRoomHost(room, userID); err != nil {
		return domain.InviteFriendToRoomResponse{}, err
	}

	if err = s.ensureRoomMemberCapacity(ctx, room, req.InvitedUserID); err != nil {
		return domain.InviteFriendToRoomResponse{}, err
	}

	if err = s.partyRepo.InviteMember(ctx, req.RoomID, req.InvitedUserID); err != nil {
		return domain.InviteFriendToRoomResponse{}, err
	}

	return domain.InviteFriendToRoomResponse{
		RoomID:        req.RoomID,
		InvitedUserID: req.InvitedUserID,
		Status:        memberStatusPending,
	}, nil
}

func validateGetRoomRequest(userID, roomID int64, repo any) error {
	if userID <= 0 {
		return domain.ErrInvalidUserID
	}

	if roomID <= 0 {
		return domain.ErrInvalidRoomID
	}

	if repo == nil {
		return domain.ErrInternal
	}

	return nil
}

func (s *service) ensureActiveRoomMember(
	ctx context.Context,
	userID int64,
	roomID int64,
	room *domain.Room,
) (*domain.Room, error) {
	if room.Visibility != roomVisibilityPrivate && !isRoomMember(room.Members, userID) {
		if err := s.ensureRoomMemberCapacity(ctx, room, userID); err != nil {
			return nil, err
		}

		return s.partyRepo.AddMember(ctx, roomID, userID)
	}

	if !isRoomMember(room.Members, userID) {
		return nil, domain.ErrAccessDenied
	}

	activated, err := s.activatePendingMemberIfNeeded(ctx, roomID, userID, room.Members)
	if err != nil {
		return nil, err
	}

	if !activated {
		return room, nil
	}

	return s.partyRepo.GetRoomByID(ctx, roomID)
}

func validateInviteFriendToRoomRequest(userID int64, req domain.InviteFriendToRoomRequest, repo any) error {
	if userID <= 0 || req.InvitedUserID <= 0 || req.InvitedUserID == userID {
		return domain.ErrInvalidUserID
	}

	if req.RoomID <= 0 {
		return domain.ErrInvalidRoomID
	}

	if repo == nil {
		return domain.ErrInternal
	}

	return nil
}

func ensureRoomHost(room *domain.Room, userID int64) error {
	if room.HostUserID != userID {
		return domain.ErrAccessDenied
	}

	return nil
}

func (s *service) CreateRoom(
	ctx context.Context,
	userID int64,
	req domain.CreateRoomRequest,
) (domain.RoomResponse, error) {
	if userID <= 0 {
		return domain.RoomResponse{}, domain.ErrInvalidUserID
	}

	if s.partyRepo == nil {
		return domain.RoomResponse{}, domain.ErrInternal
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return domain.RoomResponse{}, domain.ErrInvalidRoomName
	}

	visibility, ok := normalizeRoomVisibility(req.Visibility)
	if !ok {
		return domain.RoomResponse{}, domain.ErrInvalidVisibility
	}

	req.Visibility = visibility

	if err := s.ensureRoomCreationAllowed(ctx, userID); err != nil {
		return domain.RoomResponse{}, err
	}

	room, err := s.partyRepo.CreateRoom(ctx, userID, req)
	if err != nil {
		return domain.RoomResponse{}, err
	}

	return domain.RoomResponse{Room: *room}, nil
}

func normalizeRoomVisibility(value string) (string, bool) {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "public", "open", "opened", "открытая", "открытый", "публичная", "публичный":
		return "public", true
	case "private", "closed", "закрытая", "закрытый", "приватная", "приватный":
		return "private", true
	default:
		return "", false
	}
}

func (s *service) JoinRoom(
	ctx context.Context,
	userID int64,
	req domain.JoinRoomRequest,
) (domain.RoomResponse, error) {
	if err := validateJoinRoomRequest(userID, req, s.partyRepo); err != nil {
		return domain.RoomResponse{}, err
	}

	invite, room, err := s.joinRoomTarget(ctx, req.InviteLink)
	if err != nil {
		return domain.RoomResponse{}, err
	}

	if err = s.ensureRoomMemberCapacity(ctx, room, userID); err != nil {
		return domain.RoomResponse{}, err
	}

	room, err = s.partyRepo.AddMember(ctx, invite.RoomID, userID)
	if err != nil {
		return domain.RoomResponse{}, err
	}

	maskRoomInviteLink(room, userID)

	return domain.RoomResponse{Room: *room}, nil
}

func validateJoinRoomRequest(userID int64, req domain.JoinRoomRequest, repo any) error {
	if userID <= 0 {
		return domain.ErrInvalidUserID
	}

	if req.InviteLink == "" {
		return domain.ErrInvalidInviteLink
	}

	if repo == nil {
		return domain.ErrInternal
	}

	return nil
}

func (s *service) joinRoomTarget(
	ctx context.Context,
	inviteLink string,
) (*domain.Invite, *domain.Room, error) {
	inviteCode := normalizeInviteLink(inviteLink)
	if inviteCode == "" {
		return nil, nil, domain.ErrInvalidInviteLink
	}

	invite, err := s.partyRepo.GetInvite(ctx, inviteCode)
	if err != nil {
		return nil, nil, err
	}

	room, err := s.partyRepo.GetRoomByID(ctx, invite.RoomID)
	if err != nil {
		return nil, nil, err
	}

	return invite, room, nil
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
	if err := validateSubscribeRoomRequest(userID, req, s.eventBroker); err != nil {
		return nil, nil, err
	}

	room, err := s.partyRepo.GetRoomByID(ctx, req.RoomID)
	if err != nil {
		return nil, nil, err
	}

	if err := s.ensureRoomSubscriptionAllowed(ctx, room, userID); err != nil {
		return nil, nil, err
	}

	return s.eventBroker.Subscribe(ctx, req.RoomID, userID)
}

func validateSubscribeRoomRequest(userID int64, req domain.SubscribeRoomRequest, broker any) error {
	if userID <= 0 {
		return domain.ErrInvalidUserID
	}

	if req.RoomID <= 0 {
		return domain.ErrInvalidRoomID
	}

	if broker == nil {
		return domain.ErrNotImplemented
	}

	return nil
}

func (s *service) ensureRoomSubscriptionAllowed(ctx context.Context, room *domain.Room, userID int64) error {
	if !isRoomMember(room.Members, userID) {
		return domain.ErrAccessDenied
	}

	if _, err := s.activatePendingMemberIfNeeded(ctx, room.ID, userID, room.Members); err != nil {
		return err
	}

	return s.ensureRoomMemberCapacity(ctx, room, userID)
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
	if !ok || member.Role == memberRoleHost || member.Status != memberStatusPending {
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
