package moviegrpc

import (
	"context"
	"fmt"
	"time"

	moviev1 "github.com/go-park-mail-ru/2026_1_VKino/platform/gen/movie/v1"
	"github.com/go-park-mail-ru/2026_1_VKino/services/api-gateway/internal/dto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Config struct {
	Address        string
	RequestTimeout time.Duration
}

type Client interface {
	GetMovieByID(ctx context.Context, movieID int64) (dto.MovieResponse, error)
	GetActorByID(ctx context.Context, actorID int64) (dto.ActorResponse, error)
	GetSelectionByTitle(ctx context.Context, title string) (dto.SelectionResponse, error)
	GetAllSelections(ctx context.Context) ([]dto.SelectionResponse, error)
	Close() error
}

type GRPCClient struct {
	conn    *grpc.ClientConn
	client  moviev1.MovieServiceClient
	timeout time.Duration
}

func New(ctx context.Context, cfg Config) (*GRPCClient, error) {
	timeout := cfg.RequestTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	conn, err := grpc.DialContext(
		ctx,
		cfg.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("dial movie grpc: %w", err)
	}

	return &GRPCClient{
		conn:    conn,
		client:  moviev1.NewMovieServiceClient(conn),
		timeout: timeout,
	}, nil
}

func (c *GRPCClient) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func (c *GRPCClient) GetMovieByID(ctx context.Context, movieID int64) (dto.MovieResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.client.GetMovieByID(ctx, &moviev1.GetMovieByIDRequest{
		MovieId: movieID,
	})
	if err != nil {
		return dto.MovieResponse{}, err
	}

	result := dto.MovieResponse{
		ID:          resp.GetId(),
		Title:       resp.GetTitle(),
		Description: resp.GetDescription(),
		Year:        int(resp.GetYear()),
		Countries:   resp.GetCountries(),
		Genres:      resp.GetGenres(),
		AgeLimit:    int(resp.GetAgeLimit()),
		DurationMin: int(resp.GetDurationMin()),
		PosterURL:   resp.GetPosterUrl(),
		CardURL:     resp.GetCardUrl(),
		Actors:      make([]dto.ActorShortResponse, 0, len(resp.GetActors())),
		Episodes:    make([]dto.EpisodeResponse, 0, len(resp.GetEpisodes())),
	}

	for _, actor := range resp.GetActors() {
		result.Actors = append(result.Actors, dto.ActorShortResponse{
			ID:        actor.GetId(),
			Name:      actor.GetName(),
			AvatarURL: actor.GetAvatarUrl(),
		})
	}

	for _, episode := range resp.GetEpisodes() {
		result.Episodes = append(result.Episodes, dto.EpisodeResponse{
			ID:          episode.GetId(),
			Number:      int(episode.GetNumber()),
			Title:       episode.GetTitle(),
			DurationSec: int(episode.GetDurationSec()),
		})
	}

	return result, nil
}

func (c *GRPCClient) GetActorByID(ctx context.Context, actorID int64) (dto.ActorResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.client.GetActorByID(ctx, &moviev1.GetActorByIDRequest{
		ActorId: actorID,
	})
	if err != nil {
		return dto.ActorResponse{}, err
	}

	result := dto.ActorResponse{
		ID:          resp.GetId(),
		Name:        resp.GetName(),
		Description: resp.GetDescription(),
		AvatarURL:   resp.GetAvatarUrl(),
		Movies:      make([]dto.MovieCardResponse, 0, len(resp.GetMovies())),
	}

	for _, movie := range resp.GetMovies() {
		result.Movies = append(result.Movies, dto.MovieCardResponse{
			ID:        movie.GetId(),
			Title:     movie.GetTitle(),
			Year:      int(movie.GetYear()),
			PosterURL: movie.GetPosterUrl(),
			CardURL:   movie.GetCardUrl(),
		})
	}

	return result, nil
}

func (c *GRPCClient) GetSelectionByTitle(ctx context.Context, title string) (dto.SelectionResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.client.GetSelectionByTitle(ctx, &moviev1.GetSelectionByTitleRequest{
		Title: title,
	})
	if err != nil {
		return dto.SelectionResponse{}, err
	}

	result := dto.SelectionResponse{
		Title:  resp.GetTitle(),
		Movies: make([]dto.MovieCardResponse, 0, len(resp.GetMovies())),
	}

	for _, movie := range resp.GetMovies() {
		result.Movies = append(result.Movies, dto.MovieCardResponse{
			ID:        movie.GetId(),
			Title:     movie.GetTitle(),
			Year:      int(movie.GetYear()),
			PosterURL: movie.GetPosterUrl(),
			CardURL:   movie.GetCardUrl(),
		})
	}

	return result, nil
}

func (c *GRPCClient) GetAllSelections(ctx context.Context) ([]dto.SelectionResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.client.GetAllSelections(ctx, &moviev1.GetAllSelectionsRequest{})
	if err != nil {
		return nil, err
	}

	result := make([]dto.SelectionResponse, 0, len(resp.GetSelections()))
	for _, selection := range resp.GetSelections() {
		item := dto.SelectionResponse{
			Title:  selection.GetTitle(),
			Movies: make([]dto.MovieCardResponse, 0, len(selection.GetMovies())),
		}

		for _, movie := range selection.GetMovies() {
			item.Movies = append(item.Movies, dto.MovieCardResponse{
				ID:        movie.GetId(),
				Title:     movie.GetTitle(),
				Year:      int(movie.GetYear()),
				PosterURL: movie.GetPosterUrl(),
				CardURL:   movie.GetCardUrl(),
			})
		}

		result = append(result, item)
	}

	return result, nil
}
