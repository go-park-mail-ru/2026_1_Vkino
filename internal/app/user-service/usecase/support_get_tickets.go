package usecase

import (
	"context"
	"fmt"
	"strings"

	domain "github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/domain"
	validator "github.com/go-park-mail-ru/2026_1_VKino/pkg/validatex"
)

func (u *supportUsecase) GetTickets(
	ctx context.Context,
	actorUserID int64,
	req domain.GetSupportTicketsRequest,
) ([]domain.SupportTicketResponse, error) {
	req.Status = strings.TrimSpace(req.Status)
	req.Category = strings.TrimSpace(req.Category)
	req.UserEmail = strings.TrimSpace(req.UserEmail)

	if err := validateTicketFilters(req); err != nil {
		return nil, err
	}

	role, err := u.userRepo.GetUserRole(ctx, actorUserID)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	userIDFilter := applyTicketRoleFilters(role, actorUserID, &req)

	if req.Category != "" && !canAccessCategory(role, req.Category) {
		return []domain.SupportTicketResponse{}, nil
	}

	tickets, err := u.supportRepo.GetTickets(ctx, userIDFilter, req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrInternal, err)
	}

	return tickets, nil
}

func validateTicketFilters(req domain.GetSupportTicketsRequest) error {
	if !isValidTicketStatus(req.Status) || !isValidTicketCategory(req.Category) || !isValidSupportLine(req.SupportLine) {
		return domain.ErrInvalidTicketPayload
	}

	if req.UserEmail != "" && !validator.ValidateEmail(req.UserEmail) {
		return domain.ErrInvalidEmail
	}

	return nil
}

func applyTicketRoleFilters(
	role string,
	actorUserID int64,
	req *domain.GetSupportTicketsRequest,
) int64 {
	if isStaff(role) {
		req.AllowedCategories = allowedCategoriesForRole(role)

		return 0
	}

	req.SupportLine = 0
	req.UserEmail = ""

	return actorUserID
}
