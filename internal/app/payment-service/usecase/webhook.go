package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/payment-service/domain"
)

func parseWebhookNotification(body []byte) (webhookNotification, error) {
	var notification webhookNotification
	if err := json.Unmarshal(body, &notification); err != nil {
		return webhookNotification{}, domain.ErrWebhookInvalidPayload
	}

	if notification.Type != "notification" || notification.Object.ID == "" {
		return webhookNotification{}, domain.ErrWebhookInvalidPayload
	}

	return notification, nil
}

func webhookPayloadHash(body []byte) string {
	payloadHash := sha256.Sum256(body)

	return hex.EncodeToString(payloadHash[:])
}

func (u *Usecase) registerWebhookEvent(
	ctx context.Context,
	notification webhookNotification,
	body []byte,
) (bool, error) {
	registered, err := u.payments.TryRegisterWebhookEvent(
		ctx,
		notification.Object.ID,
		notification.Event,
		webhookPayloadHash(body),
	)
	if err != nil {
		return false, fmt.Errorf("%w: %w", domain.ErrInternal, err)
	}

	return registered, nil
}

func (u *Usecase) dispatchWebhookEvent(ctx context.Context, notification webhookNotification) error {
	ykPayment, err := u.yookassa.GetPayment(ctx, notification.Object.ID)
	if err != nil {
		return err
	}

	payment, err := u.payments.GetPaymentByYooKassaID(ctx, notification.Object.ID)
	if err != nil {
		return err
	}

	switch notification.Event {
	case "payment.succeeded":
		if ykPayment.Status != "succeeded" {
			return domain.ErrWebhookInvalidPayload
		}

		return u.finalizeSucceededPayment(ctx, payment)
	case "payment.canceled":
		return u.finalizeCanceledPayment(ctx, payment)
	default:
		return nil
	}
}
