package config

import (
	"fmt"
	"strings"

	authusecase "github.com/go-park-mail-ru/2026_1_VKino/services/auth-service/internal/usecase"

	corepostgres "github.com/go-park-mail-ru/2026_1_VKino/internal/pkg/postgres"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/logger"
	"github.com/spf13/viper"
)

type GRPCConfig struct {
	Port int `mapstructure:"port"`
}

type Config struct {
	GRPC     GRPCConfig         `mapstructure:"grpc"`
	Logger   logger.Config      `mapstructure:"logger"`
	Auth     authusecase.Config `mapstructure:"auth"`
	Postgres corepostgres.Config `mapstructure:"postgres"`
}

func Load(path string, cfg any) error {
	v := viper.New()

	const defaultConfigPath = "services/auth-service/configs/auth.yaml"

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