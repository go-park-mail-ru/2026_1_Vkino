package postgresx

import (
	"context"

	"github.com/jackc/pgx/v5/pgconn"
)

//go:generate mockgen -source=./interfaces.go -destination=./mocks/db_mock.go -package=mocks

type Row interface {
	Scan(dest ...any) error
}

type Rows interface {
	Close()
	Err() error
	Next() bool
	Scan(dest ...any) error
}

type Pool interface {
	Query(ctx context.Context, sql string, args ...any) (Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Ping(ctx context.Context) error
	Close()
}
