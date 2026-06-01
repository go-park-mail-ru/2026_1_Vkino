package postgres

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

const (
	pgUniqueViolationCode     = "23505"
	pgForeignKeyViolationCode = "23503"

	friendUniqueConstraint            = "friend_unique"
	friendUser1IDForeignKey           = "friend_user1_id_fkey"
	friendUser2IDForeignKey           = "friend_user2_id_fkey"
	friendRequestFromUserIDForeignKey = "friend_request_from_user_id_fkey"
	friendRequestToUserIDForeignKey   = "friend_request_to_user_id_fkey"
)

func isUniqueConstraintViolation(err error, constraintName string) bool {
	return hasPGConstraint(err, pgUniqueViolationCode, constraintName)
}

func isForeignKeyViolation(err error, constraintNames ...string) bool {
	for _, constraintName := range constraintNames {
		if hasPGConstraint(err, pgForeignKeyViolationCode, constraintName) {
			return true
		}
	}

	return false
}

func hasPGConstraint(err error, code string, constraintName string) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}

	return pgErr.Code == code && pgErr.ConstraintName == constraintName
}
