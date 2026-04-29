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
	req.AttachmentFileKey = strings.TrimSpace(req.AttachmentFileKey)

	if req.TicketID <= 0 {
		return domain.SupportTicketResponse{}, domain.ErrInvalidTicketID
	}

	if !isValidTicketCategory(req.Category) {
		return domain.SupportTicketResponse{}, domain.ErrInvalidTicketPayload
	}

	if !isValidTicketStatus(req.Status) {
		return domain.SupportTicketResponse{}, domain.ErrInvalidTicketPayload
	}

	if !isValidSupportLine(req.SupportLine) {
		return domain.SupportTicketResponse{}, domain.ErrInvalidTicketPayload
	}

	if req.UserEmail != "" && !validator.ValidateEmail(req.UserEmail) {
		return domain.SupportTicketResponse{}, domain.ErrInvalidEmail
	}

	if (req.Rating < 0) || (req.Rating > 5) {
		return domain.SupportTicketResponse{}, domain.ErrInvalidTicketPayload
	}

	role, err := u.userRepo.GetUserRole(ctx, actorUserID)
	if err != nil {
		return domain.SupportTicketResponse{}, domain.ErrInvalidToken
	}

	ticketBeforeUpdate, err := u.supportRepo.GetTicketByID(ctx, req.TicketID)
	if err != nil {
		if errors.Is(err, postgresrepo.ErrTicketNotFound) {
			return domain.SupportTicketResponse{}, domain.ErrTicketNotFound
		}

		return domain.SupportTicketResponse{}, fmt.Errorf("%w: %v", domain.ErrInternal, err)
	}

	switch {
	case role == "user":
		if ticketBeforeUpdate.UserID != actorUserID {
			return domain.SupportTicketResponse{}, domain.ErrAccessDenied
		}

		if req.Rating > 0 && !isTerminalTicketStatus(ticketBeforeUpdate.Status) {
			return domain.SupportTicketResponse{}, domain.ErrInvalidTicketPayload
		}

		req.Category = ""
		req.Status = ""
		req.SupportLine = 0
		req.Title = ""
		req.Description = ""
		req.AttachmentFileKey = ""
		req.UserEmail = ""

	case isStaff(role):
		req.Rating = 0

		if !canAccessCategory(role, ticketBeforeUpdate.Category) {
			return domain.SupportTicketResponse{}, domain.ErrAccessDenied
		}

		if req.Category != "" && !canAccessCategory(role, req.Category) {
			return domain.SupportTicketResponse{}, domain.ErrAccessDenied
		}

		if !isAdmin(role) {
			req.UserEmail = ""
		}

	default:
		return domain.SupportTicketResponse{}, domain.ErrAccessDenied
	}

	if req.Category != "" {
		derivedSupportLine := supportLineForCategory(req.Category)
		if req.SupportLine != 0 && req.SupportLine != derivedSupportLine {
			return domain.SupportTicketResponse{}, domain.ErrInvalidTicketPayload
		}

		req.SupportLine = derivedSupportLine
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
