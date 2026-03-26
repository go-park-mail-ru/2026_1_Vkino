package domain

import "github.com/google/uuid"

type ActorPreview struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	PictureFileKey string    `json:"img_url"`
}

type ActorResponse struct {
	ID             uuid.UUID      `json:"id"`
	FullName       string         `json:"full_name"`
	Biography      string         `json:"biography"`
	BirthDate      string         `json:"birth_date"`
	Country        string         `json:"country"`
	PictureFileKey string         `json:"img_url"`
	Movies         []MoviePreview `json:"movies"`
}

func (a *ActorResponse) Name() string {
	return "actors"
}