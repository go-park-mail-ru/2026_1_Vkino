package routes

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	paymentv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/payment/v1"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestPaymentRoutes_ListMoneyTariffs(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockPaymentServiceClient(ctrl)

	client.EXPECT().ListMoneyTariffs(gomock.Any(), &paymentv1.ListMoneyTariffsRequest{}).
		Return(&paymentv1.ListMoneyTariffsResponse{
			Tariffs: []*paymentv1.MoneyTariff{
				{
					Id:                      2,
					Code:                    "level_2",
					Title:                   "Level 2",
					PriceMoney:              299,
					PriceVkinoCoins:         20,
					IsCoinsPaymentAvailable: true,
					IsMoneyPaymentAvailable: true,
					DurationDays:            30,
					Level:                   2,
				},
			},
		}, nil)

	handler := newPaymentHandler(t, client)
	rr := doRequest(handler, http.MethodGet, "/payments/tariffs", nil)

	require.Equal(t, http.StatusOK, rr.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &body))
	tariffs, ok := body["tariffs"].([]any)
	require.True(t, ok)
	require.Len(t, tariffs, 1)
	firstTariff, ok := tariffs[0].(map[string]any)
	require.True(t, ok)
	require.EqualValues(t, 20, firstTariff["price_vkino_coins"])
}

func TestPaymentRoutes_CreatePayment(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockPaymentServiceClient(ctrl)

	client.EXPECT().CreatePayment(gomock.Any(), &paymentv1.CreatePaymentRequest{
		ProductType:  "subscription",
		ProductRefId: 2,
	}).Return(&paymentv1.CreatePaymentResponse{
		PaymentId:       42,
		Status:          "pending",
		ConfirmationUrl: "https://yoomoney.ru/checkout/42",
		PaymentMethod:   "yookassa",
	}, nil)

	handler := newPaymentHandler(t, client)
	rr := doRequest(handler, http.MethodPost, "/payments",
		bytes.NewReader([]byte(`{"product_type":"subscription","product_ref_id":2}`)))

	require.Equal(t, http.StatusCreated, rr.Code)
	require.JSONEq(t, `{
		"payment_id":42,
		"status":"pending",
		"confirmation_url":"https://yoomoney.ru/checkout/42",
		"payment_method":"yookassa"
	}`, rr.Body.String())
}

func TestPaymentRoutes_CreatePayment_VKinoCoins(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockPaymentServiceClient(ctrl)

	coinsSpent := int32(20)
	coinsBalance := int32(80)

	client.EXPECT().CreatePayment(gomock.Any(), &paymentv1.CreatePaymentRequest{
		ProductType:   "subscription",
		ProductRefId:  2,
		PaymentMethod: "vkino_coins",
	}).Return(&paymentv1.CreatePaymentResponse{
		PaymentId:         123,
		Status:            "succeeded",
		ConfirmationUrl:   "",
		PaymentMethod:     "vkino_coins",
		CoinsSpent:        &coinsSpent,
		VkinoCoinsBalance: &coinsBalance,
	}, nil)

	handler := newPaymentHandler(t, client)
	rr := doRequest(handler, http.MethodPost, "/payments",
		bytes.NewReader([]byte(`{"product_type":"subscription","product_ref_id":2,"payment_method":"vkino_coins"}`)))

	require.Equal(t, http.StatusCreated, rr.Code)
	require.JSONEq(t, `{
		"payment_id":123,
		"status":"succeeded",
		"confirmation_url":"",
		"payment_method":"vkino_coins",
		"coins_spent":20,
		"vkino_coins_balance":80
	}`, rr.Body.String())
}

func TestPaymentRoutes_CreatePayment_InvalidJSON(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockPaymentServiceClient(ctrl)

	handler := newPaymentHandler(t, client)
	rr := doRequest(handler, http.MethodPost, "/payments", bytes.NewReader([]byte(`{"product_type":1}`)))

	requireJSONError(t, rr, http.StatusBadRequest, "invalid json body")
}

func TestPaymentRoutes_CreatePayment_UnknownField(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockPaymentServiceClient(ctrl)

	handler := newPaymentHandler(t, client)
	rr := doRequest(
		handler,
		http.MethodPost,
		"/payments",
		bytes.NewReader([]byte(`{"product_type":"subscription","product_ref_id":2,"extra":"field"}`)),
	)

	requireJSONError(t, rr, http.StatusBadRequest, "invalid json body")
}

func TestPaymentRoutes_GetPayment(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockPaymentServiceClient(ctrl)

	client.EXPECT().GetPayment(gomock.Any(), &paymentv1.GetPaymentRequest{PaymentId: 42}).
		Return(&paymentv1.GetPaymentResponse{
			PaymentId:     42,
			ProductType:   "subscription",
			ProductRefId:  2,
			Status:        "succeeded",
			Amount:        "299.00",
			Currency:      "RUB",
			PaymentMethod: "yookassa",
		}, nil)

	handler := newPaymentHandler(t, client)
	rr := doRequest(handler, http.MethodGet, "/payments/42", nil)

	require.Equal(t, http.StatusOK, rr.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &body))
	require.EqualValues(t, 42, body["payment_id"])
	require.Equal(t, "succeeded", body["status"])
	require.Equal(t, "yookassa", body["payment_method"])
}

func TestPaymentRoutes_GetPayment_InvalidID(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockPaymentServiceClient(ctrl)

	handler := newPaymentHandler(t, client)
	rr := doRequest(handler, http.MethodGet, "/payments/abc", nil)

	requireJSONError(t, rr, http.StatusBadRequest, "invalid payment id")
}

func TestPaymentRoutes_GetPayment_NotFound(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockPaymentServiceClient(ctrl)

	client.EXPECT().GetPayment(gomock.Any(), &paymentv1.GetPaymentRequest{PaymentId: 99}).
		Return(nil, status.Error(codes.NotFound, "payment not found"))

	handler := newPaymentHandler(t, client)
	rr := doRequest(handler, http.MethodGet, "/payments/99", nil)

	requireJSONError(t, rr, http.StatusNotFound, "payment not found")
}

func TestPaymentRoutes_YooKassaWebhook(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockPaymentServiceClient(ctrl)

	body := []byte(`{"type":"notification","event":"payment.succeeded","object":{"id":"yk-1"}}`)

	client.EXPECT().HandleYooKassaWebhook(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ any, req *paymentv1.HandleYooKassaWebhookRequest, _ ...any) (*paymentv1.HandleYooKassaWebhookResponse, error) {
			require.Equal(t, body, req.GetBody())
			require.NotEmpty(t, req.GetClientIp())

			return &paymentv1.HandleYooKassaWebhookResponse{}, nil
		},
	)

	handler := newPaymentHandler(t, client)
	rr := doRequest(handler, http.MethodPost, "/payments/webhook/yookassa", bytes.NewReader(body))

	require.Equal(t, http.StatusOK, rr.Code)
}

func TestPaymentRoutes_YooKassaWebhook_InvalidIP(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockPaymentServiceClient(ctrl)

	client.EXPECT().HandleYooKassaWebhook(gomock.Any(), gomock.Any()).
		Return(nil, status.Error(codes.PermissionDenied, "access denied"))

	handler := newPaymentHandler(t, client)
	rr := doRequest(handler, http.MethodPost, "/payments/webhook/yookassa",
		bytes.NewReader([]byte(`{"type":"notification","event":"payment.succeeded","object":{"id":"yk-1"}}`)))

	requireJSONError(t, rr, http.StatusForbidden, "access denied")
}
