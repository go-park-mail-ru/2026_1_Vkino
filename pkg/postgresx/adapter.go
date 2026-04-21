package postgresx

import (
	"context"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type pgxPool struct {
	pool *pgxpool.Pool
}

func (p *pgxPool) Query(ctx context.Context, sql string, args ...any) (Rows, error) {
	return p.pool.Query(ctx, sql, args...)
}

func (p *pgxPool) QueryRow(ctx context.Context, sql string, args ...any) Row {
	return p.pool.QueryRow(ctx, sql, args...)
}

func (p *pgxPool) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return p.pool.Exec(ctx, sql, args...)
}

func (p *pgxPool) Ping(ctx context.Context) error {
	return p.pool.Ping(ctx)
}

func (p *pgxPool) Close() {
	p.pool.Close()
}
