package domain

type VKinoCoinsHistoryItem struct {
	ID              int64
	VKinoCoinsCount int32
	OperationType   string
	Description     string
	CreatedAt       string
}

type VKinoCoinsHistoryResponse struct {
	Items      []VKinoCoinsHistoryItem
	TotalCount int32
}
