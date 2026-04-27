package domain

type CreateSupportTicketRequest struct {
	Category          string
	Title             string
	Description       string
	UserEmail         string
	AttachmentFileKey string
	SupportLine       int64
}

type GetSupportTicketsRequest struct {
	Status            string
	Category          string
	UserEmail         string
	SupportLine       int64
	AllowedCategories []string
}

type UpdateSupportTicketRequest struct {
	TicketID          int64
	Category          string
	Status            string
	SupportLine       int64
	Title             string
	UserEmail         string
	Description       string
	AttachmentFileKey string
}

type GetSupportTicketMessagesRequest struct {
	TicketID int64
}

type CreateSupportTicketMessageRequest struct {
	TicketID       int64
	Content        string
	ContentFileKey string
}

type GetSupportTicketStatisticsRequest struct{}

type SubscribeSupportTicketRequest struct {
	TicketID int64
}

type SupportTicketResponse struct {
	ID                int64
	UserID            int64
	UserEmail         string
	Category          string
	Status            string
	SupportLine       int64
	Title             string
	Description       string
	AttachmentFileKey string
	Rating            int64
	CreatedAt         string
	UpdatedAt         string
	ClosedAt          string
}

type SupportTicketMessageResponse struct {
	ID             int64
	TicketID       int64
	SenderID       int64
	Content        string
	ContentFileKey string
	CreatedAt      string
}

type SupportTicketStatisticsResponse struct {
	Total         int64
	Open          int64
	InProgress    int64
	WaitingUser   int64
	Resolved      int64
	Closed        int64
	AverageRating float64
}

type SupportTicketEventResponse struct {
	Type    string
	Ticket  *SupportTicketResponse
	Message *SupportTicketMessageResponse
}
