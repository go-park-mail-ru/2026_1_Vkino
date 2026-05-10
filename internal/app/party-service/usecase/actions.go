package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/party-service/domain"
)

func (s *service) ApplyRoomAction(
	ctx context.Context,
	userID int64,
	req domain.ApplyRoomActionRequest,
) (domain.PlaybackState, error) {
	if userID <= 0 {
		return domain.PlaybackState{}, domain.ErrInvalidUserID
	}

	if req.RoomID <= 0 {
		return domain.PlaybackState{}, domain.ErrInvalidRoomID
	}

	if s.partyRepo == nil {
		return domain.PlaybackState{}, domain.ErrInternal
	}

	room, err := s.partyRepo.GetRoomByID(ctx, req.RoomID)
	if err != nil {
		return domain.PlaybackState{}, err
	}

	action := strings.TrimSpace(strings.ToLower(req.Action))
	if action == "" {
		return domain.PlaybackState{}, domain.ErrInvalidAction
	}

	if !isRoomMember(room.Members, userID) {
		return domain.PlaybackState{}, domain.ErrAccessDenied
	}

	if !isParticipantPlaybackAction(action) && room.HostUserID != userID {
		return domain.PlaybackState{}, domain.ErrAccessDenied
	}

	state := room.Playback
	now := time.Now().UTC()

	switch action {
	case "play":
		applyPlaybackRequest(&state, req)
		state.Status = "playing"
		if req.PositionSeconds >= 0 {
			state.PositionSeconds = req.PositionSeconds
		}
	case "pause":
		applyPlaybackRequest(&state, req)
		state.Status = "paused"
		if req.PositionSeconds >= 0 {
			state.PositionSeconds = req.PositionSeconds
		}
	case "seek":
		if req.PositionSeconds < 0 {
			return domain.PlaybackState{}, domain.ErrInvalidPlayback
		}
		applyPlaybackRequest(&state, req)
		state.PositionSeconds = req.PositionSeconds
	case "select_movie":
		if req.MovieID <= 0 {
			return domain.PlaybackState{}, domain.ErrInvalidPlayback
		}
		state.MovieID = req.MovieID
		state.EpisodeID = 0
		state.PlaybackURL = ""
		state.DurationSeconds = 0
		state.PositionSeconds = 0
		state.Status = "paused"
	case "select_episode":
		if req.EpisodeID <= 0 {
			return domain.PlaybackState{}, domain.ErrInvalidPlayback
		}
		state.EpisodeID = req.EpisodeID
		if req.MovieID > 0 {
			state.MovieID = req.MovieID
		}
		if strings.TrimSpace(req.PlaybackURL) != "" {
			state.PlaybackURL = strings.TrimSpace(req.PlaybackURL)
		}
		if req.DurationSeconds >= 0 {
			state.DurationSeconds = req.DurationSeconds
		}
		state.PositionSeconds = 0
		state.Status = "paused"
	case "sync_state":
		if req.MovieID > 0 {
			state.MovieID = req.MovieID
		}
		if req.EpisodeID > 0 {
			state.EpisodeID = req.EpisodeID
		}
		if strings.TrimSpace(req.PlaybackURL) != "" {
			state.PlaybackURL = strings.TrimSpace(req.PlaybackURL)
		}
		if req.DurationSeconds >= 0 {
			state.DurationSeconds = req.DurationSeconds
		}
		if req.PositionSeconds >= 0 {
			state.PositionSeconds = req.PositionSeconds
		}
		if strings.TrimSpace(req.Status) != "" {
			state.Status = strings.TrimSpace(strings.ToLower(req.Status))
		}
	default:
		return domain.PlaybackState{}, domain.ErrInvalidAction
	}

	state.UpdatedAt = now

	if err = s.partyRepo.SavePlaybackState(ctx, req.RoomID, state); err != nil {
		return domain.PlaybackState{}, err
	}

	if err = s.partyRepo.TouchRoom(ctx, req.RoomID); err != nil {
		return domain.PlaybackState{}, err
	}

	if s.eventBroker != nil {
		_ = s.eventBroker.Publish(ctx, domain.RoomEvent{
			Type:        action,
			RoomID:      req.RoomID,
			ActorUserID: userID,
			Playback:    &state,
			SentAt:      now,
		})
	}

	return state, nil
}

