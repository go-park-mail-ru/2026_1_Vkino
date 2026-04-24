package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/pkg/httpserver"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/logger"
	"github.com/spf13/viper"
)

type ServiceGRPCConfig struct {
	Address        string        `mapstructure:"address"`
	RequestTimeout time.Duration `mapstructure:"request_timeout"`
}

type UserAuthConfig struct {
	RefreshCookieName string        `mapstructure:"refresh_cookie_name"`
	RefreshTokenTTL   time.Duration `mapstructure:"refresh_token_ttl"`
	CookieSecure      bool          `mapstructure:"cookie_secure"`
}

type Config struct {
	Server    httpserver.Config `mapstructure:"server"`
	Logger    logger.Config     `mapstructure:"logger"`
	AuthGRPC  ServiceGRPCConfig `mapstructure:"auth_grpc"`
	UserGRPC  ServiceGRPCConfig `mapstructure:"user_grpc"`
	MovieGRPC ServiceGRPCConfig `mapstructure:"movie_grpc"`
	UserAuth  UserAuthConfig    `mapstructure:"user_auth"`
}

func (c *Config) AuthRequestTimeout() time.Duration {
	return c.AuthGRPC.RequestTimeout
}

func (c *Config) UserRequestTimeout() time.Duration {
	return c.UserGRPC.RequestTimeout
}

func (c *Config) MovieRequestTimeout() time.Duration {
	return c.MovieGRPC.RequestTimeout
}

func (c *Config) RefreshCookieName() string {
	return c.UserAuth.RefreshCookieName
}

func (c *Config) CookieSecure() bool {
	return c.UserAuth.CookieSecure
}

func Load(path string, cfg *Config) error {
	v := viper.New()

	const defaultConfigPath = "configs/gateway.yaml"

	if path != "" {
		v.SetConfigFile(path)
	} else {
		v.SetConfigFile(defaultConfigPath)
	}

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}

	if err := v.Unmarshal(cfg); err != nil {
		return fmt.Errorf("error unmarshalling config: %w", err)
	}

	return nil
}
