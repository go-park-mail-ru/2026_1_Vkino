package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	log "github.com/sirupsen/logrus"
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
		err = client.Pool.Ping(context.Background())
		if err == nil {
			break
		}

		log.Infof("trying to connect to postgres, attempts left: %d", client.connAttempts)

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
