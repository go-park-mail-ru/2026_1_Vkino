package main

import (
	"fmt"
	"strings"

	"github.com/go-park-mail-ru/2026_1_VKino/pkg/configenv"
	"github.com/spf13/viper"
)

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
