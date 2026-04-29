package usecase

import (
	"context"
	"strings"

	domain "github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie-service/domain"
)

func (u *MovieUsecase) GetSelectionByTitle(ctx context.Context, title string) (domain.SelectionResponse, error) {
	normalizedTitle := strings.TrimSpace(title)
	if !domain.ValidateSelectionTitle(normalizedTitle) {
		return domain.SelectionResponse{}, domain.ErrInvalidSelectionTitle
	}

	selection, err := u.movieRepo.GetSelectionByTitle(ctx, normalizedTitle)
	if err != nil {
		return domain.SelectionResponse{}, err
	}

	return u.buildSelectionResponse(ctx, selection)
}

func (u *MovieUsecase) GetAllSelections(ctx context.Context) ([]domain.SelectionResponse, error) {
	selections, err := u.movieRepo.GetAllSelections(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]domain.SelectionResponse, 0, len(selections))
	for _, selection := range selections {
		item, buildErr := u.buildSelectionResponse(ctx, selection)
		if buildErr != nil {
			return nil, buildErr
		}

		result = append(result, item)
	}

	return result, nil
}
