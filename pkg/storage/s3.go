package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Config struct {
	InternalEndpoint string
	PublicEndpoint   string
	Region           string
	AccessKeyID      string
	SecretAccessKey  string
	Bucket           string
	InternalUseSSL   bool
	PublicUseSSL     bool
	UsePathStyle     bool
	PresignTTL       time.Duration
}

func (c Config) WithBucket(bucket string) Config {
	return Config{
		InternalEndpoint: c.InternalEndpoint,
		PublicEndpoint:   c.PublicEndpoint,
		Region:           c.Region,
		AccessKeyID:      c.AccessKeyID,
		SecretAccessKey:  c.SecretAccessKey,
		Bucket:           bucket,
		InternalUseSSL:   c.InternalUseSSL,
		PublicUseSSL:     c.PublicUseSSL,
		UsePathStyle:     c.UsePathStyle,
		PresignTTL:       c.PresignTTL,
	}
}

type S3Storage struct {
	bucket        string
	presignTTL    time.Duration
	client        MinioClient
	presignClient MinioClient
}

func NewS3Storage(_ context.Context, cfg Config) (*S3Storage, error) {
	if cfg.Bucket == "" {
		return nil, fmt.Errorf("storage: bucket is required")
	}

	if cfg.AccessKeyID == "" {
		return nil, fmt.Errorf("storage: access key is required")
	}

	if cfg.SecretAccessKey == "" {
		return nil, fmt.Errorf("storage: secret key is required")
	}

	if cfg.InternalEndpoint == "" {
		return nil, fmt.Errorf("storage: internal endpoint is required")
	}

	if cfg.PublicEndpoint == "" {
		return nil, fmt.Errorf("storage: public endpoint is required")
	}

	if cfg.PresignTTL == 0 {
		cfg.PresignTTL = 15 * time.Minute
	}

	if cfg.PresignTTL < 0 {
		return nil, fmt.Errorf("storage: presign ttl must be positive")
	}

	bucketLookup := minio.BucketLookupAuto
	if cfg.UsePathStyle {
		bucketLookup = minio.BucketLookupPath
	}

	internalClient, err := minio.New(cfg.InternalEndpoint, &minio.Options{
		Creds: credentials.NewStaticV4(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		),
		Secure:       cfg.InternalUseSSL,
		Region:       cfg.Region,
		BucketLookup: bucketLookup,
	})
	if err != nil {
		return nil, fmt.Errorf("storage: init internal minio client: %w", err)
	}

	presignClient, err := minio.New(cfg.PublicEndpoint, &minio.Options{
		Creds: credentials.NewStaticV4(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		),
		Secure:       cfg.PublicUseSSL,
		Region:       cfg.Region,
		BucketLookup: bucketLookup,
	})
	if err != nil {
		return nil, fmt.Errorf("storage: init public minio client: %w", err)
	}

	return &S3Storage{
		bucket:        cfg.Bucket,
		presignTTL:    cfg.PresignTTL,
		client:        &minioClient{client: internalClient},
		presignClient: &minioClient{client: presignClient},
	}, nil
}

func (s *S3Storage) PutObject(
	ctx context.Context,
	key string,
	body io.Reader,
	size int64,
	contentType string,
) error {
	if key == "" {
		return fmt.Errorf("storage: empty object key")
	}

	if body == nil {
		return fmt.Errorf("storage: nil object body")
	}

	_, err := s.client.PutObject(ctx, s.bucket, key, body, size, minio.PutObjectOptions{
		ContentType: contentType,
	})

	if err != nil {
		return fmt.Errorf("%w: put object %q: %w", ErrUploadFailed, key, err)
	}

	return nil
}

func (s *S3Storage) DeleteObject(ctx context.Context, key string) error {
	if key == "" {
		return fmt.Errorf("storage: empty object key")
	}

	err := s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("storage: delete object %q: %w", key, err)
	}

	return nil
}

func (s *S3Storage) PresignGetObject(
	ctx context.Context,
	key string,
	ttl time.Duration,
) (string, error) {
	if key == "" {
		return "", fmt.Errorf("storage: empty object key")
	}

	if ttl == 0 {
		ttl = s.presignTTL
	}

	if ttl < 0 {
		return "", fmt.Errorf("storage: ttl must be positive")
	}

	u, err := s.presignClient.PresignedGetObject(ctx, s.bucket, key, ttl, nil)
	if err != nil {
		return "", fmt.Errorf("storage: presign get object %q: %w", key, err)
	}

	return u.String(), nil
}

func (s *S3Storage) GetObject(ctx context.Context, key string) (io.ReadCloser, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("storage: get object %q: %w", key, err)
	}

	return obj, nil
}
