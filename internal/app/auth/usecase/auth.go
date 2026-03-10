package usecase

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Config struct {
	JWTSecret         string        `mapstructure:"jwt_secret"`
	AccessTokenTTL    time.Duration `mapstructure:"access_token_ttl"`
	RefreshTokenTTL   time.Duration `mapstructure:"refresh_token_ttl"`
	RefreshCookieName string        `mapstructure:"refresh_cookie_name"`
	CookieSecure      bool          `mapstructure:"cookie_secure"`
}

type AuthUsecase struct {
	userRepo    repository.UserRepo
	sessionRepo repository.SessionRepo
	cfg         Config
}

func NewAuthUsecase(userRepo repository.UserRepo, sessionRepo repository.SessionRepo, cfg Config) *AuthUsecase {
	return &AuthUsecase{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		cfg:         cfg,
	}
}

func (u *AuthUsecase) SignIn(email, password string) (domain.TokenPair, error) {
	user, err := u.userRepo.GetUserByEmail(email)
	if err != nil {
		return domain.TokenPair{}, domain.ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return domain.TokenPair{}, domain.ErrInvalidCredentials
	}

	return u.tokenPairGenerate(user)
}

func (u *AuthUsecase) SignUp(email, password string) (domain.TokenPair, error) {
	_, err := u.userRepo.GetUserByEmail(email)
	if err == nil {
		return domain.TokenPair{}, domain.ErrUserAlreadyExists
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return domain.TokenPair{}, errors.Join(ErrBcryptGenerate, err)
	}

	user, err := u.userRepo.CreateUser(email, string(passwordHash))
	if err != nil {
		return domain.TokenPair{}, err
	}

	return u.tokenPairGenerate(user)
}

// обновляем протухший refresh-token

func (u *AuthUsecase) Refresh(email string) (domain.TokenPair, error) {
	if _, err := u.sessionRepo.GetSession(email); err != nil {
		return domain.TokenPair{}, domain.ErrNoSession
	}

	user, err := u.userRepo.GetUserByEmail(email)
	if err != nil {
		return domain.TokenPair{}, domain.ErrNoSession
	}

	return u.tokenPairGenerate(user)
}

func (u *AuthUsecase) ValidateRefreshToken(tokenString string) (string, error) {
	claims, err := u.parseToken(tokenString)
	if err != nil {
		return "", err
	}

	if claims.Subject == "" {
		return "", domain.ErrInvalidToken
	}

	// refresh должен совпадать с сохранённым у пользователя
	tokenPair, err := u.sessionRepo.GetSession(claims.Subject)
	if err != nil {
		return "", domain.ErrNoSession
	}

	if tokenPair.RefreshToken != tokenString {
		return "", domain.ErrInvalidToken
	}

	return claims.Subject, nil
}

func (u *AuthUsecase) LogOut(email string) error {
	err := u.sessionRepo.DeleteSession(email)
	if err != nil {
		if errors.Is(err, domain.ErrNoSession) {
			return nil
		}

		return err
	}

	return nil
}

func (u *AuthUsecase) ValidateAccessToken(tokenString string) (string, error) {
	claims, err := u.parseToken(tokenString)
	if err != nil {
		return "", err
	}

	if claims.Subject == "" {
		return "", domain.ErrInvalidToken
	}

	return claims.Subject, nil
}

func (u *AuthUsecase) GetConfig() Config {
	return u.cfg
}

func (u *AuthUsecase) tokenGenerate(user string, tokenTTL time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenTTL)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Subject:   user,
	})

	stringToken, err := token.SignedString([]byte(u.cfg.JWTSecret))
	if err != nil {
		return "", err
	}

	return stringToken, nil
}

func (u *AuthUsecase) tokenPairGenerate(user *domain.User) (domain.TokenPair, error) {
	accessToken, err := u.tokenGenerate(user.Email, u.cfg.AccessTokenTTL)
	if err != nil {
		return domain.TokenPair{}, errors.Join(ErrTokenGenerate, err)
	}

	refreshToken, err := u.tokenGenerate(user.Email, u.cfg.RefreshTokenTTL)
	if err != nil {
		return domain.TokenPair{}, errors.Join(ErrTokenGenerate, err)
	}

	tokenPair := domain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	err = u.sessionRepo.SaveSession(user.Email, tokenPair)
	if err != nil {
		return domain.TokenPair{}, errors.Join(ErrSessionSave, err)
	}

	return tokenPair, nil
}

func (u *AuthUsecase) parseToken(tokenString string) (*jwt.RegisteredClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (any, error) {
		if token.Method == nil || token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, domain.ErrInvalidToken
		}

		return []byte(u.cfg.JWTSecret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return nil, domain.ErrInvalidToken
	}

	return claims, nil
}
