package grpc

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/movie-service/domain"
	moviev1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/movie/v1"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/service/authctx"
)

func (s *Server) GetMovieByID(
	ctx context.Context,
	req *moviev1.GetMovieByIDRequest,
) (*moviev1.GetMovieByIDResponse, error) {
	if authCtx, err := s.authorize(ctx); err == nil {
		ctx = authctx.WithContext(ctx, authCtx)
	}

	movie, err := s.usecase.GetMovieByID(ctx, req.GetMovieId())
	if err != nil {
		return nil, mapError(err)
	}

	return &moviev1.GetMovieByIDResponse{
		Id:                 movie.ID,
		Title:              movie.Title,
		Description:        movie.Description,
		Director:           movie.Director,
		TrailerUrl:         movie.TrailerURL,
		ContentType:        movie.ContentType,
		ReleaseYear:        i32(movie.ReleaseYear),
		DurationSeconds:    i32(movie.DurationSeconds),
		AgeLimit:           i32(movie.AgeLimit),
		OriginalLanguageId: movie.OriginalLanguageID,
		OriginalLanguage:   movie.OriginalLanguage,
		CountryId:          movie.CountryID,
		Country:            movie.Country,
		ImgUrl:             movie.PictureFileKey,
		PosterUrl:          movie.PosterFileKey,
		Genres:             movie.Genres,
		Actors:             mapActorShorts(movie.Actors),
		Episodes:           mapEpisodeShorts(movie.Episodes),
		IsFavorite:         movie.IsFavorite,
	}, nil
}

func (s *Server) GetMoviesByIDs(
	ctx context.Context,
	req *moviev1.GetMoviesByIDsRequest,
) (*moviev1.GetMoviesByIDsResponse, error) {
	movies, err := s.usecase.GetMoviesByIDs(ctx, req.GetMovieIds())
	if err != nil {
		return nil, mapError(err)
	}

	return &moviev1.GetMoviesByIDsResponse{
		Movies: mapMovieCards(movies),
	}, nil
}

func (s *Server) GetActorByID(
	ctx context.Context,
	req *moviev1.GetActorByIDRequest,
) (*moviev1.GetActorByIDResponse, error) {
	actor, err := s.usecase.GetActorByID(ctx, req.GetActorId())
	if err != nil {
		return nil, mapError(err)
	}

	return &moviev1.GetActorByIDResponse{
		Id:        actor.ID,
		FullName:  actor.FullName,
		Biography: actor.Biography,
		Birthdate: actor.BirthDate,
		CountryId: actor.CountryID,
		ImgUrl:    actor.PictureFileKey,
		Movies:    mapMovieCards(actor.Movies),
	}, nil
}

func (s *Server) GetGenreByID(
	ctx context.Context,
	req *moviev1.GetGenreByIDRequest,
) (*moviev1.GetGenreByIDResponse, error) {
	genre, err := s.usecase.GetGenreByID(ctx, req.GetGenreId())
	if err != nil {
		return nil, mapError(err)
	}

	return &moviev1.GetGenreByIDResponse{
		Id:     genre.ID,
		Title:  genre.Title,
		Movies: mapMovieCards(genre.Movies),
	}, nil
}

func (s *Server) GetAllGenres(
	ctx context.Context,
	_ *moviev1.GetAllGenresRequest,
) (*moviev1.GetAllGenresResponse, error) {
	genres, err := s.usecase.GetAllGenres(ctx)
	if err != nil {
		return nil, mapError(err)
	}

	return &moviev1.GetAllGenresResponse{
		Genres: mapGenreShorts(genres),
	}, nil
}

func (s *Server) GetSelectionByTitle(
	ctx context.Context,
	req *moviev1.GetSelectionByTitleRequest,
) (*moviev1.GetSelectionByTitleResponse, error) {
	selection, err := s.usecase.GetSelectionByTitle(ctx, req.GetTitle())
	if err != nil {
		return nil, mapError(err)
	}

	return &moviev1.GetSelectionByTitleResponse{
		Title:  selection.Title,
		Movies: mapMovieCards(selection.Movies),
	}, nil
}

