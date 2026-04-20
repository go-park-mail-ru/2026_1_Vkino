package authgrpc

import (
	"context"
	"fmt"
	"time"

	authv1 "github.com/go-park-mail-ru/2026_1_VKino/platform/gen/auth/v1"
	"github.com/go-park-mail-ru/2026_1_VKino/services/api-gateway/internal/dto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Config struct {
	Address        string
	RequestTimeout time.Duration
}

type Client interface {
	SignUp(ctx context.Context, email, password string) (dto.TokenPair, error)
	SignIn(ctx context.Context, email, password string) (dto.TokenPair, error)
	Refresh(ctx context.Context, refreshToken string) (dto.TokenPair, error)
	Validate(ctx context.Context, accessToken string) (dto.AuthContext, error)
	Logout(ctx context.Context, email string) error
	ChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error
	Close() error
}

type GRPCClient struct {
	conn    *grpc.ClientConn
	client  authv1.AuthServiceClient
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
		return nil, fmt.Errorf("dial auth grpc: %w", err)
	}

	return &GRPCClient{
		conn:    conn,
		client:  authv1.NewAuthServiceClient(conn),
		timeout: timeout,
	}, nil
}

func (c *GRPCClient) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func (c *GRPCClient) SignUp(ctx context.Context, email, password string) (dto.TokenPair, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.client.SignUp(ctx, &authv1.SignUpRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return dto.TokenPair{}, err
	}

	return dto.TokenPair{
		AccessToken:  resp.GetAccessToken(),
		RefreshToken: resp.GetRefreshToken(),
	}, nil
}

func (c *GRPCClient) SignIn(ctx context.Context, email, password string) (dto.TokenPair, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.client.SignIn(ctx, &authv1.SignInRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return dto.TokenPair{}, err
	}

	return dto.TokenPair{
		AccessToken:  resp.GetAccessToken(),
		RefreshToken: resp.GetRefreshToken(),
	}, nil
}

func (c *GRPCClient) Refresh(ctx context.Context, refreshToken string) (dto.TokenPair, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.client.Refresh(ctx, &authv1.RefreshRequest{
		RefreshToken: refreshToken,
	})
	if err != nil {
		return dto.TokenPair{}, err
	}

	return dto.TokenPair{
		AccessToken:  resp.GetAccessToken(),
		RefreshToken: resp.GetRefreshToken(),
	}, nil
}

func (c *GRPCClient) Validate(ctx context.Context, accessToken string) (dto.AuthContext, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.client.Validate(ctx, &authv1.ValidateRequest{
		AccessToken: accessToken,
	})
	if err != nil {
		return dto.AuthContext{}, err
	}

	return dto.AuthContext{
		UserID: resp.GetUserId(),
		Email:  resp.GetEmail(),
	}, nil
}

func (c *GRPCClient) Logout(ctx context.Context, email string) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	_, err := c.client.Logout(ctx, &authv1.LogoutRequest{
		Email: email,
	})

	return err
}

func (c *GRPCClient) ChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	_, err := c.client.ChangePassword(ctx, &authv1.ChangePasswordRequest{
		UserId:      userID,
		OldPassword: oldPassword,
		NewPassword: newPassword,
	})

	return err
}
