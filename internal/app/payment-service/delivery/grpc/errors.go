package grpc

import (
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/payment-service/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/errmap/grpcx"
	"google.golang.org/grpc/codes"
)

var paymentGRPCErrorMapper = grpcx.New(
	[]error{
		domain.ErrInvalidToken,
		domain.ErrInvalidProductType,
		domain.ErrInvalidProductRef,
		domain.ErrTariffNotFound,
		domain.ErrTariffNotAvailable,
		domain.ErrPaymentNotFound,
		domain.ErrPaymentAccessDenied,
		domain.ErrWebhookInvalidIP,
		domain.ErrWebhookInvalidPayload,
		domain.ErrYooKassaUnavailable,
		domain.ErrInternal,
	},
	map[error]grpcx.ErrResponse{
		domain.ErrInvalidToken:       {Code: codes.Unauthenticated, Message: "unauthorized"},
		domain.ErrInvalidProductType: {Code: codes.InvalidArgument, Message: "invalid product type"},
		domain.ErrInvalidProductRef:  {Code: codes.InvalidArgument, Message: "invalid product reference"},
		domain.ErrTariffNotFound:     {Code: codes.NotFound, Message: "tariff not found"},
		domain.ErrTariffNotAvailable: {
			Code: codes.FailedPrecondition, Message: "tariff is not available for money payment",
		},
		domain.ErrPaymentNotFound:       {Code: codes.NotFound, Message: "payment not found"},
		domain.ErrPaymentAccessDenied:   {Code: codes.PermissionDenied, Message: "payment access denied"},
		domain.ErrWebhookInvalidIP:      {Code: codes.PermissionDenied, Message: "webhook ip is not allowed"},
		domain.ErrWebhookInvalidPayload: {Code: codes.InvalidArgument, Message: "invalid webhook payload"},
		domain.ErrYooKassaUnavailable:   {Code: codes.Unavailable, Message: "payment provider unavailable"},
		domain.ErrInternal:              {Code: codes.Internal, Message: "internal server error"},
	},
	codes.Internal,
	"internal server error",
)

func mapError(err error) error {
	return paymentGRPCErrorMapper.Map(err)
}
