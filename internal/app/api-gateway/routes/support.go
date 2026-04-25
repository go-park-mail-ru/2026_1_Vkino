package routes

import (
	"net/http"
	"strings"

	supportv1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/support/v1"
	httppkg "github.com/go-park-mail-ru/2026_1_VKino/pkg/http"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/httpserver"
)

func Support(cfg Config, supportClient supportv1.SupportServiceClient) []httpserver.Option {
	return []httpserver.Option{
		httpserver.WithRoute("POST /support/tickets", func(w http.ResponseWriter, r *http.Request) {
			var req struct {
				Category          string `json:"category"`
				Title             string `json:"title"`
				Description       string `json:"description"`
				AttachmentFileKey string `json:"attachment_file_key"`
			}
			if !readJSON(w, r, &req) {
				return
			}

			cancel := grpcContext(r, cfg.UserRequestTimeout())
			defer cancel()

			resp, err := supportClient.CreateTicket(r.Context(), &supportv1.CreateTicketRequest{
				Category:          req.Category,
				Title:             req.Title,
				Description:       req.Description,
				AttachmentFileKey: req.AttachmentFileKey,
			})
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			httppkg.Response(w, http.StatusCreated, resp)
		}),

		httpserver.WithRoute("GET /support/tickets", func(w http.ResponseWriter, r *http.Request) {
			var req struct {
				Role        string `json:"role"`
				Status      string `json:"status"`
				Category    string `json:"category"`
				SupportLine int64  `json:"support_line"`
			}
			if !readJSON(w, r, &req) {
				return
			}

			role := strings.TrimSpace(req.Role)
			if role == "" {
				httppkg.ErrResponse(w, http.StatusBadRequest, "invalid role")

				return
			}

			cancel := grpcContext(r, cfg.UserRequestTimeout())
			defer cancel()

			request := &supportv1.GetTicketsRequest{
				Status:      strings.TrimSpace(req.Status),
				Category:    strings.TrimSpace(req.Category),
				SupportLine: req.SupportLine,
			}

			switch role {
			case "user", "support_l1", "support_l2", "admin":
			default:
				httppkg.ErrResponse(w, http.StatusBadRequest, "invalid role")

				return
			}

			resp, err := supportClient.GetTickets(r.Context(), request)
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}),

		httpserver.WithRoute("PATCH /support/tickets/{id}", func(w http.ResponseWriter, r *http.Request) {
			ticketID, ok := parsePathID(w, r, "invalid ticket id")
			if !ok {
				return
			}

			var req struct {
				Category          string `json:"category"`
				Status            string `json:"status"`
				SupportLine       int64  `json:"support_line"`
				Title             string `json:"title"`
				Description       string `json:"description"`
				AttachmentFileKey string `json:"attachment_file_key"`
			}
			if !readJSON(w, r, &req) {
				return
			}

			cancel := grpcContext(r, cfg.UserRequestTimeout())
			defer cancel()

			resp, err := supportClient.UpdateTicket(r.Context(), &supportv1.UpdateTicketRequest{
				TicketId:          ticketID,
				Category:          req.Category,
				Status:            req.Status,
				SupportLine:       req.SupportLine,
				Title:             req.Title,
				Description:       req.Description,
				AttachmentFileKey: req.AttachmentFileKey,
			})
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}),

		httpserver.WithRoute("GET /support/tickets/{id}/messages", func(w http.ResponseWriter, r *http.Request) {
			ticketID, ok := parsePathID(w, r, "invalid ticket id")
			if !ok {
				return
			}

			cancel := grpcContext(r, cfg.UserRequestTimeout())
			defer cancel()

			resp, err := supportClient.GetTicketMessages(r.Context(), &supportv1.GetTicketMessagesRequest{
				TicketId: ticketID,
			})
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}),

		httpserver.WithRoute("POST /support/tickets/{id}/messages", func(w http.ResponseWriter, r *http.Request) {
			ticketID, ok := parsePathID(w, r, "invalid ticket id")
			if !ok {
				return
			}

			var req struct {
				Content        string `json:"content"`
				ContentFileKey string `json:"content_file_key"`
			}
			if !readJSON(w, r, &req) {
				return
			}

			cancel := grpcContext(r, cfg.UserRequestTimeout())
			defer cancel()

			resp, err := supportClient.CreateTicketMessage(r.Context(), &supportv1.CreateTicketMessageRequest{
				TicketId:       ticketID,
				Content:        req.Content,
				ContentFileKey: req.ContentFileKey,
			})
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			httppkg.Response(w, http.StatusCreated, resp)
		}),

		httpserver.WithRoute("GET /support/statistics", func(w http.ResponseWriter, r *http.Request) {
			cancel := grpcContext(r, cfg.UserRequestTimeout())
			defer cancel()

			resp, err := supportClient.GetTicketStatistics(r.Context(), &supportv1.GetTicketStatisticsRequest{})
			if err != nil {
				writeGRPCError(w, err)

				return
			}

			httppkg.Response(w, http.StatusOK, resp)
		}),
	}
}
