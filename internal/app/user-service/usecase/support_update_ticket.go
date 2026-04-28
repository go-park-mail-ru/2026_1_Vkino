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
	req.AttachmentFileKey = strings.TrimSpace(req.AttachmentFileKey)

	if req.TicketID <= 0 {
		return domain2.SupportTicketResponse{}, domain2.ErrInvalidTicketID
	}

	if !isValidTicketCategory(req.Category) {
		return domain2.SupportTicketResponse{}, domain2.ErrInvalidTicketPayload
	}

	if !isValidTicketStatus(req.Status) {
		return domain2.SupportTicketResponse{}, domain2.ErrInvalidTicketPayload
	}

	if !isValidSupportLine(req.SupportLine) {
		return domain2.SupportTicketResponse{}, domain2.ErrInvalidTicketPayload
	}

	if req.UserEmail != "" && !validator.ValidateEmail(req.UserEmail) {
		return domain2.SupportTicketResponse{}, domain2.ErrInvalidEmail
	}

	if (req.Rating < 0) || (req.Rating > 5) {
		return domain2.SupportTicketResponse{}, domain2.ErrInvalidTicketPayload
	}

	role, err := u.userRepo.GetUserRole(ctx, actorUserID)
	if err != nil {
		return domain2.SupportTicketResponse{}, domain2.ErrInvalidToken
	}

	ticketBeforeUpdate, err := u.supportRepo.GetTicketByID(ctx, req.TicketID)
	if err != nil {
		if errors.Is(err, postgresrepo.ErrTicketNotFound) {
			return domain2.SupportTicketResponse{}, domain2.ErrTicketNotFound
		}

		return domain2.SupportTicketResponse{}, fmt.Errorf("%w: %v", domain2.ErrInternal, err)
	}

	switch {
	case role == "user":
		if ticketBeforeUpdate.UserID != actorUserID {
			return domain2.SupportTicketResponse{}, domain2.ErrAccessDenied
		}

		req.Status = ""
		req.SupportLine = 0
		req.UserEmail = ""

	case isStaff(role):
		req.Rating = 0

		if !canAccessCategory(role, ticketBeforeUpdate.Category) {
			return domain2.SupportTicketResponse{}, domain2.ErrAccessDenied
		}

		if req.Category != "" && !canAccessCategory(role, req.Category) {
			return domain2.SupportTicketResponse{}, domain2.ErrAccessDenied
		}

		if !isAdmin(role) {
			req.UserEmail = ""
		}

	default:
		return domain2.SupportTicketResponse{}, domain2.ErrAccessDenied
	}

	if req.Category != "" {
		derivedSupportLine := supportLineForCategory(req.Category)
		if req.SupportLine != 0 && req.SupportLine != derivedSupportLine {
			return domain2.SupportTicketResponse{}, domain2.ErrInvalidTicketPayload
		}

		req.SupportLine = derivedSupportLine
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
