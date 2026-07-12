package config_test

import (
	"testing"

	"github.com/gianluca-v/filas-backend/internal/config"
)

func TestLoad_ReadsEnvVars(t *testing.T) {
	t.Setenv("DB_HOST", "db")
	t.Setenv("DB_PORT", "3306")
	t.Setenv("DB_NAME", "filas")
	t.Setenv("DB_USER", "filas")
	t.Setenv("DB_PASSWORD", "filas_local_dev")
	t.Setenv("API_PORT", "9090")
	t.Setenv("JWT_SECRET", "super-secret")
	t.Setenv("JWT_EXPIRY_HOURS", "6")
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:5173,http://localhost:4173")

	cfg := config.Load()

	if cfg.DBHost != "db" {
		t.Errorf("DBHost = %q, want %q", cfg.DBHost, "db")
	}
	if cfg.APIPort != "9090" {
		t.Errorf("APIPort = %q, want %q", cfg.APIPort, "9090")
	}
	if cfg.JWTSecret != "super-secret" {
		t.Errorf("JWTSecret = %q, want %q", cfg.JWTSecret, "super-secret")
	}
	if cfg.JWTExpiryHours != 6 {
		t.Errorf("JWTExpiryHours = %d, want %d", cfg.JWTExpiryHours, 6)
	}
	wantOrigins := []string{"http://localhost:5173", "http://localhost:4173"}
	if len(cfg.CORSAllowedOrigins) != len(wantOrigins) {
		t.Fatalf("CORSAllowedOrigins = %v, want %v", cfg.CORSAllowedOrigins, wantOrigins)
	}
	for i, o := range wantOrigins {
		if cfg.CORSAllowedOrigins[i] != o {
			t.Errorf("CORSAllowedOrigins[%d] = %q, want %q", i, cfg.CORSAllowedOrigins[i], o)
		}
	}

	wantDSN := "filas:filas_local_dev@tcp(db:3306)/filas?parseTime=true&loc=Local"
	if cfg.DSN() != wantDSN {
		t.Errorf("DSN() = %q, want %q", cfg.DSN(), wantDSN)
	}
}

func TestLoad_DefaultsWhenMissing(t *testing.T) {
	// Explicitly clear vars this test cares about so defaults are exercised
	// even if a developer's shell happens to export them.
	t.Setenv("API_PORT", "")
	t.Setenv("JWT_EXPIRY_HOURS", "")

	cfg := config.Load()

	if cfg.APIPort != "8080" {
		t.Errorf("APIPort default = %q, want %q", cfg.APIPort, "8080")
	}
	if cfg.JWTExpiryHours != 3 {
		t.Errorf("JWTExpiryHours default = %d, want %d", cfg.JWTExpiryHours, 3)
	}
}
