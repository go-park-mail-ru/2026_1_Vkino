package domain

type MovieResponse struct {
	ID          int64
	Title       string
	Description string
	Year        int
	Countries   []string
	Genres      []string
	AgeLimit    int
	DurationMin int
	PosterURL   string
	CardURL     string
	Actors      []ActorShortResponse
	Episodes    []EpisodeResponse
}

type ActorResponse struct {
	ID          int64
	Name        string
	Description string
	AvatarURL   string
	Movies      []MovieCardResponse
}

type ActorShortResponse struct {
	ID        int64
	Name      string
	AvatarURL string
}

type EpisodeResponse struct {
	ID          int64
	Number      int
	Title       string
	DurationSec int
}

type MovieCardResponse struct {
	ID        int64
	Title     string
	Year      int
	PosterURL string
	CardURL   string
}

type SelectionResponse struct {
	Title  string
	Movies []MovieCardResponse
}

type EpisodePlaybackResponse struct {
	EpisodeID   int64
	PlaybackURL string
	DurationSec int
}

type EpisodeProgressResponse struct {
	EpisodeID   int64
	PositionSec int64
}