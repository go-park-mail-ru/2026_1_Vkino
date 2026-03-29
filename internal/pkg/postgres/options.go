package postgres

import (
	"time"
)

type Option func(*Postgres)

func MaxPoolSize(size int) Option {
	return func(p *Postgres) {
		p.maxPoolSize = size
	}
}

func ConnAttempts(attempts int) Option {
	return func(c *Postgres) {
		c.connAttempts = attempts
	}
}

func ConnTimeout(timeout time.Duration) Option {
	return func(c *Postgres) {
		c.connTimeout = timeout
	}
}

func BuildPostgresOptions(cfg *Config) []Option {
	var opts []Option

	if cfg.MaxPoolSize > 0 {
		opts = append(opts, MaxPoolSize(cfg.MaxPoolSize))
	}

	if cfg.ConnAttempts > 0 {
		opts = append(opts, ConnAttempts(cfg.ConnAttempts))
	}

	if cfg.ConnTimeout > 0 {
		opts = append(opts, ConnTimeout(cfg.ConnTimeout))
	}

	return opts
}
