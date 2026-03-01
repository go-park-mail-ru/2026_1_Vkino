package auth

import (
	"fmt"
	"time"

	apperrors "github.com/go-park-mail-ru/2026_1_VKino/internal/app/errors"
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

type Service struct {
	cfg Config
	userMap map[string]string
	userSessions map[string]TokenPair
}

func NewService(cfg Config) *Service {
	// дефолты
	if cfg.JWTSecret == "" {
		cfg.JWTSecret = "dev-secret"
	}
	if cfg.AccessTokenTTL == 0 {
		cfg.AccessTokenTTL = 15 * time.Minute
	}
	if cfg.RefreshTokenTTL == 0 {
		cfg.RefreshTokenTTL = 14 * 24 * time.Hour
	}
	if cfg.RefreshCookieName == "" {
		cfg.RefreshCookieName = "refresh_token"
	}
	// Саша, здесь нужно будет подключить твои модельки!
	// Мьютексы?
	return &Service{
		cfg: cfg,
		userMap: map[string]string{},
		userSessions: map[string]TokenPair{},
	}
}

// Саша, здесь логика на достать/создать, тоже меняй
func (s *Service) getUserByEmail(email string) (User, error) {
	passwordHash, exists := s.userMap[email]
	if !exists {
		return User{}, fmt.Errorf("no user")
	}
	return User{Email: email, Password: passwordHash}, nil
}

func (s *Service) createUser(email, passwordHash string) User {
	s.userMap[email] = passwordHash
	return User{Email: email, Password: passwordHash}
}
//

func (s *Service) userToString(user User) string {
	return user.Email
}

func (s *Service) tokenGenerate(user string, tokenTTL time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenTTL)),
		IssuedAt: jwt.NewNumericDate(time.Now()),
		Subject: user,
	})
	stringToken, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", err
	}
	return stringToken, nil
}

func (s *Service) tokenPairGenerate(user User) (TokenPair, error) {
	accessToken, err := s.tokenGenerate(s.userToString(user), s.cfg.AccessTokenTTL)
	if err != nil {
		return TokenPair{}, fmt.Errorf("access token generate error: %w", err)
	}
	refreshToken, err := s.tokenGenerate(s.userToString(user), s.cfg.RefreshTokenTTL)
	if err != nil {
		return TokenPair{}, fmt.Errorf("refresh token generate error: %w", err)
	}

	tokenPair := TokenPair{
		AccessToken: accessToken,
		RefreshToken: refreshToken,
	}

	// переписать
	s.userSessions[user.Email] = tokenPair
	return tokenPair, nil
}


func (s *Service) SignIn(email, password string) (TokenPair, error) {
	user, err := s.getUserByEmail(email)
	if err != nil {
		return TokenPair{}, apperrors.ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return TokenPair{}, apperrors.ErrInvalidCredentials
	}
	return s.tokenPairGenerate(user)
}


func (s *Service) SignUp(email, password string) (TokenPair, error) {
	_, err := s.getUserByEmail(email)
	if err == nil {
		return TokenPair{}, apperrors.ErrUserAlreadyExists
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return TokenPair{}, fmt.Errorf("bcrypt generate error: %w", err)
	}

	user := s.createUser(email, string(passwordHash))
	return s.tokenPairGenerate(user)
}


// обновляем протухший refresh-token
func (s *Service) refresh(email string) (TokenPair, error) {
	// переписать
	_, exists := s.userSessions[email]
	if !exists {
		return TokenPair{}, apperrors.ErrNoSession
	}

	user, err := s.getUserByEmail(email)
	if err != nil {
		return TokenPair{}, apperrors.ErrNoSession
	}

	return s.tokenPairGenerate(user)
}

func (s *Service) parseToken(tokenString string) (*jwt.RegisteredClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method == nil || token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, apperrors.ErrInvalidToken
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", apperrors.ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return nil, apperrors.ErrInvalidToken
	}

	return claims, nil
}

func (s *Service) validateAccessToken(tokenString string) (string, error) {
	claims, err := s.parseToken(tokenString)
	if err != nil {
		return "", err
	}

	if claims.Subject == "" {
		return "", apperrors.ErrInvalidToken
	}

	return claims.Subject, nil
}

func (s *Service) validateRefreshToken(tokenString string) (string, error) {
	claims, err := s.parseToken(tokenString)
	if err != nil {
		return "", err
	}

	if claims.Subject == "" {
		return "", apperrors.ErrInvalidToken
	}

	// refresh должен совпадать с сохранённым у пользователя
	tokenPair, exists := s.userSessions[claims.Subject]
	if !exists {
		return "", apperrors.ErrNoSession
	}

	if tokenPair.RefreshToken != tokenString {
		return "", apperrors.ErrInvalidToken
	}

	return claims.Subject, nil
}