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

	role, err := u.userRepo.GetUserRole(ctx, actorUserID)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	userIDFilter := actorUserID
	if isStaff(role) {
		userIDFilter = 0

		if req.UserEmail != "" && !validator.ValidateEmail(req.UserEmail) {
			return nil, domain.ErrInvalidEmail
		}
	} else {
		req.SupportLine = 0
		req.UserEmail = ""
	}

	tickets, err := u.supportRepo.GetTickets(ctx, userIDFilter, req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInternal, err)
	}

	return tickets, nil
}
