package repository

//go:generate mockgen -source=./interface.go -destination=./mocks/user_repo_mock.go -package=mocks

import (
	"context"
	"time"

	domain "github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/domain"
)

type UserReadRepo interface {
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserByID(ctx context.Context, id int64) (*domain.User, error)
	GetFriend(ctx context.Context, userID, friendID int64) (*domain.User, error)
	GetUserRole(ctx context.Context, userID int64) (string, error)
	SearchUsersByEmail(ctx context.Context, userID int64, query string) ([]domain.UserSearchResult, error)
}

type UserProfileMutationRepo interface {
	UpdateBirthdate(ctx context.Context, userID int64, birthdate *time.Time) (*domain.User, error)
	UpdateAvatarFileKey(ctx context.Context, userID int64, avatarFileKey *string) (*domain.User, error)
}

type UserSubscriptionRepo interface {
	GetActiveSubscription(ctx context.Context, userID int64) (domain.SubscriptionInfo, error)
	GetSubscriptionTariffByCode(ctx context.Context, code string) (domain.SubscriptionInfo, error)
	GetSubscriptionTariffByID(ctx context.Context, tariffID int64) (domain.SubscriptionTariff, error)
	GetSubscriptionTariffOptions(ctx context.Context, tariffID int64) ([]domain.SubscriptionOption, error)
	DeactivateUserSubscriptions(ctx context.Context, userID int64) error
	CreateUserSubscription(
		ctx context.Context,
		userID, tariffID int64,
		startsAt, expiresAt time.Time,
	) error
	GetCoinsReceivedToday(ctx context.Context, userID int64) (int32, error)
	GetRoomsCreatedThisMonth(ctx context.Context, userID int64) (int32, error)
}

type UserCoinsRepo interface {
	GetVKinoCoinsBalance(ctx context.Context, userID int64) (int32, error)
	GrantDailyVKinoCoins(ctx context.Context, userID int64) (int32, error)
	GetVKinoCoinsHistory(
		ctx context.Context,
		userID int64,
		limit, offset int32,
	) ([]domain.VKinoCoinsHistoryItem, int32, error)
	BuySubscriptionWithVKinoCoins(
		ctx context.Context,
		userID int64,
		tariffID int64,
	) (domain.VKinoCoinsSubscriptionPurchase, error)
}

type UserMovieRepo interface {
	AddMovieToFavorites(ctx context.Context, userID, movieID int64) error
	SetMovieRating(ctx context.Context, userID, movieID int64, rating float64) error
	SetMovieReview(
		ctx context.Context,
		userID, movieID int64,
		rating *float64,
		comment *string,
	) (domain.MovieReviewResponse, error)
	DeleteMovieReview(ctx context.Context, userID, movieID int64) error
	SetReviewReaction(ctx context.Context, userID, reviewID int64, reaction string) error
	DeleteReviewReaction(ctx context.Context, userID, reviewID int64) error
	ToggleFavorite(ctx context.Context, userID, movieID int64) (bool, error)
	GetFavorites(ctx context.Context, userID int64, limit, offset int32) ([]int64, int32, error)
}

type UserFriendRepo interface {
	AddFriend(ctx context.Context, userID int64, friendID int64) error
	DeleteFriend(ctx context.Context, userID int64, friendID int64) error
	SendFriendRequest(ctx context.Context, fromUserID, toUserID int64) (int64, error)
	RespondToFriendRequest(ctx context.Context, requestID, userID int64, action string) error
	DeleteOutgoingFriendRequest(ctx context.Context, requestID, fromUserID int64) error
	GetFriendRequests(
		ctx context.Context,
		userID int64,
		direction string,
		limit int32,
	) ([]domain.FriendRequestItem, error)
	GetFriendsList(
		ctx context.Context,
		userID int64,
		limit, offset int32,
	) ([]domain.UserSearchResult, int32, error)
}

type UserRepo interface {
	UserReadRepo
	UserProfileMutationRepo
	UserSubscriptionRepo
	UserCoinsRepo
	UserMovieRepo
	UserFriendRepo
}

type SupportRepo interface {
	CreateTicket(
		ctx context.Context,
		userID int64,
		req domain.CreateSupportTicketRequest,
	) (*domain.SupportTicketResponse, error)
	GetTicketByID(ctx context.Context, ticketID int64) (*domain.SupportTicketResponse, error)
	GetTickets(
		ctx context.Context,
		userID int64,
		req domain.GetSupportTicketsRequest,
	) ([]domain.SupportTicketResponse, error)
	UpdateTicket(
		ctx context.Context,
		req domain.UpdateSupportTicketRequest,
	) (*domain.SupportTicketResponse, error)
	GetTicketMessages(ctx context.Context, ticketID int64) ([]domain.SupportTicketMessageResponse, error)
	CreateTicketMessage(
		ctx context.Context,
		senderID int64,
		req domain.CreateSupportTicketMessageRequest,
	) (*domain.SupportTicketMessageResponse, error)
	HasTicketFile(ctx context.Context, ticketID int64, fileKey string) (bool, error)
	GetTicketStatistics(
		ctx context.Context,
		allowedCategories []string,
	) (*domain.SupportTicketStatisticsResponse, error)
}
