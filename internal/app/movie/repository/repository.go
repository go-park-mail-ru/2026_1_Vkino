package repository

import (
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie/domain"
)

type MovieRepo interface {
	GetSelectionByTitle(title string) (*domain.SelectionResponse, error)
}
