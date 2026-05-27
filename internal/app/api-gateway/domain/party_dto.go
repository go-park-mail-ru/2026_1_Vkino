package domain

import userv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/user/v1"

//go:generate go run -mod=mod github.com/mailru/easyjson/easyjson party_dto.go

//easyjson:json
//nolint:recvcheck // easyjson generates Marshal* on value receiver and Unmarshal* on pointer receiver.
type PartyOverviewResponse struct {
	ActiveRooms   []PartyRoomCardHTTP `json:"active_rooms"`
	MyRooms       []PartyRoomCardHTTP `json:"my_rooms"`
	FeaturedRooms []PartyRoomCardHTTP `json:"featured_rooms"`
}

//easyjson:json
//nolint:recvcheck // easyjson generates Marshal* on value receiver and Unmarshal* on pointer receiver.
type PartyRoomCardHTTP struct {
	ID                int64                   `json:"id"`
	Name              string                  `json:"name"`
	Visibility        string                  `json:"visibility"`
	InviteLink        string                  `json:"invite_link"`
	HostUserID        int64                   `json:"host_user_id"`
	HostName          string                  `json:"host_name"`
	ParticipantsCount int32                   `json:"participants_count"`
	Playback          *PartyPlaybackStateHTTP `json:"playback,omitempty"`
	UpdatedAt         string                  `json:"updated_at"`
}

//easyjson:json
//nolint:recvcheck // easyjson generates Marshal* on value receiver and Unmarshal* on pointer receiver.
type PartyPlaybackStateHTTP struct {
	MovieID         int64   `json:"movie_id"`
	MovieTitle      *string `json:"movie_title"`
	EpisodeID       int64   `json:"episode_id"`
	ImgURL          *string `json:"img_url"`
	PlaybackURL     string  `json:"playback_url,omitempty"`
	DurationSeconds int64   `json:"duration_seconds"`
	PositionSeconds int64   `json:"position_seconds"`
	Status          string  `json:"status"`
	UpdatedAt       string  `json:"updated_at"`
}

//easyjson:skip
type PartyFriendInviteHTTP struct {
	RoomID int64                     `json:"room_id"`
	Status string                    `json:"status"`
	Friend *userv1.GetFriendResponse `json:"friend"`
}
