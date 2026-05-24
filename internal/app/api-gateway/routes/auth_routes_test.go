package routes

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	authv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/auth/v1"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestAuthRoutes_SignUp(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockAuthServiceClient(ctrl)

	cfg := testConfig{refreshCookieName: testRefreshCookieName, cookieSecure: true}

	client.EXPECT().SignUp(gomock.Any(), &authv1.SignUpRequest{
		Email:    testUserEmail,
		Password: "pass",
	}).Return(&authv1.SignUpResponse{
		AccessToken:  testAccessToken,
		RefreshToken: testRefreshTokenValue,
	}, nil)

	handler := newAuthHandler(t, cfg, client)
	rr := doRequest(handler, http.MethodPost, "/user/sign-up",
		bytes.NewReader(fmt.Appendf(nil, `{"email":%q,"password":"pass"}`, testUserEmail)))

	res := rr.Result()

	defer func() {
		_ = res.Body.Close()
	}()

	require.Equal(t, http.StatusCreated, res.StatusCode)
	require.JSONEq(t, `{"access_token":"access"}`, rr.Body.String())

	cookies := res.Cookies()
	require.Len(t, cookies, 1)
	require.Equal(t, testRefreshCookieName, cookies[0].Name)
	require.Equal(t, testRefreshTokenValue, cookies[0].Value)
	require.True(t, cookies[0].HttpOnly)
	require.True(t, cookies[0].Secure)
}

func TestAuthRoutes_SignUp_InvalidJSON(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockAuthServiceClient(ctrl)

	handler := newAuthHandler(t, testConfig{}, client)
	rr := doRequest(handler, http.MethodPost, "/user/sign-up", bytes.NewReader([]byte(`{"email":1}`)))

	requireJSONError(t, rr, http.StatusBadRequest, "invalid json body")
}

func TestAuthRoutes_SignIn(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockAuthServiceClient(ctrl)

	cfg := testConfig{refreshCookieName: testRefreshCookieName, cookieSecure: false}

	client.EXPECT().SignIn(gomock.Any(), &authv1.SignInRequest{
		Email:    testUserEmail,
		Password: "pass",
	}).Return(&authv1.SignInResponse{
		AccessToken:  testAccessToken,
		RefreshToken: testRefreshTokenValue,
	}, nil)

	handler := newAuthHandler(t, cfg, client)
	rr := doRequest(handler, http.MethodPost, "/user/sign-in",
		bytes.NewReader(fmt.Appendf(nil, `{"email":%q,"password":"pass"}`, testUserEmail)))

	require.Equal(t, http.StatusOK, rr.Code)
	require.JSONEq(t, `{"access_token":"access"}`, rr.Body.String())

	cookies := rr.Result().Cookies()
	require.Len(t, cookies, 1)
	require.Equal(t, testRefreshCookieName, cookies[0].Name)
	require.Equal(t, testRefreshTokenValue, cookies[0].Value)
}

func TestAuthRoutes_Refresh_MissingCookie(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockAuthServiceClient(ctrl)

	cfg := testConfig{refreshCookieName: testRefreshCookieName}

	handler := newAuthHandler(t, cfg, client)
	rr := doRequest(handler, http.MethodPost, "/user/refresh", nil)

	requireJSONError(t, rr, http.StatusUnauthorized, "unauthorized")
}

func TestAuthRoutes_Refresh(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockAuthServiceClient(ctrl)

	cfg := testConfig{refreshCookieName: testRefreshCookieName}

	client.EXPECT().Refresh(gomock.Any(), &authv1.RefreshRequest{RefreshToken: testRefreshTokenValue}).
		Return(&authv1.RefreshResponse{AccessToken: testAccessToken, RefreshToken: "new-refresh"}, nil)

	handler := newAuthHandler(t, cfg, client)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/user/refresh", nil)
	req.AddCookie(&http.Cookie{
		Name:     testRefreshCookieName,
		Value:    testRefreshTokenValue,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.JSONEq(t, `{"access_token":"access"}`, rr.Body.String())

	cookies := rr.Result().Cookies()
	require.Len(t, cookies, 1)
	require.Equal(t, "new-refresh", cookies[0].Value)
}

func TestAuthRoutes_Logout(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockAuthServiceClient(ctrl)

	cfg := testConfig{refreshCookieName: testRefreshCookieName, cookieSecure: true}

	client.EXPECT().Logout(gomock.Any(), &authv1.LogoutRequest{}).
		Return(&authv1.LogoutResponse{}, nil)

	handler := newAuthHandler(t, cfg, client)
	rr := doRequest(handler, http.MethodPost, "/user/logout", nil)

	require.Equal(t, http.StatusOK, rr.Code)

	cookies := rr.Result().Cookies()
	require.Len(t, cookies, 1)
	require.Equal(t, testRefreshCookieName, cookies[0].Name)
	require.Equal(t, -1, cookies[0].MaxAge)
}

func TestAuthRoutes_ChangePassword(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockAuthServiceClient(ctrl)

	client.EXPECT().ChangePassword(gomock.Any(), &authv1.ChangePasswordRequest{
		OldPassword: "old",
		NewPassword: "new",
	}).Return(&authv1.ChangePasswordResponse{}, nil)

	handler := newAuthHandler(t, testConfig{}, client)
	rr := doRequest(handler, http.MethodPost, "/user/change-password",
		bytes.NewReader([]byte(`{"old_password":"old","new_password":"new"}`)))

	require.Equal(t, http.StatusOK, rr.Code)
}
