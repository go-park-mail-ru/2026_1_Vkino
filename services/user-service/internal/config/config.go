package config

import (
	"fmt"
	"strings"

	corepostgres "github.com/go-park-mail-ru/2026_1_VKino/pkg/postgresx"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/logger"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/storage"
	"github.com/spf13/viper"
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

func Load(path string, cfg any) error {
	v := viper.New()

	const defaultConfigPath = "services/user-service/configs/user.yaml"

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
