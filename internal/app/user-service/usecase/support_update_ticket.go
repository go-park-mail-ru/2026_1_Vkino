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

const maxSupportTicketRating = 5

func (u *supportUsecase) UpdateTicket(
	ctx context.Context,
	actorUserID int64,
	req domain.UpdateSupportTicketRequest,
) (domain.SupportTicketResponse, error) {
	req = normalizeTicketUpdateRequest(req)

	if err := validateTicketUpdateRequest(req); err != nil {
		return domain.SupportTicketResponse{}, err
	}

	role, err := u.userRepo.GetUserRole(ctx, actorUserID)
	if err != nil {
		return domain.SupportTicketResponse{}, domain.ErrInvalidToken
	}

	ticketBeforeUpdate, err := u.ticketForUpdate(ctx, req.TicketID)
	if err != nil {
		return domain.SupportTicketResponse{}, err
	}

	if permErr := applyTicketUpdatePermissions(role, actorUserID, ticketBeforeUpdate, &req); permErr != nil {
		return domain.SupportTicketResponse{}, permErr
	}

	ticket, err := u.updateTicketRecord(ctx, req)
	if err != nil {
		if errors.Is(err, postgresrepo.ErrTicketNotFound) {
			return domain.SupportTicketResponse{}, domain.ErrTicketNotFound
		}

		return domain.SupportTicketResponse{}, fmt.Errorf("%w: %w", domain.ErrInternal, err)
	}

	u.broker.publish(req.TicketID, domain.SupportTicketEventResponse{
		Type:   "ticket_updated",
		Ticket: ticket,
	})

	return *ticket, nil
}

func (u *supportUsecase) updateTicketRecord(
	ctx context.Context,
	req domain.UpdateSupportTicketRequest,
) (*domain.SupportTicketResponse, error) {
	if err := applyDerivedSupportLine(&req); err != nil {
		return nil, err
	}

	return u.supportRepo.UpdateTicket(ctx, req)
}

func normalizeTicketUpdateRequest(req domain.UpdateSupportTicketRequest) domain.UpdateSupportTicketRequest {
	req.Category = strings.TrimSpace(req.Category)
	req.Status = strings.TrimSpace(req.Status)
	req.Title = strings.TrimSpace(req.Title)
	req.UserEmail = strings.TrimSpace(req.UserEmail)
	req.Description = strings.TrimSpace(req.Description)
	req.AttachmentFileKey = strings.TrimSpace(req.AttachmentFileKey)

	return req
}

func applyDerivedSupportLine(req *domain.UpdateSupportTicketRequest) error {
	if req.Category == "" {
		return nil
	}

	derivedSupportLine := supportLineForCategory(req.Category)
	if req.SupportLine != 0 && req.SupportLine != derivedSupportLine {
		return domain.ErrInvalidTicketPayload
	}

	req.SupportLine = derivedSupportLine

	return nil
}

func validateTicketUpdateRequest(req domain.UpdateSupportTicketRequest) error {
	if req.TicketID <= 0 {
		return domain.ErrInvalidTicketID
	}

	if invalidTicketUpdatePayload(req) {
		return domain.ErrInvalidTicketPayload
	}

	if req.UserEmail != "" && !validator.ValidateEmail(req.UserEmail) {
		return domain.ErrInvalidEmail
	}

	return nil
}

func applyTicketUpdatePermissions(
	role string,
	actorUserID int64,
	ticket *domain.SupportTicketResponse,
	req *domain.UpdateSupportTicketRequest,
) error {
	switch {
	case role == roleUser:
		return applyUserTicketUpdatePermissions(actorUserID, ticket, req)
	case isStaff(role):
		return applyStaffTicketUpdatePermissions(role, ticket, req)
	default:
		return domain.ErrAccessDenied
	}
}

func applyUserTicketUpdatePermissions(
	actorUserID int64,
	ticket *domain.SupportTicketResponse,
	req *domain.UpdateSupportTicketRequest,
) error {
	if ticket.UserID != actorUserID {
		return domain.ErrAccessDenied
	}

	if req.Rating > 0 && !isTerminalTicketStatus(ticket.Status) {
		return domain.ErrInvalidTicketPayload
	}

	req.Category = ""
	req.Status = ""
	req.SupportLine = 0
	req.Title = ""
	req.Description = ""
	req.AttachmentFileKey = ""
	req.UserEmail = ""

	return nil
}

func applyStaffTicketUpdatePermissions(
	role string,
	ticket *domain.SupportTicketResponse,
	req *domain.UpdateSupportTicketRequest,
) error {
	req.Rating = 0

	if !canAccessCategory(role, ticket.Category) {
		return domain.ErrAccessDenied
	}

	if req.Category != "" && !canAccessCategory(role, req.Category) {
		return domain.ErrAccessDenied
	}

	if !isAdmin(role) {
		req.UserEmail = ""
	}

	return nil
}

func invalidTicketUpdatePayload(req domain.UpdateSupportTicketRequest) bool {
	return !isValidTicketCategory(req.Category) ||
		!isValidTicketStatus(req.Status) ||
		!isValidSupportLine(req.SupportLine) ||
		req.Rating < 0 ||
		req.Rating > maxSupportTicketRating
}

func (u *supportUsecase) ticketForUpdate(
	ctx context.Context,
	ticketID int64,
) (*domain.SupportTicketResponse, error) {
	ticket, err := u.supportRepo.GetTicketByID(ctx, ticketID)
	if err != nil {
		if errors.Is(err, postgresrepo.ErrTicketNotFound) {
			return nil, domain.ErrTicketNotFound
		}

		return nil, fmt.Errorf("%w: %w", domain.ErrInternal, err)
	}

	return ticket, nil
}
