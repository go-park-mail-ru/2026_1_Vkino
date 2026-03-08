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
	Duration           time.Time // в БД DurationSeconds, но лучше поменять и там
	AgeLimit           int
	OriginalLanguageID int
	CountryID          int
	PictureFileKey     string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
