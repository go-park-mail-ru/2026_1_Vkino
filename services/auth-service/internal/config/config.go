package config

import (
	authusecase "github.com/go-park-mail-ru/2026_1_VKino/services/auth-service/internal/usecase"

	corepostgres "github.com/go-park-mail-ru/2026_1_VKino/pkg/postgresx"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/logger"
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
