package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	domain2 "github.com/go-park-mail-ru/2026_1_VKino/internal/app/user-service/domain"
	corepostgres "github.com/go-park-mail-ru/2026_1_VKino/pkg/postgresx"
	"github.com/jackc/pgx/v5"
)

var ErrTicketNotFound = errors.New("ticket not found")

type SupportRepo struct {
	db *corepostgres.Client
}

func NewSupportRepo(db *corepostgres.Client) *SupportRepo {
	return &SupportRepo{db: db}
}

func scanTicket(
	id, userID int64,
	category, status string,
	supportLine int64,
	title, description string,
	attachmentFileKey *string,
	rating *int64,
	createdAt, updatedAt time.Time,
	closedAt *time.Time,
) domain2.SupportTicketResponse {
	ticket := domain2.SupportTicketResponse{
		ID:          id,
		UserID:      userID,
		Category:    category,
		Status:      status,
		SupportLine: supportLine,
		Title:       title,
		Description: description,
		CreatedAt:   createdAt.Format(time.RFC3339),
		UpdatedAt:   updatedAt.Format(time.RFC3339),
	}

	if attachmentFileKey != nil {
		ticket.AttachmentFileKey = *attachmentFileKey
	}

	if rating != nil {
		ticket.Rating = *rating
	}

	if closedAt != nil {
		ticket.ClosedAt = closedAt.Format(time.RFC3339)
	}

	return ticket
}

func (r *SupportRepo) scanTicketRow(row pgx.Row) (*domain2.SupportTicketResponse, error) {
	var (
		id                int64
		userID            int64
		category          string
		status            string
		supportLine       int64
		title             string
		description       string
		attachmentFileKey *string
		rating            *int64
		createdAt         time.Time
		updatedAt         time.Time
		closedAt          *time.Time
	)

	err := row.Scan(
		&id, &userID, &category, &status, &supportLine,
		&title, &description, &attachmentFileKey, &rating,
		&createdAt, &updatedAt, &closedAt,
	)
	if err != nil {
		return nil, err
	}

	ticket := scanTicket(id, userID, category, status, supportLine, title, description,
		attachmentFileKey, rating, createdAt, updatedAt, closedAt)

	return &ticket, nil
}

func (r *SupportRepo) CreateTicket(
	ctx context.Context,
	userID int64,
	req domain2.CreateSupportTicketRequest,
) (*domain2.SupportTicketResponse, error) {
	row := r.db.QueryRow(ctx, sqlCreateSupportTicket,
		userID, req.Category, req.Title, req.Description, req.AttachmentFileKey,
	)

	ticket, err := r.scanTicketRow(row)
	if err != nil {
		return nil, fmt.Errorf("create support ticket: %w", err)
	}

	return ticket, nil
}

func (r *SupportRepo) GetTicketByID(ctx context.Context, ticketID int64) (*domain2.SupportTicketResponse, error) {
	row := r.db.QueryRow(ctx, sqlGetSupportTicketByID, ticketID)

	ticket, err := r.scanTicketRow(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTicketNotFound
		}

		return nil, fmt.Errorf("get support ticket by id: %w", err)
	}

	return ticket, nil
}

func (r *SupportRepo) GetTickets(
	ctx context.Context,
	userID int64,
	req domain2.GetSupportTicketsRequest,
) ([]domain2.SupportTicketResponse, error) {
	rows, err := r.db.Query(ctx, sqlGetSupportTickets, userID, req.Status, req.Category, req.SupportLine)
	if err != nil {
		return nil, fmt.Errorf("get support tickets: %w", err)
	}
	defer rows.Close()

	tickets := make([]domain2.SupportTicketResponse, 0)

	for rows.Next() {
		var (
			id                int64
			uID               int64
			category          string
			status            string
			supportLine       int64
			title             string
			description       string
			attachmentFileKey *string
			rating            *int64
			createdAt         time.Time
			updatedAt         time.Time
			closedAt          *time.Time
		)

		if err = rows.Scan(
			&id, &uID, &category, &status, &supportLine,
			&title, &description, &attachmentFileKey, &rating,
			&createdAt, &updatedAt, &closedAt,
		); err != nil {
			return nil, fmt.Errorf("scan support ticket: %w", err)
		}

		tickets = append(tickets, scanTicket(id, uID, category, status, supportLine, title, description,
			attachmentFileKey, rating, createdAt, updatedAt, closedAt))
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate support tickets: %w", err)
	}

	return tickets, nil
}