func isParticipantPlaybackAction(action string) bool {
	switch action {
	case "play", "pause", "seek", "sync_state":
		return true
	default:
		return false
	}
}

func applyPlaybackRequest(state *domain.PlaybackState, req domain.ApplyRoomActionRequest) {
	if req.MovieID > 0 {
		state.MovieID = req.MovieID
	}
	if req.EpisodeID > 0 {
		state.EpisodeID = req.EpisodeID
	}
	if strings.TrimSpace(req.PlaybackURL) != "" {
		state.PlaybackURL = strings.TrimSpace(req.PlaybackURL)
	}
	if req.DurationSeconds > 0 {
		state.DurationSeconds = req.DurationSeconds
	}
}

func (s *service) SendRoomMessage(
	ctx context.Context,
	userID int64,
	req domain.SendRoomMessageRequest,
) (domain.RoomMessage, error) {
	if userID <= 0 {
		return domain.RoomMessage{}, domain.ErrInvalidUserID
	}

	if req.RoomID <= 0 {
		return domain.RoomMessage{}, domain.ErrInvalidRoomID
	}

	content := strings.TrimSpace(req.Content)
	if content == "" {
		return domain.RoomMessage{}, domain.ErrInvalidMessage
	}

	room, err := s.getAccessibleRoom(ctx, userID, req.RoomID)
	if err != nil {
		return domain.RoomMessage{}, err
	}

	message, err := s.partyRepo.SaveMessage(ctx, domain.RoomMessage{
		RoomID:       req.RoomID,
		AuthorUserID: userID,
		AuthorName:   memberDisplayName(room.Members, userID),
		Content:      content,
	})
	if err != nil {
		return domain.RoomMessage{}, err
	}

	if err = s.partyRepo.TouchRoom(ctx, req.RoomID); err != nil {
		return domain.RoomMessage{}, err
	}

	if s.eventBroker != nil {
		_ = s.eventBroker.Publish(ctx, domain.RoomEvent{
			Type:        "chat_message",
			RoomID:      req.RoomID,
			ActorUserID: userID,
			Message:     message,
			SentAt:      time.Now().UTC(),
		})
	}

	return *message, nil
}

func (s *service) CreateRoomPoll(
	ctx context.Context,
	userID int64,
	req domain.CreateRoomPollRequest,
) (domain.Poll, error) {
	if userID <= 0 {
		return domain.Poll{}, domain.ErrInvalidUserID
	}

	if req.RoomID <= 0 {
		return domain.Poll{}, domain.ErrInvalidRoomID
	}

	question := strings.TrimSpace(req.Question)
	if question == "" {
		return domain.Poll{}, domain.ErrInvalidPoll
	}

	room, err := s.getAccessibleRoom(ctx, userID, req.RoomID)
	if err != nil {
		return domain.Poll{}, err
	}

	options := make([]domain.PollOption, 0, len(req.Options))
	for _, option := range req.Options {
		title := strings.TrimSpace(option)
		if title == "" {
			continue
		}

		options = append(options, domain.PollOption{Title: title})
	}

	if len(options) < 2 {
		return domain.Poll{}, domain.ErrInvalidPollOption
	}

	poll, err := s.partyRepo.SavePoll(ctx, domain.Poll{
		RoomID:          req.RoomID,
		Question:        question,
		Options:         options,
		CreatedByUserID: userID,
	})
	if err != nil {
		return domain.Poll{}, err
	}

	if err = s.partyRepo.TouchRoom(ctx, req.RoomID); err != nil {
		return domain.Poll{}, err
	}

	if s.eventBroker != nil {
		_ = s.eventBroker.Publish(ctx, domain.RoomEvent{
			Type:        "poll_created",
			RoomID:      req.RoomID,
			ActorUserID: userID,
			Poll:        poll,
			Member: &domain.RoomMember{
				UserID:      userID,
				DisplayName: memberDisplayName(room.Members, userID),
			},
			SentAt: time.Now().UTC(),
		})
	}

	return *poll, nil
}

