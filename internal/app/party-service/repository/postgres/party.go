package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/party-service/domain"
	corepostgres "github.com/go-park-mail-ru/2026_1_VKino/pkg/postgresx"
)

type PartyRepo struct {
	db *corepostgres.Client
}

func NewPartyRepo(db *corepostgres.Client) *PartyRepo {
	return &PartyRepo{db: db}
}

func (r *PartyRepo) GetOverview(ctx context.Context, userID int64) (domain.OverviewResponse, error) {
	activeRooms, err := r.loadRoomCards(ctx, sqlGetOverviewActiveRooms)
	if err != nil {
		return domain.OverviewResponse{}, err
	}

	myRooms, err := r.loadRoomCards(ctx, sqlGetOverviewMyRooms, userID)
	if err != nil {
		return domain.OverviewResponse{}, err
	}

	featuredRooms, err := r.loadRoomCards(ctx, sqlGetOverviewFeaturedRooms)
	if err != nil {
		return domain.OverviewResponse{}, err
	}

	return domain.OverviewResponse{
		ActiveRooms:   activeRooms,
		MyRooms:       myRooms,
		FeaturedRooms: featuredRooms,
	}, nil
}

func (r *PartyRepo) GetRoomByID(ctx context.Context, roomID int64) (*domain.Room, error) {
	var (
		room       domain.Room
		inviteCode sql.NullString
		updatedAt  time.Time
	)

	err := r.db.QueryRow(ctx, sqlGetRoomBaseByID, roomID).Scan(
		&room.ID,
		&room.Name,
		&room.Visibility,
		&room.HostUserID,
		&inviteCode,
		&updatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrRoomNotFound
		}

		return nil, fmt.Errorf("get room by id: %w", err)
	}

	if inviteCode.Valid {
		room.InviteLink = inviteCode.String
	}

	room.UpdatedAt = updatedAt

	if err = r.populateRoomDetails(ctx, &room, roomID); err != nil {
		return nil, err
	}

	return &room, nil
}

func (r *PartyRepo) CreateRoom(
	ctx context.Context,
	hostUserID int64,
	req domain.CreateRoomRequest,
) (*domain.Room, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin create room tx: %w", err)
	}

	defer func() {
		ignoreRollbackError(tx.Rollback(ctx))
	}()

	var (
		roomID    int64
		updatedAt time.Time
	)

	err = tx.QueryRow(ctx, sqlCreateRoom, hostUserID, req.Name, req.Visibility).Scan(&roomID, &updatedAt)
	if err != nil {
		return nil, fmt.Errorf("create room: %w", err)
	}

	if _, err = tx.Exec(ctx, sqlCreateRoomMember, roomID, hostUserID, "host", "active"); err != nil {
		return nil, fmt.Errorf("create host room member: %w", err)
	}

	inviteCode := uuid.NewString()
	if err = tx.QueryRow(ctx, sqlCreateRoomInvite, roomID, hostUserID, inviteCode).Scan(&inviteCode); err != nil {
		return nil, fmt.Errorf("create room invite: %w", err)
	}

	if _, err = tx.Exec(ctx, sqlCreateRoomPlaybackState, roomID, req.MovieID, req.EpisodeID); err != nil {
		return nil, fmt.Errorf("create room playback state: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit create room tx: %w", err)
	}

	room := &domain.Room{
		ID:         roomID,
		Name:       req.Name,
		Visibility: req.Visibility,
		HostUserID: hostUserID,
		InviteLink: inviteCode,
		Members: []domain.RoomMember{
			{
				UserID:      hostUserID,
				DisplayName: "",
				AvatarURL:   "",
				Role:        "host",
				Status:      "active",
				JoinedAt:    updatedAt,
			},
		},
		Playback: domain.PlaybackState{
			MovieID:     req.MovieID,
			EpisodeID:   req.EpisodeID,
			Status:      "paused",
			UpdatedAt:   updatedAt,
			PlaybackURL: "",
		},
		Messages:  []domain.RoomMessage{},
		Polls:     []domain.Poll{},
		UpdatedAt: updatedAt,
	}

	return room, nil
}

func (r *PartyRepo) InviteMember(ctx context.Context, roomID, userID int64) error {
	if _, err := r.db.Exec(ctx, sqlInviteRoomMember, roomID, userID); err != nil {
		return fmt.Errorf("invite room member: %w", err)
	}

	return nil
}

