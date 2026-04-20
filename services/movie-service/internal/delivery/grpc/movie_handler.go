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
		Id:          movie.ID,
		Title:       movie.Title,
		Description: movie.Description,
		Year:        int32(movie.Year),
		Countries:   movie.Countries,
		Genres:      movie.Genres,
		AgeLimit:    int32(movie.AgeLimit),
		DurationMin: int32(movie.DurationMin),
		PosterUrl:   movie.PosterURL,
		CardUrl:     movie.CardURL,
		Actors:      make([]*moviev1.ActorShort, 0, len(movie.Actors)),
		Episodes:    make([]*moviev1.EpisodeShort, 0, len(movie.Episodes)),
	}

	for _, actor := range movie.Actors {
		resp.Actors = append(resp.Actors, &moviev1.ActorShort{
			Id:        actor.ID,
			Name:      actor.Name,
			AvatarUrl: actor.AvatarURL,
		})
	}

	for _, episode := range movie.Episodes {
		resp.Episodes = append(resp.Episodes, &moviev1.EpisodeShort{
			Id:          episode.ID,
			Number:      int32(episode.Number),
			Title:       episode.Title,
			DurationSec: int32(episode.DurationSec),
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
		Id:          actor.ID,
		Name:        actor.Name,
		Description: actor.Description,
		AvatarUrl:   actor.AvatarURL,
		Movies:      make([]*moviev1.MovieCard, 0, len(actor.Movies)),
	}

	for _, movie := range actor.Movies {
		resp.Movies = append(resp.Movies, &moviev1.MovieCard{
			Id:        movie.ID,
			Title:     movie.Title,
			Year:      int32(movie.Year),
			PosterUrl: movie.PosterURL,
			CardUrl:   movie.CardURL,
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
			Id:        movie.ID,
			Title:     movie.Title,
			Year:      int32(movie.Year),
			PosterUrl: movie.PosterURL,
			CardUrl:   movie.CardURL,
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
				Id:        movie.ID,
				Title:     movie.Title,
				Year:      int32(movie.Year),
				PosterUrl: movie.PosterURL,
				CardUrl:   movie.CardURL,
			})
		}

		resp.Selections = append(resp.Selections, item)
	}

	return resp, nil
}
