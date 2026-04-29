package grpc

import (
	"context"
	"errors"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/domain"
	authv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/auth/v1"
	supportv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/support/v1"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/service/authctx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *SupportServer) authorize(ctx context.Context) (int64, error) {
	authCtx, err := authctx.ValidateIncomingContext(ctx, s.authClient)
	if err != nil {
		return 0, err
	}

	return authCtx.UserID, nil
}

func (s *SupportServer) authorizeOptional(ctx context.Context) (int64, error) {
	accessToken, err := authctx.AccessTokenFromIncomingContext(ctx)
	if err != nil {
		if errors.Is(err, authctx.ErrAuthorizationHeaderMissing) {
			return 0, nil
		}

		return 0, status.Error(codes.Unauthenticated, "unauthorized")
	}

	resp, err := s.authClient.Validate(ctx, &authv1.ValidateRequest{
		AccessToken: accessToken,
	})
	if err != nil {
		return 0, err
	}

	return resp.GetUserId(), nil
}

func (s *SupportServer) CreateTicket(
	ctx context.Context,
	req *supportv1.CreateTicketRequest,
) (*supportv1.TicketResponse, error) {
	userID, err := s.authorizeOptional(ctx)
	if err != nil {
		return nil, err
	}

	ticket, err := s.usecase.CreateTicket(ctx, userID, domain.CreateSupportTicketRequest{
		Category:          req.GetCategory(),
		Title:             req.GetTitle(),
		Description:       req.GetDescription(),
		UserEmail:         req.GetUserEmail(),
		AttachmentFileKey: req.GetAttachmentFileKey(),
	})
	if err != nil {
		return nil, mapSupportError(err)
	}

	return &supportv1.TicketResponse{Ticket: toProtoTicket(ticket)}, nil
}

func (s *SupportServer) GetTickets(
	ctx context.Context,
	req *supportv1.GetTicketsRequest,
) (*supportv1.TicketsResponse, error) {
	userID, err := s.authorize(ctx)
	if err != nil {
		return nil, err
	}

	tickets, err := s.usecase.GetTickets(ctx, userID, domain.GetSupportTicketsRequest{
		Status:      req.GetStatus(),
		Category:    req.GetCategory(),
		UserEmail:   req.GetUserEmail(),
		SupportLine: req.GetSupportLine(),
	})
	if err != nil {
		return nil, mapSupportError(err)
	}

	resp := &supportv1.TicketsResponse{
		Tickets: make([]*supportv1.Ticket, 0, len(tickets)),
	}

	for _, t := range tickets {
		resp.Tickets = append(resp.Tickets, toProtoTicket(t))
	}

	return resp, nil
}

func (s *SupportServer) UpdateTicket(
	ctx context.Context,
	req *supportv1.UpdateTicketRequest,
) (*supportv1.TicketResponse, error) {
	userID, err := s.authorize(ctx)
	if err != nil {
		return nil, err
	}

	ticket, err := s.usecase.UpdateTicket(ctx, userID, domain.UpdateSupportTicketRequest{
		TicketID:          req.GetTicketId(),
		Category:          req.GetCategory(),
		Status:            req.GetStatus(),
		SupportLine:       req.GetSupportLine(),
		Title:             req.GetTitle(),
		UserEmail:         req.GetUserEmail(),
		Description:       req.GetDescription(),
		AttachmentFileKey: req.GetAttachmentFileKey(),
		Rating:            req.GetRating(),
	})
	if err != nil {
		return nil, mapSupportError(err)
	}

	return &supportv1.TicketResponse{Ticket: toProtoTicket(ticket)}, nil
}

func (s *SupportServer) UploadSupportFile(
	ctx context.Context,
	req *supportv1.UploadSupportFileRequest,
) (*supportv1.UploadSupportFileResponse, error) {
	userID, err := s.authorizeOptional(ctx)
	if err != nil {
		return nil, err
	}

	file, err := s.usecase.UploadSupportFile(ctx, userID, domain.UploadSupportFileRequest{
		Content:     req.GetContent(),
		Filename:    req.GetFilename(),
		ContentType: req.GetContentType(),
		SizeBytes:   req.GetSizeBytes(),
	})
	if err != nil {
		return nil, mapSupportError(err)
	}

	return &supportv1.UploadSupportFileResponse{
		FileKey:     file.FileKey,
		FileUrl:     file.FileURL,
		ContentType: file.ContentType,
		SizeBytes:   file.SizeBytes,
	}, nil
}

