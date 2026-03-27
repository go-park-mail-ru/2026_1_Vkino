package domain

import (
	"time"

	"github.com/google/uuid"
)

type Movie struct {
	ID                 uuid.UUID   `json:"id"`
	Title              string      `json:"title"`
	Description        *string     `json:"description,omitempty"`
	Director           *string     `json:"director,omitempty"`
	ContentType        string      `json:"content_type"`
	ReleaseYear        int         `json:"release_year"`
	DurationSeconds    int         `json:"duration_seconds"`
	AgeLimit           int         `json:"age_limit"`
	OriginalLanguageID int         `json:"original_language_id"`
	CountryID          int         `json:"country_id"`
	PictureFileKey     string      `json:"picture_file_key"`
	CreatedAt          time.Time   `json:"created_at"`
	UpdatedAt          time.Time   `json:"updated_at"`
	GenreIDs           []int       `json:"genre_ids,omitempty"`
	ActorIDs           []uuid.UUID `json:"actor_ids,omitempty"`
}
