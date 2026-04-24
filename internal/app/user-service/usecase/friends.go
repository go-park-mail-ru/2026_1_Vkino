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
