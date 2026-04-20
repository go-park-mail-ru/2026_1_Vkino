package grpcx

import (
	"errors"
	"net/http"

	httppkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/http"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func HTTPStatusFromGRPC(err error) (int, string) {
	if err == nil {
		return http.StatusOK, ""
	}

	st, ok := status.FromError(err)
	if !ok {
		return http.StatusInternalServerError, "internal server error"
	}

	switch st.Code() {
	case codes.InvalidArgument:
		return http.StatusBadRequest, st.Message()
	case codes.NotFound:
		return http.StatusNotFound, st.Message()
	case codes.AlreadyExists:
		return http.StatusConflict, st.Message()
	case codes.Unauthenticated, codes.PermissionDenied:
		return http.StatusUnauthorized, "unauthorized"
	case codes.FailedPrecondition:
		return http.StatusPreconditionFailed, st.Message()
	case codes.Unavailable:
		return http.StatusBadGateway, "service unavailable"
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}

func WriteError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}

	statusCode, message := HTTPStatusFromGRPC(err)
	if message == "" {
		message = errors.New("internal server error").Error()
	}

	httppkg.ErrResponse(w, statusCode, message)
}
