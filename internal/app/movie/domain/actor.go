package domain

import (
	"time"

	"github.com/google/uuid"
)

type Actor struct {
	ID             uuid.UUID   `json:"id"`
	FullName       string      `json:"full_name"`
	Biography      string      `json:"biography"`
	BirthDate      time.Time   `json:"birth_date"`
	Country        string      `json:"country"`
	PictureFileKey string      `json:"picture_file_key"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
	MovieIDs       []uuid.UUID `json:"movie_ids"`
}