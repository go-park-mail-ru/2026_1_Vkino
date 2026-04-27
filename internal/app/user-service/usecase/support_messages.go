package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	domain2 "github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/domain"
	postgresrepo "github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/repository/postgres"
)

func (u *supportUsecase) GetTicketMessages(
	ctx context.Context,
	actorUserID int64,
	req domain2.GetSupportTicketMessagesRequest,
) ([]domain2.SupportTicketMessageResponse, error) {
	if req.TicketID <= 0 {
		return nil, domain2.ErrInvalidTicketID
	}

	if err := u.checkTicketAccess(ctx, actorUserID, req.TicketID); err != nil {
		return nil, err
	}

	messages, err := u.supportRepo.GetTicketMessages(ctx, req.TicketID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain2.ErrInternal, err)
	}

	return messages, nil
}

func (u *supportUsecase) CreateTicketMessage(
	ctx context.Context,
	actorUserID int64,
	req domain2.CreateSupportTicketMessageRequest,
) (domain2.SupportTicketMessageResponse, error) {
	req.Content = strings.TrimSpace(req.Content)
	req.ContentFileKey = strings.TrimSpace(req.ContentFileKey)

	if req.TicketID <= 0 {
		return domain2.SupportTicketMessageResponse{}, domain2.ErrInvalidTicketID
	}

	if req.Content == "" && req.ContentFileKey == "" {
		return domain2.SupportTicketMessageResponse{}, domain2.ErrInvalidMessage
	}

	if err := u.checkTicketAccess(ctx, actorUserID, req.TicketID); err != nil {
		return domain2.SupportTicketMessageResponse{}, err
	}

	msg, err := u.supportRepo.CreateTicketMessage(ctx, actorUserID, req)
	if err != nil {
		return domain2.SupportTicketMessageResponse{}, fmt.Errorf("%w: %v", domain2.ErrInternal, err)
	}

	u.broker.publish(req.TicketID, domain2.SupportTicketEventResponse{
		Type:    "message_created",
		Message: msg,
	})

	return *msg, nil
}

func (u *supportUsecase) checkTicketAccess(ctx context.Context, actorUserID, ticketID int64) error {
	ticket, err := u.supportRepo.GetTicketByID(ctx, ticketID)
	if err != nil {
		if errors.Is(err, postgresrepo.ErrTicketNotFound) {
			return domain2.ErrTicketNotFound
		}

		return fmt.Errorf("%w: %v", domain2.ErrInternal, err)
	}

	role, err := u.userRepo.GetUserRole(ctx, actorUserID)
	if err != nil {
		return domain2.ErrInvalidToken
	}

	switch {
	case role == "user":
		if ticket.UserID == actorUserID {
			return nil
		}

	case isAdmin(role):
		return nil

	case isStaff(role):
		if canAccessCategory(role, ticket.Category) {
			return nil
		}
	}

	if ticket.UserID != actorUserID {
		return domain2.ErrAccessDenied
	}

	return nil
}
