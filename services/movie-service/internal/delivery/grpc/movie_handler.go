package grpc

import (
	"context"

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

	resp := &moviev1.GetMovieByIDResponse{
		Id:                 movie.ID,
		Title:              movie.Title,
		Description:        movie.Description,
		Director:           movie.Director,
		TrailerUrl:         movie.TrailerURL,
		ContentType:        movie.ContentType,
		ReleaseYear:        int32(movie.ReleaseYear),
		DurationSeconds:    int32(movie.DurationSeconds),
		AgeLimit:           int32(movie.AgeLimit),
		OriginalLanguageId: movie.OriginalLanguageID,
		OriginalLanguage:   movie.OriginalLanguage,
		CountryId:          movie.CountryID,
		Country:            movie.Country,
		ImgUrl:             movie.PictureFileKey,
		PosterUrl:          movie.PosterFileKey,
		Genres:             movie.Genres,
		Actors:             make([]*moviev1.ActorShort, 0, len(movie.Actors)),
		Episodes:           make([]*moviev1.EpisodeShort, 0, len(movie.Episodes)),
	}

	for _, actor := range movie.Actors {
		resp.Actors = append(resp.Actors, &moviev1.ActorShort{
			Id:       actor.ID,
			FullName: actor.FullName,
			ImgUrl:   actor.PictureFileKey,
		})
	}

	for _, episode := range movie.Episodes {
		resp.Episodes = append(resp.Episodes, &moviev1.EpisodeShort{
			Id:              episode.ID,
			MovieId:         episode.MovieID,
			SeasonNumber:    int32(episode.SeasonNumber),
			EpisodeNumber:   int32(episode.EpisodeNumber),
			Title:           episode.Title,
			Description:     episode.Description,
			DurationSeconds: int32(episode.DurationSeconds),
			ImgUrl:          episode.PictureFileKey,
		})
	}

	return resp, nil
}

func (s *Server) GetActorByID(
	ctx context.Context,
	req *moviev1.GetActorByIDRequest,
) (*moviev1.GetActorByIDResponse, error) {
	actor, err := s.usecase.GetActorByID(ctx, req.GetActorId())
	if err != nil {
		return nil, mapError(err)
	}

	resp := &moviev1.GetActorByIDResponse{
		Id:        actor.ID,
		FullName:  actor.FullName,
		Biography: actor.Biography,
		Birthdate: actor.BirthDate,
		CountryId: actor.CountryID,
		ImgUrl:    actor.PictureFileKey,
		Movies:    make([]*moviev1.MovieCard, 0, len(actor.Movies)),
	}

	for _, movie := range actor.Movies {
		resp.Movies = append(resp.Movies, &moviev1.MovieCard{
			Id:     movie.ID,
			Title:  movie.Title,
			ImgUrl: movie.PictureFileKey,
		})
	}

	return resp, nil
}

func (s *Server) GetSelectionByTitle(
	ctx context.Context,
	req *moviev1.GetSelectionByTitleRequest,
) (*moviev1.GetSelectionByTitleResponse, error) {
	selection, err := s.usecase.GetSelectionByTitle(ctx, req.GetTitle())
	if err != nil {
		return nil, mapError(err)
	}

	resp := &moviev1.GetSelectionByTitleResponse{
		Title:  selection.Title,
		Movies: make([]*moviev1.MovieCard, 0, len(selection.Movies)),
	}

	for _, movie := range selection.Movies {
		resp.Movies = append(resp.Movies, &moviev1.MovieCard{
			Id:     movie.ID,
			Title:  movie.Title,
			ImgUrl: movie.PictureFileKey,
		})
	}

	return resp, nil
}

func (s *Server) GetAllSelections(
	ctx context.Context,
	_ *moviev1.GetAllSelectionsRequest,
) (*moviev1.GetAllSelectionsResponse, error) {
	selections, err := s.usecase.GetAllSelections(ctx)
	if err != nil {
		return nil, mapError(err)
	}

	resp := &moviev1.GetAllSelectionsResponse{
		Selections: make([]*moviev1.Selection, 0, len(selections)),
	}

	for _, selection := range selections {
		item := &moviev1.Selection{
			Title:  selection.Title,
			Movies: make([]*moviev1.MovieCard, 0, len(selection.Movies)),
		}

		for _, movie := range selection.Movies {
			item.Movies = append(item.Movies, &moviev1.MovieCard{
				Id:     movie.ID,
				Title:  movie.Title,
				ImgUrl: movie.PictureFileKey,
			})
		}

		resp.Selections = append(resp.Selections, item)
	}

	return resp, nil
}

func (s *Server) SearchMovies(
	ctx context.Context,
	req *moviev1.SearchMoviesRequest,
) (*moviev1.SearchMoviesResponse, error) {
	movies, err := s.usecase.SearchMovies(ctx, req.GetQuery())
	if err != nil {
		return nil, mapError(err)
	}

	resp := &moviev1.SearchMoviesResponse{
		Movies: make([]*moviev1.MovieCard, 0, len(movies)),
	}

	for _, movie := range movies {
		resp.Movies = append(resp.Movies, &moviev1.MovieCard{
			Id:     movie.ID,
			Title:  movie.Title,
			ImgUrl: movie.PictureFileKey,
		})
	}

	return resp, nil
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
		SeasonNumber:    int32(playback.SeasonNumber),
		EpisodeNumber:   int32(playback.EpisodeNumber),
		Title:           playback.Title,
		DurationSeconds: int32(playback.DurationSeconds),
		PlaybackUrl:     playback.PlaybackURL,
	}, nil
}

func (s *Server) GetEpisodeProgress(
	ctx context.Context,
	req *moviev1.GetEpisodeProgressRequest,
) (*moviev1.GetEpisodeProgressResponse, error) {
	progress, err := s.usecase.GetEpisodeProgress(ctx, req.GetUserId(), req.GetEpisodeId())
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
	progress, err := s.usecase.SaveEpisodeProgress(ctx, req.GetUserId(), req.GetEpisodeId(), req.GetPositionSeconds())
	if err != nil {
		return nil, mapError(err)
	}

	return &moviev1.SaveEpisodeProgressResponse{
		EpisodeId:       progress.EpisodeID,
		PositionSeconds: progress.PositionSeconds,
	}, nil
}