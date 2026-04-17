package storage

import (
	"context"
	"io"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
)

//go:generate mockgen -source=./client.go -destination=./mocks/client_mock.go -package=mocks

type MinioClient interface {
	PutObject(
		ctx context.Context,
		bucketName string,
		objectName string,
		reader io.Reader,
		objectSize int64,
		opts minio.PutObjectOptions,
	) (minio.UploadInfo, error)

	RemoveObject(ctx context.Context, bucketName string, objectName string, opts minio.RemoveObjectOptions) error

	PresignedGetObject(
		ctx context.Context,
		bucketName string,
		objectName string,
		expires time.Duration,
		reqParams url.Values,
	) (*url.URL, error)

	GetObject(
		ctx context.Context,
		bucketName string,
		objectName string,
		opts minio.GetObjectOptions,
	) (io.ReadCloser, error)
}

type minioClient struct {
	client *minio.Client
}

func (c *minioClient) PutObject(
	ctx context.Context,
	bucketName string,
	objectName string,
	reader io.Reader,
	objectSize int64,
	opts minio.PutObjectOptions,
) (minio.UploadInfo, error) {
	return c.client.PutObject(ctx, bucketName, objectName, reader, objectSize, opts)
}

func (c *minioClient) RemoveObject(
	ctx context.Context,
	bucketName string,
	objectName string,
	opts minio.RemoveObjectOptions,
) error {
	return c.client.RemoveObject(ctx, bucketName, objectName, opts)
}

func (c *minioClient) PresignedGetObject(
	ctx context.Context,
	bucketName string,
	objectName string,
	expires time.Duration,
	reqParams url.Values,
) (*url.URL, error) {
	return c.client.PresignedGetObject(ctx, bucketName, objectName, expires, reqParams)
}

func (c *minioClient) GetObject(
	ctx context.Context,
	bucketName string,
	objectName string,
	opts minio.GetObjectOptions,
) (io.ReadCloser, error) {
	return c.client.GetObject(ctx, bucketName, objectName, opts)
}
