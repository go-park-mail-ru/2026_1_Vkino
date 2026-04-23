package repository

import (
	"context"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/user-service/domain"
)

type UserRepo interface {
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserByID(ctx context.Context, id int64) (*domain.User, error)
	SearchUsersByEmail(ctx context.Context, userID int64, query string) ([]domain.UserSearchResult, error)
	UpdateBirthdate(ctx context.Context, userID int64, birthdate *time.Time) (*domain.User, error)
	UpdateAvatarFileKey(ctx context.Context, userID int64, avatarFileKey *string) (*domain.User, error)
	AddMovieToFavorites(ctx context.Context, userID, movieID int64) error
	AddFriend(ctx context.Context, userID int64, friendID int64) error
	DeleteFriend(ctx context.Context, userID int64, friendID int64) error
}
