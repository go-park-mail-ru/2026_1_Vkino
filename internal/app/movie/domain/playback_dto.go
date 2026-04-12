package domain

// EpisodeItemResponse Хранит данные о конкретном эпизоде
type EpisodeItemResponse struct {
	ID              int64  `json:"id"`
	MovieID         int64  `json:"movie_id"`
	SeasonNumber    int    `json:"season_number"`
	EpisodeNumber   int    `json:"episode_number"`
	Title           string `json:"title"`
	Description     string `json:"description"`
	DurationSeconds int    `json:"duration_seconds"`
	ImgURL          string `json:"img_url"`
}

// EpisodePlaybackResponse Хранит данные о видео которое запускаем
type EpisodePlaybackResponse struct {
	EpisodeID       int64  `json:"episode_id"`
	MovieID         int64  `json:"movie_id"`
	SeasonNumber    int    `json:"season_number"`
	EpisodeNumber   int    `json:"episode_number"`
	Title           string `json:"title"`
	DurationSeconds int    `json:"duration_seconds"`
	PlaybackURL     string `json:"playback_url"`
	PositionSeconds int    `json:"position_seconds,omitempty"`
}

type WatchProgressRequest struct {
	PositionSeconds int `json:"position_seconds"`
}

type WatchProgressResponse struct {
	EpisodeID       int64 `json:"episode_id"`
	PositionSeconds int   `json:"position_seconds"`
}
