package main

import (
	"fmt"
	"strings"

	"github.com/go-park-mail-ru/2026_1_VKino/pkg/configenv"
	"github.com/spf13/viper"
)

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

	if err := configenv.Bind(v, map[string]string{
		"auth.jwt_secret":   "AUTH_JWT_SECRET",
		"postgres.user":     "POSTGRES_USER",
		"postgres.password": "POSTGRES_PASSWORD",
		"postgres.dbname":   "POSTGRES_DB",
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
