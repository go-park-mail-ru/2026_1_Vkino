//nolint:gocyclo // Handler flow stays explicit and close to proto contracts.
package grpc

import (
	"context"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/party-service/domain"
	partyv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/party/v1"
	"google.golang.org/grpc"
)

func (s *Server) GetOverview(
	ctx context.Context,
	_ *partyv1.GetOverviewRequest,
) (*partyv1.GetOverviewResponse, error) {
	authCtx, err := s.authorize(ctx)
	if err != nil {
		return nil, err
	}

	overview, err := s.usecase.GetOverview(ctx, authCtx.UserID)
	if err != nil {
		return nil, mapError(err)
	}

	return &partyv1.GetOverviewResponse{
		ActiveRooms:   mapRoomCards(overview.ActiveRooms),
		MyRooms:       mapRoomCards(overview.MyRooms),
		FeaturedRooms: mapRoomCards(overview.FeaturedRooms),
	}, nil
}

func (s *Server) GetRoom(ctx context.Context, req *partyv1.GetRoomRequest) (*partyv1.GetRoomResponse, error) {
	authCtx, err := s.authorize(ctx)
	if err != nil {
		return nil, err
	}

	room, err := s.usecase.GetRoom(ctx, authCtx.UserID, req.GetRoomId())
	if err != nil {
		return nil, mapError(err)
	}

	return &partyv1.GetRoomResponse{Room: toProtoRoom(room.Room)}, nil
}

