package storage

import (
	"context"
	"errors"
	"io"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
	"go.uber.org/mock/gomock"
)

func testS3Config() Config {
	return Config{
		InternalEndpoint: "localhost:9000",
		PublicEndpoint:   "localhost:9000",
		Region:           "ru-msk",
		AccessKeyID:      "access",
		SecretAccessKey:  "secret",
		Bucket:           "avatars",
		UseSSL:           false,
		UsePathStyle:     true,
		PresignTTL:       10 * time.Minute,
	}
}

func TestS3Config_Config(t *testing.T) {
	t.Parallel()

	cfg := S3Config{
		InternalEndpoint: "localhost:9000",
		PublicEndpoint:   "localhost:9000",
		Region:           "ru-msk",
		AccessKeyID:      "access",
		SecretAccessKey:  "secret",
		BucketImages:     "images",
		BucketAvatars:    "avatars",
		BucketVideos:     "videos",
		UseSSL:           true,
		UsePathStyle:     true,
		PresignTTL:       time.Minute,
	}

	got := cfg.Config()
	if got.InternalEndpoint != cfg.InternalEndpoint || got.PublicEndpoint != cfg.PublicEndpoint ||
		got.AccessKeyID != cfg.AccessKeyID || got.SecretAccessKey != cfg.SecretAccessKey ||
		got.UseSSL != cfg.UseSSL || got.UsePathStyle != cfg.UsePathStyle || got.PresignTTL != cfg.PresignTTL {
		t.Fatalf("unexpected config conversion: %#v", got)
	}
}

func TestConfig_WithBucket(t *testing.T) {
	t.Parallel()

	cfg := testS3Config()
	got := cfg.WithBucket("images")

	if got.Bucket != "images" {
		t.Fatalf("expected bucket %q, got %q", "images", got.Bucket)
	}

	if got.InternalEndpoint != cfg.InternalEndpoint || got.PublicEndpoint != cfg.PublicEndpoint {
		t.Fatalf("expected config fields to be preserved: %#v", got)
	}
}

func TestNewS3Storage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		cfg        Config
		wantErr    string
		wantTTL    time.Duration
		wantBucket string
	}{
		{
			name:    "missing bucket",
			cfg:     Config{},
			wantErr: "storage: bucket is required",
		},
		{
			name: "negative ttl",
			cfg: func() Config {
				cfg := testS3Config()
				cfg.PresignTTL = -time.Minute

				return cfg
			}(),
			wantErr: "storage: presign ttl must be positive",
		},
		{
			name: "success with default ttl",
			cfg: func() Config {
				cfg := testS3Config()
				cfg.PresignTTL = 0

				return cfg
			}(),
			wantTTL:    15 * time.Minute,
			wantBucket: "avatars",
		},
		{
			name:       "success with explicit ttl",
			cfg:        testS3Config(),
			wantTTL:    10 * time.Minute,
			wantBucket: "avatars",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewS3Storage(context.Background(), tt.cfg)

			if tt.wantErr != "" {
				if err == nil || err.Error() != tt.wantErr {
					t.Fatalf("expected error %q, got %v", tt.wantErr, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.bucket != tt.wantBucket || got.presignTTL != tt.wantTTL || got.client == nil || got.presignClient == nil {
				t.Fatalf("unexpected storage: %#v", got)
			}
		})
	}
}

func TestS3Storage_PutObject(t *testing.T) {
	t.Parallel()

	t.Run("validation errors", func(t *testing.T) {
		store := &S3Storage{}

		if err := store.PutObject(context.Background(), "", strings.NewReader("data"), 4, "image/png"); err == nil {
			t.Fatal("expected empty key error")
		}

		if err := store.PutObject(context.Background(), "avatars/1.png", nil, 4, "image/png"); err == nil {
			t.Fatal("expected nil body error")
		}
	})

	t.Run("upload failed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		client := NewMockMinioClient(ctrl)
		client.EXPECT().
			PutObject(gomock.Any(), "avatars", "avatars/1.png", gomock.Any(), int64(4), minio.PutObjectOptions{
				ContentType: "image/png",
			}).
			Return(minio.UploadInfo{}, errors.New("upload failed"))

		store := &S3Storage{bucket: "avatars", client: client}
		err := store.PutObject(context.Background(), "avatars/1.png", strings.NewReader("data"), 4, "image/png")
		if !errors.Is(err, ErrUploadFailed) {
			t.Fatalf("expected ErrUploadFailed, got %v", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		client := NewMockMinioClient(ctrl)
		client.EXPECT().
			PutObject(gomock.Any(), "avatars", "avatars/1.png", gomock.Any(), int64(4), minio.PutObjectOptions{
				ContentType: "image/png",
			}).
			Return(minio.UploadInfo{Key: "avatars/1.png"}, nil)

		store := &S3Storage{bucket: "avatars", client: client}
		if err := store.PutObject(context.Background(), "avatars/1.png", strings.NewReader("data"), 4, "image/png"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestS3Storage_DeleteObject(t *testing.T) {
	t.Parallel()

	t.Run("empty key", func(t *testing.T) {
		store := &S3Storage{}
		if err := store.DeleteObject(context.Background(), ""); err == nil {
			t.Fatal("expected empty key error")
		}
	})

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		client := NewMockMinioClient(ctrl)
		client.EXPECT().
			RemoveObject(gomock.Any(), "avatars", "avatars/1.png", minio.RemoveObjectOptions{}).
			Return(nil)

		store := &S3Storage{bucket: "avatars", client: client}
		if err := store.DeleteObject(context.Background(), "avatars/1.png"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestS3Storage_PresignGetObject(t *testing.T) {
	t.Parallel()

	t.Run("validation errors", func(t *testing.T) {
		store := &S3Storage{presignTTL: 5 * time.Minute}

		if _, err := store.PresignGetObject(context.Background(), "", time.Minute); err == nil {
			t.Fatal("expected empty key error")
		}

		if _, err := store.PresignGetObject(context.Background(), "avatars/1.png", -time.Minute); err == nil {
			t.Fatal("expected negative ttl error")
		}
	})

	t.Run("success uses default ttl", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		client := NewMockMinioClient(ctrl)
		wantURL, _ := url.Parse("https://cdn.example/avatars/1.png")
		client.EXPECT().
			PresignedGetObject(gomock.Any(), "avatars", "avatars/1.png", 5*time.Minute, nil).
			Return(wantURL, nil)

		store := &S3Storage{bucket: "avatars", presignTTL: 5 * time.Minute, presignClient: client}
		got, err := store.PresignGetObject(context.Background(), "avatars/1.png", 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if got != wantURL.String() {
			t.Fatalf("expected url %q, got %q", wantURL.String(), got)
		}
	})
}

func TestS3Storage_GetObject(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	reader := io.NopCloser(strings.NewReader("file-body"))
	client := NewMockMinioClient(ctrl)
	client.EXPECT().
		GetObject(gomock.Any(), "avatars", "avatars/1.png", minio.GetObjectOptions{}).
		Return(reader, nil)

	store := &S3Storage{bucket: "avatars", client: client}
	got, err := store.GetObject(context.Background(), "avatars/1.png")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer got.Close()

	data, err := io.ReadAll(got)
	if err != nil {
		t.Fatalf("read object: %v", err)
	}

	if string(data) != "file-body" {
		t.Fatalf("expected body %q, got %q", "file-body", string(data))
	}
}
