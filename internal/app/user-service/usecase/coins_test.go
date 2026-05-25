package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/repository/mocks"
	"go.uber.org/mock/gomock"
)

func TestGetVKinoCoinsHistory_DefaultLimit(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := mocks.NewMockUserRepo(ctrl)

	repo.EXPECT().GetUserByID(gomock.Any(), int64(42)).Return(&domain.User{ID: 42}, nil)
	repo.EXPECT().
		GetVKinoCoinsHistory(gomock.Any(), int64(42), defaultCoinsHistoryLimit, int32(0)).
		Return(nil, int32(0), nil)

	u := NewUserUsecase(repo, nil, nil)

	_, err := u.GetVKinoCoinsHistory(context.Background(), 42, 0, 0)
	if err != nil {
		t.Fatalf("GetVKinoCoinsHistory() error = %v", err)
	}
}

func TestGetVKinoCoinsHistory_MaxLimit(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := mocks.NewMockUserRepo(ctrl)

	repo.EXPECT().GetUserByID(gomock.Any(), int64(42)).Return(&domain.User{ID: 42}, nil)
	repo.EXPECT().
		GetVKinoCoinsHistory(gomock.Any(), int64(42), maxCoinsHistoryLimit, int32(3)).
		Return(nil, int32(0), nil)

	u := NewUserUsecase(repo, nil, nil)

	_, err := u.GetVKinoCoinsHistory(context.Background(), 42, 999, 3)
	if err != nil {
		t.Fatalf("GetVKinoCoinsHistory() error = %v", err)
	}
}

func TestGetVKinoCoinsHistory_NormalizesNegativeOffset(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := mocks.NewMockUserRepo(ctrl)

	repo.EXPECT().GetUserByID(gomock.Any(), int64(42)).Return(&domain.User{ID: 42}, nil)
	repo.EXPECT().
		GetVKinoCoinsHistory(gomock.Any(), int64(42), int32(25), int32(0)).
		Return(nil, int32(0), nil)

	u := NewUserUsecase(repo, nil, nil)

	_, err := u.GetVKinoCoinsHistory(context.Background(), 42, 25, -10)
	if err != nil {
		t.Fatalf("GetVKinoCoinsHistory() error = %v", err)
	}
}

func TestGetVKinoCoinsHistory_InvalidTokenOnUserLookupError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := mocks.NewMockUserRepo(ctrl)

	repo.EXPECT().GetUserByID(gomock.Any(), int64(42)).Return(nil, domain.ErrUserNotFound)

	u := NewUserUsecase(repo, nil, nil)

	_, err := u.GetVKinoCoinsHistory(context.Background(), 42, 10, 0)
	if !errors.Is(err, domain.ErrInvalidToken) {
		t.Fatalf("GetVKinoCoinsHistory() error = %v, want %v", err, domain.ErrInvalidToken)
	}
}

func TestGetVKinoCoinsHistory_InternalErrorOnRepositoryFailure(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := mocks.NewMockUserRepo(ctrl)

	repo.EXPECT().GetUserByID(gomock.Any(), int64(42)).Return(&domain.User{ID: 42}, nil)
	repo.EXPECT().
		GetVKinoCoinsHistory(gomock.Any(), int64(42), int32(10), int32(0)).
		Return(nil, int32(0), domain.ErrInternal)

	u := NewUserUsecase(repo, nil, nil)

	_, err := u.GetVKinoCoinsHistory(context.Background(), 42, 10, 0)
	if !errors.Is(err, domain.ErrInternal) {
		t.Fatalf("GetVKinoCoinsHistory() error = %v, want %v", err, domain.ErrInternal)
	}
}

func TestGetVKinoCoinsHistory_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	repo := mocks.NewMockUserRepo(ctrl)

	items := []domain.VKinoCoinsHistoryItem{
		{
			ID:              1,
			VKinoCoinsCount: 3,
			OperationType:   "daily",
			Description:     "Ежедневное начисление",
			CreatedAt:       "2026-05-26T12:34:56Z",
		},
	}

	repo.EXPECT().GetUserByID(gomock.Any(), int64(42)).Return(&domain.User{ID: 42}, nil)
	repo.EXPECT().
		GetVKinoCoinsHistory(gomock.Any(), int64(42), int32(10), int32(5)).
		Return(items, int32(1), nil)

	u := NewUserUsecase(repo, nil, nil)

	resp, err := u.GetVKinoCoinsHistory(context.Background(), 42, 10, 5)
	if err != nil {
		t.Fatalf("GetVKinoCoinsHistory() error = %v", err)
	}

	if resp.TotalCount != 1 {
		t.Fatalf("TotalCount = %d, want 1", resp.TotalCount)
	}

	if len(resp.Items) != 1 {
		t.Fatalf("len(Items) = %d, want 1", len(resp.Items))
	}

	if resp.Items[0] != items[0] {
		t.Fatalf("Items[0] = %+v, want %+v", resp.Items[0], items[0])
	}
}
