package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	domain2 "github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/domain"
	postgresrepo "github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/repository/postgres"
	validator "github.com/go-park-mail-ru/2026_1_VKino/pkg/validatex"
)

func (u *supportUsecase) UpdateTicket(
	ctx context.Context,
	actorUserID int64,
	req domain2.UpdateSupportTicketRequest,
) (domain2.SupportTicketResponse, error) {
	req.Category = strings.TrimSpace(req.Category)
	req.Status = strings.TrimSpace(req.Status)
	req.Title = strings.TrimSpace(req.Title)
	req.UserEmail = strings.TrimSpace(req.UserEmail)
	req.Description = strings.TrimSpace(req.Description)

	if req.TicketID <= 0 {
		return domain2.SupportTicketResponse{}, domain2.ErrInvalidTicketID
	}

	if req.UserEmail != "" && !validator.ValidateEmail(req.UserEmail) {
		return domain2.SupportTicketResponse{}, domain2.ErrInvalidEmail
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
