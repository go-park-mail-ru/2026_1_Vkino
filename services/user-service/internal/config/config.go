package config

import (
	corepostgres "github.com/go-park-mail-ru/2026_1_VKino/pkg/postgresx"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/storage"
)

type GRPCConfig struct {
	Port int `mapstructure:"port"`
}

type Config struct {
	GRPC     GRPCConfig          `mapstructure:"grpc"`
	Logger   logger.Config       `mapstructure:"logger"`
	Postgres corepostgres.Config `mapstructure:"postgres"`
	S3       storage.S3Config    `mapstructure:"s3"`
}
