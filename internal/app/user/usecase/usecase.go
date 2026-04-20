package usecase

import (
	"context"
	"io"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user/domain"
)

//go:generate mockgen -source=./usecase.go -destination=./mocks/usecase_mock.go -package=mocks
type Usecase interface {
	SignIn(ctx context.Context, email, password string) (domain.TokenPair, error)
	SignUp(ctx context.Context, email, password string) (domain.TokenPair, error)
	Refresh(ctx context.Context, email string) (domain.TokenPair, error)
	ValidateRefreshToken(ctx context.Context, tokenString string) (string, error)
	LogOut(ctx context.Context, email string) error
	GetProfile(ctx context.Context, userID int64) (domain.ProfileResponse, error)
	SearchUsersByEmail(ctx context.Context, userID int64, emailQuery string) ([]domain.UserSearchResult, error)
	AddFriend(ctx context.Context, userID int64, friendID int64) (domain.FriendResponse, error)
	DeleteFriend(ctx context.Context, userID int64, friendID int64) error
	UpdateProfile(ctx context.Context, userID int64, birthdate string, body io.Reader, size int64, contentType string) (domain.ProfileResponse, error)
	ChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error
	AddMovieToFavorites(ctx context.Context, userID, movieID int64) (domain.FavoriteMovieResponse, error)

	ValidateAccessToken(tokenString string) (AuthContext, error)
	GetConfig() Config
}
