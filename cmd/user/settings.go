package main

import (
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/pkg/configenv"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/metrics"
	corepostgres "github.com/go-park-mail-ru/2026_1_VKino/pkg/postgresx"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/storage"
)

type GRPCConfig struct {
	Port int `mapstructure:"port"`
}

type ServiceGRPCConfig struct {
	Address        string        `mapstructure:"address"`
	RequestTimeout time.Duration `mapstructure:"request_timeout"`
}

type Config struct {
	GRPC     GRPCConfig          `mapstructure:"grpc"`
	AuthGRPC ServiceGRPCConfig   `mapstructure:"auth_grpc"`
	Logger   logger.Config       `mapstructure:"logger"`
	Metrics  metrics.Config      `mapstructure:"metrics"`
	Postgres corepostgres.Config `mapstructure:"postgres"`
	S3       storage.S3Config    `mapstructure:"s3"`
}

func Load(path string, cfg any) error {
	const defaultConfigPath = "configs/user.yaml"

	return configenv.Load(path, defaultConfigPath, cfg, map[string]string{
		"postgres.user":        "POSTGRES_USER",
		"postgres.password":    "POSTGRES_PASSWORD",
		"postgres.dbname":      "POSTGRES_DB",
		"s3.access_key_id":     "MINIO_ROOT_USER",
		"s3.secret_access_key": "MINIO_ROOT_PASSWORD",
	})
}
