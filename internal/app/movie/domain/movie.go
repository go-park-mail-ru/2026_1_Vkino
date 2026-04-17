package domain

import (
	"time"
)

type Movie struct {
	ID                 int64     `json:"id"`
	Title              string    `json:"title"`
	Description        *string   `json:"description,omitempty"`
	Director           *string   `json:"director,omitempty"`
	ContentType        string    `json:"content_type"`
	ReleaseYear        int       `json:"release_year"`
	DurationSeconds    int       `json:"duration_seconds"`
	AgeLimit           int       `json:"age_limit"`
	OriginalLanguageID int64     `json:"original_language_id"`
	CountryID          int64     `json:"country_id"`
	PictureFileKey     string    `json:"picture_file_key"`
	PosterFileKey      *string   `json:"poster_file_key,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	GenreIDs           []int64   `json:"genre_ids,omitempty"`
	ActorIDs           []int64   `json:"actor_ids,omitempty"`
}
