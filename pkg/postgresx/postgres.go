package postgresx

import (
	"context"
	"errors"
	"fmt"
	"math"
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

var errBeginUnsupported = errors.New("postgres: begin is not supported for this pool implementation")

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

	poolCfg.MaxConns = clampToInt32(client.maxPoolSize)

	if cfg.ApplicationName != "" {
		poolCfg.ConnConfig.RuntimeParams["application_name"] = cfg.ApplicationName
	}

	if cfg.StatementTimeout > 0 {
		poolCfg.ConnConfig.RuntimeParams["statement_timeout"] = cfg.StatementTimeout.String()
	}

	if cfg.LockTimeout > 0 {
		poolCfg.ConnConfig.RuntimeParams["lock_timeout"] = cfg.LockTimeout.String()
	}

	if cfg.IdleInTransactionSessionTimeout > 0 {
		poolCfg.ConnConfig.RuntimeParams["idle_in_transaction_session_timeout"] =
			cfg.IdleInTransactionSessionTimeout.String()
	}

	poolCfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeCacheStatement
	poolCfg.ConnConfig.StatementCacheCapacity = 512

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

		logger.FromContext(context.TODO()).
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

// Begin starts a transaction. Supported only when Pool is backed by pgxpool.Pool.
func (p *Client) Begin(ctx context.Context) (pgx.Tx, error) {
	pool, ok := p.Pool.(*pgxPool)
	if !ok {
		return nil, errBeginUnsupported
	}

	return pool.pool.Begin(ctx)
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

func clampToInt32(value int) int32 {
	if value > math.MaxInt32 {
		return math.MaxInt32
	}

	if value < math.MinInt32 {
		return math.MinInt32
	}

	return int32(value)
}