func (s *Server) GetAllSelections(
	ctx context.Context,
	_ *moviev1.GetAllSelectionsRequest,
) (*moviev1.GetAllSelectionsResponse, error) {
	selections, err := s.usecase.GetAllSelections(ctx)
	if err != nil {
		return nil, mapError(err)
	}

	return &moviev1.GetAllSelectionsResponse{
		Selections: mapSelections(selections),
	}, nil
}

func (s *Server) SearchMovies(
	ctx context.Context,
	req *moviev1.SearchMoviesRequest,
) (*moviev1.SearchMoviesResponse, error) {
	result, err := s.usecase.SearchMovies(ctx, req.GetQuery())
	if err != nil {
		return nil, mapError(err)
	}

	return &moviev1.SearchMoviesResponse{
		Movies: mapMovieCards(result.Movies),
		Actors: mapActorShorts(result.Actors),
	}, nil
}

func (s *Server) GetEpisodePlayback(
	ctx context.Context,
	req *moviev1.GetEpisodePlaybackRequest,
) (*moviev1.GetEpisodePlaybackResponse, error) {
	playback, err := s.usecase.GetEpisodePlayback(ctx, req.GetEpisodeId())
	if err != nil {
		return nil, mapError(err)
	}

	return &moviev1.GetEpisodePlaybackResponse{
		EpisodeId:       playback.EpisodeID,
		MovieId:         playback.MovieID,
		SeasonNumber:    i32(playback.SeasonNumber),
		EpisodeNumber:   i32(playback.EpisodeNumber),
		Title:           playback.Title,
		DurationSeconds: i32(playback.DurationSeconds),
		PlaybackUrl:     playback.PlaybackURL,
	}, nil
}

func (s *Server) GetEpisodeProgress(
	ctx context.Context,
	req *moviev1.GetEpisodeProgressRequest,
) (*moviev1.GetEpisodeProgressResponse, error) {
	authCtx, err := s.authorize(ctx)
	if err != nil {
		return nil, err
	}

	progress, err := s.usecase.GetEpisodeProgress(ctx, authCtx.UserID, req.GetEpisodeId())
	if err != nil {
		return nil, mapError(err)
	}

	return &moviev1.GetEpisodeProgressResponse{
		EpisodeId:       progress.EpisodeID,
		PositionSeconds: progress.PositionSeconds,
	}, nil
}

func (s *Server) SaveEpisodeProgress(
	ctx context.Context,
	req *moviev1.SaveEpisodeProgressRequest,
) (*moviev1.SaveEpisodeProgressResponse, error) {
	authCtx, err := s.authorize(ctx)
	if err != nil {
		return nil, err
	}

	progress, err := s.usecase.SaveEpisodeProgress(
		ctx,
		authCtx.UserID,
		req.GetEpisodeId(),
		req.GetPositionSeconds(),
	)
	if err != nil {
		return nil, mapError(err)
	}

	return &moviev1.SaveEpisodeProgressResponse{
		EpisodeId:       progress.EpisodeID,
		PositionSeconds: progress.PositionSeconds,
	}, nil
}

func (s *Server) GetContinueWatching(
	ctx context.Context,
	req *moviev1.GetContinueWatchingRequest,
) (*moviev1.GetContinueWatchingResponse, error) {
	authCtx, err := s.authorize(ctx)
	if err != nil {
		return nil, err
	}

	items, err := s.usecase.GetContinueWatching(ctx, authCtx.UserID, req.GetLimit())
	if err != nil {
		return nil, mapError(err)
	}

	return &moviev1.GetContinueWatchingResponse{
		Items: mapWatchProgressItems(items),
	}, nil
}

