package domain

type OverviewResponse struct {
	ActiveRooms   []RoomCard
	MyRooms       []RoomCard
	FeaturedRooms []RoomCard
}

type RoomResponse struct {
	Room Room
}

type CreateRoomRequest struct {
	Name       string
	Visibility string
	MovieID    int64
	EpisodeID  int64
}

type JoinRoomRequest struct {
	InviteLink string
}

type DeleteRoomResponse struct {
	RoomID  int64
	Success bool
}

type SubscribeRoomRequest struct {
	RoomID int64
}

type ApplyRoomActionRequest struct {
	RoomID          int64
	Action          string
	MovieID         int64
	EpisodeID       int64
	PlaybackURL     string
	DurationSeconds int64
	PositionSeconds int64
	Status          string
}

type SendRoomMessageRequest struct {
	RoomID  int64
	Content string
}

type CreateRoomPollRequest struct {
	RoomID   int64
	Question string
	Options  []string
}

type VoteRoomPollRequest struct {
	RoomID   int64
	PollID   int64
	OptionID int64
}
