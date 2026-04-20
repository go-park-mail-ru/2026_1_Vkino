package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/services/movie-service/internal/domain"
)

func (u *MovieUsecase) buildMovieResponse(ctx context.Context, movie *domain.Movie) (domain.MovieResponse, error) {
	resp := domain.MovieResponse{
		ID:          movie.ID,
		Title:       movie.Title,
		Description: movie.Description,
		Year:        movie.Year,
		Countries:   movie.Countries,
		Genres:      movie.Genres,
		AgeLimit:    movie.AgeLimit,
		DurationMin: movie.DurationMin,
		Actors:      make([]domain.ActorShortResponse, 0, len(movie.Actors)),
		Episodes:    make([]domain.EpisodeResponse, 0, len(movie.Episodes)),
	}

	var err error
	resp.PosterURL, err = u.presignPoster(ctx, movie.PosterFileKey)
	if err != nil {
		return domain.MovieResponse{}, err
	}

	resp.CardURL, err = u.presignCard(ctx, movie.CardFileKey)
	if err != nil {
		return domain.MovieResponse{}, err
	}

	for _, actor := range movie.Actors {
		avatarURL, avatarErr := u.presignActor(ctx, actor.AvatarFileKey)
		if avatarErr != nil {
			return domain.MovieResponse{}, avatarErr
		}

		resp.Actors = append(resp.Actors, domain.ActorShortResponse{
			ID:        actor.ID,
			Name:      actor.Name,
			AvatarURL: avatarURL,
		})
	}

	for _, episode := range movie.Episodes {
		resp.Episodes = append(resp.Episodes, domain.EpisodeResponse{
			ID:          episode.ID,
			Number:      episode.Number,
			Title:       episode.Title,
			DurationSec: episode.DurationSec,
		})
	}

	return resp, nil
}

func (u *MovieUsecase) buildActorResponse(ctx context.Context, actor *domain.Actor) (domain.ActorResponse, error) {
	resp := domain.ActorResponse{
		ID:          actor.ID,
		Name:        actor.Name,
		Description: actor.Description,
		Movies:      make([]domain.MovieCardResponse, 0, len(actor.Movies)),
	}

	var err error
	resp.AvatarURL, err = u.presignActor(ctx, actor.AvatarFileKey)
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

func (u *MovieUsecase) buildMovieCardResponse(
	ctx context.Context,
	movie domain.MovieCard,
) (domain.MovieCardResponse, error) {
	posterURL, err := u.presignPoster(ctx, movie.PosterFileKey)
	if err != nil {
		return domain.MovieCardResponse{}, err
	}

	cardURL, err := u.presignCard(ctx, movie.CardFileKey)
	if err != nil {
		return domain.MovieCardResponse{}, err
	}

	return domain.MovieCardResponse{
		ID:        movie.ID,
		Title:     movie.Title,
		Year:      movie.Year,
		PosterURL: posterURL,
		CardURL:   cardURL,
	}, nil
}

func (u *MovieUsecase) presignPoster(ctx context.Context, key *string) (string, error) {
	return presignIfExists(ctx, u.posterStore, key)
}

func (u *MovieUsecase) presignCard(ctx context.Context, key *string) (string, error) {
	return presignIfExists(ctx, u.cardStore, key)
}

func (u *MovieUsecase) presignActor(ctx context.Context, key *string) (string, error) {
	return presignIfExists(ctx, u.actorStore, key)
}

type presignStorage interface {
	PresignGetObject(ctx context.Context, key string, expires time.Duration) (string, error)
}

func presignIfExists(ctx context.Context, store presignStorage, key *string) (string, error) {
	if key == nil || *key == "" {
		return "", nil
	}

	url, err := store.PresignGetObject(ctx, *key, 0)
	if err != nil {
		return "", fmt.Errorf("%w: presign object key=%q: %v", domain.ErrInternal, *key, err)
	}

	return url, nil
}
