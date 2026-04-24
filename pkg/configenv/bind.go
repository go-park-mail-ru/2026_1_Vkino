package configenv

import (
	"fmt"

	"github.com/spf13/viper"
)

func Bind(v *viper.Viper, bindings map[string]string) error {
	for key, envName := range bindings {
		if err := v.BindEnv(key, envName); err != nil {
			return fmt.Errorf("bind env %q to key %q: %w", envName, key, err)
		}
	}

	return nil
}
