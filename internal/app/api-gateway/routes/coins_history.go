package routes

import (
	"net/http"

	userv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/user/v1"
	httppkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/http"
)

type vkinoCoinsHistoryHTTPResponse struct {
	Items      []vkinoCoinsHistoryHTTPItem `json:"items"`
	TotalCount int32                       `json:"total_count"`
}

type vkinoCoinsHistoryHTTPItem struct {
	ID              int64  `json:"id"`
	VKinoCoinsCount int32  `json:"vkino_coins_count"`
	OperationType   string `json:"operation_type"`
	Description     string `json:"description"`
	CreatedAt       string `json:"created_at"`
}

func newUserVKinoCoinsHistoryHandler(cfg Config, userClient UserClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cancel := grpcContext(r, cfg.UserRequestTimeout())
		defer cancel()

		resp, err := userClient.GetVKinoCoinsHistory(r.Context(), &userv1.GetVKinoCoinsHistoryRequest{
			Limit:  parseInt32Query(r, "limit", defaultCollectionLimit),
			Offset: parseInt32Query(r, "offset", 0),
		})
		if err != nil {
			writeGRPCError(w, err)

			return
		}

		httppkg.Response(w, http.StatusOK, convertVKinoCoinsHistoryResponse(resp))
	}
}

func convertVKinoCoinsHistoryResponse(
	resp *userv1.GetVKinoCoinsHistoryResponse,
) vkinoCoinsHistoryHTTPResponse {
	if resp == nil {
		return vkinoCoinsHistoryHTTPResponse{
			Items: make([]vkinoCoinsHistoryHTTPItem, 0),
		}
	}

	items := make([]vkinoCoinsHistoryHTTPItem, 0, len(resp.GetItems()))

	for _, item := range resp.GetItems() {
		items = append(items, vkinoCoinsHistoryHTTPItem{
			ID:              item.GetId(),
			VKinoCoinsCount: item.GetVkinoCoinsCount(),
			OperationType:   item.GetOperationType(),
			Description:     item.GetDescription(),
			CreatedAt:       item.GetCreatedAt(),
		})
	}

	return vkinoCoinsHistoryHTTPResponse{
		Items:      items,
		TotalCount: resp.GetTotalCount(),
	}
}
