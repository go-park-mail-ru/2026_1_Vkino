package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie-service/domain"
)

func (u *MovieUsecase) buildMovieResponse(ctx context.Context, movie *domain.Movie) (domain.MovieResponse, error) {
	resp := domain.MovieResponse{
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
		Actors:             make([]domain.ActorShortResponse, 0, len(movie.Actors)),
		Episodes:           make([]domain.EpisodeResponse, 0, len(movie.Episodes)),
	}

	var err error

	resp.PictureFileKey, err = u.presignCard(ctx, movie.PictureFileKey)
	if err != nil {
		return domain.MovieResponse{}, err
	}

	resp.PosterFileKey, err = u.presignPoster(ctx, movie.PosterFileKey)
	if err != nil {
		return domain.MovieResponse{}, err
	}

	for _, actor := range movie.Actors {
		pictureURL, pictureErr := u.presignActor(ctx, actor.PictureFileKey)
		if pictureErr != nil {
			return domain.MovieResponse{}, pictureErr
		}

		resp.Actors = append(resp.Actors, domain.ActorShortResponse{
			ID:             actor.ID,
			FullName:       actor.FullName,
			PictureFileKey: pictureURL,
		})
	}

	for _, episode := range movie.Episodes {
		episodeImageURL, imgErr := u.presignCard(ctx, episode.PictureFileKey)
		if imgErr != nil {
			return domain.MovieResponse{}, imgErr
		}

		videoURL, videoErr := u.presignVideo(ctx, episode.VideoFileKey)
		if videoErr != nil {
			return domain.MovieResponse{}, videoErr
		}

		resp.Episodes = append(resp.Episodes, domain.EpisodeResponse{
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

func (u *MovieUsecase) buildActorResponse(ctx context.Context, actor *domain.Actor) (domain.ActorResponse, error) {
	resp := domain.ActorResponse{
		ID:        actor.ID,
		FullName:  actor.FullName,
		Biography: actor.Biography,
		BirthDate: formatDate(actor.BirthDate),
		CountryID: actor.CountryID,
		Movies:    make([]domain.MovieCardResponse, 0, len(actor.Movies)),
	}

	var err error

	resp.PictureFileKey, err = u.presignActor(ctx, actor.PictureFileKey)
	if err != nil {
		return domain.ActorResponse{}, err
	}

	for _, movie := range actor.Movies {
		card, buildErr := u.buildMovieCardResponse(ctx, movie)
		if buildErr != nil {
			return domain.ActorResponse{}, buildErr
		}

		resp.Movies = append(resp.Movies, card)
	}

	return resp, nil
}

func (u *MovieUsecase) buildSelectionResponse(
	ctx context.Context,
	selection domain.Selection,
) (domain.SelectionResponse, error) {
	resp := domain.SelectionResponse{
		Title:  selection.Title,
		Movies: make([]domain.MovieCardResponse, 0, len(selection.Movies)),
	}

	for _, movie := range selection.Movies {
		card, err := u.buildMovieCardResponse(ctx, movie)
		if err != nil {
			return domain.SelectionResponse{}, err
		}

		resp.Movies = append(resp.Movies, card)
	}

	return resp, nil
}

func (u *MovieUsecase) buildGenreResponse(
	ctx context.Context,
	genre domain.Genre,
) (domain.GenreResponse, error) {
	resp := domain.GenreResponse{
		ID:     genre.ID,
		Title:  genre.Title,
		Movies: make([]domain.MovieCardResponse, 0, len(genre.Movies)),
	}

	for _, movie := range genre.Movies {
		card, err := u.buildMovieCardResponse(ctx, movie)
		if err != nil {
			return domain.GenreResponse{}, err
		}

		resp.Movies = append(resp.Movies, card)
	}

	return resp, nil
}

func (u *MovieUsecase) buildMovieCardResponse(
	ctx context.Context,
	movie domain.MovieCard,
) (domain.MovieCardResponse, error) {
	imageURL, err := u.presignCard(ctx, movie.PictureFileKey)
	if err != nil {
		return domain.MovieCardResponse{}, err
	}

	return domain.MovieCardResponse{
		ID:             movie.ID,
		Title:          movie.Title,
		PictureFileKey: imageURL,
	}, nil
}

func (u *MovieUsecase) buildActorShortResponse(
	ctx context.Context,
	actor domain.ActorShort,
) (domain.ActorShortResponse, error) {
	imageURL, err := u.presignActor(ctx, actor.PictureFileKey)
	if err != nil {
		return domain.ActorShortResponse{}, err
	}

	return domain.ActorShortResponse{
		ID:             actor.ID,
		FullName:       actor.FullName,
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
		return "", fmt.Errorf("%w: presign object key=%q: %v", domain.ErrInternal, key, err)
	}

	return url, nil
}

func formatDate(value *time.Time) string {
	if value == nil {
		return ""
	}

	return value.Format("2006-01-02")
}
