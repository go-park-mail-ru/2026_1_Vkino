//nolint:gocyclo,lll // Friend flows stay explicit to keep branching readable.
package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	domain "github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/domain"
)

func (u *UserUsecase) SearchUsersByEmail(
	ctx context.Context,
	userID int64,
	emailQuery string,
) ([]domain.UserSearchResult, error) {
	if _, err := u.userRepo.GetUserByID(ctx, userID); err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrUserNotFound
		}

		return nil, fmt.Errorf("get user by id: %w", err)
	}

	normalizedQuery := strings.TrimSpace(emailQuery)
	if !domain.ValidateEmailQuery(normalizedQuery) {
		return nil, domain.ErrInvalidSearchQuery
	}

	users, err := u.userRepo.SearchUsersByEmail(ctx, userID, normalizedQuery)
	if err != nil {
		return nil, fmt.Errorf("%w: search users by email: %w", domain.ErrInternal, err)
	}

	u.enrichUserSearchAvatarKeys(ctx, users)

	return users, nil
}

func (u *UserUsecase) SearchUsers(
	ctx context.Context,
	userID int64,
	query string,
	limit int32,
) ([]domain.UserSearchResult, error) {
	if limit <= 0 {
		limit = 10
	}

	users, err := u.SearchUsersByEmail(ctx, userID, query)
	if err != nil {
		return nil, err
	}

	if len(users) <= int(limit) {
		return users, nil
	}

	return users[:limit], nil
}

func (u *UserUsecase) AddFriend(ctx context.Context, userID int64, friendID int64) (domain.FriendResponse, error) {
	if _, err := u.userRepo.GetUserByID(ctx, userID); err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return domain.FriendResponse{}, domain.ErrUserNotFound
		}

		return domain.FriendResponse{}, fmt.Errorf("get user by id: %w", err)
	}

	if userID == friendID {
		return domain.FriendResponse{}, domain.ErrSelfFriendship
	}

	friend, err := u.userRepo.GetUserByID(ctx, friendID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return domain.FriendResponse{}, domain.ErrUserNotFound
		}

		return domain.FriendResponse{}, fmt.Errorf("get friend by id: %w", err)
	}

	if err = u.userRepo.AddFriend(ctx, userID, friendID); err != nil {
		if errors.Is(err, domain.ErrAlreadyFriends) {
			return domain.FriendResponse{}, err
		}

		return domain.FriendResponse{}, fmt.Errorf("add friend: %w", err)
	}

	return domain.FriendResponse{
		ID:        friend.ID,
		Email:     friend.Email,
		AvatarURL: u.friendAvatarPresignedURL(ctx, friend),
	}, nil
}

func (u *UserUsecase) DeleteFriend(ctx context.Context, userID int64, friendID int64) error {
	if _, err := u.userRepo.GetUserByID(ctx, userID); err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return domain.ErrUserNotFound
		}

		return fmt.Errorf("get user by id: %w", err)
	}

	if userID == friendID {
		return domain.ErrSelfFriendship
	}

	if err := u.userRepo.DeleteFriend(ctx, userID, friendID); err != nil {
		if errors.Is(err, domain.ErrFriendNotFound) {
			return err
		}

		return fmt.Errorf("delete friend: %w", err)
	}

	return nil
}

func (u *UserUsecase) SendFriendRequest(ctx context.Context, userID, toUserID int64) (int64, error) {
	if _, err := u.userRepo.GetUserByID(ctx, userID); err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return 0, domain.ErrUserNotFound
		}

		return 0, fmt.Errorf("get user by id: %w", err)
	}

	if userID == toUserID {
		return 0, domain.ErrSelfFriendship
	}

	requestID, err := u.userRepo.SendFriendRequest(ctx, userID, toUserID)
	if err != nil {
		if errors.Is(err, domain.ErrAlreadyFriends) {
			return 0, domain.ErrAlreadyFriends
		}

		return 0, fmt.Errorf("send friend request: %w", err)
	}

	return requestID, nil
}

func (u *UserUsecase) RespondToFriendRequest(ctx context.Context, userID, requestID int64, action string) error {
	if _, err := u.userRepo.GetUserByID(ctx, userID); err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return domain.ErrUserNotFound
		}

		return fmt.Errorf("get user by id: %w", err)
	}

	if action != "accept" && action != "decline" && action != "cancel" {
		return domain.ErrInvalidSearchQuery
	}

	if err := u.userRepo.RespondToFriendRequest(ctx, requestID, userID, action); err != nil {
		if errors.Is(err, domain.ErrFriendNotFound) {
			return err
		}

		return fmt.Errorf("respond to friend request: %w", err)
	}

	return nil
}

func (u *UserUsecase) DeleteOutgoingFriendRequest(ctx context.Context, userID, requestID int64) error {
	if _, err := u.userRepo.GetUserByID(ctx, userID); err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return domain.ErrUserNotFound
		}

		return fmt.Errorf("get user by id: %w", err)
	}

	if err := u.userRepo.DeleteOutgoingFriendRequest(ctx, requestID, userID); err != nil {
		if errors.Is(err, domain.ErrFriendNotFound) {
			return err
		}

		return fmt.Errorf("delete outgoing friend request: %w", err)
	}

	return nil
}

func (u *UserUsecase) GetFriendRequests(ctx context.Context, userID int64, direction string, limit int32) ([]domain.FriendRequestItem, error) {
	if _, err := u.userRepo.GetUserByID(ctx, userID); err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrUserNotFound
		}

		return nil, fmt.Errorf("get user by id: %w", err)
	}

	if direction != "incoming" && direction != "outgoing" {
		direction = "incoming"
	}

	if limit <= 0 {
		limit = 50
	}

	items, err := u.userRepo.GetFriendRequests(ctx, userID, direction, limit)
	if err != nil {
		return nil, fmt.Errorf("get friend requests: %w", err)
	}

	return items, nil
}

func (u *UserUsecase) GetFriendsList(ctx context.Context, userID int64, limit, offset int32) (domain.FriendsListResponse, error) {
	if _, err := u.userRepo.GetUserByID(ctx, userID); err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return domain.FriendsListResponse{}, domain.ErrUserNotFound
		}

		return domain.FriendsListResponse{}, fmt.Errorf("get user by id: %w", err)
	}

	if limit <= 0 {
		limit = 50
	}

	if offset < 0 {
		offset = 0
	}

	friends, total, err := u.userRepo.GetFriendsList(ctx, userID, limit, offset)
	if err != nil {
		return domain.FriendsListResponse{}, fmt.Errorf("get friends list: %w", err)
	}

	u.enrichUserSearchAvatarKeys(ctx, friends)

	return domain.FriendsListResponse{
		Friends:    friends,
		TotalCount: total,
	}, nil
}
