package usecase

import (
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie/domain"
)

type Usecase interface {
	GetAllSelections() ([]domain.SelectionResponse, error)
	GetSelectionByTitle(title string) (domain.SelectionResponse, error)
}
