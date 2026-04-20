package dto

type MovieResponse struct {
	ID          int64                `json:"id"`
	Title       string               `json:"title"`
	Description string               `json:"description"`
	Year        int                  `json:"year"`
	Countries   []string             `json:"countries"`
	Genres      []string             `json:"genres"`
	AgeLimit    int                  `json:"age_limit"`
	DurationMin int                  `json:"duration_min"`
	PosterURL   string               `json:"poster_url,omitempty"`
	CardURL     string               `json:"card_url,omitempty"`
	Actors      []ActorShortResponse `json:"actors"`
	Episodes    []EpisodeResponse    `json:"episodes"`
}

type ActorResponse struct {
	ID          int64               `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	AvatarURL   string              `json:"avatar_url,omitempty"`
	Movies      []MovieCardResponse `json:"movies"`
}

type ActorShortResponse struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

type EpisodeResponse struct {
	ID          int64  `json:"id"`
	Number      int    `json:"number"`
	Title       string `json:"title"`
	DurationSec int    `json:"duration_sec"`
}

type MovieCardResponse struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Year      int    `json:"year"`
	PosterURL string `json:"poster_url,omitempty"`
	CardURL   string `json:"card_url,omitempty"`
}

type SelectionResponse struct {
	Title  string              `json:"title"`
	Movies []MovieCardResponse `json:"movies"`
}

type SelectionsResponse struct {
	Selections []SelectionResponse `json:"selections"`
}
