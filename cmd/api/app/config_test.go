package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	t.Parallel()

	t.Run("explicit path", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "api.yaml")

		configContent := `
server:
  port: 9090
auth:
  jwt_secret: test-secret
postgres:
  host: localhost
`

		if err := os.WriteFile(path, []byte(configContent), 0o644); err != nil {
			t.Fatalf("write config: %v", err)
		}

		var cfg Config
		if err := LoadConfig(path, &cfg); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cfg.Server.Port != 9090 || cfg.User.JWTSecret != "test-secret" || cfg.Postgres.Host != "localhost" {
			t.Fatalf("unexpected config: %+v", cfg)
		}
	})

	t.Run("default path", func(t *testing.T) {
		wd, err := os.Getwd()
		if err != nil {
			t.Fatalf("getwd: %v", err)
		}

		root := filepath.Clean(filepath.Join(wd, "../../.."))
		if err = os.Chdir(root); err != nil {
			t.Fatalf("chdir to root: %v", err)
		}
		t.Cleanup(func() {
			_ = os.Chdir(wd)
		})

		var cfg Config
		if err = LoadConfig("", &cfg); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cfg.Server.Port == 0 || cfg.User.JWTSecret == "" {
			t.Fatalf("unexpected default config: %+v", cfg)
		}
	})

	t.Run("missing file", func(t *testing.T) {
		var cfg Config
		err := LoadConfig(filepath.Join(t.TempDir(), "missing.yaml"), &cfg)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
