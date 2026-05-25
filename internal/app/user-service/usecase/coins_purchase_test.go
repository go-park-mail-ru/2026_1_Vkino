package usecase

import (
	"context"
	"testing"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/repository/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestBuySubscriptionWithVKinoCoins_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockUserRepo(ctrl)
	repo.EXPECT().
		BuySubscriptionWithVKinoCoins(gomock.Any(), int64(42), int64(2)).
		Return(domain.VKinoCoinsSubscriptionPurchase{
			PaymentID:         123,
			CoinsSpent:        20,
			VKinoCoinsBalance: 80,
			Subscription: domain.SubscriptionInfo{
				ID:    2,
				Code:  "level_2",
				Name:  "Level 2",
				Level: 2,
			},
		}, nil)

	u := NewUserUsecase(repo, nil, nil)

	purchase, err := u.BuySubscriptionWithVKinoCoins(context.Background(), 42, 2)
	require.NoError(t, err)
	require.Equal(t, int64(123), purchase.PaymentID)
	require.Equal(t, int32(20), purchase.CoinsSpent)
	require.Equal(t, int32(80), purchase.VKinoCoinsBalance)
}

func TestBuySubscriptionWithVKinoCoins_InsufficientBalance(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockUserRepo(ctrl)
	repo.EXPECT().
		BuySubscriptionWithVKinoCoins(gomock.Any(), int64(42), int64(2)).
		Return(domain.VKinoCoinsSubscriptionPurchase{}, domain.ErrInsufficientVKinoCoins)

	u := NewUserUsecase(repo, nil, nil)

	_, err := u.BuySubscriptionWithVKinoCoins(context.Background(), 42, 2)
	require.ErrorIs(t, err, domain.ErrInsufficientVKinoCoins)
}
