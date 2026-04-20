package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/go-park-mail-ru/2026_1_VKino/services/user-service/internal/domain"
)

func (u *UserUsecase) SearchUsersByEmail(
	ctx context.Context,
	userID int64,
	emailQuery string,
) ([]domain.UserSearchResult, error) {
	if _, err := u.userRepo.GetUserByID(ctx, userID); err != nil {
		return nil, domain.ErrInvalidToken
	}

	normalizedQuery := strings.TrimSpace(emailQuery)
	if !domain.ValidateEmailQuery(normalizedQuery) {
		return nil, domain.ErrInvalidSearchQuery
	}

	users, err := u.userRepo.SearchUsersByEmail(ctx, userID, normalizedQuery)
	if err != nil {
		return nil, fmt.Errorf("%w: search users by email: %v", domain.ErrInternal, err)
	}

	return users, nil
}

func (u *UserUsecase) AddFriend(ctx context.Context, userID int64, friendID int64) (domain.FriendResponse, error) {
	if _, err := u.userRepo.GetUserByID(ctx, userID); err != nil {
		return domain.FriendResponse{}, domain.ErrInvalidToken
	}

	if userID == friendID {
		return domain.FriendResponse{}, domain.ErrSelfFriendship
	}

	friend, err := u.userRepo.GetUserByID(ctx, friendID)
	if err != nil {
		return domain.FriendResponse{}, domain.ErrUserNotFound
	}

	if err = u.userRepo.AddFriend(ctx, userID, friendID); err != nil {
		if errors.Is(err, domain.ErrAlreadyFriends) {
			return domain.FriendResponse{}, err
		}

		return domain.FriendResponse{}, fmt.Errorf("%w: add friend: %v", domain.ErrInternal, err)
	}

	return domain.FriendResponse{
		ID:    friend.ID,
		Email: friend.Email,
	}, nil
}

func (u *UserUsecase) DeleteFriend(ctx context.Context, userID int64, friendID int64) error {
	if _, err := u.userRepo.GetUserByID(ctx, userID); err != nil {
		return domain.ErrInvalidToken
	}

	if userID == friendID {
		return domain.ErrSelfFriendship
	}

	if err := u.userRepo.DeleteFriend(ctx, userID, friendID); err != nil {
		if errors.Is(err, domain.ErrFriendNotFound) {
			return err
		}

		return fmt.Errorf("%w: delete friend: %v", domain.ErrInternal, err)
	}

	return nil
}
