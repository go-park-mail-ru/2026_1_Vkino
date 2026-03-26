package domain

import (
	"time"

	"github.com/google/uuid"
)

type Movie struct {
	// собственные поля
	ID                 uuid.UUID `json:"id"`
	Title              string `json:"title"`
	Description        string `json:"description"`
	ContentType        string `json:"content_type"`
	ReleaseYear        int `json:"release_year"`
	Duration           time.Time `json:"duration"` // в БД DurationSeconds, но лучше поменять и там
	AgeLimit           int `json:"age_limit"`
	OriginalLanguageID int `json:"original_language_id"`
	CountryID          int `json:"country_id"`
	PictureFileKey     string `json:"picture_file_key"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	// связанные сущности
	Genres      []string `json:"genres"`
	
	// Actors	[]Actor `json:"actors"`
}
