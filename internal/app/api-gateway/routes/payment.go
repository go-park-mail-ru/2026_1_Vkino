package routes

import (
	"io"
	"net/http"

	paymentv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/payment/v1"
	httppkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/http"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/httpserver"
)

//go:generate go run -mod=mod github.com/mailru/easyjson/easyjson -all -disallow_unknown_fields payment.go

//nolint:recvcheck // easyjson generates Marshal* on value receiver and Unmarshal* on pointer receiver.
type createPaymentRequest struct {
	ProductType   string `json:"product_type"`
	ProductRefID  int64  `json:"product_ref_id"`
	PaymentMethod string `json:"payment_method,omitempty"`
}

//nolint:recvcheck // easyjson generates Marshal* on value receiver and Unmarshal* on pointer receiver.
type createPaymentResponse struct {
	PaymentID         int64  `json:"payment_id"`
	Status            string `json:"status"`
	ConfirmationURL   string `json:"confirmation_url"`
	PaymentMethod     string `json:"payment_method"`
	CoinsSpent        *int32 `json:"coins_spent,omitempty"`
	VkinoCoinsBalance *int32 `json:"vkino_coins_balance,omitempty"`
}

//nolint:recvcheck // easyjson generates Marshal* on value receiver and Unmarshal* on pointer receiver.
type getPaymentResponse struct {
	PaymentID       int64   `json:"payment_id"`
	ProductType     string  `json:"product_type"`
	ProductRefID    int64   `json:"product_ref_id"`
	Status          string  `json:"status"`
	Amount          string  `json:"amount"`
	Currency        string  `json:"currency"`
	PaymentMethod   string  `json:"payment_method"`
	ConfirmationURL *string `json:"confirmation_url,omitempty"`
	PaidAt          *string `json:"paid_at,omitempty"`
	CoinsSpent      *int32  `json:"coins_spent,omitempty"`
}

//nolint:recvcheck // easyjson generates Marshal* on value receiver and Unmarshal* on pointer receiver.
type moneyTariffResponse struct {
	ID                      int64  `json:"id"`
	Code                    string `json:"code"`
	Title                   string `json:"title"`
	PriceMoney              int32  `json:"price_money"`
	PriceVkinoCoins         int32  `json:"price_vkino_coins"`
	IsMoneyPaymentAvailable bool   `json:"is_money_payment_available"`
	IsCoinsPaymentAvailable bool   `json:"is_coins_payment_available"`
	DurationDays            int32  `json:"duration_days"`
	Level                   int32  `json:"level"`
}

//nolint:recvcheck // easyjson generates Marshal* on value receiver and Unmarshal* on pointer receiver.
type listMoneyTariffsResponse struct {
	Tariffs []moneyTariffResponse `json:"tariffs"`
}

func Payment(cfg Config, client paymentv1.PaymentServiceClient) []httpserver.Option {
	return []httpserver.Option{
		route("POST /payments", createPaymentHandler(cfg, client)),
		route("GET /payments/{id}", getPaymentHandler(cfg, client)),
		route("GET /payments/tariffs", listMoneyTariffsHandler(cfg, client)),
		route("POST /payments/webhook/yookassa", yooKassaWebhookHandler(cfg, client)),
	}
}

func createPaymentHandler(cfg Config, client paymentv1.PaymentServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cancel := grpcContext(r, cfg.PaymentRequestTimeout())
		defer cancel()

		var req createPaymentRequest
		if !readJSON(w, r, &req) {
			return
		}

		resp, err := client.CreatePayment(r.Context(), &paymentv1.CreatePaymentRequest{
			ProductType:   req.ProductType,
			ProductRefId:  req.ProductRefID,
			PaymentMethod: req.PaymentMethod,
		})
		if err != nil {
			writeGRPCError(w, err)

			return
		}

		httppkg.Response(w, http.StatusCreated, createPaymentResponse{
			PaymentID:         resp.GetPaymentId(),
			Status:            resp.GetStatus(),
			ConfirmationURL:   resp.GetConfirmationUrl(),
			PaymentMethod:     resp.GetPaymentMethod(),
			CoinsSpent:        resp.CoinsSpent,
			VkinoCoinsBalance: resp.VkinoCoinsBalance,
		})
	}
}

func getPaymentHandler(cfg Config, client paymentv1.PaymentServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cancel := grpcContext(r, cfg.PaymentRequestTimeout())
		defer cancel()

		paymentID, ok := parsePathID(w, r, "invalid payment id")
		if !ok {
			return
		}

		resp, err := client.GetPayment(r.Context(), &paymentv1.GetPaymentRequest{
			PaymentId: paymentID,
		})
		if err != nil {
			writeGRPCError(w, err)

			return
		}

		httppkg.Response(w, http.StatusOK, getPaymentResponse{
			PaymentID:       resp.GetPaymentId(),
			ProductType:     resp.GetProductType(),
			ProductRefID:    resp.GetProductRefId(),
			Status:          resp.GetStatus(),
			Amount:          resp.GetAmount(),
			Currency:        resp.GetCurrency(),
			PaymentMethod:   resp.GetPaymentMethod(),
			ConfirmationURL: resp.ConfirmationUrl,
			PaidAt:          resp.PaidAt,
			CoinsSpent:      resp.CoinsSpent,
		})
	}
}

func listMoneyTariffsHandler(cfg Config, client paymentv1.PaymentServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cancel := grpcContext(r, cfg.PaymentRequestTimeout())
		defer cancel()

		resp, err := client.ListMoneyTariffs(r.Context(), &paymentv1.ListMoneyTariffsRequest{})
		if err != nil {
			writeGRPCError(w, err)

			return
		}

		tariffs := make([]moneyTariffResponse, 0, len(resp.GetTariffs()))
		for _, tariff := range resp.GetTariffs() {
			tariffs = append(tariffs, moneyTariffResponse{
				ID:                      tariff.GetId(),
				Code:                    tariff.GetCode(),
				Title:                   tariff.GetTitle(),
				PriceMoney:              tariff.GetPriceMoney(),
				PriceVkinoCoins:         tariff.GetPriceVkinoCoins(),
				IsMoneyPaymentAvailable: tariff.GetIsMoneyPaymentAvailable(),
				IsCoinsPaymentAvailable: tariff.GetIsCoinsPaymentAvailable(),
				DurationDays:            tariff.GetDurationDays(),
				Level:                   tariff.GetLevel(),
			})
		}

		httppkg.Response(w, http.StatusOK, listMoneyTariffsResponse{Tariffs: tariffs})
	}
}

func yooKassaWebhookHandler(cfg Config, client paymentv1.PaymentServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cancel := grpcContext(r, cfg.PaymentRequestTimeout())
		defer cancel()

		body, err := io.ReadAll(r.Body)
		if err != nil {
			httppkg.ErrResponse(w, http.StatusBadRequest, "invalid request body")

			return
		}

		clientIP := r.RemoteAddr
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			clientIP = forwarded
		}

		_, err = client.HandleYooKassaWebhook(r.Context(), &paymentv1.HandleYooKassaWebhookRequest{
			Body:     body,
			ClientIp: clientIP,
		})
		if err != nil {
			writeGRPCError(w, err)

			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
