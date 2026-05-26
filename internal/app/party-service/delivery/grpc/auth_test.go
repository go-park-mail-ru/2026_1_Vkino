package grpc

import (
	"context"
	"testing"

	authv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/auth/v1"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/service/authctx"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestAuthorizeStoresAuthContext(t *testing.T) {
	t.Parallel()

	server := &Server{
		authClient: authClientStub{
			validateResp: &authv1.ValidateResponse{
				UserId: 42,
				Email:  "user@example.com",
			},
		},
	}

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(authctx.MetadataKey, "Bearer token-123"))

	nextCtx, got, err := server.authorize(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(42), got.UserID)
	require.Equal(t, "user@example.com", got.Email)
	require.Equal(t, "Bearer token-123", got.Authorization)

	stored, err := authctx.FromContext(nextCtx)
	require.NoError(t, err)
	require.Equal(t, got, stored)
}

type authClientStub struct {
	validateResp *authv1.ValidateResponse
	validateErr  error
}

func (s authClientStub) SignUp(
	context.Context,
	*authv1.SignUpRequest,
	...grpc.CallOption,
) (*authv1.SignUpResponse, error) {
	panic("unexpected call")
}

func (s authClientStub) SignIn(
	context.Context,
	*authv1.SignInRequest,
	...grpc.CallOption,
) (*authv1.SignInResponse, error) {
	panic("unexpected call")
}

func (s authClientStub) Refresh(
	context.Context,
	*authv1.RefreshRequest,
	...grpc.CallOption,
) (*authv1.RefreshResponse, error) {
	panic("unexpected call")
}

func (s authClientStub) Validate(
	context.Context,
	*authv1.ValidateRequest,
	...grpc.CallOption,
) (*authv1.ValidateResponse, error) {
	return s.validateResp, s.validateErr
}

func (s authClientStub) Logout(
	context.Context,
	*authv1.LogoutRequest,
	...grpc.CallOption,
) (*authv1.LogoutResponse, error) {
	panic("unexpected call")
}

func (s authClientStub) ChangePassword(
	context.Context,
	*authv1.ChangePasswordRequest,
	...grpc.CallOption,
) (*authv1.ChangePasswordResponse, error) {
	panic("unexpected call")
}
