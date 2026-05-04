package requestid

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
)

const (
	HeaderName    = "X-Request-ID"
	MetadataKey   = "x-request-id"
	contextKeyID  = contextKey("request_id")
	requestIDSize = 16
)

type contextKey string

func Generate() string {
	buf := make([]byte, requestIDSize)

	_, err := rand.Read(buf)
	if err != nil {
		return "unknown-request-id"
	}

	return hex.EncodeToString(buf)
}

func Normalize(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return Generate()
	}

	return trimmed
}

func ContextWithID(ctx context.Context, id string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	id = Normalize(id)

	return context.WithValue(ctx, contextKeyID, id)
}

func FromContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}

	value, ok := ctx.Value(contextKeyID).(string)
	if !ok || strings.TrimSpace(value) == "" {
		return "", false
	}

	return value, true
}
