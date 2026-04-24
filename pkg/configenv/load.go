package configenv

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

func Load(path, defaultPath string, cfg any, bindings map[string]string) error {
	v := viper.New()

	if path != "" {
		v.SetConfigFile(path)
	} else {
		v.SetConfigFile(defaultPath)
	}

	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if len(bindings) > 0 {
		if err := Bind(v, bindings); err != nil {
			return fmt.Errorf("error binding env: %w", err)
		}
	}

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}

	if err := v.Unmarshal(cfg); err != nil {
		return fmt.Errorf("error unmarshalling config: %w", err)
	}

	return nil
}
