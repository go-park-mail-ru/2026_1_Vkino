package grpc

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/movie-service/domain"
	moviev1 "github.com/go-park-mail-ru/2026_1_VKino/platform/gen/movie/v1"
)

func (s *Server) GetMovieByID(
	ctx context.Context,
	req *moviev1.GetMovieByIDRequest,
) (*moviev1.GetMovieByIDResponse, error) {
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
	movies, err := s.usecase.SearchMovies(ctx, req.GetQuery())
	if err != nil {
		return nil, mapError(err)
	}

	return &moviev1.SearchMoviesResponse{
		Movies: mapMovieCards(movies),
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