func (s *SupportServer) GetSupportFileURL(
	ctx context.Context,
	req *supportv1.GetSupportFileURLRequest,
) (*supportv1.GetSupportFileURLResponse, error) {
	userID, err := s.authorize(ctx)
	if err != nil {
		return nil, err
	}

	file, err := s.usecase.GetSupportFileURL(ctx, userID, domain.GetSupportFileURLRequest{
		FileKey:  req.GetFileKey(),
		TicketID: req.GetTicketId(),
	})
	if err != nil {
		return nil, mapSupportError(err)
	}

	return &supportv1.GetSupportFileURLResponse{
		FileKey: file.FileKey,
		FileUrl: file.FileURL,
	}, nil
}

func (s *SupportServer) GetTicketMessages(
	ctx context.Context,
	req *supportv1.GetTicketMessagesRequest,
) (*supportv1.TicketMessagesResponse, error) {
	userID, err := s.authorize(ctx)
	if err != nil {
		return nil, err
	}

	messages, err := s.usecase.GetTicketMessages(ctx, userID, domain.GetSupportTicketMessagesRequest{
		TicketID: req.GetTicketId(),
	})
	if err != nil {
		return nil, mapSupportError(err)
	}

	resp := &supportv1.TicketMessagesResponse{
		Messages: make([]*supportv1.TicketMessage, 0, len(messages)),
	}

	for _, m := range messages {
		resp.Messages = append(resp.Messages, toProtoTicketMessage(m))
	}

	return resp, nil
}

func (s *SupportServer) CreateTicketMessage(
	ctx context.Context,
	req *supportv1.CreateTicketMessageRequest,
) (*supportv1.TicketMessageResponse, error) {
	userID, err := s.authorize(ctx)
	if err != nil {
		return nil, err
	}

	msg, err := s.usecase.CreateTicketMessage(ctx, userID, domain.CreateSupportTicketMessageRequest{
		TicketID:       req.GetTicketId(),
		Content:        req.GetContent(),
		ContentFileKey: req.GetContentFileKey(),
	})
	if err != nil {
		return nil, mapSupportError(err)
	}

	return &supportv1.TicketMessageResponse{Message: toProtoTicketMessage(msg)}, nil
}

func (s *SupportServer) GetTicketStatistics(
	ctx context.Context,
	_ *supportv1.GetTicketStatisticsRequest,
) (*supportv1.TicketStatisticsResponse, error) {
	userID, err := s.authorize(ctx)
	if err != nil {
		return nil, err
	}

	stats, err := s.usecase.GetTicketStatistics(ctx, userID, domain.GetSupportTicketStatisticsRequest{})
	if err != nil {
		return nil, mapSupportError(err)
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

func (s *SupportServer) SubscribeTicket(
	req *supportv1.SubscribeTicketRequest,
	stream grpc.ServerStreamingServer[supportv1.TicketEvent],
) error {
	userID, err := s.authorize(stream.Context())
	if err != nil {
		return err
	}

	events, unsubscribe, err := s.usecase.SubscribeTicket(
		stream.Context(),
		userID,
		domain.SubscribeSupportTicketRequest{TicketID: req.GetTicketId()},
	)
	if err != nil {
		return mapSupportError(err)
	}
	defer unsubscribe()

	for {
		select {
		case event, ok := <-events:
			if !ok {
				return nil
			}

			pbEvent := &supportv1.TicketEvent{Type: event.Type}

			if event.Ticket != nil {
				pbEvent.Ticket = toProtoTicket(*event.Ticket)
			}

			if event.Message != nil {
				pbEvent.Message = toProtoTicketMessage(*event.Message)
			}

			if err := stream.Send(pbEvent); err != nil {
				return err
			}

		case <-stream.Context().Done():
			return nil
		}
	}
}

func toProtoTicket(t domain.SupportTicketResponse) *supportv1.Ticket {
	return &supportv1.Ticket{
		Id:                t.ID,
		UserId:            t.UserID,
		UserEmail:         t.UserEmail,
		SenderEmail:       t.SenderEmail,
		Category:          t.Category,
		Status:            t.Status,
		SupportLine:       t.SupportLine,
		Title:             t.Title,
		Description:       t.Description,
		AttachmentFileKey: t.AttachmentFileKey,
		Rating:            t.Rating,
		CreatedAt:         t.CreatedAt,
		UpdatedAt:         t.UpdatedAt,
		ClosedAt:          t.ClosedAt,
	}
}

func toProtoTicketMessage(m domain.SupportTicketMessageResponse) *supportv1.TicketMessage {
	return &supportv1.TicketMessage{
		Id:             m.ID,
		TicketId:       m.TicketID,
		SenderId:       m.SenderID,
		Content:        m.Content,
		ContentFileKey: m.ContentFileKey,
		CreatedAt:      m.CreatedAt,
	}
}
