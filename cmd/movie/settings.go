package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2026_1_VKino/pkg/configenv"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/logger"
	corepostgres "github.com/go-park-mail-ru/2026_1_VKino/pkg/postgresx"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/storage"
	"github.com/spf13/viper"
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
	Postgres corepostgres.Config `mapstructure:"postgres"`
	S3       storage.S3Config    `mapstructure:"s3"`
}

func Load(path string, cfg any) error {
	v := viper.New()

	const defaultConfigPath = "services/movie/configs/config.yaml"

	if path != "" {
		v.SetConfigFile(path)
	} else {
		v.SetConfigFile(defaultConfigPath)
	}

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := configenv.Bind(v, map[string]string{
		"postgres.user":        "POSTGRES_USER",
		"postgres.password":    "POSTGRES_PASSWORD",
		"postgres.dbname":      "POSTGRES_DB",
		"s3.access_key_id":     "MINIO_ROOT_USER",
		"s3.secret_access_key": "MINIO_ROOT_PASSWORD",
	}); err != nil {
		return fmt.Errorf("error binding env: %w", err)
	}

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}

	if err := v.Unmarshal(cfg); err != nil {
		return fmt.Errorf("error unmarshalling config: %w", err)
	}

	return nil
}
