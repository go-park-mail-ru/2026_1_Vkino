package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	log "github.com/sirupsen/logrus"
)

const (
	defaultMaxPoolSize  = 1
	defaultConnAttempts = 10
	defaultConnTimeout  = time.Second
)

type Postgres struct {
	Pool *pgxpool.Pool

	maxPoolSize  int
	connAttempts int
	connTimeout  time.Duration
}

func New(dsn string, opts ...Option) (*Postgres, error) {
	pg := &Postgres{
		maxPoolSize:  defaultMaxPoolSize,
		connAttempts: defaultConnAttempts,
		connTimeout:  defaultConnTimeout,
	}

	for _, opt := range opts {
		opt(pg)
	}

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.ParseConfig failes: %w", err)
	}

	poolCfg.MaxConns = int32(pg.maxPoolSize)

	pg.Pool, err = pgxpool.NewWithConfig(context.Background(), poolCfg)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.NewWithConfig failed: %w", err)
	}

	for pg.connAttempts > 0 {
		err = pg.Pool.Ping(context.Background())
		if err == nil {
			break
		}

		log.Infof("trying to connect to postgres, attempts left: %d", pg.connAttempts)

		time.Sleep(pg.connTimeout)
		pg.connAttempts--
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	return pg, nil
}

func (p *Postgres) Close() {
	p.Pool.Close()
}
