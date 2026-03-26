package domain

import (
	"time"

	"github.com/google/uuid"
)

type Selection struct {
	ID        uuid.UUID
	Title     string
	CreatedAt time.Time
	UpdatedAt time.Time

	Movies []Movie `json:"movies"`
}
