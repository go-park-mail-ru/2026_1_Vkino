package domain

import "github.com/google/uuid"

type MoviePreview struct {
	ID             uuid.UUID `json:"id"`
	Title          string    `json:"title"`
	PictureFileKey string    `json:"img_url"`
}
