package grpc

import (
	"bytes"
	"context"

	userv1 "github.com/go-park-mail-ru/2026_1_VKino/platform/gen/user/v1"
)

func (s *Server) GetProfile(ctx context.Context, req *userv1.GetProfileRequest) (*userv1.GetProfileResponse, error) {
	profile, err := s.usecase.GetProfile(ctx, req.GetUserId())
	if err != nil {
		return nil, mapError(err)
	}

	resp := &userv1.GetProfileResponse{
		Email:     profile.Email,
		AvatarUrl: profile.AvatarURL,
	}

	if profile.Birthdate != nil {
		resp.Birthdate = *profile.Birthdate
	}

	return resp, nil
}

func (s *Server) UpdateProfile(ctx context.Context, req *userv1.UpdateProfileRequest) (*userv1.UpdateProfileResponse, error) {
	var body *bytes.Reader
	
	var size int64

	if len(req.GetAvatar()) > 0 {
		body = bytes.NewReader(req.GetAvatar())
		size = int64(len(req.GetAvatar()))
	}

	profile, err := s.usecase.UpdateProfile(
		ctx,
		req.GetUserId(),
		req.GetBirthdate(),
		body,
		size,
		req.GetAvatarContentType(),
	)
	if err != nil {
		return nil, mapError(err)
	}

	resp := &userv1.UpdateProfileResponse{
		Email:     profile.Email,
		AvatarUrl: profile.AvatarURL,
	}

	if profile.Birthdate != nil {
		resp.Birthdate = *profile.Birthdate
	}

	return resp, nil
}

func (s *Server) SearchUsersByEmail(
	ctx context.Context,
	req *userv1.SearchUsersByEmailRequest,
) (*userv1.SearchUsersByEmailResponse, error) {
	users, err := s.usecase.SearchUsersByEmail(ctx, req.GetUserId(), req.GetEmailQuery())
	if err != nil {
		return nil, mapError(err)
	}

	resp := &userv1.SearchUsersByEmailResponse{
		Users: make([]*userv1.UserSearchResult, 0, len(users)),
	}

	for _, user := range users {
		resp.Users = append(resp.Users, &userv1.UserSearchResult{
			Id:       user.ID,
			Email:    user.Email,
			IsFriend: user.IsFriend,
		})
	}

	return resp, nil
}

func (s *Server) AddFriend(ctx context.Context, req *userv1.AddFriendRequest) (*userv1.AddFriendResponse, error) {
	friend, err := s.usecase.AddFriend(ctx, req.GetUserId(), req.GetFriendId())
	if err != nil {
		return nil, mapError(err)
	}

	return &userv1.AddFriendResponse{
		Id:    friend.ID,
		Email: friend.Email,
	}, nil
}

func (s *Server) DeleteFriend(
	ctx context.Context,
	req *userv1.DeleteFriendRequest,
) (*userv1.DeleteFriendResponse, error) {
	err := s.usecase.DeleteFriend(ctx, req.GetUserId(), req.GetFriendId())
	if err != nil {
		return nil, mapError(err)
	}

	return &userv1.DeleteFriendResponse{Success: true}, nil
}

func (s *Server) AddMovieToFavorites(
	ctx context.Context,
	req *userv1.AddMovieToFavoritesRequest,
) (*userv1.AddMovieToFavoritesResponse, error) {
	favorite, err := s.usecase.AddMovieToFavorites(ctx, req.GetUserId(), req.GetMovieId())
	if err != nil {
		return nil, mapError(err)
	}

	return &userv1.AddMovieToFavoritesResponse{
		MovieId:    favorite.MovieID,
		IsFavorite: favorite.IsFavorite,
	}, nil
}
