package usecase

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/domain"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/repository/mocks"
	"github.com/stretchr/testify/require"
)

var errDeleteFriendFailed = errors.New("delete friend failed")

func TestUserUsecaseAddFriend(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockUserRepo(ctrl)
	u := NewUserUsecase(repo, nil, nil)

	friend := &domain.User{
		ID:    9,
		Email: "friend@example.com",
	}

	repo.EXPECT().AddFriend(gomock.Any(), int64(42), int64(9)).Return(friend, nil)

	result, err := u.AddFriend(context.Background(), 42, 9)
	require.NoError(t, err)
	require.Equal(t, domain.FriendResponse{
		ID:    9,
		Email: "friend@example.com",
	}, result)
}

func TestUserUsecaseAddFriendReturnsDomainErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockUserRepo(ctrl)
	u := NewUserUsecase(repo, nil, nil)

	repo.EXPECT().AddFriend(gomock.Any(), int64(42), int64(9)).Return(nil, domain.ErrAlreadyFriends)

	_, err := u.AddFriend(context.Background(), 42, 9)
	require.ErrorIs(t, err, domain.ErrAlreadyFriends)
}

func TestUserUsecaseDeleteFriendUsesRepositoryDirectly(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockUserRepo(ctrl)
	u := NewUserUsecase(repo, nil, nil)

	repo.EXPECT().DeleteFriend(gomock.Any(), int64(42), int64(9)).Return(nil)

	err := u.DeleteFriend(context.Background(), 42, 9)
	require.NoError(t, err)
}

func TestUserUsecaseDeleteFriendWrapsUnexpectedError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockUserRepo(ctrl)
	u := NewUserUsecase(repo, nil, nil)

	repo.EXPECT().DeleteFriend(gomock.Any(), int64(42), int64(9)).Return(errDeleteFriendFailed)

	err := u.DeleteFriend(context.Background(), 42, 9)
	require.ErrorIs(t, err, errDeleteFriendFailed)
}
