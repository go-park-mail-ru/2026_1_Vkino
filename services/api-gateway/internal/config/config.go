package config

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

type LegacyAPIConfig struct {
	BaseURL string `mapstructure:"base_url"`
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
	LegacyAPI LegacyAPIConfig   `mapstructure:"legacy_api"`
	UserAuth  UserAuthConfig    `mapstructure:"user_auth"`
}

func Load(path string, cfg any) error {
	v := viper.New()

	const defaultConfigPath = "services/api-gateway/configs/gateway.yaml"

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
