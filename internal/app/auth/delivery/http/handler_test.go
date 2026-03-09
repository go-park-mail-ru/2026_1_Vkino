package http

import (
	"context"
	"encoding/json"
	stdhttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/usecase"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/pkg/middleware"
)

type mockUsecase struct {
	signInFn               func(email, password string) (domain.TokenPair, error)
	signUpFn               func(email, password string) (domain.TokenPair, error)
	refreshFn              func(email string) (domain.TokenPair, error)
	validateRefreshTokenFn func(token string) (string, error)
	validateAccessTokenFn  func(token string) (string, error)
	getConfigFn            func() usecase.Config
}

func (m *mockUsecase) SignIn(email, password string) (domain.TokenPair, error) {
	if m.signInFn == nil {
		panic("unexpected call: SignIn")
	}
	return m.signInFn(email, password)
}

func (m *mockUsecase) SignUp(email, password string) (domain.TokenPair, error) {
	if m.signUpFn == nil {
		panic("unexpected call: SignUp")
	}
	return m.signUpFn(email, password)
}

func (m *mockUsecase) Refresh(email string) (domain.TokenPair, error) {
	if m.refreshFn == nil {
		panic("unexpected call: Refresh")
	}
	return m.refreshFn(email)
}

func (m *mockUsecase) ValidateRefreshToken(token string) (string, error) {
	if m.validateRefreshTokenFn == nil {
		panic("unexpected call: ValidateRefreshToken")
	}
	return m.validateRefreshTokenFn(token)
}

func (m *mockUsecase) ValidateAccessToken(token string) (string, error) {
	if m.validateAccessTokenFn == nil {
		panic("unexpected call: ValidateAccessToken")
	}
	return m.validateAccessTokenFn(token)
}

func (m *mockUsecase) GetConfig() usecase.Config {
	if m.getConfigFn == nil {
		panic("unexpected call: GetConfig")
	}
	return m.getConfigFn()
}

func testConfig() usecase.Config {
	return usecase.Config{
		RefreshTokenTTL:   24 * time.Hour,
		RefreshCookieName: "refresh_token",
		CookieSecure:      true,
	}
}

func mustJSON(t *testing.T, v any) string {
	t.Helper()

	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal json: %v", err)
	}

	return string(data)
}

func decodeBody[T any](t *testing.T, rr *httptest.ResponseRecorder) T {
	t.Helper()

	var v T
	if err := json.Unmarshal(rr.Body.Bytes(), &v); err != nil {
		t.Fatalf("decode body: %v; body=%s", err, rr.Body.String())
	}

	return v
}

func assertCookie(t *testing.T, rr *httptest.ResponseRecorder, wantName, wantValue string) {
	t.Helper()

	cookies := rr.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}

	got := cookies[0]
	if got.Name != wantName {
		t.Fatalf("expected cookie name %q, got %q", wantName, got.Name)
	}
	if got.Value != wantValue {
		t.Fatalf("expected cookie value %q, got %q", wantValue, got.Value)
	}
}

func assertNoCookies(t *testing.T, rr *httptest.ResponseRecorder) {
	t.Helper()

	if len(rr.Result().Cookies()) != 0 {
		t.Fatalf("expected no cookies, got %d", len(rr.Result().Cookies()))
	}
}

func assertJSONContainsStringValue(t *testing.T, rr *httptest.ResponseRecorder, want string) {
	t.Helper()

	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode generic body: %v; body=%s", err, rr.Body.String())
	}

	for _, value := range body {
		if s, ok := value.(string); ok && s == want {
			return
		}
	}

	t.Fatalf("expected body to contain string value %q, got %v", want, body)
}

