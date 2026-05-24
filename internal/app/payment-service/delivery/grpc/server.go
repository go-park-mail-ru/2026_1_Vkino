package grpc

import (
	"context"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/payment-service/usecase"
	authv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/auth/v1"
	paymentv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/payment/v1"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/service/authctx"
)

type Server struct {
	paymentv1.UnimplementedPaymentServiceServer

	usecase    *usecase.Usecase
	authClient authv1.AuthServiceClient
}

func NewServer(u *usecase.Usecase, authClient authv1.AuthServiceClient) *Server {
	return &Server{
		usecase:    u,
		authClient: authClient,
	}
}

func (s *Server) authorize(ctx context.Context) (authctx.Context, error) {
	authCtx, err := authctx.ValidateIncomingContext(ctx, s.authClient)
	if err != nil {
		return authctx.Context{}, err
	}

	return authCtx, nil
}

func (s *Server) CreatePayment(
	ctx context.Context,
	req *paymentv1.CreatePaymentRequest,
) (*paymentv1.CreatePaymentResponse, error) {
	authCtx, err := s.authorize(ctx)
	if err != nil {
		return nil, err
	}

	result, err := s.usecase.CreatePayment(ctx, usecase.CreatePaymentInput{
		UserID:       authCtx.UserID,
		ProductType:  req.GetProductType(),
		ProductRefID: req.GetProductRefId(),
	})
	if err != nil {
		return nil, mapError(err)
	}

	return &paymentv1.CreatePaymentResponse{
		PaymentId:       result.PaymentID,
		Status:          result.Status,
		ConfirmationUrl: result.ConfirmationURL,
	}, nil
}

func (s *Server) GetPayment(
	ctx context.Context,
	req *paymentv1.GetPaymentRequest,
) (*paymentv1.GetPaymentResponse, error) {
	authCtx, err := s.authorize(ctx)
	if err != nil {
		return nil, err
	}

	payment, err := s.usecase.GetPayment(ctx, authCtx.UserID, req.GetPaymentId())
	if err != nil {
		return nil, mapError(err)
	}

	resp := &paymentv1.GetPaymentResponse{
		PaymentId:    payment.ID,
		ProductType:  string(payment.ProductType),
		ProductRefId: payment.ProductRefID,
		Status:       string(payment.Status),
		Amount:       payment.Amount,
		Currency:     payment.Currency,
	}

	if payment.ConfirmationURL != nil {
		resp.ConfirmationUrl = payment.ConfirmationURL
	}

	if payment.PaidAt != nil {
		paidAt := payment.PaidAt.Format(time.RFC3339)
		resp.PaidAt = &paidAt
	}

	return resp, nil
}

func (s *Server) ListMoneyTariffs(
	ctx context.Context,
	_ *paymentv1.ListMoneyTariffsRequest,
) (*paymentv1.ListMoneyTariffsResponse, error) {
	tariffs, err := s.usecase.ListMoneyTariffs(ctx)
	if err != nil {
		return nil, mapError(err)
	}

	resp := &paymentv1.ListMoneyTariffsResponse{
		Tariffs: make([]*paymentv1.MoneyTariff, 0, len(tariffs)),
	}

	for _, tariff := range tariffs {
		resp.Tariffs = append(resp.Tariffs, &paymentv1.MoneyTariff{
			Id:           tariff.ID,
			Code:         tariff.Code,
			Title:        tariff.Title,
			PriceMoney:   tariff.PriceMoney,
			DurationDays: tariff.DurationDays,
			Level:        tariff.Level,
		})
	}

	return resp, nil
}

func (s *Server) HandleYooKassaWebhook(
	ctx context.Context,
	req *paymentv1.HandleYooKassaWebhookRequest,
) (*paymentv1.HandleYooKassaWebhookResponse, error) {
	if err := s.usecase.HandleYooKassaWebhook(ctx, req.GetBody(), req.GetClientIp()); err != nil {
		return nil, mapError(err)
	}

	return &paymentv1.HandleYooKassaWebhookResponse{}, nil
}