func (s *Server) GetRoomInvite(
	ctx context.Context,
	req *partyv1.GetRoomInviteRequest,
) (*partyv1.GetRoomInviteResponse, error) {
	authCtx, err := s.authorize(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := s.usecase.GetRoomInvite(ctx, authCtx.UserID, req.GetRoomId())
	if err != nil {
		return nil, mapError(err)
	}

	return &partyv1.GetRoomInviteResponse{
		RoomId:     resp.RoomID,
		InviteLink: resp.InviteLink,
	}, nil
}

func (s *Server) InviteFriendToRoom(
	ctx context.Context,
	req *partyv1.InviteFriendToRoomRequest,
) (*partyv1.InviteFriendToRoomResponse, error) {
	authCtx, err := s.authorize(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := s.usecase.InviteFriendToRoom(ctx, authCtx.UserID, domain.InviteFriendToRoomRequest{
		RoomID:        req.GetRoomId(),
		InvitedUserID: req.GetInvitedUserId(),
	})
	if err != nil {
		return nil, mapError(err)
	}

	return &partyv1.InviteFriendToRoomResponse{
		RoomId:        resp.RoomID,
		InvitedUserId: resp.InvitedUserID,
		Status:        resp.Status,
	}, nil
}

func (s *Server) CreateRoom(
	ctx context.Context,
	req *partyv1.CreateRoomRequest,
) (*partyv1.CreateRoomResponse, error) {
	authCtx, err := s.authorize(ctx)
	if err != nil {
		return nil, err
	}

	room, err := s.usecase.CreateRoom(ctx, authCtx.UserID, domain.CreateRoomRequest{
		Name:       req.GetName(),
		Visibility: req.GetVisibility(),
		MovieID:    req.GetMovieId(),
		EpisodeID:  req.GetEpisodeId(),
	})
	if err != nil {
		return nil, mapError(err)
	}

	return &partyv1.CreateRoomResponse{Room: toProtoRoom(room.Room)}, nil
}

func (s *Server) JoinRoom(
	ctx context.Context,
	req *partyv1.JoinRoomRequest,
) (*partyv1.JoinRoomResponse, error) {
	authCtx, err := s.authorize(ctx)
	if err != nil {
		return nil, err
	}

	room, err := s.usecase.JoinRoom(ctx, authCtx.UserID, domain.JoinRoomRequest{
		InviteLink: req.GetInviteLink(),
	})
	if err != nil {
		return nil, mapError(err)
	}

	return &partyv1.JoinRoomResponse{Room: toProtoRoom(room.Room)}, nil
}

func (s *Server) DeleteRoom(
	ctx context.Context,
	req *partyv1.DeleteRoomRequest,
) (*partyv1.DeleteRoomResponse, error) {
	authCtx, err := s.authorize(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := s.usecase.DeleteRoom(ctx, authCtx.UserID, req.GetRoomId())
	if err != nil {
		return nil, mapError(err)
	}

	return &partyv1.DeleteRoomResponse{
		RoomId:  resp.RoomID,
		Success: resp.Success,
	}, nil
}

func (s *Server) ApplyRoomAction(
	ctx context.Context,
	req *partyv1.ApplyRoomActionRequest,
) (*partyv1.ApplyRoomActionResponse, error) {
	authCtx, err := s.authorize(ctx)
	if err != nil {
		return nil, err
	}

	playback, err := s.usecase.ApplyRoomAction(ctx, authCtx.UserID, domain.ApplyRoomActionRequest{
		RoomID:          req.GetRoomId(),
		Action:          req.GetAction(),
		MovieID:         req.GetMovieId(),
		EpisodeID:       req.GetEpisodeId(),
		PlaybackURL:     req.GetPlaybackUrl(),
		DurationSeconds: req.GetDurationSeconds(),
		PositionSeconds: req.GetPositionSeconds(),
		Status:          req.GetStatus(),
	})
	if err != nil {
		return nil, mapError(err)
	}

	return &partyv1.ApplyRoomActionResponse{
		Playback: toProtoPlaybackState(playback),
	}, nil
}

func (s *Server) SendRoomMessage(
	ctx context.Context,
	req *partyv1.SendRoomMessageRequest,
) (*partyv1.SendRoomMessageResponse, error) {
	authCtx, err := s.authorize(ctx)
	if err != nil {
		return nil, err
	}

	message, err := s.usecase.SendRoomMessage(ctx, authCtx.UserID, domain.SendRoomMessageRequest{
		RoomID:  req.GetRoomId(),
		Content: req.GetContent(),
	})
	if err != nil {
		return nil, mapError(err)
	}

	return &partyv1.SendRoomMessageResponse{
		Message: toProtoRoomMessage(message),
	}, nil
}

func (s *Server) CreateRoomPoll(
	ctx context.Context,
	req *partyv1.CreateRoomPollRequest,
) (*partyv1.CreateRoomPollResponse, error) {
	authCtx, err := s.authorize(ctx)
	if err != nil {
		return nil, err
	}

	poll, err := s.usecase.CreateRoomPoll(ctx, authCtx.UserID, domain.CreateRoomPollRequest{
		RoomID:   req.GetRoomId(),
		Question: req.GetQuestion(),
		Options:  req.GetOptions(),
	})
	if err != nil {
		return nil, mapError(err)
	}

	return &partyv1.CreateRoomPollResponse{
		Poll: toProtoPoll(poll),
	}, nil
}

func (s *Server) VoteRoomPoll(
	ctx context.Context,
	req *partyv1.VoteRoomPollRequest,
) (*partyv1.VoteRoomPollResponse, error) {
	authCtx, err := s.authorize(ctx)
	if err != nil {
		return nil, err
	}

	vote, poll, err := s.usecase.VoteRoomPoll(ctx, authCtx.UserID, domain.VoteRoomPollRequest{
		RoomID:   req.GetRoomId(),
		PollID:   req.GetPollId(),
		OptionID: req.GetOptionId(),
	})
	if err != nil {
		return nil, mapError(err)
	}

	return &partyv1.VoteRoomPollResponse{
		Vote: &partyv1.PollVote{
			PollId:   vote.PollID,
			OptionId: vote.OptionID,
			UserId:   vote.UserID,
		},
		Poll: toProtoPoll(poll),
	}, nil
}

func (s *Server) SubscribeRoom(
	req *partyv1.SubscribeRoomRequest,
	stream grpc.ServerStreamingServer[partyv1.RoomEvent],
) error {
	authCtx, err := s.authorize(stream.Context())
	if err != nil {
		return err
	}

	events, unsubscribe, err := s.usecase.SubscribeRoom(
		stream.Context(),
		authCtx.UserID,
		domain.SubscribeRoomRequest{RoomID: req.GetRoomId()},
	)
	if err != nil {
		return mapError(err)
	}
	defer unsubscribe()

	for {
		select {
		case event, ok := <-events:
			if !ok {
				return nil
			}

			if err := stream.Send(toProtoRoomEvent(event)); err != nil {
				return err
			}
		case <-stream.Context().Done():
			return nil
		}
	}
}

func mapRoomCards(items []domain.RoomCard) []*partyv1.RoomCard {
	result := make([]*partyv1.RoomCard, 0, len(items))
	for _, item := range items {
		result = append(result, toProtoRoomCard(item))
	}

	return result
}

func mapRoomMembers(items []domain.RoomMember) []*partyv1.RoomMember {
	result := make([]*partyv1.RoomMember, 0, len(items))
	for _, item := range items {
		result = append(result, toProtoRoomMember(item))
	}

	return result
}

func mapRoomMessages(items []domain.RoomMessage) []*partyv1.RoomMessage {
	result := make([]*partyv1.RoomMessage, 0, len(items))
	for _, item := range items {
		result = append(result, toProtoRoomMessage(item))
	}

	return result
}

func mapPolls(items []domain.Poll) []*partyv1.Poll {
	result := make([]*partyv1.Poll, 0, len(items))
	for _, item := range items {
		result = append(result, toProtoPoll(item))
	}

	return result
}

func toProtoRoomCard(item domain.RoomCard) *partyv1.RoomCard {
	return &partyv1.RoomCard{
		Id:                item.ID,
		Name:              item.Name,
		Visibility:        item.Visibility,
		InviteLink:        item.InviteLink,
		HostUserId:        item.HostUserID,
		HostName:          item.HostName,
		ParticipantsCount: item.ParticipantsCount,
		Playback:          toProtoPlaybackState(item.Playback),
		UpdatedAt:         formatTime(item.UpdatedAt),
	}
}

func toProtoRoom(item domain.Room) *partyv1.Room {
	return &partyv1.Room{
		Id:         item.ID,
		Name:       item.Name,
		Visibility: item.Visibility,
		HostUserId: item.HostUserID,
		InviteLink: item.InviteLink,
		Members:    mapRoomMembers(item.Members),
		Playback:   toProtoPlaybackState(item.Playback),
		Messages:   mapRoomMessages(item.Messages),
		Polls:      mapPolls(item.Polls),
		UpdatedAt:  formatTime(item.UpdatedAt),
	}
}

func toProtoRoomMember(item domain.RoomMember) *partyv1.RoomMember {
	return &partyv1.RoomMember{
		UserId:      item.UserID,
		DisplayName: item.DisplayName,
		AvatarUrl:   item.AvatarURL,
		Role:        item.Role,
		JoinedAt:    formatTime(item.JoinedAt),
		Status:      item.Status,
	}
}

func toProtoPlaybackState(item domain.PlaybackState) *partyv1.PlaybackState {
	return &partyv1.PlaybackState{
		MovieId:         item.MovieID,
		EpisodeId:       item.EpisodeID,
		PlaybackUrl:     item.PlaybackURL,
		DurationSeconds: item.DurationSeconds,
		PositionSeconds: item.PositionSeconds,
		Status:          item.Status,
		UpdatedAt:       formatTime(item.UpdatedAt),
	}
}

func toProtoRoomMessage(item domain.RoomMessage) *partyv1.RoomMessage {
	return &partyv1.RoomMessage{
		Id:           item.ID,
		RoomId:       item.RoomID,
		AuthorUserId: item.AuthorUserID,
		AuthorName:   item.AuthorName,
		Content:      item.Content,
		CreatedAt:    formatTime(item.CreatedAt),
	}
}

func toProtoPoll(item domain.Poll) *partyv1.Poll {
	result := &partyv1.Poll{
		Id:              item.ID,
		RoomId:          item.RoomID,
		Question:        item.Question,
		CreatedByUserId: item.CreatedByUserID,
		CreatedAt:       formatTime(item.CreatedAt),
		ClosedAt:        formatTimePtr(item.ClosedAt),
		Options:         make([]*partyv1.PollOption, 0, len(item.Options)),
	}

	for _, option := range item.Options {
		result.Options = append(result.Options, &partyv1.PollOption{
			Id:         option.ID,
			Title:      option.Title,
			VotesCount: option.VotesCount,
		})
	}

	return result
}

func toProtoRoomEvent(item domain.RoomEvent) *partyv1.RoomEvent {
	result := &partyv1.RoomEvent{
		Type:        item.Type,
		RoomId:      item.RoomID,
		ActorUserId: item.ActorUserID,
		SentAt:      formatTime(item.SentAt),
	}

	if item.Playback != nil {
		result.Playback = toProtoPlaybackState(*item.Playback)
	}

	if item.Message != nil {
		result.Message = toProtoRoomMessage(*item.Message)
	}

	if item.Poll != nil {
		result.Poll = toProtoPoll(*item.Poll)
	}
	if item.Member != nil {
		result.Member = toProtoRoomMember(*item.Member)
	}

	if item.Vote != nil {
		result.Vote = &partyv1.PollVote{
			PollId:   item.Vote.PollID,
			OptionId: item.Vote.OptionID,
			UserId:   item.Vote.UserID,
		}
	}

	return result
}

func formatTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}

	return value.UTC().Format(time.RFC3339)
}

func formatTimePtr(value *time.Time) string {
	if value == nil {
		return ""
	}

	return formatTime(*value)
}
