package app

import (
	"fmt"
	"strings"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/usecase"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/pkg/middleware"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/pkg/postgres"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/httpserver"
	"github.com/spf13/viper"
)

type Config struct {
	Server   httpserver.Config     `mapstructure:"server"`
	Auth     usecase.Config        `mapstructure:"auth"`
	Postgres postgres.Config       `mapstructure:"postgres"`
	CORS     middleware.CORSConfig `mapstructure:"cors"`
}

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
