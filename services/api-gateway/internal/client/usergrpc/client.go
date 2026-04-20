package usergrpc

import (
	"context"
	"fmt"
	"time"

	userv1 "github.com/go-park-mail-ru/2026_1_VKino/platform/gen/user/v1"
	"github.com/go-park-mail-ru/2026_1_VKino/services/api-gateway/internal/dto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Config struct {
	Address        string
	RequestTimeout time.Duration
}

type Client interface {
	GetProfile(ctx context.Context, userID int64) (dto.ProfileResponse, error)
	UpdateProfile(ctx context.Context, userID int64, birthdate string) (dto.ProfileResponse, error)
	SearchUsersByEmail(ctx context.Context, userID int64, emailQuery string) ([]dto.UserSearchResult, error)
	AddFriend(ctx context.Context, userID, friendID int64) (dto.FriendResponse, error)
	DeleteFriend(ctx context.Context, userID, friendID int64) error
	AddMovieToFavorites(ctx context.Context, userID, movieID int64) (dto.FavoriteMovieResponse, error)
	Close() error
}

type GRPCClient struct {
	conn    *grpc.ClientConn
	client  userv1.UserServiceClient
	timeout time.Duration
}

func New(ctx context.Context, cfg Config) (*GRPCClient, error) {
	timeout := cfg.RequestTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	conn, err := grpc.DialContext(
		ctx,
		cfg.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("dial user grpc: %w", err)
	}

	return &GRPCClient{
		conn:    conn,
		client:  userv1.NewUserServiceClient(conn),
		timeout: timeout,
	}, nil
}

func (c *GRPCClient) Close() error {
	if c.conn == nil {
		return nil
	}

	return c.conn.Close()
}

func (c *GRPCClient) GetProfile(ctx context.Context, userID int64) (dto.ProfileResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.client.GetProfile(ctx, &userv1.GetProfileRequest{
		UserId: userID,
	})
	if err != nil {
		return dto.ProfileResponse{}, err
	}

	var birthdate *string
	if resp.GetBirthdate() != "" {
		value := resp.GetBirthdate()
		birthdate = &value
	}

	return dto.ProfileResponse{
		Email:     resp.GetEmail(),
		Birthdate: birthdate,
		AvatarURL: resp.GetAvatarUrl(),
	}, nil
}

func (c *GRPCClient) UpdateProfile(ctx context.Context, userID int64, birthdate string) (dto.ProfileResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.client.UpdateProfile(ctx, &userv1.UpdateProfileRequest{
		UserId:    userID,
		Birthdate: birthdate,
	})
	if err != nil {
		return dto.ProfileResponse{}, err
	}

	var respBirthdate *string
	if resp.GetBirthdate() != "" {
		value := resp.GetBirthdate()
		respBirthdate = &value
	}

	return dto.ProfileResponse{
		Email:     resp.GetEmail(),
		Birthdate: respBirthdate,
		AvatarURL: resp.GetAvatarUrl(),
	}, nil
}

func (c *GRPCClient) SearchUsersByEmail(ctx context.Context, userID int64, emailQuery string) ([]dto.UserSearchResult, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.client.SearchUsersByEmail(ctx, &userv1.SearchUsersByEmailRequest{
		UserId:     userID,
		EmailQuery: emailQuery,
	})
	if err != nil {
		return nil, err
	}

	result := make([]dto.UserSearchResult, 0, len(resp.GetUsers()))
	for _, user := range resp.GetUsers() {
		result = append(result, dto.UserSearchResult{
			ID:       user.GetId(),
			Email:    user.GetEmail(),
			IsFriend: user.GetIsFriend(),
		})
	}

	return result, nil
}

func (c *GRPCClient) AddFriend(ctx context.Context, userID, friendID int64) (dto.FriendResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.client.AddFriend(ctx, &userv1.AddFriendRequest{
		UserId:   userID,
		FriendId: friendID,
	})
	if err != nil {
		return dto.FriendResponse{}, err
	}

	return dto.FriendResponse{
		ID:    resp.GetId(),
		Email: resp.GetEmail(),
	}, nil
}

func (c *GRPCClient) DeleteFriend(ctx context.Context, userID, friendID int64) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	_, err := c.client.DeleteFriend(ctx, &userv1.DeleteFriendRequest{
		UserId:   userID,
		FriendId: friendID,
	})

	return err
}

func (c *GRPCClient) AddMovieToFavorites(ctx context.Context, userID, movieID int64) (dto.FavoriteMovieResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.client.AddMovieToFavorites(ctx, &userv1.AddMovieToFavoritesRequest{
		UserId:  userID,
		MovieId: movieID,
	})
	if err != nil {
		return dto.FavoriteMovieResponse{}, err
	}

	return dto.FavoriteMovieResponse{
		MovieID:    resp.GetMovieId(),
		IsFavorite: resp.GetIsFavorite(),
	}, nil
}
