package repository

import (
	"github.com/google/uuid"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie/domain"
)

type MovieRepo interface {
	GetAllSelections() ([]domain.SelectionResponse, error)
	GetSelectionByTitle(title string) (domain.SelectionResponse, error)
	GetMovieByID(id uuid.UUID) (domain.MovieResponse, error)
	GetActorByID(id uuid.UUID) (domain.ActorResponse, error)
}
