package repository

import (
	"context"
	"time"

	domain2 "github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/domain"
)

type UserRepo interface {
	GetUserByEmail(ctx context.Context, email string) (*domain2.User, error)
	GetUserByID(ctx context.Context, id int64) (*domain2.User, error)
	GetUserRole(ctx context.Context, userID int64) (string, error)
	SearchUsersByEmail(ctx context.Context, userID int64, query string) ([]domain2.UserSearchResult, error)
	UpdateBirthdate(ctx context.Context, userID int64, birthdate *time.Time) (*domain2.User, error)
	UpdateAvatarFileKey(ctx context.Context, userID int64, avatarFileKey *string) (*domain2.User, error)
	AddMovieToFavorites(ctx context.Context, userID, movieID int64) error
	ToggleFavorite(ctx context.Context, userID, movieID int64) (bool, error)
	GetFavorites(ctx context.Context, userID int64, limit, offset int32) ([]domain2.MovieCardResponse, int32, error)
	AddFriend(ctx context.Context, userID int64, friendID int64) error
	DeleteFriend(ctx context.Context, userID int64, friendID int64) error
	SendFriendRequest(ctx context.Context, fromUserID, toUserID int64) (int64, error)
	RespondToFriendRequest(ctx context.Context, requestID, userID int64, action string) error
	DeleteOutgoingFriendRequest(ctx context.Context, requestID, fromUserID int64) error
	GetFriendRequests(ctx context.Context, userID int64, direction string, limit int32) ([]domain2.FriendRequestItem, error)
	GetFriendsList(ctx context.Context, userID int64, limit, offset int32) ([]domain2.UserSearchResult, int32, error)
}

type SupportRepo interface {
	CreateTicket(ctx context.Context, userID int64, req domain2.CreateSupportTicketRequest) (*domain2.SupportTicketResponse, error)
	GetTicketByID(ctx context.Context, ticketID int64) (*domain2.SupportTicketResponse, error)
	GetTickets(ctx context.Context, userID int64, req domain2.GetSupportTicketsRequest) ([]domain2.SupportTicketResponse, error)
	UpdateTicket(ctx context.Context, req domain2.UpdateSupportTicketRequest) (*domain2.SupportTicketResponse, error)
	GetTicketMessages(ctx context.Context, ticketID int64) ([]domain2.SupportTicketMessageResponse, error)
	CreateTicketMessage(ctx context.Context, senderID int64, req domain2.CreateSupportTicketMessageRequest) (*domain2.SupportTicketMessageResponse, error)
	GetTicketStatistics(ctx context.Context) (*domain2.SupportTicketStatisticsResponse, error)
}
