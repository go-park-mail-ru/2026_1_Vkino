package grpcx

import (
	"errors"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	errSentinel = errors.New("sentinel")
	errUnknown  = errors.New("unknown")
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

	m := New([]error{errSentinel}, map[error]ErrResponse{
		errSentinel: {Code: codes.NotFound, Message: "not found"},
	}, codes.Internal, "internal")

	err := m.Map(errSentinel)

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
	err := m.Map(errUnknown)

	st, _ := status.FromError(err)
	if st.Code() != codes.Internal {
		t.Fatalf("expected internal code, got %v", st.Code())
	}
}
