package postgres

import (
	"fmt"
	"time"
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
