package storage

import (
    "context"
    "io"
    "time"
)

type FileStorage interface {
    PutObject(
        ctx context.Context,
        key string,
        body io.Reader,
        size int64,
        contentType string,
    ) error

    DeleteObject(ctx context.Context, key string) error

    PresignGetObject(
        ctx context.Context,
        key string,
        ttl time.Duration,
    ) (string, error)
}