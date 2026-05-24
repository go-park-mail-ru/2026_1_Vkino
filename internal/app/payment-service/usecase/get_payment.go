package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/payment-service/domain"
)

func (u *Usecase) refreshPendingPayment(
	ctx context.Context,
	paymentID int64,
	payment domain.Payment,
) (domain.Payment, error) {
	//nolint:errcheck // best-effort sync; re-read reflects canceled/succeeded when sync succeeds
	u.syncPendingPaymentFromYooKassa(ctx, payment)

	return u.payments.GetPaymentByID(ctx, paymentID)
}
