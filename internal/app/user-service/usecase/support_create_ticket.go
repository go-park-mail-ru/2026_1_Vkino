package usecase

import (
	"context"
	"fmt"
	"strings"

	domain2 "github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/domain"
	validator "github.com/go-park-mail-ru/2026_1_VKino/pkg/validatex"
)

func (u *supportUsecase) CreateTicket(
	ctx context.Context,
	actorUserID int64,
	req domain2.CreateSupportTicketRequest,
) (domain2.SupportTicketResponse, error) {
	req.Category = strings.TrimSpace(req.Category)
	req.Title = strings.TrimSpace(req.Title)
	req.Description = strings.TrimSpace(req.Description)
	req.UserEmail = strings.TrimSpace(req.UserEmail)
	req.AttachmentFileKey = strings.TrimSpace(req.AttachmentFileKey)

	if req.Title == "" || req.Description == "" || !isValidTicketCategory(req.Category) {
		return domain2.SupportTicketResponse{}, domain2.ErrInvalidTicket
	}

	if actorUserID > 0 {
		req.UserEmail = ""
	} else if !validator.ValidateEmail(req.UserEmail) {
		return domain2.SupportTicketResponse{}, domain2.ErrInvalidEmail
	}

	req.SupportLine = supportLineForCategory(req.Category)

	ticket, err := u.supportRepo.CreateTicket(ctx, actorUserID, req)
	if err != nil {
		return domain2.SupportTicketResponse{}, fmt.Errorf("%w: %v", domain2.ErrInternal, err)
	}

	return *ticket, nil
}
