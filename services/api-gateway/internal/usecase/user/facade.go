package user

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	usergrpc "github.com/go-park-mail-ru/2026_1_VKino/services/api-gateway/internal/client/usergrpc"
	authmiddleware "github.com/go-park-mail-ru/2026_1_VKino/services/api-gateway/internal/delivery/http/middleware"
	"github.com/go-park-mail-ru/2026_1_VKino/services/api-gateway/internal/dto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Facade struct {
	userClient usergrpc.Client
}

func NewFacade(userClient usergrpc.Client) *Facade {
	return &Facade{
		userClient: userClient,
	}
}

func (f *Facade) GetProfile(ctx context.Context, r *http.Request) (dto.ProfileResponse, int, error) {
	authCtx, err := authmiddleware.AuthFromContext(r.Context())
	if err != nil {
		return dto.ProfileResponse{}, http.StatusUnauthorized, errors.New("unauthorized")
	}

	resp, err := f.userClient.GetProfile(ctx, authCtx.UserID)
	if err != nil {
		statusCode, mappedErr := grpcToHTTPError(err)
		return dto.ProfileResponse{}, statusCode, mappedErr
	}

	return resp, http.StatusOK, nil
}

func (f *Facade) UpdateProfile(
	ctx context.Context,
	r *http.Request,
	req dto.UpdateProfileRequest,
) (dto.ProfileResponse, int, error) {
	authCtx, err := authmiddleware.AuthFromContext(r.Context())
	if err != nil {
		return dto.ProfileResponse{}, http.StatusUnauthorized, errors.New("unauthorized")
	}

	resp, err := f.userClient.UpdateProfile(ctx, authCtx.UserID, req.Birthdate)
	if err != nil {
		statusCode, mappedErr := grpcToHTTPError(err)
		return dto.ProfileResponse{}, statusCode, mappedErr
	}

	return resp, http.StatusOK, nil
}

func (f *Facade) SearchUsersByEmail(ctx context.Context, r *http.Request) (dto.SearchUsersResponse, int, error) {
	authCtx, err := authmiddleware.AuthFromContext(r.Context())
	if err != nil {
		return dto.SearchUsersResponse{}, http.StatusUnauthorized, errors.New("unauthorized")
	}

	query := r.URL.Query().Get("email")
	users, err := f.userClient.SearchUsersByEmail(ctx, authCtx.UserID, query)
	if err != nil {
		statusCode, mappedErr := grpcToHTTPError(err)
		return dto.SearchUsersResponse{}, statusCode, mappedErr
	}

	return dto.SearchUsersResponse{Users: users}, http.StatusOK, nil
}

func (f *Facade) AddFriend(
	ctx context.Context,
	r *http.Request,
	req dto.AddFriendRequest,
) (dto.FriendResponse, int, error) {
	authCtx, err := authmiddleware.AuthFromContext(r.Context())
	if err != nil {
		return dto.FriendResponse{}, http.StatusUnauthorized, errors.New("unauthorized")
	}

	resp, err := f.userClient.AddFriend(ctx, authCtx.UserID, req.FriendID)
	if err != nil {
		statusCode, mappedErr := grpcToHTTPError(err)
		return dto.FriendResponse{}, statusCode, mappedErr
	}

	return resp, http.StatusOK, nil
}

func (f *Facade) DeleteFriend(
	ctx context.Context,
	r *http.Request,
	req dto.DeleteFriendRequest,
) (dto.DeleteFriendResponse, int, error) {
	authCtx, err := authmiddleware.AuthFromContext(r.Context())
	if err != nil {
		return dto.DeleteFriendResponse{}, http.StatusUnauthorized, errors.New("unauthorized")
	}

	err = f.userClient.DeleteFriend(ctx, authCtx.UserID, req.FriendID)
	if err != nil {
		statusCode, mappedErr := grpcToHTTPError(err)
		return dto.DeleteFriendResponse{}, statusCode, mappedErr
	}

	return dto.DeleteFriendResponse{Success: true}, http.StatusOK, nil
}

func (f *Facade) AddMovieToFavorites(
	ctx context.Context,
	r *http.Request,
	req dto.AddMovieToFavoritesRequest,
) (dto.FavoriteMovieResponse, int, error) {
	authCtx, err := authmiddleware.AuthFromContext(r.Context())
	if err != nil {
		return dto.FavoriteMovieResponse{}, http.StatusUnauthorized, errors.New("unauthorized")
	}

	resp, err := f.userClient.AddMovieToFavorites(ctx, authCtx.UserID, req.MovieID)
	if err != nil {
		statusCode, mappedErr := grpcToHTTPError(err)
		return dto.FavoriteMovieResponse{}, statusCode, mappedErr
	}

	return resp, http.StatusOK, nil
}

func (f *Facade) AddMovieToFavoritesFromPath(
	ctx context.Context,
	r *http.Request,
) (dto.FavoriteMovieResponse, int, error) {
	movieIDStr := r.PathValue("id")
	movieID, err := strconv.ParseInt(movieIDStr, 10, 64)
	if err != nil {
		return dto.FavoriteMovieResponse{}, http.StatusBadRequest, errors.New("invalid movie id")
	}

	return f.AddMovieToFavorites(ctx, r, dto.AddMovieToFavoritesRequest{
		MovieID: movieID,
	})
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
	case codes.AlreadyExists:
		return http.StatusConflict, errors.New(st.Message())
	case codes.NotFound:
		return http.StatusNotFound, errors.New(st.Message())
	case codes.InvalidArgument:
		return http.StatusBadRequest, errors.New(st.Message())
	case codes.Unauthenticated, codes.PermissionDenied:
		return http.StatusUnauthorized, errors.New("unauthorized")
	case codes.FailedPrecondition:
		return http.StatusPreconditionFailed, errors.New(st.Message())
	case codes.Unavailable:
		return http.StatusBadGateway, errors.New("user service unavailable")
	default:
		return http.StatusInternalServerError, errors.New("internal server error")
	}
}
