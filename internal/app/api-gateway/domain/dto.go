package domain

import (
	moviev1 "github.com/go-park-mail-ru/2026_1_VKino/pkg/gen/movie/v1"
	"github.com/mailru/easyjson/jlexer"
)

//go:generate go run -mod=mod github.com/mailru/easyjson/easyjson -disallow_unknown_fields dto.go

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

func (r *SignInRequest) UnmarshalEasyJSON(l *jlexer.Lexer) {
	decodeCredentials(l, &r.Email, &r.passphrase)
}

func (r *SignInRequest) UnmarshalJSON(data []byte) error {
	lexer := jlexer.Lexer{Data: data}
	r.UnmarshalEasyJSON(&lexer)
	lexer.Consumed()

	return lexer.Error()
}

func (r *SignUpRequest) Password() string {
	return r.passphrase
}

func (r *SignUpRequest) UnmarshalEasyJSON(l *jlexer.Lexer) {
	decodeCredentials(l, &r.Email, &r.passphrase)
}

func (r *SignUpRequest) UnmarshalJSON(data []byte) error {
	lexer := jlexer.Lexer{Data: data}
	r.UnmarshalEasyJSON(&lexer)
	lexer.Consumed()

	return lexer.Error()
}

//easyjson:json
//nolint:recvcheck // easyjson generates Marshal* on value receiver and Unmarshal* on pointer receiver.
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

//easyjson:json
//nolint:recvcheck // easyjson generates Marshal* on value receiver and Unmarshal* on pointer receiver.
type SupportCreateTicketRequest struct {
	Category          string `json:"category"`
	Title             string `json:"title"`
	Description       string `json:"description"`
	UserEmail         string `json:"user_email"`
	AttachmentFileKey string `json:"attachment_file_key"`
}

//easyjson:json
//nolint:recvcheck // easyjson generates Marshal* on value receiver and Unmarshal* on pointer receiver.
type SupportGetTicketsRequest struct {
	Status      string `json:"status"`
	Category    string `json:"category"`
	UserEmail   string `json:"user_email"`
	SupportLine int64  `json:"support_line"`
}

//easyjson:json
//nolint:recvcheck // easyjson generates Marshal* on value receiver and Unmarshal* on pointer receiver.
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

//easyjson:json
//nolint:recvcheck // easyjson generates Marshal* on value receiver and Unmarshal* on pointer receiver.
type SupportCreateTicketMessageRequest struct {
	Content        string `json:"content"`
	ContentFileKey string `json:"content_file_key"`
}

type FavoritesHTTPResponse struct {
	MovieIDs   []int64              `json:"movie_ids"`
	TotalCount int32                `json:"total_count"`
	Movies     []*moviev1.MovieCard `json:"movies"`
}

func decodeCredentials(l *jlexer.Lexer, email, password *string) {
	if skipNullCredentials(l) {
		return
	}

	l.Delim('{')

	for !l.IsDelim('}') {
		decodeCredentialField(l, email, password)
	}

	l.Delim('}')
	consumeCredentialsTopLevel(l)
}

func skipNullCredentials(l *jlexer.Lexer) bool {
	if !l.IsNull() {
		return false
	}

	consumeCredentialsTopLevel(l)
	l.Skip()

	return true
}

func consumeCredentialsTopLevel(l *jlexer.Lexer) {
	if l.IsStart() {
		l.Consumed()
	}
}

func decodeCredentialField(l *jlexer.Lexer, email, password *string) {
	key := l.UnsafeFieldName(false)
	l.WantColon()

	if l.IsNull() {
		l.Skip()
		l.WantComma()

		return
	}

	assignCredentialField(l, key, email, password)
	l.WantComma()
}

func assignCredentialField(l *jlexer.Lexer, key string, email, password *string) {
	switch key {
	case "email":
		*email = l.String()
	case "password":
		*password = l.String()
	default:
		l.AddError(&jlexer.LexerError{
			Offset: l.GetPos(),
			Reason: "unknown field",
			Data:   key,
		})
	}
}
