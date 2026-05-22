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

type Config struct {
	GRPC      GRPCConfig          `mapstructure:"grpc"`
	AuthGRPC  ServiceGRPCConfig   `mapstructure:"auth_grpc"`
	MovieGRPC ServiceGRPCConfig   `mapstructure:"movie_grpc"`
	UserGRPC  ServiceGRPCConfig   `mapstructure:"user_grpc"`
	Logger    logger.Config       `mapstructure:"logger"`
	Metrics   metrics.Config      `mapstructure:"metrics"`
	Postgres  corepostgres.Config `mapstructure:"postgres"`
}

func Load(path string, cfg any) error {
	const defaultConfigPath = "configs/party.yaml"

	return configenv.Load(path, defaultConfigPath, cfg, map[string]string{
		"postgres.user":     envName("POSTGRES", "USER"),
		"postgres.password": envName("POSTGRES", "PASSWORD"),
		"postgres.dbname":   envName("POSTGRES", "DB"),
	})
}

func envName(parts ...string) string {
	return strings.Join(parts, "_")
}
