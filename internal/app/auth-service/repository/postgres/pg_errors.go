package postgres

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

const (
	pgUniqueViolationCode = "23505"

	userEmailUniqueConstraint = "user_email_unique"
)

func isUniqueConstraintViolation(err error, constraintName string) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}

	return pgErr.Code == pgUniqueViolationCode && pgErr.ConstraintName == constraintName
}
