package repository

import (
	"context"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/domain"
	"github.com/google/uuid"
)

//go:generate mockgen -source=./repository.go -destination=../usecase/mocks/repository_mock.go -package=mocks
type UserRepo interface {
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	CreateUser(ctx context.Context, login string, password string) (*domain.User, error)
	UpdateUser(ctx context.Context, login string, password string) (*domain.User, error)
	DeleteUser(ctx context.Context, login string) error
	// GetAllUsers(ctx context.Context) ([]*domain.User, error)
}

type SessionRepo interface {
	SaveSession(ctx context.Context, userID uuid.UUID, refreshToken string, expiresAt time.Time) error
	GetSession(ctx context.Context, userID uuid.UUID) (string, error)
	DeleteSession(ctx context.Context, userID uuid.UUID) error
}
