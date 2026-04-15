package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/pkg/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Client struct {
	Pool Pool

	maxPoolSize  int
	connAttempts int
	connTimeout  time.Duration
}

func New(cfg Config, opts ...Option) (*Client, error) {
	cfg.SetDefaults()
	dsn := cfg.DSN()

	client := &Client{
		maxPoolSize:  cfg.MaxPoolSize,
		connAttempts: cfg.ConnAttempts,
		connTimeout:  cfg.ConnTimeout,
	}

	for _, opt := range opts {
		opt(client)
	}

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.ParseConfig failes: %w", err)
	}

	poolCfg.MaxConns = int32(client.maxPoolSize)

	rawPool, err := pgxpool.NewWithConfig(context.Background(), poolCfg)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.NewWithConfig failed: %w", err)
	}

	client.Pool = &pgxPool{pool: rawPool}

	for client.connAttempts > 0 {
		err = client.Ping(context.Background())
		if err == nil {
			break
		}

		logger.FromContext(nil).
			WithField("component", "postgres").
			WithField("attempts_left", client.connAttempts).
			WithField("error", err).
			Warn("trying to connect to postgres")

		time.Sleep(client.connTimeout)
		client.connAttempts--
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	return client, nil
}

func (p *Client) Close() {
	p.Pool.Close()
}

func (p *Client) Ping(ctx context.Context) error {
	startedAt := time.Now()
	err := p.Pool.Ping(ctx)
	p.logCall(ctx, "ping", "", 0, time.Since(startedAt), err, nil)

	return err
}

func (p *Client) Query(ctx context.Context, query string, args ...any) (Rows, error) {
	startedAt := time.Now()
	rows, err := p.Pool.Query(ctx, query, args...)
	p.logCall(ctx, "query", query, len(args), time.Since(startedAt), err, nil)

	return rows, err
}

func (p *Client) QueryRow(ctx context.Context, query string, args ...any) Row {
	return &loggingRow{
		row:       p.Pool.QueryRow(ctx, query, args...),
		client:    p,
		ctx:       ctx,
		query:     query,
		argsCount: len(args),
		startedAt: time.Now(),
	}
}

func (p *Client) Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	startedAt := time.Now()
	tag, err := p.Pool.Exec(ctx, query, args...)

	extraFields := map[string]any{
		"rows_affected": tag.RowsAffected(),
	}
	p.logCall(ctx, "exec", query, len(args), time.Since(startedAt), err, extraFields)

	return tag, err
}

type loggingRow struct {
	row       pgx.Row
	client    *Client
	ctx       context.Context
	query     string
	argsCount int
	startedAt time.Time
}

func (r *loggingRow) Scan(dest ...any) error {
	err := r.row.Scan(dest...)
	r.client.logCall(r.ctx, "query_row", r.query, r.argsCount, time.Since(r.startedAt), err, nil)

	return err
}

func (p *Client) logCall(
	ctx context.Context,
	operation string,
	query string,
	argsCount int,
	duration time.Duration,
	err error,
	extraFields map[string]any,
) {
	dbLogger := logger.FromContext(ctx).
		WithField("component", "postgres").
		WithField("db_operation", operation).
		WithField("duration", duration.String()).
		WithField("args_count", argsCount)

	if query != "" {
		dbLogger = dbLogger.WithField("query", compactQuery(query))
	}

	for key, value := range extraFields {
		dbLogger = dbLogger.WithField(key, value)
	}

	switch {
	case err == nil:
		dbLogger.Info("db call completed")
	case errors.Is(err, pgx.ErrNoRows):
		dbLogger.Info("db call returned no rows")
	default:
		dbLogger.WithField("error", err).Error("db call failed")
	}
}

func compactQuery(query string) string {
	return strings.Join(strings.Fields(query), " ")
}
