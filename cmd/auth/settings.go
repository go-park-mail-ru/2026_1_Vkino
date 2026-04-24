package main

import (
	authusecase "github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth-service/usecase"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/configenv"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/logger"
	corepostgres "github.com/go-park-mail-ru/2026_1_VKino/pkg/postgresx"
)

type GRPCConfig struct {
	Port int `mapstructure:"port"`
}

type Config struct {
	GRPC     GRPCConfig          `mapstructure:"grpc"`
	Logger   logger.Config       `mapstructure:"logger"`
	Auth     authusecase.Config  `mapstructure:"auth"`
	Postgres corepostgres.Config `mapstructure:"postgres"`
}

func Load(path string, cfg any) error {
	const defaultConfigPath = "configs/auth.yaml"

	return configenv.Load(path, defaultConfigPath, cfg, map[string]string{
		"auth.jwt_secret":   "AUTH_JWT_SECRET",
		"postgres.user":     "POSTGRES_USER",
		"postgres.password": "POSTGRES_PASSWORD",
		"postgres.dbname":   "POSTGRES_DB",
	})
}
