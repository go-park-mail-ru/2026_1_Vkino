package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie-service/domain"
)

func (u *MovieUsecase) GetAllGenres(ctx context.Context) ([]domain.GenreShortResponse, error) {
	genres, err := u.movieRepo.GetAllGenres(ctx)
	if err != nil {
		return nil, err
	}

	return u.buildGenreShortResponses(genres), nil
}
