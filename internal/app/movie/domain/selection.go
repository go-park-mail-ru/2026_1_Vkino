package domain

import (
	"time"
)

type Selection struct {
	ID        int64
	Title     string
	CreatedAt time.Time
	UpdatedAt time.Time
	Movies    []Movie `json:"movies"`
}
