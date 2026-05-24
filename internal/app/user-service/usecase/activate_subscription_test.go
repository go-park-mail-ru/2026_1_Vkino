package usecase

import (
	"context"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/repository/mocks"
)

type fixedClock struct {
	now time.Time
}

func (c fixedClock) Now() time.Time {
	return c.now
}

func TestActivateSubscription_UpgradeStartsFromNow(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	activeUntil := now.AddDate(0, 2, 0).Format(time.RFC3339)
	repo := mocks.NewMockUserRepo(ctrl)

	repo.EXPECT().GetUserByID(gomock.Any(), int64(42)).Return(&domain.User{ID: 42}, nil)
	repo.EXPECT().
		GetSubscriptionTariffByID(gomock.Any(), int64(4)).
		Return(domain.SubscriptionTariff{
			ID:           4,
			Code:         "level_4",
			Title:        "Level 4",
			Level:        4,
			DurationDays: 30,
		}, nil)
	repo.EXPECT().
		GetActiveSubscription(gomock.Any(), int64(42)).
		Return(domain.SubscriptionInfo{
			ID:          2,
			Level:       2,
			ActiveUntil: &activeUntil,
		}, nil)
	repo.EXPECT().DeactivateUserSubscriptions(gomock.Any(), int64(42)).Return(nil)
	repo.EXPECT().
		CreateUserSubscription(gomock.Any(), int64(42), int64(4), now, now.AddDate(0, 0, 30)).
		Return(nil)

	u := NewUserUsecase(repo, nil, fixedClock{now: now})

	info, err := u.ActivateSubscription(context.Background(), 42, 4, 100)
	if err != nil {
		t.Fatalf("ActivateSubscription() error = %v", err)
	}

	wantUntil := now.AddDate(0, 0, 30).Format(time.RFC3339)
	if info.ActiveUntil == nil || *info.ActiveUntil != wantUntil {
		t.Fatalf("ActiveUntil = %v, want %v", info.ActiveUntil, wantUntil)
	}
}

func TestActivateSubscription_SameTierExtendsFromActiveUntil(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	activeUntilTime := now.AddDate(0, 2, 0)
	activeUntil := activeUntilTime.Format(time.RFC3339)
	repo := mocks.NewMockUserRepo(ctrl)

	repo.EXPECT().GetUserByID(gomock.Any(), int64(42)).Return(&domain.User{ID: 42}, nil)
	repo.EXPECT().
		GetSubscriptionTariffByID(gomock.Any(), int64(2)).
		Return(domain.SubscriptionTariff{
			ID:           2,
			Code:         "level_2",
			Title:        "Level 2",
			Level:        2,
			DurationDays: 30,
		}, nil)
	repo.EXPECT().
		GetActiveSubscription(gomock.Any(), int64(42)).
		Return(domain.SubscriptionInfo{
			ID:          2,
			Level:       2,
			ActiveUntil: &activeUntil,
		}, nil)
	repo.EXPECT().DeactivateUserSubscriptions(gomock.Any(), int64(42)).Return(nil)
	repo.EXPECT().
		CreateUserSubscription(
			gomock.Any(),
			int64(42),
			int64(2),
			now,
			activeUntilTime.AddDate(0, 0, 30),
		).
		Return(nil)

	u := NewUserUsecase(repo, nil, fixedClock{now: now})

	info, err := u.ActivateSubscription(context.Background(), 42, 2, 100)
	if err != nil {
		t.Fatalf("ActivateSubscription() error = %v", err)
	}

	wantUntil := activeUntilTime.AddDate(0, 0, 30).Format(time.RFC3339)
	if info.ActiveUntil == nil || *info.ActiveUntil != wantUntil {
		t.Fatalf("ActiveUntil = %v, want %v", info.ActiveUntil, wantUntil)
	}
}

func TestActivateSubscription_NoActiveSubscriptionStartsFromNow(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	repo := mocks.NewMockUserRepo(ctrl)

	repo.EXPECT().GetUserByID(gomock.Any(), int64(42)).Return(&domain.User{ID: 42}, nil)
	repo.EXPECT().
		GetSubscriptionTariffByID(gomock.Any(), int64(2)).
		Return(domain.SubscriptionTariff{
			ID:           2,
			Code:         "level_2",
			Title:        "Level 2",
			Level:        2,
			DurationDays: 30,
		}, nil)
	repo.EXPECT().
		GetActiveSubscription(gomock.Any(), int64(42)).
		Return(domain.SubscriptionInfo{}, nil)
	repo.EXPECT().DeactivateUserSubscriptions(gomock.Any(), int64(42)).Return(nil)
	repo.EXPECT().
		CreateUserSubscription(gomock.Any(), int64(42), int64(2), now, now.AddDate(0, 0, 30)).
		Return(nil)

	u := NewUserUsecase(repo, nil, fixedClock{now: now})

	info, err := u.ActivateSubscription(context.Background(), 42, 2, 100)
	if err != nil {
		t.Fatalf("ActivateSubscription() error = %v", err)
	}

	wantUntil := now.AddDate(0, 0, 30).Format(time.RFC3339)
	if info.ActiveUntil == nil || *info.ActiveUntil != wantUntil {
		t.Fatalf("ActiveUntil = %v, want %v", info.ActiveUntil, wantUntil)
	}
}
