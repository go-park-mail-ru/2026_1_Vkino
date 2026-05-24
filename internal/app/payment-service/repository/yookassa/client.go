package yookassa

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/payment-service/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/payment-service/repository"
)

const defaultHTTPTimeout = 30 * time.Second

type Config struct {
	APIURL    string
	ShopID    string
	SecretKey string
	ReturnURL string
	Capture   bool
	Timeout   time.Duration
}

type Client struct {
	httpClient *http.Client
	cfg        Config
}

func NewClient(cfg Config) *Client {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = defaultHTTPTimeout
	}

	return &Client{
		httpClient: &http.Client{Timeout: timeout},
		cfg:        cfg,
	}
}

type createPaymentRequest struct {
	Amount struct {
		Value    string `json:"value"`
		Currency string `json:"currency"`
	} `json:"amount"`
	Capture      bool `json:"capture"`
	Confirmation struct {
		Type      string `json:"type"`
		ReturnURL string `json:"return_url"`
	} `json:"confirmation"`
	Description string            `json:"description"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type paymentResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Paid   bool   `json:"paid"`
	Amount struct {
		Value    string `json:"value"`
		Currency string `json:"currency"`
	} `json:"amount"`
	Confirmation struct {
		Type            string `json:"type"`
		ConfirmationURL string `json:"confirmation_url"`
	} `json:"confirmation"`
}

func (c *Client) CreatePayment(
	ctx context.Context,
	req repository.YooKassaCreateRequest,
) (repository.YooKassaPayment, error) {
	body := createPaymentRequest{}
	body.Amount.Value = req.AmountValue
	body.Amount.Currency = req.AmountCurrency
	body.Capture = req.Capture
	body.Confirmation.Type = "redirect"
	body.Confirmation.ReturnURL = req.ReturnURL
	body.Description = req.Description
	body.Metadata = req.Metadata

	var resp paymentResponse

	if err := c.do(ctx, http.MethodPost, "/payments", req.IdempotencyKey, body, &resp); err != nil {
		return repository.YooKassaPayment{}, err
	}

	return repository.YooKassaPayment{
		ID:              resp.ID,
		Status:          resp.Status,
		Paid:            resp.Paid,
		ConfirmationURL: resp.Confirmation.ConfirmationURL,
	}, nil
}

func (c *Client) GetPayment(ctx context.Context, paymentID string) (repository.YooKassaPayment, error) {
	var resp paymentResponse

	if err := c.do(ctx, http.MethodGet, "/payments/"+paymentID, "", nil, &resp); err != nil {
		return repository.YooKassaPayment{}, err
	}

	return repository.YooKassaPayment{
		ID:              resp.ID,
		Status:          resp.Status,
		Paid:            resp.Paid,
		ConfirmationURL: resp.Confirmation.ConfirmationURL,
	}, nil
}

func (c *Client) do(
	ctx context.Context,
	method, path, idempotencyKey string,
	body any,
	dst any,
) error {
	req, err := c.newHTTPRequest(ctx, method, path, idempotencyKey, body)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %w", domain.ErrYooKassaUnavailable, err)
	}
	defer func() { _ = resp.Body.Close() }()

	return c.decodeHTTPResponse(resp, dst)
}

func (c *Client) newHTTPRequest(
	ctx context.Context,
	method, path, idempotencyKey string,
	body any,
) (*http.Request, error) {
	apiURL := strings.TrimRight(c.cfg.APIURL, "/")
	if apiURL == "" {
		apiURL = "https://api.yookassa.ru/v3"
	}

	var reader io.Reader

	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("%w: marshal request: %w", domain.ErrInternal, err)
		}

		reader = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(ctx, method, apiURL+path, reader)
	if err != nil {
		return nil, fmt.Errorf("%w: build request: %w", domain.ErrInternal, err)
	}

	req.SetBasicAuth(c.cfg.ShopID, c.cfg.SecretKey)
	req.Header.Set("Content-Type", "application/json")

	if idempotencyKey != "" {
		req.Header.Set("Idempotence-Key", idempotencyKey)
	}

	return req, nil
}

func (c *Client) decodeHTTPResponse(resp *http.Response, dst any) error {
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("%w: read response: %w", domain.ErrInternal, err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("%w: status %d: %s", domain.ErrYooKassaUnavailable, resp.StatusCode, string(respBody))
	}

	if dst != nil && len(respBody) > 0 {
		if err = json.Unmarshal(respBody, dst); err != nil {
			return fmt.Errorf("%w: decode response: %w", domain.ErrInternal, err)
		}
	}

	return nil
}
