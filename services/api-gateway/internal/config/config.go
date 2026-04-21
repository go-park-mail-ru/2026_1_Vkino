package config

import (
    "time"

    "github.com/go-park-mail-ru/2026_1_VKino/pkg/httpserver"
    "github.com/go-park-mail-ru/2026_1_VKino/pkg/logger"
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