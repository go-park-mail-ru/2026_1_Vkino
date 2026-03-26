package domain

import "github.com/google/uuid"

type MoviePreview struct {
	ID             uuid.UUID `json:"id"`
	Title          string    `json:"title"`
	PictureFileKey string    `json:"img_url"`
}


type MovieResponse struct {
	ID              uuid.UUID      `json:"id"`
	Title           string         `json:"title"`
	Description     string         `json:"description"`
	PictureFileKey  string         `json:"img_url"`
	CoverFileKey    string         `json:"cover_img_url"`
	DurationMinutes int            `json:"duration_minutes"`
	AgeLimit        int            `json:"age_limit"`
	ReleaseYear     int            `json:"release_year"`
	Country         string         `json:"country"`
	Director        string         `json:"director"`
	Genres          []string       `json:"genres"`
	Actors          []ActorPreview `json:"actors"`
}

func (m *MovieResponse) Name() string {
	return "movies"
}