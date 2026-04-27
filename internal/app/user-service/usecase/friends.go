package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	domain2 "github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/domain"
)

func (u *UserUsecase) SearchUsersByEmail(
	ctx context.Context,
	userID int64,
	emailQuery string,
) ([]domain2.UserSearchResult, error) {
	if _, err := u.userRepo.GetUserByID(ctx, userID); err != nil {
		return nil, domain2.ErrInvalidToken
	}

	normalizedQuery := strings.TrimSpace(emailQuery)
	if !domain2.ValidateEmailQuery(normalizedQuery) {
		return nil, domain2.ErrInvalidSearchQuery
	}

	users, err := u.userRepo.SearchUsersByEmail(ctx, userID, normalizedQuery)
	if err != nil {
		return nil, fmt.Errorf("%w: search users by email: %v", domain2.ErrInternal, err)
	}

	return users, nil
}

func (u *UserUsecase) SearchUsers(
	ctx context.Context,
	userID int64,
	query string,
	limit int32,
) ([]domain2.UserSearchResult, error) {
	if limit <= 0 {
		limit = 10
	}
	users, err := u.SearchUsersByEmail(ctx, userID, query)
	if err != nil {
		return nil, err
	}
	if int32(len(users)) <= limit {
		return users, nil
	}
	return users[:limit], nil
}

func (u *UserUsecase) AddFriend(ctx context.Context, userID int64, friendID int64) (domain2.FriendResponse, error) {
	if _, err := u.userRepo.GetUserByID(ctx, userID); err != nil {
		return domain2.FriendResponse{}, domain2.ErrInvalidToken
	}

	if userID == friendID {
		return domain2.FriendResponse{}, domain2.ErrSelfFriendship
	}

	friend, err := u.userRepo.GetUserByID(ctx, friendID)
	if err != nil {
		return domain2.FriendResponse{}, domain2.ErrUserNotFound
	}

	if err = u.userRepo.AddFriend(ctx, userID, friendID); err != nil {
		if errors.Is(err, domain2.ErrAlreadyFriends) {
			return domain2.FriendResponse{}, err
		}

		return domain2.FriendResponse{}, fmt.Errorf("%w: add friend: %v", domain2.ErrInternal, err)
	}

	return domain2.FriendResponse{
		ID:    friend.ID,
		Email: friend.Email,
	}, nil
}

func (u *UserUsecase) DeleteFriend(ctx context.Context, userID int64, friendID int64) error {
	if _, err := u.userRepo.GetUserByID(ctx, userID); err != nil {
		return domain2.ErrInvalidToken
	}

	if userID == friendID {
		return domain2.ErrSelfFriendship
	}

	if err := u.userRepo.DeleteFriend(ctx, userID, friendID); err != nil {
		if errors.Is(err, domain2.ErrFriendNotFound) {
			return err
		}

		return fmt.Errorf("%w: delete friend: %v", domain2.ErrInternal, err)
	}

	return nil
}

func (u *UserUsecase) SendFriendRequest(ctx context.Context, userID, toUserID int64) (int64, error) {
	if _, err := u.userRepo.GetUserByID(ctx, userID); err != nil {
		return 0, domain2.ErrInvalidToken
	}
	if userID == toUserID {
		return 0, domain2.ErrSelfFriendship
	}
	requestID, err := u.userRepo.SendFriendRequest(ctx, userID, toUserID)
	if err != nil {
		if errors.Is(err, domain2.ErrAlreadyFriends) {
			return 0, domain2.ErrAlreadyFriends
		}
		return 0, domain2.ErrInternal
	}
	return requestID, nil
}

func (u *UserUsecase) RespondToFriendRequest(ctx context.Context, userID, requestID int64, action string) error {
	if _, err := u.userRepo.GetUserByID(ctx, userID); err != nil {
		return domain2.ErrInvalidToken
	}
	if action != "accept" && action != "decline" && action != "cancel" {
		return domain2.ErrInvalidSearchQuery
	}
	if err := u.userRepo.RespondToFriendRequest(ctx, requestID, userID, action); err != nil {
		if errors.Is(err, domain2.ErrFriendNotFound) {
			return err
		}
		return domain2.ErrInternal
	}
	return nil
}

func (u *UserUsecase) DeleteOutgoingFriendRequest(ctx context.Context, userID, requestID int64) error {
	if _, err := u.userRepo.GetUserByID(ctx, userID); err != nil {
		return domain2.ErrInvalidToken
	}
	if err := u.userRepo.DeleteOutgoingFriendRequest(ctx, requestID, userID); err != nil {
		if errors.Is(err, domain2.ErrFriendNotFound) {
			return err
		}
		return domain2.ErrInternal
	}
	return nil
}

func (u *UserUsecase) GetFriendRequests(ctx context.Context, userID int64, direction string, limit int32) ([]domain2.FriendRequestItem, error) {
	if _, err := u.userRepo.GetUserByID(ctx, userID); err != nil {
		return nil, domain2.ErrInvalidToken
	}
	if direction != "incoming" && direction != "outgoing" {
		direction = "incoming"
	}
	if limit <= 0 {
		limit = 50
	}
	items, err := u.userRepo.GetFriendRequests(ctx, userID, direction, limit)
	if err != nil {
		return nil, domain2.ErrInternal
	}
	return items, nil
}

func (u *UserUsecase) GetFriendsList(ctx context.Context, userID int64, limit, offset int32) (domain2.FriendsListResponse, error) {
	if _, err := u.userRepo.GetUserByID(ctx, userID); err != nil {
		return domain2.FriendsListResponse{}, domain2.ErrInvalidToken
	}
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	friends, total, err := u.userRepo.GetFriendsList(ctx, userID, limit, offset)
	if err != nil {
		return domain2.FriendsListResponse{}, domain2.ErrInternal
	}
	return domain2.FriendsListResponse{
		Friends:    friends,
		TotalCount: total,
	}, nil
}
