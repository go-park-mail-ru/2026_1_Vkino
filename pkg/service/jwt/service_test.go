package jwt

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateAndParseToken(t *testing.T) {
	t.Parallel()

	svc := New(Config{Secret: "secret", Issuer: "issuer"})
	token, err := svc.GenerateToken("user@example.com", 42, time.Minute)
	if err != nil {
		t.Fatalf("GenerateToken error: %v", err)
	}

	ctx, err := svc.ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken error: %v", err)
	}

	if ctx.UserID != 42 || ctx.Email != "user@example.com" {
		t.Fatalf("unexpected auth context: %+v", ctx)
	}
}

func TestParseTokenInvalidMethod(t *testing.T) {
	t.Parallel()

	svc := New(Config{Secret: "secret", Issuer: "issuer"})

	claims := CustomClaims{RegisteredClaims: jwt.RegisteredClaims{Subject: "user@example.com"}}
	token := jwt.NewWithClaims(jwt.SigningMethodHS384, claims)
	signed, err := token.SignedString([]byte("secret"))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	_, err = svc.ParseToken(signed)
	if err == nil || !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected invalid token error, got %v", err)
	}
}

func TestParseTokenEmptySubject(t *testing.T) {
	t.Parallel()

	svc := New(Config{Secret: "secret", Issuer: "issuer"})

	claims := CustomClaims{RegisteredClaims: jwt.RegisteredClaims{Subject: ""}}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte("secret"))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	_, err = svc.ParseToken(signed)
	if err == nil {
		t.Fatal("expected error for empty subject")
	}
}
