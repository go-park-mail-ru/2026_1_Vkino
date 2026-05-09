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
	RoomID     int64
}

type DeleteRoomResponse struct {
	RoomID  int64
	Success bool
}

type SubscribeRoomRequest struct {
	RoomID int64
}
