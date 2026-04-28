package usecase

import (
	"context"
	"errors"
	"fmt"

	domain "github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/domain"
	postgresrepo "github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/repository/postgres"
)

func (u *supportUsecase) GetTicketMessages(
	ctx context.Context,
	actorUserID int64,
	req domain.GetSupportTicketMessagesRequest,
) ([]domain.SupportTicketMessageResponse, error) {
	if req.TicketID <= 0 {
		return nil, domain.ErrInvalidTicketID
	}

	if err := u.checkTicketAccess(ctx, actorUserID, req.TicketID); err != nil {
		return nil, err
	}

	messages, err := u.supportRepo.GetTicketMessages(ctx, req.TicketID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInternal, err)
	}

	return messages, nil
}

func (u *supportUsecase) CreateTicketMessage(
	ctx context.Context,
	actorUserID int64,
	req domain.CreateSupportTicketMessageRequest,
) (domain.SupportTicketMessageResponse, error) {
	if req.TicketID <= 0 {
		return domain.SupportTicketMessageResponse{}, domain.ErrInvalidTicketID
	}

	if req.Content == "" && req.ContentFileKey == "" {
		return domain.SupportTicketMessageResponse{}, domain.ErrInvalidTicketID
	}

	if err := u.checkTicketAccess(ctx, actorUserID, req.TicketID); err != nil {
		return domain.SupportTicketMessageResponse{}, err
	}

	msg, err := u.supportRepo.CreateTicketMessage(ctx, actorUserID, req)
	if err != nil {
		return domain.SupportTicketMessageResponse{}, fmt.Errorf("%w: %v", domain.ErrInternal, err)
	}

	u.broker.publish(req.TicketID, domain.SupportTicketEventResponse{
		Type:    "message_created",
		Message: msg,
	})

	return *msg, nil
}

func (u *supportUsecase) checkTicketAccess(ctx context.Context, actorUserID, ticketID int64) error {
	role, err := u.userRepo.GetUserRole(ctx, actorUserID)
	if err != nil {
		return domain.ErrInvalidToken
	}

	if isStaff(role) {
		return nil
	}

	ticket, err := u.supportRepo.GetTicketByID(ctx, ticketID)
	if err != nil {
		if errors.Is(err, postgresrepo.ErrTicketNotFound) {
			return domain.ErrTicketNotFound
		}

		return fmt.Errorf("%w: %v", domain.ErrInternal, err)
	}

	if ticket.UserID != actorUserID {
		return domain.ErrAccessDenied
	}

	return nil
}
