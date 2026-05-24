package domain

import "time"

type ProductType string

const (
	ProductTypeSubscription ProductType = "subscription"
	ProductTypeCoins        ProductType = "coins"
	ProductTypePaidContent  ProductType = "paid_content"
)

type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusSucceeded PaymentStatus = "succeeded"
	PaymentStatusCanceled  PaymentStatus = "canceled"
)

type Payment struct {
	ID                int64
	UserID            int64
	ProductType       ProductType
	ProductRefID      int64
	Amount            string
	Currency          string
	Status            PaymentStatus
	YooKassaPaymentID *string
	IdempotencyKey    string
	ConfirmationURL   *string
	PaidAt            *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type SubscriptionTariff struct {
	ID                     int64
	Code                   string
	Title                  string
	PriceMoney             int32
	IsMoneyPaymentAvailable bool
	DurationDays           int32
	Level                  int32
}

type MoneyTariff struct {
	ID           int64
	Code         string
	Title        string
	PriceMoney   int32
	DurationDays int32
	Level        int32
}