func TestHandler_SignUp(t *testing.T) {
	t.Parallel()

	cfg := testConfig()

	tests := []struct {
		name              string
		body              string
		signUpResp        domain.TokenPair
		signUpErr         error
		wantStatus        int
		wantErrorValue    string
		wantAccessToken   string
		wantCookie        bool
		wantUsecaseCalled bool
		wantEmail         string
		wantPassword      string
	}{
		{
			name:              "invalid json body",
			body:              `{"email":"user@example.com",`,
			wantStatus:        stdhttp.StatusBadRequest,
			wantErrorValue:    "invalid json body",
			wantUsecaseCalled: false,
		},
		{
			name: "user already exists",
			body: mustJSON(t, domain.SignUpRequest{
				Email:    "user@example.com",
				Password: "qwerty",
			}),
			signUpErr:         domain.ErrUserAlreadyExists,
			wantStatus:        stdhttp.StatusConflict,
			wantErrorValue:    "user already exists",
			wantUsecaseCalled: true,
			wantEmail:         "user@example.com",
			wantPassword:      "qwerty",
		},
		{
			name: "success",
			body: mustJSON(t, domain.SignUpRequest{
				Email:    "user@example.com",
				Password: "qwerty",
			}),
			signUpResp: domain.TokenPair{
				AccessToken:  "access-1",
				RefreshToken: "refresh-1",
			},
			wantStatus:        stdhttp.StatusCreated,
			wantAccessToken:   "access-1",
			wantCookie:        true,
			wantUsecaseCalled: true,
			wantEmail:         "user@example.com",
			wantPassword:      "qwerty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var called bool
			var gotEmail string
			var gotPassword string

			mu := &mockUsecase{
				getConfigFn: func() usecase.Config { return cfg },
				signUpFn: func(email, password string) (domain.TokenPair, error) {
					called = true
					gotEmail = email
					gotPassword = password
					return tt.signUpResp, tt.signUpErr
				},
			}

			h := NewHandler(mu)

			req := httptest.NewRequest(stdhttp.MethodPost, "/signup", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			h.SignUp(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rr.Code)
			}

			if called != tt.wantUsecaseCalled {
				t.Fatalf("expected SignUp called=%v, got %v", tt.wantUsecaseCalled, called)
			}

			if called {
				if gotEmail != tt.wantEmail {
					t.Fatalf("expected email %q, got %q", tt.wantEmail, gotEmail)
				}
				if gotPassword != tt.wantPassword {
					t.Fatalf("expected password %q, got %q", tt.wantPassword, gotPassword)
				}
			}

			if tt.wantErrorValue != "" {
				assertJSONContainsStringValue(t, rr, tt.wantErrorValue)
				assertNoCookies(t, rr)
				return
			}

			got := decodeBody[domain.AccessTokenResponse](t, rr)
			if got.AccessToken != tt.wantAccessToken {
				t.Fatalf("expected access token %q, got %q", tt.wantAccessToken, got.AccessToken)
			}

			if tt.wantCookie {
				assertCookie(t, rr, cfg.RefreshCookieName, tt.signUpResp.RefreshToken)
			} else {
				assertNoCookies(t, rr)
			}
		})
	}
}

func TestHandler_SignIn(t *testing.T) {
	t.Parallel()

	cfg := testConfig()

	tests := []struct {
		name              string
		body              string
		signInResp        domain.TokenPair
		signInErr         error
		wantStatus        int
		wantErrorValue    string
		wantAccessToken   string
		wantCookie        bool
		wantUsecaseCalled bool
		wantEmail         string
		wantPassword      string
	}{
		{
			name:              "invalid json body",
			body:              `{"email":"user@example.com"`,
			wantStatus:        stdhttp.StatusBadRequest,
			wantErrorValue:    "invalid json body",
			wantUsecaseCalled: false,
		},
		{
			name: "invalid credentials",
			body: mustJSON(t, domain.SignInRequest{
				Email:    "user@example.com",
				Password: "wrong",
			}),
			signInErr:         domain.ErrInvalidCredentials,
			wantStatus:        stdhttp.StatusUnauthorized,
			wantErrorValue:    "invalid credentials",
			wantUsecaseCalled: true,
			wantEmail:         "user@example.com",
			wantPassword:      "wrong",
		},
		{
			name: "success",
			body: mustJSON(t, domain.SignInRequest{
				Email:    "user@example.com",
				Password: "qwerty",
			}),
			signInResp: domain.TokenPair{
				AccessToken:  "access-2",
				RefreshToken: "refresh-2",
			},
			wantStatus:        stdhttp.StatusOK,
			wantAccessToken:   "access-2",
			wantCookie:        true,
			wantUsecaseCalled: true,
			wantEmail:         "user@example.com",
			wantPassword:      "qwerty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var called bool
			var gotEmail string
			var gotPassword string

			mu := &mockUsecase{
				getConfigFn: func() usecase.Config { return cfg },
				signInFn: func(email, password string) (domain.TokenPair, error) {
					called = true
					gotEmail = email
					gotPassword = password
					return tt.signInResp, tt.signInErr
				},
			}

			h := NewHandler(mu)

			req := httptest.NewRequest(stdhttp.MethodPost, "/signin", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			h.SignIn(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rr.Code)
			}

			if called != tt.wantUsecaseCalled {
				t.Fatalf("expected SignIn called=%v, got %v", tt.wantUsecaseCalled, called)
			}

			if called {
				if gotEmail != tt.wantEmail {
					t.Fatalf("expected email %q, got %q", tt.wantEmail, gotEmail)
				}
				if gotPassword != tt.wantPassword {
					t.Fatalf("expected password %q, got %q", tt.wantPassword, gotPassword)
				}
			}

			if tt.wantErrorValue != "" {
				assertJSONContainsStringValue(t, rr, tt.wantErrorValue)
				assertNoCookies(t, rr)
				return
			}

			got := decodeBody[domain.AccessTokenResponse](t, rr)
			if got.AccessToken != tt.wantAccessToken {
				t.Fatalf("expected access token %q, got %q", tt.wantAccessToken, got.AccessToken)
			}

			if tt.wantCookie {
				assertCookie(t, rr, cfg.RefreshCookieName, tt.signInResp.RefreshToken)
			} else {
				assertNoCookies(t, rr)
			}
		})
	}
}

