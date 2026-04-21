package main

import (
    "fmt"
    "strings"

    configpkg "github.com/go-park-mail-ru/2026_1_VKino/services/api-gateway/internal/config"
    "github.com/spf13/viper"
)

func Load(path string, cfg *configpkg.Config) error {
    v := viper.New()

    const defaultConfigPath = "services/api-gateway/configs/gateway.yaml"

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