package grpcx

import (
	"errors"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestMapNil(t *testing.T) {
	t.Parallel()

	m := New([]error{}, map[error]ErrResponse{}, codes.Internal, "internal")
	if err := m.Map(nil); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestMapKnownError(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("sentinel")
	m := New([]error{sentinel}, map[error]ErrResponse{
		sentinel: {Code: codes.NotFound, Message: "not found"},
	}, codes.Internal, "internal")

	err := m.Map(sentinel)
	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("expected grpc status error")
	}

	if st.Code() != codes.NotFound || st.Message() != "not found" {
		t.Fatalf("unexpected status: %v", st)
	}
}

func TestMapDefault(t *testing.T) {
	t.Parallel()

	m := New([]error{}, map[error]ErrResponse{}, codes.Internal, "internal")
	err := m.Map(errors.New("unknown"))
	st, _ := status.FromError(err)
	if st.Code() != codes.Internal {
		t.Fatalf("expected internal code, got %v", st.Code())
	}
}
