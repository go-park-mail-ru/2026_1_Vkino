package repository

import (
	"context"
	"time"

	domain2 "github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/domain"
)

type UserRepo interface {
	GetUserByEmail(ctx context.Context, email string) (*domain2.User, error)
	GetUserByID(ctx context.Context, id int64) (*domain2.User, error)
	SearchUsersByEmail(ctx context.Context, userID int64, query string) ([]domain2.UserSearchResult, error)
	UpdateBirthdate(ctx context.Context, userID int64, birthdate *time.Time) (*domain2.User, error)
	UpdateAvatarFileKey(ctx context.Context, userID int64, avatarFileKey *string) (*domain2.User, error)
	AddMovieToFavorites(ctx context.Context, userID, movieID int64) error
	AddFriend(ctx context.Context, userID int64, friendID int64) error
	DeleteFriend(ctx context.Context, userID int64, friendID int64) error
}
