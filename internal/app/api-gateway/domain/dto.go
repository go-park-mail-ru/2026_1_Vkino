package domain

import "encoding/json"

import moviev1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/movie/v1"

type SignInRequest struct {
	Email      string `json:"email"`
	passphrase string
}

type SignUpRequest struct {
	Email      string `json:"email"`
	passphrase string
}

func (r *SignInRequest) Password() string {
	return r.passphrase
}

func (r *SignInRequest) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if value, ok := raw["email"]; ok {
		if err := json.Unmarshal(value, &r.Email); err != nil {
			return err
		}
	}

	if value, ok := raw["password"]; ok {
		if err := json.Unmarshal(value, &r.passphrase); err != nil {
			return err
		}
	}

	return nil
}

func (r *SignUpRequest) Password() string {
	return r.passphrase
}

func (r *SignUpRequest) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if value, ok := raw["email"]; ok {
		if err := json.Unmarshal(value, &r.Email); err != nil {
			return err
		}
	}

	if value, ok := raw["password"]; ok {
		if err := json.Unmarshal(value, &r.passphrase); err != nil {
			return err
		}
	}

	return nil
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type SupportCreateTicketRequest struct {
	Category          string `json:"category"`
	Title             string `json:"title"`
	Description       string `json:"description"`
	UserEmail         string `json:"user_email"`
	AttachmentFileKey string `json:"attachment_file_key"`
}

type SupportGetTicketsRequest struct {
	Status      string `json:"status"`
	Category    string `json:"category"`
	UserEmail   string `json:"user_email"`
	SupportLine int64  `json:"support_line"`
}

type SupportUpdateTicketRequest struct {
	Category          string `json:"category"`
	Status            string `json:"status"`
	SupportLine       int64  `json:"support_line"`
	Title             string `json:"title"`
	UserEmail         string `json:"user_email"`
	Description       string `json:"description"`
	AttachmentFileKey string `json:"attachment_file_key"`
	Rating            int64  `json:"rating"`
}

type SupportCreateTicketMessageRequest struct {
	Content        string `json:"content"`
	ContentFileKey string `json:"content_file_key"`
}

type FavoritesHTTPResponse struct {
	MovieIDs   []int64              `json:"movie_ids"`
	TotalCount int32                `json:"total_count"`
	Movies     []*moviev1.MovieCard `json:"movies"`
}