func (r *PartyRepo) AddMember(ctx context.Context, roomID, userID int64) (*domain.Room, error) {
	tag, err := r.db.Exec(ctx, sqlAddRoomMember, roomID, userID)
	if err != nil {
		return nil, fmt.Errorf("add room member: %w", err)
	}

	if tag.RowsAffected() == 0 {
		room, getErr := r.GetRoomByID(ctx, roomID)
		if getErr != nil {
			return nil, getErr
		}

		return room, nil
	}

	return r.GetRoomByID(ctx, roomID)
}

func (r *PartyRepo) ActivateMember(ctx context.Context, roomID, userID int64) error {
	if _, err := r.db.Exec(ctx, sqlActivateRoomMember, roomID, userID); err != nil {
		return fmt.Errorf("activate room member: %w", err)
	}

	return nil
}

func (r *PartyRepo) DeleteRoom(ctx context.Context, roomID int64) error {
	tag, err := r.db.Exec(ctx, sqlDeleteRoom, roomID)
	if err != nil {
		return fmt.Errorf("delete room: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrRoomNotFound
	}

	return nil
}

func (r *PartyRepo) GetInvite(ctx context.Context, inviteLink string) (*domain.Invite, error) {
	var (
		invite    domain.Invite
		expiresAt sql.NullTime
	)

	err := r.db.QueryRow(ctx, sqlGetInviteByCode, inviteLink).Scan(
		&invite.RoomID,
		&invite.Link,
		&invite.CreatedBy,
		&invite.CreatedAt,
		&expiresAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrInviteNotFound
		}

		return nil, fmt.Errorf("get invite by code: %w", err)
	}

	if expiresAt.Valid {
		invite.ExpiresAt = &expiresAt.Time
	}

	return &invite, nil
}

func (r *PartyRepo) SavePlaybackState(ctx context.Context, roomID int64, state domain.PlaybackState) error {
	_, err := r.db.Exec(
		ctx,
		sqlUpsertPlaybackState,
		roomID,
		state.MovieID,
		state.EpisodeID,
		state.PlaybackURL,
		state.DurationSeconds,
		state.PositionSeconds,
		state.Status,
	)
	if err != nil {
		return fmt.Errorf("save playback state: %w", err)
	}

	return nil
}

func (r *PartyRepo) TouchRoom(ctx context.Context, roomID int64) error {
	_, err := r.db.Exec(ctx, sqlTouchRoom, roomID)
	if err != nil {
		return fmt.Errorf("touch room: %w", err)
	}

	return nil
}

func (r *PartyRepo) SaveMessage(ctx context.Context, message domain.RoomMessage) (*domain.RoomMessage, error) {
	var createdAt time.Time

	err := r.db.QueryRow(ctx, sqlInsertRoomMessage, message.AuthorUserID, message.RoomID, message.Content).Scan(
		&message.ID,
		&createdAt,
	)
	if err != nil {
		return nil, fmt.Errorf("save room message: %w", err)
	}

	message.CreatedAt = createdAt

	return &message, nil
}

