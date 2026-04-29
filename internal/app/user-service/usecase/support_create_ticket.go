package usecase

import (
	"context"
	"fmt"
	"strings"

	domain "github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/domain"
	validator "github.com/go-park-mail-ru/2026_1_VKino/pkg/validatex"
)

func (u *supportUsecase) CreateTicket(
	ctx context.Context,
	actorUserID int64,
	req domain.CreateSupportTicketRequest,
) (domain.SupportTicketResponse, error) {
	req.Category = strings.TrimSpace(req.Category)
	req.Title = strings.TrimSpace(req.Title)
	req.Description = strings.TrimSpace(req.Description)
	req.UserEmail = strings.TrimSpace(req.UserEmail)
	req.AttachmentFileKey = strings.TrimSpace(req.AttachmentFileKey)

	if req.Title == "" || req.Description == "" || req.Category == "" || !isValidTicketCategory(req.Category) {
		return domain.SupportTicketResponse{}, domain.ErrInvalidTicketPayload
	}

	if actorUserID > 0 {
		role, err := u.userRepo.GetUserRole(ctx, actorUserID)
		if err != nil {
			return domain.SupportTicketResponse{}, domain.ErrInvalidToken
		}

		if isStaff(role) {
			return domain.SupportTicketResponse{}, domain.ErrAccessDenied
		}

		req.UserEmail = ""
	} else if !validator.ValidateEmail(req.UserEmail) {
		return domain.SupportTicketResponse{}, domain.ErrInvalidEmail
	}

	req.SupportLine = supportLineForCategory(req.Category)

	ticket, err := u.supportRepo.CreateTicket(ctx, actorUserID, req)
	if err != nil {
		return domain.SupportTicketResponse{}, fmt.Errorf("%w: %v", domain.ErrInternal, err)
	}

	return *ticket, nil
}