func (r *SupportRepo) UpdateTicket(
	ctx context.Context,
	req domain2.UpdateSupportTicketRequest,
) (*domain2.SupportTicketResponse, error) {
	row := r.db.QueryRow(ctx, sqlUpdateSupportTicket,
		req.TicketID, req.Category, req.Status, req.SupportLine,
		req.Title, req.Description, req.AttachmentFileKey,
	)

	ticket, err := r.scanTicketRow(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTicketNotFound
		}

		return nil, fmt.Errorf("update support ticket: %w", err)
	}

	return ticket, nil
}

func (r *SupportRepo) GetTicketMessages(
	ctx context.Context,
	ticketID int64,
) ([]domain2.SupportTicketMessageResponse, error) {
	rows, err := r.db.Query(ctx, sqlGetSupportTicketMessages, ticketID)
	if err != nil {
		return nil, fmt.Errorf("get support ticket messages: %w", err)
	}
	defer rows.Close()

	messages := make([]domain2.SupportTicketMessageResponse, 0)

	for rows.Next() {
		var (
			id             int64
			tID            int64
			senderID       int64
			content        *string
			contentFileKey *string
			createdAt      time.Time
		)

		if err = rows.Scan(&id, &tID, &senderID, &content, &contentFileKey, &createdAt); err != nil {
			return nil, fmt.Errorf("scan support ticket message: %w", err)
		}

		msg := domain2.SupportTicketMessageResponse{
			ID:        id,
			TicketID:  tID,
			SenderID:  senderID,
			CreatedAt: createdAt.Format(time.RFC3339),
		}

		if content != nil {
			msg.Content = *content
		}

		if contentFileKey != nil {
			msg.ContentFileKey = *contentFileKey
		}

		messages = append(messages, msg)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate support ticket messages: %w", err)
	}

	return messages, nil
}

func (r *SupportRepo) CreateTicketMessage(
	ctx context.Context,
	senderID int64,
	req domain2.CreateSupportTicketMessageRequest,
) (*domain2.SupportTicketMessageResponse, error) {
	var contentPtr *string
	if req.Content != "" {
		contentPtr = &req.Content
	}

	var contentFileKeyPtr *string
	if req.ContentFileKey != "" {
		contentFileKeyPtr = &req.ContentFileKey
	}

	var (
		id             int64
		tID            int64
		sndrID         int64
		content        *string
		contentFileKey *string
		createdAt      time.Time
	)

	err := r.db.QueryRow(ctx, sqlCreateSupportTicketMessage,
		req.TicketID, senderID, contentPtr, contentFileKeyPtr,
	).Scan(&id, &tID, &sndrID, &content, &contentFileKey, &createdAt)
	if err != nil {
		return nil, fmt.Errorf("create support ticket message: %w", err)
	}

	msg := &domain2.SupportTicketMessageResponse{
		ID:        id,
		TicketID:  tID,
		SenderID:  sndrID,
		CreatedAt: createdAt.Format(time.RFC3339),
	}

	if content != nil {
		msg.Content = *content
	}

	if contentFileKey != nil {
		msg.ContentFileKey = *contentFileKey
	}

	return msg, nil
}

func (r *SupportRepo) GetTicketStatistics(ctx context.Context) (*domain2.SupportTicketStatisticsResponse, error) {
	var stats domain2.SupportTicketStatisticsResponse

	err := r.db.QueryRow(ctx, sqlGetSupportStatistics).Scan(
		&stats.Total,
		&stats.Open,
		&stats.InProgress,
		&stats.WaitingUser,
		&stats.Resolved,
		&stats.Closed,
		&stats.AverageRating,
	)
	if err != nil {
		return nil, fmt.Errorf("get support ticket statistics: %w", err)
	}

	return &stats, nil
}
