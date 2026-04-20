package auth

import (
	"context"
	"errors"
	"net/http"
	"time"

	authgrpc "github.com/go-park-mail-ru/2026_1_VKino/services/api-gateway/internal/client/authgrpc"
	"github.com/go-park-mail-ru/2026_1_VKino/services/api-gateway/internal/config"
	"github.com/go-park-mail-ru/2026_1_VKino/services/api-gateway/internal/delivery/http/middleware"
	"github.com/go-park-mail-ru/2026_1_VKino/services/api-gateway/internal/dto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Facade struct {
	authClient authgrpc.Client
	cfg        config.UserAuthConfig
}

func NewFacade(authClient authgrpc.Client, cfg config.UserAuthConfig) *Facade {
	return &Facade{
		authClient: authClient,
		cfg:        cfg,
	}
}

func (f *Facade) SignUp(ctx context.Context, req dto.SignUpRequest) (dto.TokenPair, *http.Cookie, int, error) {
	tokens, err := f.authClient.SignUp(ctx, req.Email, req.Password)
	if err != nil {
		statusCode, mappedErr := grpcToHTTPError(err)
		return dto.TokenPair{}, nil, statusCode, mappedErr
	}

	return tokens, f.refreshCookie(tokens.RefreshToken), http.StatusCreated, nil
}

func (f *Facade) SignIn(ctx context.Context, req dto.SignInRequest) (dto.TokenPair, *http.Cookie, int, error) {
	tokens, err := f.authClient.SignIn(ctx, req.Email, req.Password)
	if err != nil {
		statusCode, mappedErr := grpcToHTTPError(err)
		return dto.TokenPair{}, nil, statusCode, mappedErr
	}

	return tokens, f.refreshCookie(tokens.RefreshToken), http.StatusOK, nil
}

func (f *Facade) Refresh(ctx context.Context, r *http.Request) (dto.AccessTokenResponse, *http.Cookie, int, error) {
	cookie, err := r.Cookie(f.cfg.RefreshCookieName)
	if err != nil {
		return dto.AccessTokenResponse{}, nil, http.StatusUnauthorized, errors.New("unauthorized")
	}

	tokens, err := f.authClient.Refresh(ctx, cookie.Value)
	if err != nil {
		statusCode, mappedErr := grpcToHTTPError(err)
		return dto.AccessTokenResponse{}, nil, statusCode, mappedErr
	}

	return dto.AccessTokenResponse{
		AccessToken: tokens.AccessToken,
	}, f.refreshCookie(tokens.RefreshToken), http.StatusOK, nil
}

func (f *Facade) LogOut(ctx context.Context, r *http.Request) (dto.Response, *http.Cookie, int, error) {
	authCtx, err := middleware.AuthFromContext(r.Context())
	if err != nil {
		return dto.Response{}, nil, http.StatusUnauthorized, errors.New("unauthorized")
	}

	if err := f.authClient.Logout(ctx, authCtx.Email); err != nil {
		statusCode, mappedErr := grpcToHTTPError(err)
		return dto.Response{}, nil, statusCode, mappedErr
	}

	return dto.Response{
		Message: "successfully log out",
	}, f.deleteRefreshCookie(), http.StatusOK, nil
}

func (f *Facade) ChangePassword(
	ctx context.Context,
	r *http.Request,
	req dto.ChangePasswordRequest,
) (dto.Response, int, error) {
	authCtx, err := middleware.AuthFromContext(r.Context())
	if err != nil {
		return dto.Response{}, http.StatusUnauthorized, errors.New("unauthorized")
	}

	if err := f.authClient.ChangePassword(ctx, authCtx.UserID, req.OldPassword, req.NewPassword); err != nil {
		statusCode, mappedErr := grpcToHTTPError(err)
		return dto.Response{}, statusCode, mappedErr
	}

	return dto.Response{
		Message: "password updated",
	}, http.StatusOK, nil
}

func (f *Facade) refreshCookie(refreshToken string) *http.Cookie {
	return &http.Cookie{
		Name:     f.cfg.RefreshCookieName,
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   f.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(f.cfg.RefreshTokenTTL),
	}
}

func (f *Facade) deleteRefreshCookie() *http.Cookie {
	return &http.Cookie{
		Name:     f.cfg.RefreshCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   f.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	}
}

func grpcToHTTPError(err error) (int, error) {
	if err == nil {
		return http.StatusOK, nil
	}

	st, ok := status.FromError(err)
	if !ok {
		return http.StatusInternalServerError, errors.New("internal server error")
	}

	switch st.Code() {
	case codes.AlreadyExists:
		return http.StatusConflict, errors.New(st.Message())
	case codes.NotFound:
		return http.StatusNotFound, errors.New(st.Message())
	case codes.InvalidArgument:
		return http.StatusBadRequest, errors.New(st.Message())
	case codes.Unauthenticated, codes.PermissionDenied:
		return http.StatusUnauthorized, errors.New("unauthorized")
	case codes.Unavailable:
		return http.StatusBadGateway, errors.New("auth service unavailable")
	default:
		return http.StatusInternalServerError, errors.New("internal server error")
	}
}
