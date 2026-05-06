package authctx

import (
	"context"
	"testing"

	"google.golang.org/grpc/metadata"
)

func TestParseBearerToken(t *testing.T) {
	t.Parallel()

	token, err := ParseBearerToken("Bearer abc")
	if err != nil || token != "abc" {
		t.Fatalf("expected token abc, got %q, err=%v", token, err)
	}
}

func TestParseBearerTokenInvalid(t *testing.T) {
	t.Parallel()

	if _, err := ParseBearerToken("Token abc"); err == nil {
		t.Fatal("expected error")
	}
}

func TestAccessTokenFromIncomingContext(t *testing.T) {
	t.Parallel()

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(MetadataKey, "Bearer token"))
	got, err := AccessTokenFromIncomingContext(ctx)
	if err != nil || got != "token" {
		t.Fatalf("expected token, got %q, err=%v", got, err)
	}
}

func TestAppendOutgoing(t *testing.T) {
	t.Parallel()

	ctx := AppendOutgoing(context.Background(), "Bearer token")
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		t.Fatal("expected metadata")
	}

	values := md.Get(MetadataKey)
	if len(values) != 1 || values[0] != "Bearer token" {
		t.Fatalf("unexpected metadata: %v", values)
	}
}