func TestHandler_Refresh(t *testing.T) {
	t.Parallel()

	cfg := testConfig()

	tests := []struct {
		name               string
		refreshCookieValue string
		validateEmail      string
		validateErr        error
		refreshResp        domain.TokenPair
		refreshErr         error
		wantStatus         int
		wantErrorValue     string
		wantAccessToken    string
		wantCookie         bool
		wantValidateCalled bool
		wantRefreshCalled  bool
	}{
		{
			name:               "no refresh cookie",
			wantStatus:         stdhttp.StatusUnauthorized,
			wantErrorValue:     "unauthorized",
			wantValidateCalled: false,
			wantRefreshCalled:  false,
		},
		{
			name:               "invalid refresh token",
			refreshCookieValue: "bad-token",
			validateErr:        domain.ErrInvalidToken,
			wantStatus:         stdhttp.StatusUnauthorized,
			wantErrorValue:     "unauthorized",
			wantValidateCalled: true,
			wantRefreshCalled:  false,
		},
		{
			name:               "no session",
			refreshCookieValue: "good-token",
			validateEmail:      "user@example.com",
			refreshErr:         domain.ErrNoSession,
			wantStatus:         stdhttp.StatusUnauthorized,
			wantErrorValue:     "unauthorized",
			wantValidateCalled: true,
			wantRefreshCalled:  true,
		},
		{
			name:               "success",
			refreshCookieValue: "good-token",
			validateEmail:      "user@example.com",
			refreshResp: domain.TokenPair{
				AccessToken:  "new-access",
				RefreshToken: "new-refresh",
			},
			wantStatus:         stdhttp.StatusOK,
			wantAccessToken:    "new-access",
			wantCookie:         true,
			wantValidateCalled: true,
			wantRefreshCalled:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var validateCalled bool
			var refreshCalled bool
			var gotRefreshEmail string
			var gotValidateToken string

			mu := &mockUsecase{
				getConfigFn: func() usecase.Config { return cfg },
				validateRefreshTokenFn: func(token string) (string, error) {
					validateCalled = true
					gotValidateToken = token
					return tt.validateEmail, tt.validateErr
				},
				refreshFn: func(email string) (domain.TokenPair, error) {
					refreshCalled = true
					gotRefreshEmail = email
					return tt.refreshResp, tt.refreshErr
				},
			}

			h := NewHandler(mu)

			req := httptest.NewRequest(stdhttp.MethodPost, "/refresh", nil)
			if tt.refreshCookieValue != "" {
				req.AddCookie(&stdhttp.Cookie{
					Name:  cfg.RefreshCookieName,
					Value: tt.refreshCookieValue,
				})
			}

			rr := httptest.NewRecorder()
			h.Refresh(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rr.Code)
			}

			if validateCalled != tt.wantValidateCalled {
				t.Fatalf("expected ValidateRefreshToken called=%v, got %v", tt.wantValidateCalled, validateCalled)
			}
			if refreshCalled != tt.wantRefreshCalled {
				t.Fatalf("expected Refresh called=%v, got %v", tt.wantRefreshCalled, refreshCalled)
			}

			if tt.wantValidateCalled && gotValidateToken != tt.refreshCookieValue {
				t.Fatalf("expected refresh token %q, got %q", tt.refreshCookieValue, gotValidateToken)
			}
			if tt.wantRefreshCalled && gotRefreshEmail != tt.validateEmail {
				t.Fatalf("expected refresh email %q, got %q", tt.validateEmail, gotRefreshEmail)
			}

			if tt.wantErrorValue != "" {
				assertJSONContainsStringValue(t, rr, tt.wantErrorValue)
				assertNoCookies(t, rr)
				return
			}

			got := decodeBody[domain.AccessTokenResponse](t, rr)
			if got.AccessToken != tt.wantAccessToken {
				t.Fatalf("expected access token %q, got %q", tt.wantAccessToken, got.AccessToken)
			}

			if tt.wantCookie {
				assertCookie(t, rr, cfg.RefreshCookieName, tt.refreshResp.RefreshToken)
			} else {
				assertNoCookies(t, rr)
			}
		})
	}
}

