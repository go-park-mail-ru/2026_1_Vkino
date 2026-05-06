package postgresx

import (
	"net"
	"net/url"
	"strconv"
	"time"
)

const (
	DefaultMaxPoolSize  = 1
	DefaultConnAttempts = 10
	DefaultConnTimeout  = time.Second
)

type Config struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	DBName          string `mapstructure:"dbname"`
	SSLMode         string `mapstructure:"sslmode"`
	ApplicationName string `mapstructure:"application_name"`

	MaxPoolSize  int           `mapstructure:"max_pool_size"`
	ConnAttempts int           `mapstructure:"conn_attempts"`
	ConnTimeout  time.Duration `mapstructure:"conn_timeout"`

	StatementTimeout                time.Duration `mapstructure:"statement_timeout"`
	LockTimeout                     time.Duration `mapstructure:"lock_timeout"`
	IdleInTransactionSessionTimeout time.Duration `mapstructure:"idle_in_transaction_session_timeout"`
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

	if c.SSLMode == "" {
		c.SSLMode = "disable"
	}
}

func (c *Config) DSN() string {
	hostPort := net.JoinHostPort(c.Host, strconv.Itoa(c.Port))

	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.User, c.Password),
		Host:   hostPort,
		Path:   c.DBName,
	}

	q := u.Query()
	q.Set("sslmode", c.SSLMode)
	u.RawQuery = q.Encode()

	return u.String()
}
