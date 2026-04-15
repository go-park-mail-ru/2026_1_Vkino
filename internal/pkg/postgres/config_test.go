package postgres

import (
	"strings"
	"testing"
	"time"

	"go.uber.org/mock/gomock"
)

func TestConfig_SetDefaults(t *testing.T) {
	t.Parallel()

	cfg := &Config{}
	cfg.SetDefaults()

	if cfg.MaxPoolSize != DefaultMaxPoolSize {
		t.Fatalf("expected max pool size %d, got %d", DefaultMaxPoolSize, cfg.MaxPoolSize)
	}

	if cfg.ConnAttempts != DefaultConnAttempts {
		t.Fatalf("expected conn attempts %d, got %d", DefaultConnAttempts, cfg.ConnAttempts)
	}

	if cfg.ConnTimeout != DefaultConnTimeout {
		t.Fatalf("expected conn timeout %v, got %v", DefaultConnTimeout, cfg.ConnTimeout)
	}
}

func TestConfig_DSN(t *testing.T) {
	t.Parallel()

	cfg := &Config{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "secret",
		DBName:   "vkino",
		SSLMode:  "disable",
	}

	want := "postgres://postgres:secret@localhost:5432/vkino?sslmode=disable"
	if got := cfg.DSN(); got != want {
		t.Fatalf("expected DSN %q, got %q", want, got)
	}
}

func TestBuildPostgresOptions(t *testing.T) {
	t.Parallel()

	cfg := &Config{
		MaxPoolSize:  4,
		ConnAttempts: 5,
		ConnTimeout:  2 * time.Second,
	}

	client := &Client{}
	for _, opt := range BuildPostgresOptions(cfg) {
		opt(client)
	}

	if client.maxPoolSize != 4 || client.connAttempts != 5 || client.connTimeout != 2*time.Second {
		t.Fatalf("unexpected client options: %+v", client)
	}
}

func TestBuildPostgresOptions_Empty(t *testing.T) {
	t.Parallel()

	if got := BuildPostgresOptions(&Config{}); len(got) != 0 {
		t.Fatalf("expected no options, got %d", len(got))
	}
}

func TestClient_Close(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pool := NewMockPool(ctrl)
	pool.EXPECT().Close()

	client := &Client{Pool: pool}
	client.Close()
}

func TestNew_FailedConnection(t *testing.T) {
	t.Parallel()

	cfg := Config{
		Host:         "127.0.0.1",
		Port:         1,
		User:         "postgres",
		Password:     "secret",
		DBName:       "vkino",
		SSLMode:      "disable",
		MaxPoolSize:  1,
		ConnAttempts: 1,
		ConnTimeout:  time.Nanosecond,
	}

	_, err := New(cfg)
	if err == nil {
		t.Fatal("expected connection error, got nil")
	}

	if !strings.Contains(err.Error(), "failed to connect to postgres") {
		t.Fatalf("unexpected error: %v", err)
	}
}
