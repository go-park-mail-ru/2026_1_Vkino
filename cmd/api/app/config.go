package app

import (
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/usecase"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/httpserver"
)

type Config struct {
	Server httpserver.Config `mapstructure:"server"`
	Auth   usecase.Config    `mapstructure:"auth"`
}
