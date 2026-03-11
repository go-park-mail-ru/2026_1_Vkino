package auth

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

func LoadConfig(path string, cfg any) error {
	v := viper.New()

	// если не запускаем конкретный конфиг - используем локальный
	if path != "" {
		v.SetConfigFile(path)
	} else {
		v.AddConfigPath("configs/")
		v.SetConfigName("api")
		v.SetConfigType("yaml")
	}

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("error reading config file, %w", err)

	}

	if err := v.Unmarshal(cfg); err != nil {
		return fmt.Errorf("error unmarshalling config, %w", err)
	}

	return nil
}
