package usecase

import (
	"context"
	"strings"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie/domain"
)

const maxSearchQueryLength = 256

func (m *MovieUsecase) Search(ctx context.Context, query string) (domain.SearchResponse, error) {
	query = strings.TrimSpace(query)
	if query == "" || len(query) > maxSearchQueryLength {
		return domain.SearchResponse{}, domain.ErrInvalidSearchQuery
	}

	result, err := m.movieRepo.Search(ctx, query)
	if err != nil {
		return domain.SearchResponse{}, err
	}

	for i := range result.Movies {
		result.Movies[i].ImgUrl, err = m.presignCardURL(ctx, result.Movies[i].ImgUrl)
		if err != nil {
			return domain.SearchResponse{}, err
		}
	}

	for i := range result.Actors {
		result.Actors[i].PictureFileKey, err = m.presignActorURL(ctx, result.Actors[i].PictureFileKey)
		if err != nil {
			return domain.SearchResponse{}, err
		}
	}

	return result, nil
}
