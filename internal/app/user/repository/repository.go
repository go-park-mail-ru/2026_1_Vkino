package repository

import (
	"context"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user/domain"
)

//go:generate mockgen -source=./repository.go -destination=../usecase/mocks/repository_mock.go -package=mocks
type UserRepo interface {
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserByID(ctx context.Context, id int64) (*domain.User, error)
	SearchUsersByEmail(ctx context.Context, userID int64, query string) ([]domain.UserSearchResult, error)
	CreateUser(ctx context.Context, login string, password string) (*domain.User, error)
	UpdateUser(ctx context.Context, login string, password string) (*domain.User, error)
	UpdateBirthdate(ctx context.Context, userID int64, birthdate *time.Time) (*domain.User, error)
	UpdateAvatarFileKey(ctx context.Context, userID int64, avatarFileKey *string) (*domain.User, error)
	UpdatePassword(ctx context.Context, userID int64, passwordHash string) error
	AddFriend(ctx context.Context, userID int64, friendID int64) error
	DeleteFriend(ctx context.Context, userID int64, friendID int64) error
	DeleteUser(ctx context.Context, login string) error
	// GetAllUsers(ctx context.Context) ([]*domain.User, error)
}

type SessionRepo interface {
	SaveSession(ctx context.Context, userID int64, refreshToken string, expiresAt time.Time) error
	GetSession(ctx context.Context, userID int64) (string, error)
	DeleteSession(ctx context.Context, userID int64) error
}
