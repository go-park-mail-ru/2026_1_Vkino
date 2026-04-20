package movie

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	moviegrpc "github.com/go-park-mail-ru/2026_1_VKino/services/api-gateway/internal/client/moviegrpc"
	"github.com/go-park-mail-ru/2026_1_VKino/services/api-gateway/internal/dto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Facade struct {
	movieClient moviegrpc.Client
}

func NewFacade(movieClient moviegrpc.Client) *Facade {
	return &Facade{
		movieClient: movieClient,
	}
}

func (f *Facade) GetMovieByID(ctx context.Context, r *http.Request) (dto.MovieResponse, int, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		return dto.MovieResponse{}, http.StatusBadRequest, errors.New("invalid movie id")
	}

	resp, err := f.movieClient.GetMovieByID(ctx, id)
	if err != nil {
		statusCode, mappedErr := grpcToHTTPError(err)
		return dto.MovieResponse{}, statusCode, mappedErr
	}

	return resp, http.StatusOK, nil
}

func (f *Facade) GetActorByID(ctx context.Context, r *http.Request) (dto.ActorResponse, int, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		return dto.ActorResponse{}, http.StatusBadRequest, errors.New("invalid actor id")
	}

	resp, err := f.movieClient.GetActorByID(ctx, id)
	if err != nil {
		statusCode, mappedErr := grpcToHTTPError(err)
		return dto.ActorResponse{}, statusCode, mappedErr
	}

	return resp, http.StatusOK, nil
}

func (f *Facade) GetSelectionByTitle(ctx context.Context, r *http.Request) (dto.SelectionResponse, int, error) {
	title := strings.TrimSpace(r.PathValue("selection"))
	if title == "" {
		return dto.SelectionResponse{}, http.StatusBadRequest, errors.New("invalid selection title")
	}

	resp, err := f.movieClient.GetSelectionByTitle(ctx, title)
	if err != nil {
		statusCode, mappedErr := grpcToHTTPError(err)
		return dto.SelectionResponse{}, statusCode, mappedErr
	}

	return resp, http.StatusOK, nil
}

func (f *Facade) GetAllSelections(ctx context.Context) (dto.SelectionsResponse, int, error) {
	resp, err := f.movieClient.GetAllSelections(ctx)
	if err != nil {
		statusCode, mappedErr := grpcToHTTPError(err)
		return dto.SelectionsResponse{}, statusCode, mappedErr
	}

	return dto.SelectionsResponse{Selections: resp}, http.StatusOK, nil
}

func grpcToHTTPError(err error) (int, error) {
	if err == nil {
		return http.StatusOK, nil
	}

	st, ok := status.FromError(err)
	if !ok {
		return http.StatusInternalServerError, errors.New("internal server error")
	}

	switch st.Code() {
	case codes.NotFound:
		return http.StatusNotFound, errors.New(st.Message())
	case codes.InvalidArgument:
		return http.StatusBadRequest, errors.New(st.Message())
	case codes.Unavailable:
		return http.StatusBadGateway, errors.New("movie service unavailable")
	default:
		return http.StatusInternalServerError, errors.New("internal server error")
	}
}