func (r *PartyRepo) SavePoll(ctx context.Context, poll domain.Poll) (*domain.Poll, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin save poll tx: %w", err)
	}

	defer func() {
		ignoreRollbackError(tx.Rollback(ctx))
	}()

	err = tx.QueryRow(ctx, sqlInsertRoomPoll, poll.CreatedByUserID, poll.RoomID, 1, poll.Question).Scan(
		&poll.ID,
		&poll.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert poll: %w", err)
	}

	for i := range poll.Options {
		if err = tx.QueryRow(
			ctx,
			sqlInsertRoomPollOption,
			poll.ID,
			poll.Options[i].Title,
		).Scan(&poll.Options[i].ID); err != nil {
			return nil, fmt.Errorf("insert poll option: %w", err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit save poll tx: %w", err)
	}

	return &poll, nil
}

func (r *PartyRepo) SaveVote(ctx context.Context, vote domain.PollVote) error {
	_, err := r.db.Exec(ctx, sqlInsertRoomVote, vote.UserID, vote.OptionID, vote.CoinsAmount)
	if err != nil {
		return fmt.Errorf("save vote: %w", err)
	}

	return nil
}

func (r *PartyRepo) GetPollOptionStakes(
	ctx context.Context,
	pollID, optionID int64,
) ([]domain.PollOptionStake, error) {
	rows, err := r.db.Query(ctx, sqlGetPollOptionStakes, pollID, optionID)
	if err != nil {
		return nil, fmt.Errorf("get poll option stakes: %w", err)
	}
	defer rows.Close()

	stakes := make([]domain.PollOptionStake, 0)

	for rows.Next() {
		var stake domain.PollOptionStake
		if err = rows.Scan(&stake.UserID, &stake.OptionID, &stake.CoinsAmount); err != nil {
			return nil, fmt.Errorf("scan poll option stake: %w", err)
		}

		stakes = append(stakes, stake)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate poll option stakes: %w", err)
	}

	return stakes, nil
}

func (r *PartyRepo) ResolvePoll(
	ctx context.Context,
	roomID, pollID, optionID, resolvedByUserID int64,
) (*domain.Poll, error) {
	tag, err := r.db.Exec(ctx, sqlResolvePoll, pollID, roomID, optionID, resolvedByUserID)
	if err != nil {
		return nil, fmt.Errorf("resolve poll: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return nil, domain.ErrPollAlreadyResolved
	}

	if err = r.TouchRoom(ctx, roomID); err != nil {
		return nil, err
	}

	room, err := r.GetRoomByID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	for _, poll := range room.Polls {
		if poll.ID == pollID {
			pollCopy := poll

			return &pollCopy, nil
		}
	}

	return nil, domain.ErrInvalidPoll
}

func (r *PartyRepo) populateRoomDetails(ctx context.Context, room *domain.Room, roomID int64) error {
	members, err := r.getRoomMembers(ctx, roomID)
	if err != nil {
		return err
	}

	room.Members = members

	playback, err := r.getRoomPlaybackState(ctx, roomID)
	if err != nil {
		return err
	}

	room.Playback = playback

	messages, err := r.getRoomMessages(ctx, roomID)
	if err != nil {
		return err
	}

	room.Messages = messages

	polls, err := r.getRoomPolls(ctx, roomID)
	if err != nil {
		return err
	}

	room.Polls = polls

	return nil
}

func (r *PartyRepo) loadRoomCards(ctx context.Context, query string, args ...any) ([]domain.RoomCard, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("load room cards: %w", err)
	}
	defer rows.Close()

	rooms := make([]domain.RoomCard, 0)

	for rows.Next() {
		var (
			room              domain.RoomCard
			playbackUpdatedAt time.Time
			roomUpdatedAt     time.Time
		)

		if err = rows.Scan(
			&room.ID,
			&room.Name,
			&room.Visibility,
			&room.InviteLink,
			&room.HostUserID,
			&room.HostName,
			&room.ParticipantsCount,
			&room.Playback.MovieID,
			&room.Playback.EpisodeID,
			&room.Playback.PlaybackURL,
			&room.Playback.DurationSeconds,
			&room.Playback.PositionSeconds,
			&room.Playback.Status,
			&playbackUpdatedAt,
			&roomUpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan room card: %w", err)
		}

		room.Playback.UpdatedAt = playbackUpdatedAt
		room.UpdatedAt = roomUpdatedAt
		rooms = append(rooms, room)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate room cards: %w", err)
	}

	return rooms, nil
}

func (r *PartyRepo) getRoomMembers(ctx context.Context, roomID int64) ([]domain.RoomMember, error) {
	rows, err := r.db.Query(ctx, sqlGetRoomMembers, roomID)
	if err != nil {
		return nil, fmt.Errorf("get room members: %w", err)
	}
	defer rows.Close()

	members := make([]domain.RoomMember, 0)

	for rows.Next() {
		var member domain.RoomMember
		if err = rows.Scan(
			&member.UserID,
			&member.DisplayName,
			&member.AvatarURL,
			&member.Role,
			&member.Status,
			&member.JoinedAt,
		); err != nil {
			return nil, fmt.Errorf("scan room member: %w", err)
		}

		members = append(members, member)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate room members: %w", err)
	}

	return members, nil
}

func (r *PartyRepo) getRoomPlaybackState(ctx context.Context, roomID int64) (domain.PlaybackState, error) {
	var state domain.PlaybackState

	err := r.db.QueryRow(ctx, sqlGetRoomPlaybackState, roomID).Scan(
		&state.MovieID,
		&state.EpisodeID,
		&state.PlaybackURL,
		&state.DurationSeconds,
		&state.PositionSeconds,
		&state.Status,
		&state.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.PlaybackState{}, nil
		}

		return domain.PlaybackState{}, fmt.Errorf("get room playback state: %w", err)
	}

	return state, nil
}

