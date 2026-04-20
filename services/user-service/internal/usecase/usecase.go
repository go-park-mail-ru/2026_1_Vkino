package usecase

import (
	"context"
	"io"

	"github.com/go-park-mail-ru/2026_1_VKino/pkg/storage"
	"github.com/go-park-mail-ru/2026_1_VKino/services/user-service/internal/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/services/user-service/internal/repository"
	clocksvc "github.com/go-park-mail-ru/2026_1_VKino/services/user-service/internal/service/clock"
)

type Usecase interface {
	GetProfile(ctx context.Context, userID int64) (domain.ProfileResponse, error)
	SearchUsersByEmail(ctx context.Context, userID int64, emailQuery string) ([]domain.UserSearchResult, error)
	AddFriend(ctx context.Context, userID int64, friendID int64) (domain.FriendResponse, error)
	DeleteFriend(ctx context.Context, userID int64, friendID int64) error
	UpdateProfile(ctx context.Context, userID int64, birthdate string, body io.Reader, size int64, contentType string) (domain.ProfileResponse, error)
	AddMovieToFavorites(ctx context.Context, userID, movieID int64) (domain.FavoriteMovieResponse, error)
}

type UserUsecase struct {
	userRepo     repository.UserRepo
	avatarStore  storage.FileStorage
	clockService clocksvc.Service
}
