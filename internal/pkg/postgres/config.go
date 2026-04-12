package postgres

import (
	"fmt"
	"time"
)

const (
	DefaultMaxPoolSize  = 1
	DefaultConnAttempts = 10
	DefaultConnTimeout  = time.Second
)

type Config struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	User         string        `mapstructure:"user"`
	Password     string        `mapstructure:"password"`
	DBName       string        `mapstructure:"dbname"`
	SSLMode      string        `mapstructure:"sslmode"`
	MaxPoolSize  int           `mapstructure:"max_pool_size"`
	ConnAttempts int           `mapstructure:"conn_attempts"`
	ConnTimeout  time.Duration `mapstructure:"conn_timeout"`
}

func (c *Config) SetDefaults() {
	if c.MaxPoolSize == 0 {
		c.MaxPoolSize = DefaultMaxPoolSize
	}

	if c.ConnAttempts == 0 {
		c.ConnAttempts = DefaultConnAttempts
	}

	if c.ConnTimeout == 0 {
		c.ConnTimeout = DefaultConnTimeout
	}
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.DBName,
		c.SSLMode,
	)
}
