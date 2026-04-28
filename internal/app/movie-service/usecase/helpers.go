package usecase

import (
	"context"
	"fmt"
	"time"

	domain2 "github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie-service/domain"
)

func (u *MovieUsecase) buildMovieResponse(ctx context.Context, movie *domain2.Movie) (domain2.MovieResponse, error) {
	resp := domain2.MovieResponse{
		ID:                 movie.ID,
		Title:              movie.Title,
		Description:        movie.Description,
		Director:           movie.Director,
		TrailerURL:         movie.TrailerURL,
		ContentType:        movie.ContentType,
		ReleaseYear:        movie.ReleaseYear,
		DurationSeconds:    movie.DurationSeconds,
		AgeLimit:           movie.AgeLimit,
		OriginalLanguageID: movie.OriginalLanguageID,
		OriginalLanguage:   movie.OriginalLanguage,
		CountryID:          movie.CountryID,
		Country:            movie.Country,
		Genres:             movie.Genres,
		Actors:             make([]domain2.ActorShortResponse, 0, len(movie.Actors)),
		Episodes:           make([]domain2.EpisodeResponse, 0, len(movie.Episodes)),
	}

	var err error

	resp.PictureFileKey, err = u.presignCard(ctx, movie.PictureFileKey)
	if err != nil {
		return domain2.MovieResponse{}, err
	}

	resp.PosterFileKey, err = u.presignPoster(ctx, movie.PosterFileKey)
	if err != nil {
		return domain2.MovieResponse{}, err
	}

	for _, actor := range movie.Actors {
		pictureURL, pictureErr := u.presignActor(ctx, actor.PictureFileKey)
		if pictureErr != nil {
			return domain2.MovieResponse{}, pictureErr
		}

		resp.Actors = append(resp.Actors, domain2.ActorShortResponse{
			ID:             actor.ID,
			FullName:       actor.FullName,
			PictureFileKey: pictureURL,
		})
	}

	for _, episode := range movie.Episodes {
		episodeImageURL, imgErr := u.presignCard(ctx, episode.PictureFileKey)
		if imgErr != nil {
			return domain2.MovieResponse{}, imgErr
		}

		videoURL, videoErr := u.presignVideo(ctx, episode.VideoFileKey)
		if videoErr != nil {
			return domain2.MovieResponse{}, videoErr
		}

		resp.Episodes = append(resp.Episodes, domain2.EpisodeResponse{
			ID:              episode.ID,
			MovieID:         episode.MovieID,
			SeasonNumber:    episode.SeasonNumber,
			EpisodeNumber:   episode.EpisodeNumber,
			Title:           episode.Title,
			Description:     episode.Description,
			DurationSeconds: episode.DurationSeconds,
			PictureFileKey:  episodeImageURL,
			VideoURL:        videoURL,
		})
	}

	return resp, nil
}

func (u *MovieUsecase) buildActorResponse(ctx context.Context, actor *domain2.Actor) (domain2.ActorResponse, error) {
	resp := domain2.ActorResponse{
		ID:        actor.ID,
		FullName:  actor.FullName,
		Biography: actor.Biography,
		BirthDate: formatDate(actor.BirthDate),
		CountryID: actor.CountryID,
		Movies:    make([]domain2.MovieCardResponse, 0, len(actor.Movies)),
	}

	var err error

	resp.PictureFileKey, err = u.presignActor(ctx, actor.PictureFileKey)
	if err != nil {
		return domain2.ActorResponse{}, err
	}

	for _, movie := range actor.Movies {
		card, buildErr := u.buildMovieCardResponse(ctx, movie)
		if buildErr != nil {
			return domain2.ActorResponse{}, buildErr
		}

		resp.Movies = append(resp.Movies, card)
	}

	return resp, nil
}

func (u *MovieUsecase) buildSelectionResponse(
	ctx context.Context,
	selection domain2.Selection,
) (domain2.SelectionResponse, error) {
	resp := domain2.SelectionResponse{
		Title:  selection.Title,
		Movies: make([]domain2.MovieCardResponse, 0, len(selection.Movies)),
	}

	for _, movie := range selection.Movies {
		card, err := u.buildMovieCardResponse(ctx, movie)
		if err != nil {
			return domain2.SelectionResponse{}, err
		}

		resp.Movies = append(resp.Movies, card)
	}

	return resp, nil
}

func (u *MovieUsecase) buildGenreResponse(
	ctx context.Context,
	genre domain2.Genre,
) (domain2.GenreResponse, error) {
	resp := domain2.GenreResponse{
		ID:     genre.ID,
		Title:  genre.Title,
		Movies: make([]domain2.MovieCardResponse, 0, len(genre.Movies)),
	}

	for _, movie := range genre.Movies {
		card, err := u.buildMovieCardResponse(ctx, movie)
		if err != nil {
			return domain2.GenreResponse{}, err
		}

		resp.Movies = append(resp.Movies, card)
	}

	return resp, nil
}

func (u *MovieUsecase) buildMovieCardResponse(
	ctx context.Context,
	movie domain2.MovieCard,
) (domain2.MovieCardResponse, error) {
	imageURL, err := u.presignCard(ctx, movie.PictureFileKey)
	if err != nil {
		return domain2.MovieCardResponse{}, err
	}

	return domain2.MovieCardResponse{
		ID:             movie.ID,
		Title:          movie.Title,
		PictureFileKey: imageURL,
	}, nil
}

func (u *MovieUsecase) presignPoster(ctx context.Context, key string) (string, error) {
	return presignIfExists(ctx, u.posterStore, key)
}

func (u *MovieUsecase) presignCard(ctx context.Context, key string) (string, error) {
	return presignIfExists(ctx, u.cardStore, key)
}

func (u *MovieUsecase) presignActor(ctx context.Context, key string) (string, error) {
	return presignIfExists(ctx, u.actorStore, key)
}

func (u *MovieUsecase) presignVideo(ctx context.Context, key string) (string, error) {
	return presignIfExists(ctx, u.videoStore, key)
}

type presignStorage interface {
	PresignGetObject(ctx context.Context, key string, expires time.Duration) (string, error)
}

func presignIfExists(ctx context.Context, store presignStorage, key string) (string, error) {
	if key == "" {
		return "", nil
	}

	url, err := store.PresignGetObject(ctx, key, 0)
	if err != nil {
		return "", fmt.Errorf("%w: presign object key=%q: %v", domain2.ErrInternal, key, err)
	}

	return url, nil
}

func formatDate(value *time.Time) string {
	if value == nil {
		return ""
	}

	return value.Format("2006-01-02")
}
