package grpc

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/domain"
	supportv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/support/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const supportNotImplementedMessage = "support usecase not implemented"

func (s *Server) CreateTicket(
	ctx context.Context,
	req *supportv1.CreateTicketRequest,
) (*supportv1.TicketResponse, error) {
	authCtx, err := s.authorize(ctx)
	if err != nil {
		return nil, err
	}

	if s.supportUsecase == nil {
		return nil, status.Error(codes.Unimplemented, supportNotImplementedMessage)
	}

	ticket, err := s.supportUsecase.CreateTicket(ctx, authCtx.UserID, domain.CreateSupportTicketRequest{
		Category:          req.GetCategory(),
		Title:             req.GetTitle(),
		Description:       req.GetDescription(),
		AttachmentFileKey: req.GetAttachmentFileKey(),
	})
	if err != nil {
		return nil, mapError(err)
	}

	return &supportv1.TicketResponse{Ticket: mapSupportTicket(ticket)}, nil
}

func (s *Server) GetTickets(ctx context.Context, req *supportv1.GetTicketsRequest) (*supportv1.TicketsResponse, error) {
	authCtx, err := s.authorize(ctx)
	if err != nil {
		return nil, err
	}

	if s.supportUsecase == nil {
		return nil, status.Error(codes.Unimplemented, supportNotImplementedMessage)
	}

	tickets, err := s.supportUsecase.GetTickets(ctx, authCtx.UserID, domain.GetSupportTicketsRequest{
		Status:      req.GetStatus(),
		Category:    req.GetCategory(),
		SupportLine: req.GetSupportLine(),
	})
	if err != nil {
		return nil, mapError(err)
	}

	return &supportv1.TicketsResponse{Tickets: mapSupportTickets(tickets)}, nil
}

func (s *Server) UpdateTicket(
	ctx context.Context,
	req *supportv1.UpdateTicketRequest,
) (*supportv1.TicketResponse, error) {
	authCtx, err := s.authorize(ctx)
	if err != nil {
		return nil, err
	}

	if s.supportUsecase == nil {
		return nil, status.Error(codes.Unimplemented, supportNotImplementedMessage)
	}

	ticket, err := s.supportUsecase.UpdateTicket(ctx, authCtx.UserID, domain.UpdateSupportTicketRequest{
		TicketID:          req.GetTicketId(),
		Category:          req.GetCategory(),
		Status:            req.GetStatus(),
		SupportLine:       req.GetSupportLine(),
		Title:             req.GetTitle(),
		Description:       req.GetDescription(),
		AttachmentFileKey: req.GetAttachmentFileKey(),
	})
	if err != nil {
		return nil, mapError(err)
	}

	return &supportv1.TicketResponse{Ticket: mapSupportTicket(ticket)}, nil
}

func (s *Server) GetTicketMessages(
	ctx context.Context,
	req *supportv1.GetTicketMessagesRequest,
) (*supportv1.TicketMessagesResponse, error) {
	authCtx, err := s.authorize(ctx)
	if err != nil {
		return nil, err
	}

	if s.supportUsecase == nil {
		return nil, status.Error(codes.Unimplemented, supportNotImplementedMessage)
	}

	messages, err := s.supportUsecase.GetTicketMessages(ctx, authCtx.UserID, domain.GetSupportTicketMessagesRequest{
		TicketID: req.GetTicketId(),
	})
	if err != nil {
		return nil, mapError(err)
	}

	return &supportv1.TicketMessagesResponse{Messages: mapSupportMessages(messages)}, nil
}

func (s *Server) CreateTicketMessage(
	ctx context.Context,
	req *supportv1.CreateTicketMessageRequest,
) (*supportv1.TicketMessageResponse, error) {
	authCtx, err := s.authorize(ctx)
	if err != nil {
		return nil, err
	}

	if s.supportUsecase == nil {
		return nil, status.Error(codes.Unimplemented, supportNotImplementedMessage)
	}

	message, err := s.supportUsecase.CreateTicketMessage(ctx, authCtx.UserID, domain.CreateSupportTicketMessageRequest{
		TicketID:       req.GetTicketId(),
		Content:        req.GetContent(),
		ContentFileKey: req.GetContentFileKey(),
	})
	if err != nil {
		return nil, mapError(err)
	}

	return &supportv1.TicketMessageResponse{Message: mapSupportMessage(message)}, nil
}

