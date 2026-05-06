package configenv

import (
	"os"
	"path/filepath"
	"testing"
)

type loadConfig struct {
	Server struct {
		Port int `mapstructure:"port"`
	} `mapstructure:"server"`
	Auth struct {
		JWTSecret string `mapstructure:"jwt_secret"`
	} `mapstructure:"auth"`
}

func TestLoadWithBindings(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	configBody := []byte("server:\n  port: 8080\nauth:\n  jwt_secret: file-secret\n")
	if err := os.WriteFile(path, configBody, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	t.Setenv("AUTH_JWT_SECRET", "env-secret")

	var cfg loadConfig
	err := Load(path, "unused", &cfg, map[string]string{
		"auth.jwt_secret": "AUTH_JWT_SECRET",
	})
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Server.Port != 8080 {
		t.Fatalf("port = %d, want 8080", cfg.Server.Port)
	}

	if cfg.Auth.JWTSecret != "env-secret" {
		t.Fatalf("jwt_secret = %q, want env-secret", cfg.Auth.JWTSecret)
	}
}
