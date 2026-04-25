package usecase

import (
	"context"
	"errors"
	"fmt"

	domain2 "github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/domain"
	postgresrepo "github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/repository/postgres"
)

func (u *supportUsecase) UpdateTicket(
	ctx context.Context,
	actorUserID int64,
	req domain2.UpdateSupportTicketRequest,
) (domain2.SupportTicketResponse, error) {
	if req.TicketID <= 0 {
		return domain2.SupportTicketResponse{}, domain2.ErrInvalidTicketID
	}

	role, err := u.userRepo.GetUserRole(ctx, actorUserID)
	if err != nil {
		return domain2.SupportTicketResponse{}, domain2.ErrInvalidToken
	}

	if !isStaff(role) {
		return domain2.SupportTicketResponse{}, domain2.ErrAccessDenied
	}

	ticket, err := u.supportRepo.UpdateTicket(ctx, req)
	if err != nil {
		if errors.Is(err, postgresrepo.ErrTicketNotFound) {
			return domain2.SupportTicketResponse{}, domain2.ErrTicketNotFound
		}

		return domain2.SupportTicketResponse{}, fmt.Errorf("%w: %v", domain2.ErrInternal, err)
	}

	u.broker.publish(req.TicketID, domain2.SupportTicketEventResponse{
		Type:   "ticket_updated",
		Ticket: ticket,
	})

	return *ticket, nil
}