func (s *Server) GetWatchHistory(
	ctx context.Context,
	req *moviev1.GetWatchHistoryRequest,
) (*moviev1.GetWatchHistoryResponse, error) {
	authCtx, err := s.authorize(ctx)
	if err != nil {
		return nil, err
	}

	items, err := s.usecase.GetWatchHistory(ctx, authCtx.UserID, req.GetLimit(), req.GetMinProgress())
	if err != nil {
		return nil, mapError(err)
	}

	return &moviev1.GetWatchHistoryResponse{
		Items: mapWatchProgressItems(items),
	}, nil
}

func mapActorShorts(actors []domain.ActorShortResponse) []*moviev1.ActorShort {
	if len(actors) == 0 {
		return []*moviev1.ActorShort{}
	}

	result := make([]*moviev1.ActorShort, 0, len(actors))
	for _, actor := range actors {
		result = append(result, &moviev1.ActorShort{
			Id:       actor.ID,
			FullName: actor.FullName,
			ImgUrl:   actor.PictureFileKey,
		})
	}

	return result
}

func mapEpisodeShorts(episodes []domain.EpisodeResponse) []*moviev1.EpisodeShort {
	if len(episodes) == 0 {
		return []*moviev1.EpisodeShort{}
	}

	result := make([]*moviev1.EpisodeShort, 0, len(episodes))
	for _, episode := range episodes {
		result = append(result, &moviev1.EpisodeShort{
			Id:              episode.ID,
			MovieId:         episode.MovieID,
			SeasonNumber:    i32(episode.SeasonNumber),
			EpisodeNumber:   i32(episode.EpisodeNumber),
			Title:           episode.Title,
			Description:     episode.Description,
			DurationSeconds: i32(episode.DurationSeconds),
			ImgUrl:          episode.PictureFileKey,
		})
	}

	return result
}

func mapGenreShorts(genres []domain.GenreShortResponse) []*moviev1.GenreShort {
	if len(genres) == 0 {
		return []*moviev1.GenreShort{}
	}

	result := make([]*moviev1.GenreShort, 0, len(genres))
	for _, genre := range genres {
		result = append(result, &moviev1.GenreShort{
			Id:    genre.ID,
			Title: genre.Title,
		})
	}

	return result
}

func mapMovieCards(movies []domain.MovieCardResponse) []*moviev1.MovieCard {
	if len(movies) == 0 {
		return []*moviev1.MovieCard{}
	}

	result := make([]*moviev1.MovieCard, 0, len(movies))
	for _, movie := range movies {
		result = append(result, &moviev1.MovieCard{
			Id:     movie.ID,
			Title:  movie.Title,
			ImgUrl: movie.PictureFileKey,
		})
	}

	return result
}

func mapSelections(selections []domain.SelectionResponse) []*moviev1.Selection {
	if len(selections) == 0 {
		return []*moviev1.Selection{}
	}

	result := make([]*moviev1.Selection, 0, len(selections))
	for _, selection := range selections {
		result = append(result, &moviev1.Selection{
			Title:  selection.Title,
			Movies: mapMovieCards(selection.Movies),
		})
	}

	return result
}

func i32(v int) int32 {
	return int32(v)
}

func mapWatchProgressItems(items []domain.WatchProgressItemResponse) []*moviev1.WatchProgressItem {
	result := make([]*moviev1.WatchProgressItem, 0, len(items))
	for _, item := range items {
		result = append(result, &moviev1.WatchProgressItem{
			EpisodeId:       item.EpisodeID,
			MovieId:         item.MovieID,
			MovieTitle:      item.MovieTitle,
			PosterUrl:       item.PosterURL,
			ContentType:     item.ContentType,
			SeasonNumber:    i32(item.SeasonNumber),
			EpisodeNumber:   i32(item.EpisodeNumber),
			EpisodeTitle:    item.EpisodeTitle,
			PositionSeconds: item.PositionSeconds,
			DurationSeconds: item.DurationSeconds,
			UpdatedAt:       item.UpdatedAt,
		})
	}

	return result
}
