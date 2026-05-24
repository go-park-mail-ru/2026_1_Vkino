package routes

import (
	"io"
	"net/http"

	paymentv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/payment/v1"
	httppkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/http"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/httpserver"
)

type createPaymentRequest struct {
	ProductType  string `json:"product_type"`
	ProductRefID int64  `json:"product_ref_id"`
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
			ProductType:  req.ProductType,
			ProductRefId: req.ProductRefID,
		})
		if err != nil {
			writeGRPCError(w, err)

			return
		}

		httppkg.Response(w, http.StatusCreated, map[string]any{
			"payment_id":       resp.GetPaymentId(),
			"status":           resp.GetStatus(),
			"confirmation_url": resp.GetConfirmationUrl(),
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

		body := map[string]any{
			"payment_id":     resp.GetPaymentId(),
			"product_type":   resp.GetProductType(),
			"product_ref_id": resp.GetProductRefId(),
			"status":         resp.GetStatus(),
			"amount":         resp.GetAmount(),
			"currency":       resp.GetCurrency(),
		}

		if resp.ConfirmationUrl != nil {
			body["confirmation_url"] = resp.GetConfirmationUrl()
		}

		if resp.PaidAt != nil {
			body["paid_at"] = resp.GetPaidAt()
		}

		httppkg.Response(w, http.StatusOK, body)
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

		tariffs := make([]map[string]any, 0, len(resp.GetTariffs()))
		for _, tariff := range resp.GetTariffs() {
			tariffs = append(tariffs, map[string]any{
				"id":            tariff.GetId(),
				"code":          tariff.GetCode(),
				"title":         tariff.GetTitle(),
				"price_money":   tariff.GetPriceMoney(),
				"duration_days": tariff.GetDurationDays(),
				"level":         tariff.GetLevel(),
			})
		}

		httppkg.Response(w, http.StatusOK, map[string]any{"tariffs": tariffs})
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
