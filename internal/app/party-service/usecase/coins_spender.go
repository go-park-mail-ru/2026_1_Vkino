package usecase

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/party-service/domain"
	userv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/user/v1"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/service/authctx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const betPlaceOperationType = "bet_place"

type CoinsSpender interface {
	SpendForPollVote(
		ctx context.Context,
		userID, roomID, pollID, optionID int64,
		coinsAmount int32,
	) error
	RewardForPollWin(
		ctx context.Context,
		userID, roomID, pollID, optionID int64,
		coinsAmount int32,
	) error
}

type userServiceCoinsSpender struct {
	client userv1.UserServiceClient
}

func NewCoinsSpender(client userv1.UserServiceClient) CoinsSpender {
	return &userServiceCoinsSpender{client: client}
}

func (s *userServiceCoinsSpender) SpendForPollVote(
	ctx context.Context,
	userID, roomID, pollID, optionID int64,
	coinsAmount int32,
) error {
	if s == nil || s.client == nil || coinsAmount <= 0 {
		return nil
	}

	ctx = appendOutgoingAuthorization(ctx)

	_, err := s.client.SpendVKinoCoins(ctx, &userv1.SpendVKinoCoinsRequest{
		UserId:        userID,
		CoinsAmount:   coinsAmount,
		OperationType: betPlaceOperationType,
		Description:   pollVoteSpendDescription(coinsAmount, pollID, roomID, optionID),
	})
	if err == nil {
		return nil
	}

	return mapSpendForPollVoteError(err)
}

func (s *userServiceCoinsSpender) RewardForPollWin(
	ctx context.Context,
	userID, roomID, pollID, optionID int64,
	coinsAmount int32,
) error {
	if s == nil || s.client == nil || coinsAmount <= 0 {
		return nil
	}

	ctx = appendOutgoingAuthorization(ctx)

	description := fmt.Sprintf(
		"Выигрыш %d VKino coins в опросе %d комнаты %d за вариант %d",
		coinsAmount,
		pollID,
		roomID,
		optionID,
	)

	referenceKey := fmt.Sprintf("poll-win:%d:%d:%d", pollID, optionID, userID)

	_, err := s.client.GrantVKinoCoins(ctx, &userv1.GrantVKinoCoinsRequest{
		UserId:        userID,
		CoinsAmount:   coinsAmount,
		OperationType: "bet_win",
		Description:   description,
		ReferenceKey:  referenceKey,
	})
	if err != nil {
		return fmt.Errorf("grant vkino coins for poll win: %w", err)
	}

	return nil
}

func appendOutgoingAuthorization(ctx context.Context) context.Context {
	if auth, err := authctx.FromContext(ctx); err == nil && auth.Authorization != "" {
		return authctx.AppendOutgoing(ctx, auth.Authorization)
	}

	return ctx
}

func pollVoteSpendDescription(coinsAmount int32, pollID, roomID, optionID int64) string {
	return fmt.Sprintf(
		"Ставка %d VKino coins в опросе %d комнаты %d на вариант %d",
		coinsAmount,
		pollID,
		roomID,
		optionID,
	)
}

func mapSpendForPollVoteError(err error) error {
	st, ok := status.FromError(err)
	if ok && st.Code() == codes.FailedPrecondition && st.Message() == "insufficient vkino coins" {
		return domain.ErrInsufficientVKinoCoins
	}

	return fmt.Errorf("spend vkino coins for poll vote: %w", err)
}
