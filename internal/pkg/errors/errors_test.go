package errors

import (
	"errors"
	"net/http"
	"testing"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/domain"
	"github.com/stretchr/testify/assert"
)

func TestMapError(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		httpErr HttpErr
	}{
		{
			name: "err_found",
			err:  domain.ErrNoSession,
			httpErr: HttpErr{
				status:  http.StatusUnauthorized,
				message: "unauthorized",
			},
		},
		{
			name: "err_not_found",
			err:  errors.New("unknown error"),
			httpErr: HttpErr{
				status:  http.StatusInternalServerError,
				message: "internal server error",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, message := MapError(tt.err)
			assert.Equal(t, tt.httpErr.status, status)
			assert.Equal(t, tt.httpErr.message, message)
		})
	}
}
