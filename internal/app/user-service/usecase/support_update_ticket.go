package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	domain "github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/domain"
	postgresrepo "github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/repository/postgres"
	validator "github.com/go-park-mail-ru/2026_1_VKino/pkg/validatex"
)

func (u *supportUsecase) UpdateTicket(
	ctx context.Context,
	actorUserID int64,
	req domain.UpdateSupportTicketRequest,
) (domain.SupportTicketResponse, error) {
	req.Category = strings.TrimSpace(req.Category)
	req.Status = strings.TrimSpace(req.Status)
	req.Title = strings.TrimSpace(req.Title)
	req.UserEmail = strings.TrimSpace(req.UserEmail)
	req.Description = strings.TrimSpace(req.Description)

	if req.TicketID <= 0 {
		return domain.SupportTicketResponse{}, domain.ErrInvalidTicketID
	}

	if req.UserEmail != "" && !validator.ValidateEmail(req.UserEmail) {
		return domain.SupportTicketResponse{}, domain.ErrInvalidEmail
	}

	role, err := u.userRepo.GetUserRole(ctx, actorUserID)
	if err != nil {
		return domain.SupportTicketResponse{}, domain.ErrInvalidToken
	}

	if !isStaff(role) {
		return domain.SupportTicketResponse{}, domain.ErrAccessDenied
	}

	ticket, err := u.supportRepo.UpdateTicket(ctx, req)
	if err != nil {
		if errors.Is(err, postgresrepo.ErrTicketNotFound) {
			return domain.SupportTicketResponse{}, domain.ErrTicketNotFound
		}

		return domain.SupportTicketResponse{}, fmt.Errorf("%w: %v", domain.ErrInternal, err)
	}

	u.broker.publish(req.TicketID, domain.SupportTicketEventResponse{
		Type:   "ticket_updated",
		Ticket: ticket,
	})

	return *ticket, nil
}
