package usecase

import (
	"context"
	"strings"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/movie-service/domain"
)

func (u *MovieUsecase) SearchMovies(ctx context.Context, query string) ([]domain.MovieCardResponse, error) {
	normalized := strings.TrimSpace(query)
	if !domain.ValidateSearchQuery(normalized) {
		return nil, domain.ErrInvalidSearchQuery
	}

	movies, err := u.movieRepo.SearchMovies(ctx, normalized)
	if err != nil {
		return nil, err
	}

	result := make([]domain.MovieCardResponse, 0, len(movies))
	for _, movie := range movies {
		card, buildErr := u.buildMovieCardResponse(ctx, movie)
		if buildErr != nil {
			return nil, buildErr
		}
		result = append(result, card)
	}

	return result, nil
}
