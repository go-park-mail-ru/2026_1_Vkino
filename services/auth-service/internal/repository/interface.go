package repository

import (
	"context"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/services/auth-service/internal/domain"
)

type UserRepo interface {
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserByID(ctx context.Context, id int64) (*domain.User, error)
	CreateUser(ctx context.Context, email string, passwordHash string) (*domain.User, error)
	UpdatePassword(ctx context.Context, userID int64, passwordHash string) error
}

type SessionRepo interface {
	SaveSession(ctx context.Context, userID int64, refreshToken string, expiresAt time.Time) error
	GetSession(ctx context.Context, userID int64) (string, error)
	DeleteSession(ctx context.Context, userID int64) error
}
