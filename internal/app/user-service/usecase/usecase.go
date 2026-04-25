package usecase

import (
	"context"
	"io"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/repository"
	clocksvc "github.com/go-park-mail-ru/2026_1_VKino/pkg/service/clock"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/storage"
)

type Usecase interface {
	GetProfile(ctx context.Context, userID int64) (domain.ProfileResponse, error)
	SearchUsersByEmail(ctx context.Context, userID int64, emailQuery string) ([]domain.UserSearchResult, error)
	AddFriend(ctx context.Context, userID int64, friendID int64) (domain.FriendResponse, error)
	DeleteFriend(ctx context.Context, userID int64, friendID int64) error
	UpdateProfile(ctx context.Context, userID int64, birthdate string, body io.Reader, size int64,
		contentType string) (domain.ProfileResponse, error)
	AddMovieToFavorites(ctx context.Context, userID, movieID int64) (domain.FavoriteMovieResponse, error)
}

type SupportUsecase interface {
	CreateTicket(ctx context.Context, actorUserID int64, req domain.CreateSupportTicketRequest) (domain.SupportTicketResponse, error)
	GetTickets(
		ctx context.Context,
		actorUserID int64,
		req domain.GetSupportTicketsRequest,
	) ([]domain.SupportTicketResponse, error)
	UpdateTicket(ctx context.Context, actorUserID int64, req domain.UpdateSupportTicketRequest) (domain.SupportTicketResponse, error)
	GetTicketMessages(
		ctx context.Context,
		actorUserID int64,
		req domain.GetSupportTicketMessagesRequest,
	) ([]domain.SupportTicketMessageResponse, error)
	CreateTicketMessage(
		ctx context.Context,
		actorUserID int64,
		req domain.CreateSupportTicketMessageRequest,
	) (domain.SupportTicketMessageResponse, error)
	GetTicketStatistics(
		ctx context.Context,
		actorUserID int64,
		req domain.GetSupportTicketStatisticsRequest,
	) (domain.SupportTicketStatisticsResponse, error)
	SubscribeTicket(
		ctx context.Context,
		actorUserID int64,
		req domain.SubscribeSupportTicketRequest,
	) (<-chan domain.SupportTicketEventResponse, func(), error)
}

type UserUsecase struct {
	userRepo     repository.UserRepo
	avatarStore  storage.FileStorage
	clockService clocksvc.Service
}
