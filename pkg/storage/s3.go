package storage

import (
	"context"
	"errors"
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

const defaultPresignTTL = 15 * time.Minute

var (
	errBucketRequired           = errors.New("storage: bucket is required")
	errAccessKeyRequired        = errors.New("storage: access key is required")
	errSecretKeyRequired        = errors.New("storage: secret key is required")
	errInternalEndpointRequired = errors.New("storage: internal endpoint is required")
	errPublicEndpointRequired   = errors.New("storage: public endpoint is required")
	errPresignTTLInvalid        = errors.New("storage: presign ttl must be positive")
	errEmptyObjectKey           = errors.New("storage: empty object key")
	errNilObjectBody            = errors.New("storage: nil object body")
	errTTLInvalid               = errors.New("storage: ttl must be positive")
)

func NewS3Storage(_ context.Context, cfg Config) (*S3Storage, error) {
	if err := validateS3Config(cfg); err != nil {
		return nil, err
	}

	if cfg.PresignTTL == 0 {
		cfg.PresignTTL = defaultPresignTTL
	}

	if cfg.PresignTTL < 0 {
		return nil, errPresignTTLInvalid
	}

	bucketLookup := minio.BucketLookupAuto
	if cfg.UsePathStyle {
		bucketLookup = minio.BucketLookupPath
	}

	internalClient, err := newMinioStorageClient(
		cfg.InternalEndpoint,
		cfg.AccessKeyID,
		cfg.SecretAccessKey,
		cfg.InternalUseSSL,
		cfg.Region,
		bucketLookup,
	)
	if err != nil {
		return nil, fmt.Errorf("storage: init internal minio client: %w", err)
	}

	presignClient, err := newMinioStorageClient(
		cfg.PublicEndpoint,
		cfg.AccessKeyID,
		cfg.SecretAccessKey,
		cfg.PublicUseSSL,
		cfg.Region,
		bucketLookup,
	)
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

func validateS3Config(cfg Config) error {
	switch {
	case cfg.Bucket == "":
		return errBucketRequired
	case cfg.AccessKeyID == "":
		return errAccessKeyRequired
	case cfg.SecretAccessKey == "":
		return errSecretKeyRequired
	case cfg.InternalEndpoint == "":
		return errInternalEndpointRequired
	case cfg.PublicEndpoint == "":
		return errPublicEndpointRequired
	default:
		return nil
	}
}

func newMinioStorageClient(
	endpoint, accessKeyID, secretAccessKey string,
	secure bool,
	region string,
	bucketLookup minio.BucketLookupType,
) (*minio.Client, error) {
	return minio.New(endpoint, &minio.Options{
		Creds: credentials.NewStaticV4(
			accessKeyID,
			secretAccessKey,
			"",
		),
		Secure:       secure,
		Region:       region,
		BucketLookup: bucketLookup,
	})
}

func (s *S3Storage) EnsureBucket(ctx context.Context, region string) error {
	client, ok := s.client.(*minioClient)
	if !ok {
		return nil
	}

	exists, err := client.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("storage: check bucket %q: %w", s.bucket, err)
	}

	if exists {
		return nil
	}

	if err = client.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{
		Region: region,
	}); err != nil {
		return fmt.Errorf("storage: create bucket %q: %w", s.bucket, err)
	}

	return nil
}

func (s *S3Storage) PutObject(
	ctx context.Context,
	key string,
	body io.Reader,
	size int64,
	contentType string,
) error {
	if key == "" {
		return errEmptyObjectKey
	}

	if body == nil {
		return errNilObjectBody
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
		return errEmptyObjectKey
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
		return "", errEmptyObjectKey
	}

	if ttl == 0 {
		ttl = s.presignTTL
	}

	if ttl < 0 {
		return "", errTTLInvalid
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
