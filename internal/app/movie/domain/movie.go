package domain

import (
	"time"

	"github.com/google/uuid"
)

type Movie struct {
	ID                 uuid.UUID
	Title              string
	Description        string
	ContentType        string
	ReleaseYear        int
	DurationSeconds    int
	AgeLimit           int
	OriginalLanguageID int
	CountryID          int
	PictureFileKey     string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