func (s *Server) GetTicketStatistics(
	ctx context.Context,
	_ *supportv1.GetTicketStatisticsRequest,
) (*supportv1.TicketStatisticsResponse, error) {
	authCtx, err := s.authorize(ctx)
	if err != nil {
		return nil, err
	}

	if s.supportUsecase == nil {
		return nil, status.Error(codes.Unimplemented, supportNotImplementedMessage)
	}

	stats, err := s.supportUsecase.GetTicketStatistics(ctx, authCtx.UserID, domain.GetSupportTicketStatisticsRequest{})
	if err != nil {
		return nil, mapError(err)
	}

	return &supportv1.TicketStatisticsResponse{
		Total:         stats.Total,
		Open:          stats.Open,
		InProgress:    stats.InProgress,
		WaitingUser:   stats.WaitingUser,
		Resolved:      stats.Resolved,
		Closed:        stats.Closed,
		AverageRating: stats.AverageRating,
	}, nil
}

func (s *Server) SubscribeTicket(
	req *supportv1.SubscribeTicketRequest,
	stream supportv1.SupportService_SubscribeTicketServer,
) error {
	authCtx, err := s.authorize(stream.Context())
	if err != nil {
		return err
	}

	if s.supportUsecase == nil {
		return status.Error(codes.Unimplemented, supportNotImplementedMessage)
	}

	eventCh, unsubscribe, err := s.supportUsecase.SubscribeTicket(
		stream.Context(),
		authCtx.UserID,
		domain.SubscribeSupportTicketRequest{
			TicketID: req.GetTicketId(),
		},
	)
	if err != nil {
		return mapError(err)
	}
	defer unsubscribe()

	for {
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()
		case event, ok := <-eventCh:
			if !ok {
				return nil
			}

			if err := stream.Send(mapSupportEvent(event)); err != nil {
				return err
			}
		}
	}
}

func mapSupportTicket(ticket domain.SupportTicketResponse) *supportv1.Ticket {
	return &supportv1.Ticket{
		Id:                ticket.ID,
		UserId:            ticket.UserID,
		Category:          ticket.Category,
		Status:            ticket.Status,
		SupportLine:       ticket.SupportLine,
		Title:             ticket.Title,
		Description:       ticket.Description,
		AttachmentFileKey: ticket.AttachmentFileKey,
		Rating:            ticket.Rating,
		CreatedAt:         ticket.CreatedAt,
		UpdatedAt:         ticket.UpdatedAt,
		ClosedAt:          ticket.ClosedAt,
	}
}

func mapSupportTickets(tickets []domain.SupportTicketResponse) []*supportv1.Ticket {
	if len(tickets) == 0 {
		return []*supportv1.Ticket{}
	}

	result := make([]*supportv1.Ticket, 0, len(tickets))
	for _, ticket := range tickets {
		result = append(result, mapSupportTicket(ticket))
	}

	return result
}

func mapSupportMessage(message domain.SupportTicketMessageResponse) *supportv1.TicketMessage {
	return &supportv1.TicketMessage{
		Id:             message.ID,
		TicketId:       message.TicketID,
		SenderId:       message.SenderID,
		Content:        message.Content,
		ContentFileKey: message.ContentFileKey,
		CreatedAt:      message.CreatedAt,
	}
}

func mapSupportMessages(messages []domain.SupportTicketMessageResponse) []*supportv1.TicketMessage {
	if len(messages) == 0 {
		return []*supportv1.TicketMessage{}
	}

	result := make([]*supportv1.TicketMessage, 0, len(messages))
	for _, message := range messages {
		result = append(result, mapSupportMessage(message))
	}

	return result
}

func mapSupportEvent(event domain.SupportTicketEventResponse) *supportv1.TicketEvent {
	var ticket *supportv1.Ticket
	if event.Ticket != nil {
		ticket = mapSupportTicket(*event.Ticket)
	}

	var message *supportv1.TicketMessage
	if event.Message != nil {
		message = mapSupportMessage(*event.Message)
	}

	return &supportv1.TicketEvent{
		Type:    event.Type,
		Ticket:  ticket,
		Message: message,
	}
}
