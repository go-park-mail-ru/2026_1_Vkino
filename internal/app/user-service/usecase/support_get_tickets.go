package usecase

import (
	"context"
	"fmt"
	"strings"

	domain2 "github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/domain"
	validator "github.com/go-park-mail-ru/2026_1_VKino/pkg/validatex"
)

func (u *supportUsecase) GetTickets(
	ctx context.Context,
	actorUserID int64,
	req domain2.GetSupportTicketsRequest,
) ([]domain2.SupportTicketResponse, error) {
	req.Status = strings.TrimSpace(req.Status)
	req.Category = strings.TrimSpace(req.Category)
	req.UserEmail = strings.TrimSpace(req.UserEmail)

	if !isValidTicketStatus(req.Status) {
		return nil, domain2.ErrInvalidTicketPayload
	}

	if !isValidTicketCategory(req.Category) {
		return nil, domain2.ErrInvalidTicketPayload
	}

	if !isValidSupportLine(req.SupportLine) {
		return nil, domain2.ErrInvalidTicketPayload
	}

	if req.UserEmail != "" && !validator.ValidateEmail(req.UserEmail) {
		return nil, domain2.ErrInvalidEmail
	}

	role, err := u.userRepo.GetUserRole(ctx, actorUserID)
	if err != nil {
		return nil, domain2.ErrInvalidToken
	}

	userIDFilter := actorUserID
	if isStaff(role) {
		userIDFilter = 0
		req.AllowedCategories = allowedCategoriesForRole(role)
	} else {
		req.SupportLine = 0
		req.UserEmail = ""
	}

	if req.Category != "" && !canAccessCategory(role, req.Category) {
		return []domain2.SupportTicketResponse{}, nil
	}

	tickets, err := u.supportRepo.GetTickets(ctx, userIDFilter, req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain2.ErrInternal, err)
	}

	return tickets, nil
}
