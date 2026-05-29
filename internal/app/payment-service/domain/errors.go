package domain

import "errors"

var (
	ErrInvalidToken               = errors.New("invalid token")
	ErrInvalidProductType         = errors.New("invalid product type")
	ErrInvalidProductRef          = errors.New("invalid product reference")
	ErrInvalidPaymentMethod       = errors.New("invalid payment method")
	ErrTariffNotFound             = errors.New("tariff not found")
	ErrCoinsPackNotFound          = errors.New("coins pack not found")
	ErrTariffNotAvailable         = errors.New("tariff is not available for money payment")
	ErrTariffNotAvailableForCoins = errors.New("tariff is not available for vkino coins payment")
	ErrInsufficientVKinoCoins     = errors.New("insufficient vkino coins")
	ErrPaymentNotFound            = errors.New("payment not found")
	ErrPaymentAccessDenied        = errors.New("payment access denied")
	ErrPaymentAlreadyFinal        = errors.New("payment is already finalized")
	ErrWebhookInvalidIP           = errors.New("webhook ip is not allowed")
	ErrWebhookInvalidPayload      = errors.New("webhook payload is invalid")
	ErrYooKassaUnavailable        = errors.New("yookassa unavailable")
	ErrInternal                   = errors.New("internal error")
)
