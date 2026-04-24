package httpx

import (
	"net/http"

	"google.golang.org/grpc/codes"
)

var DefaultMapper = New(
	[]codes.Code{
		codes.InvalidArgument,
		codes.NotFound,
		codes.AlreadyExists,
		codes.Unauthenticated,
		codes.PermissionDenied,
		codes.FailedPrecondition,
		codes.Unavailable,
		codes.Internal,
	},
	map[codes.Code]ErrResponse{
		codes.InvalidArgument:    {Status: http.StatusBadRequest, Message: ""},
		codes.NotFound:           {Status: http.StatusNotFound, Message: ""},
		codes.AlreadyExists:      {Status: http.StatusConflict, Message: ""},
		codes.Unauthenticated:    {Status: http.StatusUnauthorized, Message: "unauthorized"},
		codes.PermissionDenied:   {Status: http.StatusUnauthorized, Message: "unauthorized"},
		codes.FailedPrecondition: {Status: http.StatusPreconditionFailed, Message: ""},
		codes.Unavailable:        {Status: http.StatusBadGateway, Message: "service unavailable"},
		codes.Internal:           {Status: http.StatusInternalServerError, Message: "internal server error"},
	},
	http.StatusInternalServerError,
	"internal server error",
)
