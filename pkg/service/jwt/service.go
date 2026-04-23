package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("invalid token")

type Config struct {
	Secret string
	Issuer string
}

type AuthContext struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
}

type Service interface {
	GenerateToken(userEmail string, userID int64, tokenTTL time.Duration) (string, error)
	ParseToken(tokenString string) (AuthContext, error)
}

type JWTService struct {
	cfg Config
}

type CustomClaims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

func New(cfg Config) *JWTService {
	return &JWTService{cfg: cfg}
}

func (s *JWTService) GenerateToken(userEmail string, userID int64, tokenTTL time.Duration) (string, error) {
	now := time.Now()

	claims := CustomClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   userEmail,
			Issuer:    s.cfg.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	stringToken, err := token.SignedString([]byte(s.cfg.Secret))
	if err != nil {
		return "", err
	}

	return stringToken, nil
}

func (s *JWTService) ParseToken(tokenString string) (AuthContext, error) {
	claims := &CustomClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if token.Method == nil || token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, ErrInvalidToken
		}

		return []byte(s.cfg.Secret), nil
	})
	if err != nil {
		return AuthContext{}, fmt.Errorf("%w: %w", ErrInvalidToken, err)
	}

	parsedClaims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return AuthContext{}, ErrInvalidToken
	}

	if parsedClaims.Subject == "" {
		return AuthContext{}, ErrInvalidToken
	}

	return AuthContext{
		UserID: parsedClaims.UserID,
		Email:  parsedClaims.Subject,
	}, nil
}