func TestHandler_Me(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		ctxEmail       string
		wantStatus     int
		wantStringBody string
	}{
		{
			name:           "unauthorized",
			wantStatus:     stdhttp.StatusUnauthorized,
			wantStringBody: "unauthorized",
		},
		{
			name:           "success",
			ctxEmail:       "user@example.com",
			wantStatus:     stdhttp.StatusOK,
			wantStringBody: "user@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler(&mockUsecase{})

			req := httptest.NewRequest(stdhttp.MethodGet, "/me", nil)
			if tt.ctxEmail != "" {
				ctx := context.WithValue(req.Context(), middleware.UserEmailKey, tt.ctxEmail)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()
			h.Me(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rr.Code)
			}

			assertJSONContainsStringValue(t, rr, tt.wantStringBody)
		})
	}
}

func TestHandler_Logout(t *testing.T) {
	t.Parallel()

	cfg := testConfig()

	tests := []struct {
		name             string
		hasRefreshCookie bool
		wantStatus       int
		wantMessage      string
		wantDeleteCookie bool
	}{
		{
			name:             "user was not authorized",
			hasRefreshCookie: false,
			wantStatus:       stdhttp.StatusOK,
			wantMessage:      "user wasn't authorized",
			wantDeleteCookie: false,
		},
		{
			name:             "successfully log out",
			hasRefreshCookie: true,
			wantStatus:       stdhttp.StatusOK,
			wantMessage:      "successfully log out",
			wantDeleteCookie: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler(&mockUsecase{
				getConfigFn: func() usecase.Config { return cfg },
			})

			req := httptest.NewRequest(stdhttp.MethodPost, "/logout", nil)
			if tt.hasRefreshCookie {
				req.AddCookie(&stdhttp.Cookie{
					Name:  cfg.RefreshCookieName,
					Value: "refresh-token",
				})
			}

			rr := httptest.NewRecorder()
			h.Logout(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rr.Code)
			}

			assertJSONContainsStringValue(t, rr, tt.wantMessage)

			setCookieHeader := rr.Header().Get("Set-Cookie")
			if tt.wantDeleteCookie {
				if setCookieHeader == "" {
					t.Fatal("expected delete cookie header, got empty")
				}
				if !strings.Contains(setCookieHeader, cfg.RefreshCookieName+"=") {
					t.Fatalf("expected delete cookie for %q, got %q", cfg.RefreshCookieName, setCookieHeader)
				}
				if !strings.Contains(setCookieHeader, "Expires=") {
					t.Fatalf("expected Expires in delete cookie, got %q", setCookieHeader)
				}
			} else if setCookieHeader != "" {
				t.Fatalf("expected no Set-Cookie header, got %q", setCookieHeader)
			}
		})
	}
}

func TestHandler_RefreshCookie(t *testing.T) {
	t.Parallel()

	cfg := testConfig()
	h := NewHandler(&mockUsecase{
		getConfigFn: func() usecase.Config { return cfg },
	})

	before := time.Now()
	cookie := h.RefreshCookie("refresh-value")
	after := time.Now()

	if cookie.Name != cfg.RefreshCookieName {
		t.Fatalf("expected cookie name %q, got %q", cfg.RefreshCookieName, cookie.Name)
	}
	if cookie.Value != "refresh-value" {
		t.Fatalf("expected cookie value %q, got %q", "refresh-value", cookie.Value)
	}
	if cookie.Path != "/" {
		t.Fatalf("expected path %q, got %q", "/", cookie.Path)
	}
	if !cookie.HttpOnly {
		t.Fatal("expected HttpOnly=true")
	}
	if cookie.Secure != cfg.CookieSecure {
		t.Fatalf("expected Secure=%v, got %v", cfg.CookieSecure, cookie.Secure)
	}
	if cookie.SameSite != stdhttp.SameSiteLaxMode {
		t.Fatalf("expected SameSite=%v, got %v", stdhttp.SameSiteLaxMode, cookie.SameSite)
	}

	minExpires := before.Add(cfg.RefreshTokenTTL).Add(-time.Second)
	maxExpires := after.Add(cfg.RefreshTokenTTL).Add(time.Second)

	if cookie.Expires.Before(minExpires) || cookie.Expires.After(maxExpires) {
		t.Fatalf("expected expires in [%v, %v], got %v", minExpires, maxExpires, cookie.Expires)
	}
}
