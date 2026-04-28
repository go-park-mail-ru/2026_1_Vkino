package usecase

import (
	"context"

	domain2 "github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie-service/domain"
)

func (u *MovieUsecase) GetGenreByID(ctx context.Context, genreID int64) (domain2.GenreResponse, error) {
	if genreID <= 0 {
		return domain2.GenreResponse{}, domain2.ErrInvalidGenreID
	}

	genre, err := u.movieRepo.GetGenreByID(ctx, genreID)
	if err != nil {
		return domain2.GenreResponse{}, err
	}

	return u.buildGenreResponse(ctx, genre)
}
