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

type PaymentMethod string

const (
	PaymentMethodYooKassa   PaymentMethod = "yookassa"
	PaymentMethodVKinoCoins PaymentMethod = "vkino_coins"
)

type Payment struct {
	ID                int64
	UserID            int64
	ProductType       ProductType
	ProductRefID      int64
	Amount            string
	Currency          string
	Status            PaymentStatus
	PaymentMethod     PaymentMethod
	CoinsSpent        *int32
	YooKassaPaymentID *string
	IdempotencyKey    string
	ConfirmationURL   *string
	PaidAt            *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type SubscriptionTariff struct {
	ID                      int64
	Code                    string
	Title                   string
	PriceMoney              int32
	PriceVKinoCoins         int32
	IsCoinsPaymentAvailable bool
	IsMoneyPaymentAvailable bool
	DurationDays            int32
	Level                   int32
}

type MoneyTariff struct {
	ID                      int64
	Code                    string
	Title                   string
	PriceMoney              int32
	PriceVKinoCoins         int32
	IsCoinsPaymentAvailable bool
	IsMoneyPaymentAvailable bool
	DurationDays            int32
	Level                   int32
}

type CoinsSubscriptionPurchase struct {
	PaymentID         int64
	CoinsSpent        int32
	VKinoCoinsBalance int32
}

type CoinsPack struct {
	ID          int64
	Code        string
	Title       string
	CoinsAmount int32
	PriceMoney  int32
}
