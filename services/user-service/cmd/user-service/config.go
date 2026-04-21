package main

import (
	"fmt"
	"strings"
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

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}

	if err := v.Unmarshal(cfg); err != nil {
		return fmt.Errorf("error unmarshalling config: %w", err)
	}

	return nil
}
