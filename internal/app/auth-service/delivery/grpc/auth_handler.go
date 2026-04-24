package grpc

import (
	"context"

	authv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/auth/v1"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/service/authctx"
)

func (s *Server) SignUp(ctx context.Context, req *authv1.SignUpRequest) (*authv1.SignUpResponse, error) {
	tokens, err := s.usecase.SignUp(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, mapError(err)
	}

	return &authv1.SignUpResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (s *Server) SignIn(ctx context.Context, req *authv1.SignInRequest) (*authv1.SignInResponse, error) {
	tokens, err := s.usecase.SignIn(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, mapError(err)
	}

	return &authv1.SignInResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (s *Server) Refresh(ctx context.Context, req *authv1.RefreshRequest) (*authv1.RefreshResponse, error) {
	email, err := s.usecase.ValidateRefreshToken(ctx, req.GetRefreshToken())
	if err != nil {
		return nil, mapError(err)
	}

	tokens, err := s.usecase.Refresh(ctx, email)
	if err != nil {
		return nil, mapError(err)
	}

	return &authv1.RefreshResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (s *Server) Validate(ctx context.Context, req *authv1.ValidateRequest) (*authv1.ValidateResponse, error) {
	authCtx, err := s.usecase.ValidateAccessToken(req.GetAccessToken())
	if err != nil {
		return nil, mapError(err)
	}

	return &authv1.ValidateResponse{
		Valid:  true,
		UserId: authCtx.UserID,
		Email:  authCtx.Email,
	}, nil
}

func (s *Server) Logout(ctx context.Context, req *authv1.LogoutRequest) (*authv1.LogoutResponse, error) {
	accessToken, err := authctx.AccessTokenFromIncomingContext(ctx)
	if err != nil {
		return nil, mapError(err)
	}

	authData, err := s.usecase.ValidateAccessToken(accessToken)
	if err != nil {
		return nil, mapError(err)
	}

	err = s.usecase.LogOut(ctx, authData.Email)
	if err != nil {
		return nil, mapError(err)
	}

	return &authv1.LogoutResponse{Success: true}, nil
}

func (s *Server) ChangePassword(
	ctx context.Context,
	req *authv1.ChangePasswordRequest,
) (*authv1.ChangePasswordResponse, error) {
	accessToken, err := authctx.AccessTokenFromIncomingContext(ctx)
	if err != nil {
		return nil, mapError(err)
	}

	authData, err := s.usecase.ValidateAccessToken(accessToken)
	if err != nil {
		return nil, mapError(err)
	}

	err = s.usecase.ChangePassword(ctx, authData.UserID, req.GetOldPassword(), req.GetNewPassword())
	if err != nil {
		return nil, mapError(err)
	}

	return &authv1.ChangePasswordResponse{Success: true}, nil
}
