package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/party-service/domain"
)

const (
	playbackStatusPaused = "paused"
	roomActionPlay       = "play"
	roomActionPause      = "pause"
	roomActionSeek       = "seek"
	roomActionSyncState  = "sync_state"
	memberRoleHost       = "host"
	memberRoleMember     = "member"
	memberStatusActive   = "active"
	memberStatusPending  = "pending"
)

func (s *service) ApplyRoomAction(
	ctx context.Context,
	userID int64,
	req domain.ApplyRoomActionRequest,
) (domain.PlaybackState, error) {
	action, room, err := s.prepareRoomAction(ctx, userID, req)
	if err != nil {
		return domain.PlaybackState{}, err
	}

	state := room.Playback
	now := time.Now().UTC()

	if err = applyRoomPlaybackAction(action, &state, req); err != nil {
		return domain.PlaybackState{}, err
	}

	state.UpdatedAt = now

	if err = s.partyRepo.SavePlaybackState(ctx, req.RoomID, state); err != nil {
		return domain.PlaybackState{}, err
	}

	if err = s.partyRepo.TouchRoom(ctx, req.RoomID); err != nil {
		return domain.PlaybackState{}, err
	}

	if err = s.publishRoomEvent(ctx, domain.RoomEvent{
		Type:        action,
		RoomID:      req.RoomID,
		ActorUserID: userID,
		Playback:    &state,
		SentAt:      now,
	}); err != nil {
		return domain.PlaybackState{}, err
	}

	return state, nil
}

func (s *service) prepareRoomAction(
	ctx context.Context,
	userID int64,
	req domain.ApplyRoomActionRequest,
) (string, *domain.Room, error) {
	if err := validateRoomActionRequest(userID, req.RoomID, s.partyRepo); err != nil {
		return "", nil, err
	}

	room, err := s.partyRepo.GetRoomByID(ctx, req.RoomID)
	if err != nil {
		return "", nil, err
	}

	action := normalizeRoomAction(req.Action)
	if err = ensureRoomActionAllowed(room, userID, action); err != nil {
		return "", nil, err
	}

	return action, room, nil
}

func applyRoomPlaybackAction(action string, state *domain.PlaybackState, req domain.ApplyRoomActionRequest) error {
	handler, ok := roomPlaybackActionHandlers()[action]
	if !ok {
		return domain.ErrInvalidAction
	}

	return handler(state, req)
}

