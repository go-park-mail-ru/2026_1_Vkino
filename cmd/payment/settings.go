package main

import (
	"strings"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/pkg/configenv"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/metrics"
	corepostgres "github.com/go-park-mail-ru/2026_1_VKino/pkg/postgresx"
)

type GRPCConfig struct {
	Port int `mapstructure:"port"`
}

type ServiceGRPCConfig struct {
	Address        string        `mapstructure:"address"`
	RequestTimeout time.Duration `mapstructure:"request_timeout"`
}

type YooKassaConfig struct {
	APIURL    string        `mapstructure:"api_url"`
	ShopID    string        `mapstructure:"shop_id"`
	SecretKey string        `mapstructure:"secret_key"`
	ReturnURL string        `mapstructure:"return_url"`
	Capture   bool          `mapstructure:"capture"`
	Timeout   time.Duration `mapstructure:"timeout"`
}

type Config struct {
	GRPC     GRPCConfig          `mapstructure:"grpc"`
	AuthGRPC ServiceGRPCConfig   `mapstructure:"auth_grpc"`
	UserGRPC ServiceGRPCConfig   `mapstructure:"user_grpc"`
	Logger   logger.Config       `mapstructure:"logger"`
	Metrics  metrics.Config      `mapstructure:"metrics"`
	Postgres corepostgres.Config `mapstructure:"postgres"`
	YooKassa YooKassaConfig      `mapstructure:"yookassa"`
}

func Load(path string, cfg any) error {
	const defaultConfigPath = "configs/payment.yaml"

	return configenv.Load(path, defaultConfigPath, cfg, map[string]string{
		"postgres.user":       envName("POSTGRES", "USER"),
		"postgres.password":   envName("POSTGRES", "PASSWORD"),
		"postgres.dbname":     envName("POSTGRES", "DB"),
		"yookassa.shop_id":    envName("YOOKASSA", "SHOP", "ID"),
		"yookassa.secret_key": envName("YOOKASSA", "SECRET", "KEY"),
	})
}

func envName(parts ...string) string {
	return strings.Join(parts, "_")
}
