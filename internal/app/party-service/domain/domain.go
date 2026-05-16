package domain

import "time"

type Room struct {
	ID         int64
	Name       string
	Visibility string
	HostUserID int64
	InviteLink string
	Members    []RoomMember
	Playback   PlaybackState
	Messages   []RoomMessage
	Polls      []Poll
	UpdatedAt  time.Time
}

type RoomCard struct {
	ID                int64
	Name              string
	Visibility        string
	InviteLink        string
	HostUserID        int64
	HostName          string
	ParticipantsCount int32
	Playback          PlaybackState
	UpdatedAt         time.Time
}

type RoomMember struct {
	UserID      int64
	DisplayName string
	AvatarURL   string
	Role        string
	Status      string
	JoinedAt    time.Time
}

type Invite struct {
	RoomID    int64
	Link      string
	CreatedBy int64
	CreatedAt time.Time
	ExpiresAt *time.Time
}

type PlaybackState struct {
	MovieID         int64
	EpisodeID       int64
	PlaybackURL     string
	DurationSeconds int64
	PositionSeconds int64
	Status          string
	UpdatedAt       time.Time
}

type RoomMessage struct {
	ID           int64
	RoomID       int64
	AuthorUserID int64
	AuthorName   string
	Content      string
	CreatedAt    time.Time
}

type Poll struct {
	ID              int64
	RoomID          int64
	Question        string
	Options         []PollOption
	CreatedByUserID int64
	CreatedAt       time.Time
	ClosedAt        *time.Time
}

type PollOption struct {
	ID         int64
	Title      string
	VotesCount int64
}

type PollVote struct {
	PollID   int64
	OptionID int64
	UserID   int64
}

type RoomEvent struct {
	Type        string
	RoomID      int64
	ActorUserID int64
	Playback    *PlaybackState
	Message     *RoomMessage
	Poll        *Poll
	Member      *RoomMember
	Vote        *PollVote
	SentAt      time.Time
}