func isParticipantPlaybackAction(action string) bool {
	switch action {
	case roomActionPlay, roomActionPause, roomActionSeek, roomActionSyncState:
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

func applyPositionIfProvided(state *domain.PlaybackState, positionSeconds int64) {
	if positionSeconds >= 0 {
		state.PositionSeconds = positionSeconds
	}
}

func selectMovie(state *domain.PlaybackState, movieID int64) {
	state.MovieID = movieID
	state.EpisodeID = 0
	state.PlaybackURL = ""
	state.DurationSeconds = 0
	state.PositionSeconds = 0
	state.Status = playbackStatusPaused
}

func selectEpisode(state *domain.PlaybackState, req domain.ApplyRoomActionRequest) {
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
	state.Status = playbackStatusPaused
}

func syncPlaybackState(state *domain.PlaybackState, req domain.ApplyRoomActionRequest) {
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
}

type roomPlaybackActionHandler func(state *domain.PlaybackState, req domain.ApplyRoomActionRequest) error

func roomPlaybackActionHandlers() map[string]roomPlaybackActionHandler {
	return map[string]roomPlaybackActionHandler{
		roomActionPlay:      applyPlayAction,
		roomActionPause:     applyPauseAction,
		roomActionSeek:      applySeekAction,
		"select_movie":      applySelectMovieAction,
		"select_episode":    applySelectEpisodeAction,
		roomActionSyncState: applySyncStateAction,
	}
}

func applyPlayAction(state *domain.PlaybackState, req domain.ApplyRoomActionRequest) error {
	applyPlaybackRequest(state, req)
	state.Status = "playing"
	applyPositionIfProvided(state, req.PositionSeconds)

	return nil
}

func applyPauseAction(state *domain.PlaybackState, req domain.ApplyRoomActionRequest) error {
	applyPlaybackRequest(state, req)
	state.Status = playbackStatusPaused
	applyPositionIfProvided(state, req.PositionSeconds)

	return nil
}

func applySeekAction(state *domain.PlaybackState, req domain.ApplyRoomActionRequest) error {
	if req.PositionSeconds < 0 {
		return domain.ErrInvalidPlayback
	}

	applyPlaybackRequest(state, req)
	state.PositionSeconds = req.PositionSeconds

	return nil
}

func applySelectMovieAction(state *domain.PlaybackState, req domain.ApplyRoomActionRequest) error {
	if req.MovieID <= 0 {
		return domain.ErrInvalidPlayback
	}

	selectMovie(state, req.MovieID)

	return nil
}

func applySelectEpisodeAction(state *domain.PlaybackState, req domain.ApplyRoomActionRequest) error {
	if req.EpisodeID <= 0 {
		return domain.ErrInvalidPlayback
	}

	selectEpisode(state, req)

	return nil
}

func applySyncStateAction(state *domain.PlaybackState, req domain.ApplyRoomActionRequest) error {
	syncPlaybackState(state, req)

	return nil
}

func (s *service) SendRoomMessage(
	ctx context.Context,
	userID int64,
	req domain.SendRoomMessageRequest,
) (domain.RoomMessage, error) {
	content, err := validateRoomMessageRequest(userID, req)
	if err != nil {
		return domain.RoomMessage{}, err
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

	if err = s.publishRoomEvent(ctx, domain.RoomEvent{
		Type:        "chat_message",
		RoomID:      req.RoomID,
		ActorUserID: userID,
		Message:     message,
		SentAt:      time.Now().UTC(),
	}); err != nil {
		return domain.RoomMessage{}, err
	}

	return *message, nil
}

func (s *service) CreateRoomPoll(
	ctx context.Context,
	userID int64,
	req domain.CreateRoomPollRequest,
) (domain.Poll, error) {
	question, err := validateRoomPollRequest(userID, req)
	if err != nil {
		return domain.Poll{}, err
	}

	room, err := s.getAccessibleRoom(ctx, userID, req.RoomID)
	if err != nil {
		return domain.Poll{}, err
	}

	options, err := buildPollOptions(req.Options)
	if err != nil {
		return domain.Poll{}, err
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

	if err = s.publishRoomEvent(ctx, domain.RoomEvent{
		Type:        "poll_created",
		RoomID:      req.RoomID,
		ActorUserID: userID,
		Poll:        poll,
		Member: &domain.RoomMember{
			UserID:      userID,
			DisplayName: memberDisplayName(room.Members, userID),
		},
		SentAt: time.Now().UTC(),
	}); err != nil {
		return domain.Poll{}, err
	}

	return *poll, nil
}

func (s *service) VoteRoomPoll(
	ctx context.Context,
	userID int64,
	req domain.VoteRoomPollRequest,
) (domain.PollVote, domain.Poll, error) {
	if err := validateVoteRoomPollRequest(userID, req); err != nil {
		return domain.PollVote{}, domain.Poll{}, err
	}

	if err := s.ensurePollCanBeVoted(ctx, userID, req); err != nil {
		return domain.PollVote{}, domain.Poll{}, err
	}

	vote := domain.PollVote{
		PollID:      req.PollID,
		OptionID:    req.OptionID,
		UserID:      userID,
		CoinsAmount: req.CoinsAmount,
	}

	if err := s.spendPollVoteCoins(ctx, req, vote); err != nil {
		return domain.PollVote{}, domain.Poll{}, err
	}

	updatedPoll, err := s.saveVoteAndLoadPoll(ctx, req.RoomID, req.PollID, vote)
	if err != nil {
		return domain.PollVote{}, domain.Poll{}, err
	}

	if err = s.publishVoteRoomPollEvent(ctx, req.RoomID, userID, updatedPoll, vote); err != nil {
		return domain.PollVote{}, domain.Poll{}, err
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

	if err = ensureAccessibleRoom(room, userID); err != nil {
		return nil, err
	}

	if activated, err := s.activatePendingMemberIfNeeded(ctx, roomID, userID, room.Members); err != nil {
		return nil, err
	} else if activated {
		room, err = s.partyRepo.GetRoomByID(ctx, roomID)
		if err != nil {
			return nil, err
		}
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

func (s *service) publishRoomEvent(ctx context.Context, event domain.RoomEvent) error {
	if s.eventBroker == nil {
		return nil
	}

	if err := s.eventBroker.Publish(ctx, event); err != nil {
		return fmt.Errorf("publish room event: %w", err)
	}

	return nil
}

func validateRoomActionRequest(userID, roomID int64, repo any) error {
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

func normalizeRoomAction(action string) string {
	return strings.TrimSpace(strings.ToLower(action))
}

func ensureRoomActionAllowed(room *domain.Room, userID int64, action string) error {
	if action == "" {
		return domain.ErrInvalidAction
	}

	if !isRoomMember(room.Members, userID) {
		return domain.ErrAccessDenied
	}

	if !isParticipantPlaybackAction(action) && room.HostUserID != userID {
		return domain.ErrAccessDenied
	}

	return nil
}

func validateRoomMessageRequest(userID int64, req domain.SendRoomMessageRequest) (string, error) {
	if userID <= 0 {
		return "", domain.ErrInvalidUserID
	}

	if req.RoomID <= 0 {
		return "", domain.ErrInvalidRoomID
	}

	content := strings.TrimSpace(req.Content)
	if content == "" {
		return "", domain.ErrInvalidMessage
	}

	return content, nil
}

func validateRoomPollRequest(userID int64, req domain.CreateRoomPollRequest) (string, error) {
	if userID <= 0 {
		return "", domain.ErrInvalidUserID
	}

	if req.RoomID <= 0 {
		return "", domain.ErrInvalidRoomID
	}

	question := strings.TrimSpace(req.Question)
	if question == "" {
		return "", domain.ErrInvalidPoll
	}

	return question, nil
}

func buildPollOptions(rawOptions []string) ([]domain.PollOption, error) {
	options := make([]domain.PollOption, 0, len(rawOptions))
	for _, option := range rawOptions {
		title := strings.TrimSpace(option)
		if title == "" {
			continue
		}

		options = append(options, domain.PollOption{Title: title})
	}

	if len(options) < 2 {
		return nil, domain.ErrInvalidPollOption
	}

	return options, nil
}

func ensureAccessibleRoom(room *domain.Room, userID int64) error {
	if room.Visibility == "private" && !isRoomMember(room.Members, userID) {
		return domain.ErrAccessDenied
	}

	return nil
}

func validateVoteRoomPollRequest(userID int64, req domain.VoteRoomPollRequest) error {
	if userID <= 0 {
		return domain.ErrInvalidUserID
	}

	if req.RoomID <= 0 {
		return domain.ErrInvalidRoomID
	}

	if req.PollID <= 0 || req.OptionID <= 0 || req.CoinsAmount <= 0 {
		return domain.ErrInvalidPollOption
	}

	return nil
}

func (s *service) spendPollVoteCoins(
	ctx context.Context,
	req domain.VoteRoomPollRequest,
	vote domain.PollVote,
) error {
	if s.coinsSpender == nil {
		return nil
	}

	return s.coinsSpender.SpendForPollVote(
		ctx,
		vote.UserID,
		req.RoomID,
		req.PollID,
		req.OptionID,
		req.CoinsAmount,
	)
}

func (s *service) publishVoteRoomPollEvent(
	ctx context.Context,
	roomID int64,
	userID int64,
	poll domain.Poll,
	vote domain.PollVote,
) error {
	if s.eventBroker == nil {
		return nil
	}

	pollCopy := poll
	voteCopy := vote

	return s.publishRoomEvent(ctx, domain.RoomEvent{
		Type:        "poll_voted",
		RoomID:      roomID,
		ActorUserID: userID,
		Poll:        &pollCopy,
		Vote:        &voteCopy,
		SentAt:      time.Now().UTC(),
	})
}

func (s *service) ensurePollCanBeVoted(ctx context.Context, userID int64, req domain.VoteRoomPollRequest) error {
	room, err := s.getAccessibleRoom(ctx, userID, req.RoomID)
	if err != nil {
		return err
	}

	poll, ok := findPoll(room.Polls, req.PollID)
	if !ok {
		return domain.ErrInvalidPoll
	}

	if !pollHasOption(poll, req.OptionID) {
		return domain.ErrInvalidPollOption
	}

	return nil
}

func (s *service) saveVoteAndLoadPoll(
	ctx context.Context,
	roomID int64,
	pollID int64,
	vote domain.PollVote,
) (domain.Poll, error) {
	if err := s.partyRepo.SaveVote(ctx, vote); err != nil {
		return domain.Poll{}, err
	}

	if err := s.partyRepo.TouchRoom(ctx, roomID); err != nil {
		return domain.Poll{}, err
	}

	updatedRoom, err := s.partyRepo.GetRoomByID(ctx, roomID)
	if err != nil {
		return domain.Poll{}, err
	}

	updatedPoll, ok := findPoll(updatedRoom.Polls, pollID)
	if !ok {
		return domain.Poll{}, domain.ErrInvalidPoll
	}

	return updatedPoll, nil
}
