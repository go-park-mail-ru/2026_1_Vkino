package postgresx

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func newClosedPgxPool(t *testing.T) *pgxPool {
	t.Helper()

	cfg, err := pgxpool.ParseConfig("postgres://postgres:secret@127.0.0.1:1/vkino?sslmode=disable")
	if err != nil {
		t.Fatalf("parse config: %v", err)
	}

	rawPool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		t.Fatalf("new pool: %v", err)
	}

	t.Cleanup(rawPool.Close)

	return &pgxPool{pool: rawPool}
}

func TestPgxPoolAdapter(t *testing.T) {
	t.Parallel()

	pool := newClosedPgxPool(t)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	if _, err := pool.Query(ctx, "select 1"); err == nil {
		t.Fatal("expected query error")
	}

	if _, err := pool.Exec(ctx, "select 1"); err == nil {
		t.Fatal("expected exec error")
	}

	row := pool.QueryRow(ctx, "select 1")

	var value int
	if err := row.Scan(&value); err == nil {
		t.Fatal("expected query row scan error")
	}

	pool.Close()
}
