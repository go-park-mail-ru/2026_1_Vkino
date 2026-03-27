package app

import (
	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth/usecase"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/httpserver"
	"github.com/go-park-mail-ru/2026_1_VKino/pkg/storage"
)

type Config struct {
	Server httpserver.Config `mapstructure:"server"`
	Auth   usecase.Config    `mapstructure:"auth"`
	S3     storage.S3Config  `mapstructure:"s3"`
}