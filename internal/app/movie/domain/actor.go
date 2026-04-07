package domain

import (
	"time"
)

type Actor struct {
	ID             int64       `json:"id"`
	FullName       string      `json:"full_name"`
	BirthDate      *time.Time  `json:"birth_date,omitempty"`
	Biography      *string     `json:"biography,omitempty"`
	CountryID      int64       `json:"country_id"`
	PictureFileKey string      `json:"picture_file_key"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
	MovieIDs       []int64     `json:"movie_ids,omitempty"`
}