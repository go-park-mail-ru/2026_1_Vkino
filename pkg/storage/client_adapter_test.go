package storage

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func newClosedMinioClient(t *testing.T) *minioClient {
	t.Helper()

	client, err := minio.New("127.0.0.1:1", &minio.Options{
		Creds:  credentials.NewStaticV4("access", "secret", ""),
		Secure: false,
	})
	if err != nil {
		t.Fatalf("new minio client: %v", err)
	}

	return &minioClient{client: client}
}

func TestMinioClientAdapter(t *testing.T) {
	t.Parallel()

	client := newClosedMinioClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	if _, err := client.PutObject(ctx, "avatars", "avatars/1.png", strings.NewReader("data"), 4, minio.PutObjectOptions{}); err == nil {
		t.Fatal("expected put object error")
	}

	if err := client.RemoveObject(ctx, "avatars", "avatars/1.png", minio.RemoveObjectOptions{}); err == nil {
		t.Fatal("expected remove object error")
	}

	u, err := client.PresignedGetObject(ctx, "avatars", "avatars/1.png", time.Minute, nil)
	if u == nil && err == nil {
		t.Fatal("expected presigned url or error")
	}

	if u != nil && !strings.Contains(u.String(), "avatars") {
		t.Fatalf("unexpected presigned url: %q", u.String())
	}

	obj, err := client.GetObject(ctx, "avatars", "avatars/1.png", minio.GetObjectOptions{})
	if obj == nil && err == nil {
		t.Fatal("expected object or error from get object")
	}
}
