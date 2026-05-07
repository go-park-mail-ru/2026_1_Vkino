package grpc

import (
	"bytes"
	"context"

	userv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/user/v1"
)

func (s *Server) GetProfile(ctx context.Context, req *userv1.GetProfileRequest) (*userv1.GetProfileResponse, error) {
	authCtx, err := s.authorize(ctx)
	if err != nil {
		return nil, err
	}

	profile, err := s.usecase.GetProfile(ctx, authCtx.UserID)
	if err != nil {
		return nil, mapError(err)
	}

	resp := &userv1.GetProfileResponse{
		Email:     profile.Email,
		AvatarUrl: profile.AvatarURL,
		Role:      profile.Role,
	}

	if profile.Birthdate != nil {
		resp.Birthdate = *profile.Birthdate
	}

	return resp, nil
}

func (s *Server) UpdateProfile(
	ctx context.Context,
	req *userv1.UpdateProfileRequest,
) (*userv1.UpdateProfileResponse, error) {
	authCtx, err := s.authorize(ctx)
	if err != nil {
		return nil, err
	}

	var body *bytes.Reader

	var size int64

	if len(req.GetAvatar()) > 0 {
		body = bytes.NewReader(req.GetAvatar())
		size = int64(len(req.GetAvatar()))
	}

	profile, err := s.usecase.UpdateProfile(
		ctx,
		authCtx.UserID,
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
	authCtx, err := s.authorize(ctx)
	if err != nil {
		return nil, err
	}

	users, err := s.usecase.SearchUsersByEmail(ctx, authCtx.UserID, req.GetEmailQuery())
	if err != nil {
		return nil, mapError(err)
	}

	resp := &userv1.SearchUsersByEmailResponse{
		Users: make([]*userv1.UserSearchResult, 0, len(users)),
	}

	for _, user := range users {
		resp.Users = append(resp.Users, &userv1.UserSearchResult{
			Id:        user.ID,
			Email:     user.Email,
			AvatarUrl: user.AvatarURL,
			IsFriend:  user.IsFriend,
		})
	}

	return resp, nil
}

func (s *Server) AddFriend(ctx context.Context, req *userv1.AddFriendRequest) (*userv1.AddFriendResponse, error) {
	authCtx, err := s.authorize(ctx)
	if err != nil {
		return nil, err
	}

	friend, err := s.usecase.AddFriend(ctx, authCtx.UserID, req.GetFriendId())
	if err != nil {
		return nil, mapError(err)
	}

	return &userv1.AddFriendResponse{
		Id:        friend.ID,
		Email:     friend.Email,
		AvatarUrl: friend.AvatarURL,
	}, nil
}

func (s *Server) DeleteFriend(
	ctx context.Context,
	req *userv1.DeleteFriendRequest,
) (*userv1.DeleteFriendResponse, error) {
	authCtx, err := s.authorize(ctx)
	if err != nil {
		return nil, err
	}

	err = s.usecase.DeleteFriend(ctx, authCtx.UserID, req.GetFriendId())
	if err != nil {
		return nil, mapError(err)
	}

	return &userv1.DeleteFriendResponse{Success: true}, nil
}

func (s *Server) AddMovieToFavorites(
	ctx context.Context,
	req *userv1.AddMovieToFavoritesRequest,
) (*userv1.AddMovieToFavoritesResponse, error) {
	authCtx, err := s.authorize(ctx)
	if err != nil {
		return nil, err
	}

	favorite, err := s.usecase.AddMovieToFavorites(ctx, authCtx.UserID, req.GetMovieId())
	if err != nil {
		return nil, mapError(err)
	}

	return &userv1.AddMovieToFavoritesResponse{
		MovieId:    favorite.MovieID,
		IsFavorite: favorite.IsFavorite,
	}, nil
}

func (s *Server) SetMovieRating(
	ctx context.Context,
	req *userv1.SetMovieRatingRequest,
) (*userv1.SetMovieRatingResponse, error) {
	authCtx, err := s.authorize(ctx)
	if err != nil {
		return nil, err
	}

	ratedMovie, err := s.usecase.SetMovieRating(ctx, authCtx.UserID, req.GetMovieId(), req.GetRating())
	if err != nil {
		return nil, mapError(err)
	}

	return &userv1.SetMovieRatingResponse{
		MovieId: ratedMovie.MovieID,
		Rating:  ratedMovie.Rating,
	}, nil
}

func (s *Server) SetMovieReview(
	ctx context.Context,
	req *userv1.SetMovieReviewRequest,
) (*userv1.SetMovieReviewResponse, error) {
	authCtx, err := s.authorize(ctx)
	if err != nil {
		return nil, err
	}

	review, err := s.usecase.SetMovieReview(ctx, authCtx.UserID, req.GetMovieId(), req.Rating, req.Comment)
	if err != nil {
		return nil, mapError(err)
	}

	return &userv1.SetMovieReviewResponse{
		ReviewId: review.ReviewID,
		MovieId:  review.MovieID,
		Rating:   review.Rating,
		Comment:  review.Comment,
	}, nil
}

func (s *Server) DeleteMovieReview(
	ctx context.Context,
	req *userv1.DeleteMovieReviewRequest,
) (*userv1.DeleteMovieReviewResponse, error) {
	authCtx, err := s.authorize(ctx)
	if err != nil {
		return nil, err
	}

	if err = s.usecase.DeleteMovieReview(ctx, authCtx.UserID, req.GetMovieId()); err != nil {
		return nil, mapError(err)
	}

	return &userv1.DeleteMovieReviewResponse{Success: true}, nil
}

func (s *Server) SetReviewReaction(
	ctx context.Context,
	req *userv1.SetReviewReactionRequest,
) (*userv1.SetReviewReactionResponse, error) {
	authCtx, err := s.authorize(ctx)
	if err != nil {
		return nil, err
	}

	reaction, err := s.usecase.SetReviewReaction(ctx, authCtx.UserID, req.GetReviewId(), req.GetReaction())
	if err != nil {
		return nil, mapError(err)
	}

	return &userv1.SetReviewReactionResponse{
		ReviewId: reaction.ReviewID,
		Reaction: reaction.Reaction,
	}, nil
}

func (s *Server) DeleteReviewReaction(
	ctx context.Context,
	req *userv1.DeleteReviewReactionRequest,
) (*userv1.DeleteReviewReactionResponse, error) {
	authCtx, err := s.authorize(ctx)
	if err != nil {
		return nil, err
	}

	if err = s.usecase.DeleteReviewReaction(ctx, authCtx.UserID, req.GetReviewId()); err != nil {
		return nil, mapError(err)
	}

	return &userv1.DeleteReviewReactionResponse{Success: true}, nil
}

func (s *Server) ToggleFavorite(
	ctx context.Context,
	req *userv1.ToggleFavoriteRequest,
) (*userv1.ToggleFavoriteResponse, error) {
	authCtx, err := s.authorize(ctx)
	if err != nil {
		return nil, err
	}

	favorite, err := s.usecase.ToggleFavorite(ctx, authCtx.UserID, req.GetMovieId())
	if err != nil {
		return nil, mapError(err)
	}

	return &userv1.ToggleFavoriteResponse{
		MovieId:    favorite.MovieID,
		IsFavorite: favorite.IsFavorite,
	}, nil
}

func (s *Server) GetFavorites(
	ctx context.Context,
	req *userv1.GetFavoritesRequest,
) (*userv1.GetFavoritesResponse, error) {
	authCtx, err := s.authorize(ctx)
	if err != nil {
		return nil, err
	}

	favorites, err := s.usecase.GetFavorites(ctx, authCtx.UserID, req.GetLimit(), req.GetOffset())
	if err != nil {
		return nil, mapError(err)
	}

	return &userv1.GetFavoritesResponse{
		MovieIds:   favorites.MovieIDs,
		TotalCount: favorites.TotalCount,
	}, nil
}

func (s *Server) SearchUsers(ctx context.Context, req *userv1.SearchUsersRequest) (*userv1.SearchUsersResponse, error) {
	authCtx, err := s.authorize(ctx)
	if err != nil {
		return nil, err
	}

	users, err := s.usecase.SearchUsers(ctx, authCtx.UserID, req.GetQuery(), req.GetLimit())
	if err != nil {
		return nil, mapError(err)
	}

	resp := &userv1.SearchUsersResponse{Users: make([]*userv1.UserSearchResult, 0, len(users))}
	for _, user := range users {
		resp.Users = append(resp.Users, &userv1.UserSearchResult{
			Id:        user.ID,
			Email:     user.Email,
			AvatarUrl: user.AvatarURL,
			IsFriend:  user.IsFriend,
		})
	}

	return resp, nil
}

func (s *Server) SendFriendRequest(
	ctx context.Context,
	req *userv1.SendFriendRequestRequest,
) (*userv1.SendFriendRequestResponse, error) {
	authCtx, err := s.authorize(ctx)
	if err != nil {
		return nil, err
	}

	requestID, err := s.usecase.SendFriendRequest(ctx, authCtx.UserID, req.GetToUserId())
	if err != nil {
		return nil, mapError(err)
	}

	return &userv1.SendFriendRequestResponse{RequestId: requestID}, nil
}

func (s *Server) RespondToFriendRequest(
	ctx context.Context,
	req *userv1.RespondToFriendRequestRequest,
) (*userv1.RespondToFriendRequestResponse, error) {
	authCtx, err := s.authorize(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.usecase.RespondToFriendRequest(ctx, authCtx.UserID, req.GetRequestId(), req.GetAction()); err != nil {
		return nil, mapError(err)
	}

	return &userv1.RespondToFriendRequestResponse{Success: true}, nil
}

func (s *Server) DeleteOutgoingFriendRequest(
	ctx context.Context,
	req *userv1.DeleteOutgoingFriendRequestRequest,
) (*userv1.DeleteOutgoingFriendRequestResponse, error) {
	authCtx, err := s.authorize(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.usecase.DeleteOutgoingFriendRequest(ctx, authCtx.UserID, req.GetRequestId()); err != nil {
		return nil, mapError(err)
	}

	return &userv1.DeleteOutgoingFriendRequestResponse{Success: true}, nil
}

func (s *Server) GetFriendRequests(
	ctx context.Context,
	req *userv1.GetFriendRequestsRequest,
) (*userv1.GetFriendRequestsResponse, error) {
	authCtx, err := s.authorize(ctx)
	if err != nil {
		return nil, err
	}

	items, err := s.usecase.GetFriendRequests(ctx, authCtx.UserID, req.GetDirection(), req.GetLimit())
	if err != nil {
		return nil, mapError(err)
	}

	resp := &userv1.GetFriendRequestsResponse{Requests: make([]*userv1.FriendRequestItem, 0, len(items))}
	for _, item := range items {
		resp.Requests = append(resp.Requests, &userv1.FriendRequestItem{
			Id:        item.ID,
			UserId:    item.UserID,
			Email:     item.Email,
			CreatedAt: item.CreatedAt,
		})
	}

	return resp, nil
}

func (s *Server) GetFriendsList(
	ctx context.Context,
	req *userv1.GetFriendsListRequest,
) (*userv1.GetFriendsListResponse, error) {
	authCtx, err := s.authorize(ctx)
	if err != nil {
		return nil, err
	}

	friendsResp, err := s.usecase.GetFriendsList(ctx, authCtx.UserID, req.GetLimit(), req.GetOffset())
	if err != nil {
		return nil, mapError(err)
	}

	friends := make([]*userv1.UserSearchResult, 0, len(friendsResp.Friends))
	for _, friend := range friendsResp.Friends {
		friends = append(friends, &userv1.UserSearchResult{
			Id:        friend.ID,
			Email:     friend.Email,
			AvatarUrl: friend.AvatarURL,
			IsFriend:  true,
		})
	}

	return &userv1.GetFriendsListResponse{
		Friends:    friends,
		TotalCount: friendsResp.TotalCount,
	}, nil
}
