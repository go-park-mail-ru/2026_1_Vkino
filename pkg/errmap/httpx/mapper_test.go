package httpx

import (
	"errors"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestMapStatusError(t *testing.T) {
	t.Parallel()

	m := New([]codes.Code{codes.NotFound}, map[codes.Code]ErrResponse{
		codes.NotFound: {Status: 404, Message: ""},
	}, 500, "internal")

	statusCode, message := m.Map(status.Error(codes.NotFound, "missing"))
	if statusCode != 404 || message != "missing" {
		t.Fatalf("unexpected result: %d %q", statusCode, message)
	}
}

func TestMapDefault(t *testing.T) {
	t.Parallel()

	m := New([]codes.Code{codes.NotFound}, map[codes.Code]ErrResponse{}, 500, "internal")

	statusCode, message := m.Map(errors.New("unknown"))
	if statusCode != 500 || message != "internal" {
		t.Fatalf("unexpected default: %d %q", statusCode, message)
	}
}
