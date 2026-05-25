package routes

import (
	"net/http"
	"testing"

	userv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/user/v1"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUserRoutes_GetVKinoCoinsHistory(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockUserClient(ctrl)

	client.EXPECT().
		GetVKinoCoinsHistory(gomock.Any(), &userv1.GetVKinoCoinsHistoryRequest{Limit: 10, Offset: 5}).
		Return(&userv1.GetVKinoCoinsHistoryResponse{
			Items: []*userv1.VKinoCoinsHistoryItem{
				{
					Id:              123,
					VkinoCoinsCount: 3,
					OperationType:   "daily",
					Description:     "Ежедневное начисление",
					CreatedAt:       "2026-05-26T12:34:56Z",
				},
			},
			TotalCount: 1,
		}, nil)

	handler := newUserHandler(t, client)
	rr := doRequest(handler, http.MethodGet, "/user/coins/history?limit=10&offset=5", nil)

	require.Equal(t, http.StatusOK, rr.Code)
	require.JSONEq(t, `{
		"items": [
			{
				"id": 123,
				"vkino_coins_count": 3,
				"operation_type": "daily",
				"description": "Ежедневное начисление",
				"created_at": "2026-05-26T12:34:56Z"
			}
		],
		"total_count": 1
	}`, rr.Body.String())
}

func TestUserRoutes_GetVKinoCoinsHistory_GRPCError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockUserClient(ctrl)

	client.EXPECT().
		GetVKinoCoinsHistory(gomock.Any(), &userv1.GetVKinoCoinsHistoryRequest{Limit: defaultCollectionLimit, Offset: 0}).
		Return(nil, status.Error(codes.Internal, "internal server error"))

	handler := newUserHandler(t, client)
	rr := doRequest(handler, http.MethodGet, "/user/coins/history", nil)

	requireJSONError(t, rr, http.StatusInternalServerError, "internal server error")
}