func (r *PartyRepo) getRoomMessages(ctx context.Context, roomID int64) ([]domain.RoomMessage, error) {
	rows, err := r.db.Query(ctx, sqlGetRoomMessages, roomID)
	if err != nil {
		return nil, fmt.Errorf("get room messages: %w", err)
	}
	defer rows.Close()

	messages := make([]domain.RoomMessage, 0)

	for rows.Next() {
		var message domain.RoomMessage
		if err = rows.Scan(
			&message.ID,
			&message.RoomID,
			&message.AuthorUserID,
			&message.AuthorName,
			&message.Content,
			&message.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan room message: %w", err)
		}

		messages = append(messages, message)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate room messages: %w", err)
	}

	return messages, nil
}

func (r *PartyRepo) getRoomPolls(ctx context.Context, roomID int64) ([]domain.Poll, error) {
	rows, err := r.db.Query(ctx, sqlGetRoomPolls, roomID)
	if err != nil {
		return nil, fmt.Errorf("get room polls: %w", err)
	}
	defer rows.Close()

	pollMap := make(map[int64]*domain.Poll)
	order := make([]int64, 0)

	for rows.Next() {
		poll, optionID, optionText, votesCount, coinsTotal, scanErr := scanRoomPollRow(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scan room poll: %w", scanErr)
		}

		existing := ensureRoomPoll(pollMap, &order, poll)
		appendRoomPollOption(existing, optionID, optionText, votesCount, coinsTotal)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate room polls: %w", err)
	}

	polls := make([]domain.Poll, 0, len(order))
	for _, id := range order {
		polls = append(polls, *pollMap[id])
	}

	return polls, nil
}

func scanRoomPollRow(rows corepostgres.Rows) (
	domain.Poll,
	sql.NullInt64,
	sql.NullString,
	sql.NullInt64,
	sql.NullInt64,
	error,
) {
	var (
		poll             domain.Poll
		resolvedAt       sql.NullTime
		correctOptionID  sql.NullInt64
		resolvedByUserID sql.NullInt64
		optionID         sql.NullInt64
		optionText       sql.NullString
		votesCount       sql.NullInt64
		coinsTotal       sql.NullInt64
	)

	if err := rows.Scan(
		&poll.ID,
		&poll.RoomID,
		&poll.Question,
		&poll.CreatedByUserID,
		&poll.CreatedAt,
		&resolvedAt,
		&correctOptionID,
		&resolvedByUserID,
		&optionID,
		&optionText,
		&votesCount,
		&coinsTotal,
	); err != nil {
		return domain.Poll{}, sql.NullInt64{}, sql.NullString{}, sql.NullInt64{}, sql.NullInt64{}, err
	}

	applyRoomPollNullFields(&poll, resolvedAt, correctOptionID, resolvedByUserID)

	return poll, optionID, optionText, votesCount, coinsTotal, nil
}

func applyRoomPollNullFields(
	poll *domain.Poll,
	resolvedAt sql.NullTime,
	correctOptionID sql.NullInt64,
	resolvedByUserID sql.NullInt64,
) {
	if resolvedAt.Valid {
		value := resolvedAt.Time
		poll.ClosedAt = &value
	}

	if correctOptionID.Valid && correctOptionID.Int64 > 0 {
		value := correctOptionID.Int64
		poll.CorrectOptionID = &value
	}

	if resolvedByUserID.Valid && resolvedByUserID.Int64 > 0 {
		value := resolvedByUserID.Int64
		poll.ResolvedByUserID = &value
	}
}

func ensureRoomPoll(pollMap map[int64]*domain.Poll, order *[]int64, poll domain.Poll) *domain.Poll {
	existing, ok := pollMap[poll.ID]
	if ok {
		return existing
	}

	poll.Options = make([]domain.PollOption, 0)
	pollMap[poll.ID] = &poll
	*order = append(*order, poll.ID)

	return &poll
}

func appendRoomPollOption(
	poll *domain.Poll,
	optionID sql.NullInt64,
	optionText sql.NullString,
	votesCount sql.NullInt64,
	coinsTotal sql.NullInt64,
) {
	if !optionID.Valid {
		return
	}

	option := domain.PollOption{
		ID:    optionID.Int64,
		Title: optionText.String,
	}
	if votesCount.Valid {
		option.VotesCount = votesCount.Int64
	}

	if coinsTotal.Valid {
		option.CoinsTotal = coinsTotal.Int64
	}

	poll.Options = append(poll.Options, option)
}

func ignoreRollbackError(err error) {
	if err == nil {
		return
	}

	if errors.Is(err, pgx.ErrTxClosed) {
		return
	}
}
