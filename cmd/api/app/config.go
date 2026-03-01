package app

import (
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/server"
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth"
)

type Config struct {
	Server server.Config `mapstructure:"server"`
	Auth   auth.Config   `mapstructure:"auth"`
}
