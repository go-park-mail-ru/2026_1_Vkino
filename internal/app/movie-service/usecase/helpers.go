package usecase

import (
	"context"
	"fmt"
	"strings"
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
		ContentType:        localizeMovieContentType(movie.ContentType),
		ReleaseYear:        movie.ReleaseYear,
		DurationSeconds:    movie.DurationSeconds,
		AgeLimit:           movie.AgeLimit,
		OriginalLanguageID: movie.OriginalLanguageID,
		OriginalLanguage:   movie.OriginalLanguage,
		Country:            movie.Country,
		Genres:             movie.Genres,
		Actors:             make([]domain.ActorShortResponse, 0, len(movie.Actors)),
		Episodes:           make([]domain.EpisodeResponse, 0, len(movie.Episodes)),
		ExternalRatings:    make([]domain.ExternalRating, 0, len(movie.ExternalRatings)),
		Reviews:            make([]domain.MovieReviewDTO, 0, len(movie.Reviews)),
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

	resp.Actors, err = u.buildMovieActorResponses(ctx, movie.Actors)
	if err != nil {
		return domain.MovieResponse{}, err
	}

	resp.Episodes, err = u.buildEpisodeResponses(ctx, movie.Episodes)
	if err != nil {
		return domain.MovieResponse{}, err
	}

	resp.ExternalRatings = append(resp.ExternalRatings, movie.ExternalRatings...)
	resp.Reviews = buildMovieReviewDTOs(movie.Reviews)

	return resp, nil
}

func (u *MovieUsecase) buildActorResponse(ctx context.Context, actor *domain.Actor) (domain.ActorResponse, error) {
	resp := domain.ActorResponse{
		ID:        actor.ID,
		FullName:  actor.FullName,
		Biography: actor.Biography,
		BirthDate: formatDate(actor.BirthDate),
		Country:   actor.Country,
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
		Rating: selection.Rating,
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

func (u *MovieUsecase) buildGenreShortResponses(genres []domain.GenreShort) []domain.GenreShortResponse {
	if len(genres) == 0 {
		return []domain.GenreShortResponse{}
	}

	result := make([]domain.GenreShortResponse, 0, len(genres))
	for _, genre := range genres {
		result = append(result, domain.GenreShortResponse(genre))
	}

	return result
}

func (u *MovieUsecase) buildMovieCardResponses(
	ctx context.Context,
	movies []domain.MovieCard,
) ([]domain.MovieCardResponse, error) {
	result := make([]domain.MovieCardResponse, 0, len(movies))
	for _, movie := range movies {
		card, err := u.buildMovieCardResponse(ctx, movie)
		if err != nil {
			return nil, err
		}

		result = append(result, card)
	}

	return result, nil
}

func (u *MovieUsecase) buildActorShortResponses(
	ctx context.Context,
	actors []domain.ActorShort,
) ([]domain.ActorShortResponse, error) {
	result := make([]domain.ActorShortResponse, 0, len(actors))
	for _, actor := range actors {
		item, err := u.buildActorShortResponse(ctx, actor)
		if err != nil {
			return nil, err
		}

		result = append(result, item)
	}

	return result, nil
}

func (u *MovieUsecase) buildMovieActorResponses(
	ctx context.Context,
	actors []domain.ActorShort,
) ([]domain.ActorShortResponse, error) {
	return u.buildActorShortResponses(ctx, actors)
}

func (u *MovieUsecase) buildEpisodeResponses(
	ctx context.Context,
	episodes []domain.Episode,
) ([]domain.EpisodeResponse, error) {
	result := make([]domain.EpisodeResponse, 0, len(episodes))
	for _, episode := range episodes {
		episodeImageURL, err := u.presignCard(ctx, episode.PictureFileKey)
		if err != nil {
			return nil, err
		}

		videoURL, err := u.presignVideo(ctx, episode.VideoFileKey)
		if err != nil {
			return nil, err
		}

		result = append(result, domain.EpisodeResponse{
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

	return result, nil
}

func buildMovieReviewDTOs(reviews []domain.MovieReview) []domain.MovieReviewDTO {
	result := make([]domain.MovieReviewDTO, 0, len(reviews))
	for _, review := range reviews {
		result = append(result, domain.MovieReviewDTO{
			ID:             review.ID,
			AuthorUserID:   review.AuthorUserID,
			Author:         maskEmail(review.AuthorEmail),
			Rating:         review.Rating,
			Comment:        review.Comment,
			LikesCount:     review.LikesCount,
			DislikesCount:  review.DislikesCount,
			ViewerReaction: review.ViewerReaction,
			CreatedAt:      review.CreatedAt.Format(time.RFC3339),
			UpdatedAt:      review.UpdatedAt.Format(time.RFC3339),
		})
	}

	return result
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

func localizeMovieContentType(contentType string) string {
	switch contentType {
	case "film":
		return "Фильм"
	case "series":
		return "Сериал"
	default:
		return contentType
	}
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
		return "", fmt.Errorf("%w: presign object key=%q: %w", domain.ErrInternal, key, err)
	}

	return url, nil
}

func formatDate(value *time.Time) string {
	if value == nil {
		return ""
	}

	return value.Format("2006-01-02")
}

func maskEmail(email string) string {
	at := strings.Index(email, "@")
	if at <= 0 {
		return "user"
	}

	local := email[:at]

	domainPart := email[at:]
	if len(local) <= 2 {
		return local[:1] + "***" + domainPart
	}

	return local[:2] + "***" + domainPart
}
