package domain

import "github.com/google/uuid"

type MoviePreview struct {
	ID             uuid.UUID `json:"id"`
	Title          string    `json:"title"`
	PictureFileKey string    `json:"img_url"`
}


type MovieResponse struct {
	ID                 uuid.UUID      `json:"id"`
	Title              string         `json:"title"`
	Description        string         `json:"description"`
	Director           string         `json:"director"`
	ContentType        string         `json:"content_type"`
	ReleaseYear        int            `json:"release_year"`
	DurationSeconds    int            `json:"duration_seconds"`
	AgeLimit           int            `json:"age_limit"`
	OriginalLanguageID int            `json:"original_language_id"`
	CountryID          int            `json:"country_id"`
	PictureFileKey     string         `json:"img_url"`
	Genres             []string       `json:"genres"`
	Actors             []ActorPreview `json:"actors"`
}

func (m *MovieResponse) Name() string {
	return "movies"
}