func (s *service) VoteRoomPoll(
	ctx context.Context,
	userID int64,
	req domain.VoteRoomPollRequest,
) (domain.PollVote, domain.Poll, error) {
	if userID <= 0 {
		return domain.PollVote{}, domain.Poll{}, domain.ErrInvalidUserID
	}

	if req.RoomID <= 0 {
		return domain.PollVote{}, domain.Poll{}, domain.ErrInvalidRoomID
	}

	if req.PollID <= 0 || req.OptionID <= 0 {
		return domain.PollVote{}, domain.Poll{}, domain.ErrInvalidPollOption
	}

	room, err := s.getAccessibleRoom(ctx, userID, req.RoomID)
	if err != nil {
		return domain.PollVote{}, domain.Poll{}, err
	}

	poll, ok := findPoll(room.Polls, req.PollID)
	if !ok {
		return domain.PollVote{}, domain.Poll{}, domain.ErrInvalidPoll
	}

	if !pollHasOption(poll, req.OptionID) {
		return domain.PollVote{}, domain.Poll{}, domain.ErrInvalidPollOption
	}

	vote := domain.PollVote{
		PollID:   req.PollID,
		OptionID: req.OptionID,
		UserID:   userID,
	}

	if err = s.partyRepo.SaveVote(ctx, vote); err != nil {
		return domain.PollVote{}, domain.Poll{}, err
	}

	if err = s.partyRepo.TouchRoom(ctx, req.RoomID); err != nil {
		return domain.PollVote{}, domain.Poll{}, err
	}

	updatedRoom, err := s.partyRepo.GetRoomByID(ctx, req.RoomID)
	if err != nil {
		return domain.PollVote{}, domain.Poll{}, err
	}

	updatedPoll, ok := findPoll(updatedRoom.Polls, req.PollID)
	if !ok {
		return domain.PollVote{}, domain.Poll{}, domain.ErrInvalidPoll
	}

	if s.eventBroker != nil {
		pollCopy := updatedPoll
		voteCopy := vote
		_ = s.eventBroker.Publish(ctx, domain.RoomEvent{
			Type:        "poll_voted",
			RoomID:      req.RoomID,
			ActorUserID: userID,
			Poll:        &pollCopy,
			Vote:        &voteCopy,
			SentAt:      time.Now().UTC(),
		})
	}

	return vote, updatedPoll, nil
}

func (s *service) getAccessibleRoom(ctx context.Context, userID, roomID int64) (*domain.Room, error) {
	if s.partyRepo == nil {
		return nil, domain.ErrInternal
	}

	room, err := s.partyRepo.GetRoomByID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	if room.Visibility == "private" && !isRoomMember(room.Members, userID) {
		return nil, domain.ErrAccessDenied
	}

	return room, nil
}

func memberDisplayName(members []domain.RoomMember, userID int64) string {
	for _, member := range members {
		if member.UserID == userID {
			return member.DisplayName
		}
	}

	return ""
}

func findPoll(items []domain.Poll, pollID int64) (domain.Poll, bool) {
	for _, item := range items {
		if item.ID == pollID {
			return item, true
		}
	}

	return domain.Poll{}, false
}

func pollHasOption(poll domain.Poll, optionID int64) bool {
	for _, option := range poll.Options {
		if option.ID == optionID {
			return true
		}
	}

	return false
}
