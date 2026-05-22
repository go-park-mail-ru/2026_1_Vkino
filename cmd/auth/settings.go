package main

import (
	"strings"

	authusecase "github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth-service/usecase"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/configenv"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/metrics"
	corepostgres "github.com/go-park-mail-ru/2026_1_VKino/pkg/postgresx"
)

type GRPCConfig struct {
	Port int `mapstructure:"port"`
}

type Config struct {
	GRPC     GRPCConfig          `mapstructure:"grpc"`
	Logger   logger.Config       `mapstructure:"logger"`
	Metrics  metrics.Config      `mapstructure:"metrics"`
	Auth     authusecase.Config  `mapstructure:"auth"`
	Postgres corepostgres.Config `mapstructure:"postgres"`
}

func Load(path string, cfg any) error {
	const defaultConfigPath = "configs/auth.yaml"

	return configenv.Load(path, defaultConfigPath, cfg, map[string]string{
		"auth.jwt_secret":   envName("AUTH", "JWT", "SECRET"),
		"postgres.user":     envName("POSTGRES", "USER"),
		"postgres.password": envName("POSTGRES", "PASSWORD"),
		"postgres.dbname":   envName("POSTGRES", "DB"),
	})
}

func envName(parts ...string) string {
	return strings.Join(parts, "_")
}
