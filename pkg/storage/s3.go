package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
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
	UseSSL           bool
	UsePathStyle     bool
	PresignTTL       time.Duration
}

type S3Storage struct {
	bucket       string
	presignTTL   time.Duration
	client       *minio.Client
	presignClient *minio.Client
}

func NewS3Storage(_ context.Context, cfg Config) (*S3Storage, error) {
	if strings.TrimSpace(cfg.InternalEndpoint) == "" {
		return nil, fmt.Errorf("storage: internal endpoint is required")
	}
	if strings.TrimSpace(cfg.PublicEndpoint) == "" {
		return nil, fmt.Errorf("storage: public endpoint is required")
	}
	if strings.TrimSpace(cfg.Bucket) == "" {
		return nil, fmt.Errorf("storage: bucket is required")
	}
	if strings.TrimSpace(cfg.AccessKeyID) == "" {
		return nil, fmt.Errorf("storage: access key is required")
	}
	if strings.TrimSpace(cfg.SecretAccessKey) == "" {
		return nil, fmt.Errorf("storage: secret key is required")
	}
	if cfg.Region == "" {
		cfg.Region = "us-east-1"
	}
	if cfg.PresignTTL <= 0 {
		cfg.PresignTTL = 15 * time.Minute
	}

	internalEndpoint, internalSecure, err := normalizeEndpoint(cfg.InternalEndpoint, cfg.UseSSL)
	if err != nil {
		return nil, fmt.Errorf("storage: invalid internal endpoint: %w", err)
	}

	publicEndpoint, publicSecure, err := normalizeEndpoint(cfg.PublicEndpoint, cfg.UseSSL)
	if err != nil {
		return nil, fmt.Errorf("storage: invalid public endpoint: %w", err)
	}

	bucketLookup := minio.BucketLookupAuto
	if cfg.UsePathStyle {
		bucketLookup = minio.BucketLookupPath
	}

	internalClient, err := minio.New(internalEndpoint, &minio.Options{
		Creds: credentials.NewStaticV4(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		),
		Secure:       internalSecure,
		Region:       cfg.Region,
		BucketLookup: bucketLookup,
	})
	if err != nil {
		return nil, fmt.Errorf("storage: init internal minio client: %w", err)
	}

	publicClient, err := minio.New(publicEndpoint, &minio.Options{
		Creds: credentials.NewStaticV4(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		),
		Secure:       publicSecure,
		Region:       cfg.Region,
		BucketLookup: bucketLookup,
	})
	if err != nil {
		return nil, fmt.Errorf("storage: init public minio client: %w", err)
	}

	return &S3Storage{
		bucket:        cfg.Bucket,
		presignTTL:    cfg.PresignTTL,
		client:        internalClient,
		presignClient: publicClient,
	}, nil
}

func (s *S3Storage) PutObject(
	ctx context.Context,
	key string,
	body io.Reader,
	size int64,
	contentType string,
) error {
	if strings.TrimSpace(key) == "" {
		return fmt.Errorf("storage: empty object key")
	}
	if body == nil {
		return fmt.Errorf("storage: nil object body")
	}

	opts := minio.PutObjectOptions{
		ContentType: contentType,
	}

	_, err := s.client.PutObject(ctx, s.bucket, key, body, size, opts)
	if err != nil {
		return fmt.Errorf("%w: put object %q: %w", ErrUploadFailed, key, err)
	}

	return nil
}

func (s *S3Storage) DeleteObject(ctx context.Context, key string) error {
	if strings.TrimSpace(key) == "" {
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
	if strings.TrimSpace(key) == "" {
		return "", fmt.Errorf("storage: empty object key")
	}
	if ttl <= 0 {
		ttl = s.presignTTL
	}

	u, err := s.presignClient.PresignedGetObject(ctx, s.bucket, key, ttl, nil)
	if err != nil {
		return "", fmt.Errorf("storage: presign get object %q: %w", key, err)
	}

	return u.String(), nil
}

func normalizeEndpoint(raw string, defaultUseSSL bool) (endpoint string, secure bool, err error) {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimRight(raw, "/")
	if raw == "" {
		return "", false, fmt.Errorf("empty endpoint")
	}

	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		u, err := url.Parse(raw)
		if err != nil {
			return "", false, fmt.Errorf("parse url %q: %w", raw, err)
		}
		if u.Host == "" {
			return "", false, fmt.Errorf("endpoint %q has empty host", raw)
		}
		return u.Host, u.Scheme == "https", nil
	}

	return raw, defaultUseSSL, nil
}

func IsNotFound(err error) bool {
	if err == nil {
		return false
	}

	resp := minio.ToErrorResponse(err)
	switch resp.Code {
	case "NoSuchKey", "NoSuchBucket", "ResourceNotFound", "NoSuchUpload":
		return true
	default:
		return false
	}
}