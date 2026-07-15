package config_test

import (
	"bytes"
	"log"
	"os"
	"strings"
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

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

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

	wantDSN := "filas:filas_local_dev@tcp(db:3306)/filas?parseTime=true&loc=Local&timeout=5s&readTimeout=5s&writeTimeout=5s"
	if cfg.DSN() != wantDSN {
		t.Errorf("DSN() = %q, want %q", cfg.DSN(), wantDSN)
	}
}

func TestLoad_DefaultsWhenMissing(t *testing.T) {
	setRequiredEnv(t)
	// Explicitly clear vars this test cares about so defaults are exercised
	// even if a developer's shell happens to export them.
	t.Setenv("API_PORT", "")
	t.Setenv("JWT_EXPIRY_HOURS", "")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	if cfg.APIPort != "8080" {
		t.Errorf("APIPort default = %q, want %q", cfg.APIPort, "8080")
	}
	if cfg.JWTExpiryHours != 3 {
		t.Errorf("JWTExpiryHours default = %d, want %d", cfg.JWTExpiryHours, 3)
	}
}

func TestLoad_ReturnsErrorListingAllMissingRequiredVars(t *testing.T) {
	// Clear every required var regardless of ambient shell environment.
	for _, k := range []string{"DB_HOST", "DB_NAME", "DB_USER", "DB_PASSWORD", "JWT_SECRET"} {
		t.Setenv(k, "")
	}

	_, err := config.Load()
	if err == nil {
		t.Fatal("Load() error = nil, want error listing missing required vars")
	}

	for _, want := range []string{"DB_HOST", "DB_NAME", "DB_USER", "DB_PASSWORD", "JWT_SECRET"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("Load() error = %q, want it to name missing var %q", err.Error(), want)
		}
	}
}

func TestLoad_ReturnsErrorListingOnlyTheMissingSubset(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("JWT_SECRET", "")

	_, err := config.Load()
	if err == nil {
		t.Fatal("Load() error = nil, want error naming JWT_SECRET")
	}
	if !strings.Contains(err.Error(), "JWT_SECRET") {
		t.Errorf("Load() error = %q, want it to name JWT_SECRET", err.Error())
	}
	if strings.Contains(err.Error(), "DB_HOST") {
		t.Errorf("Load() error = %q, must not name DB_HOST which was set", err.Error())
	}
}

func TestLoad_WarnsOnMalformedJWTExpiryHoursAndFallsBackToDefault(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("JWT_EXPIRY_HOURS", "not-a-number")

	var buf bytes.Buffer
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil (malformed optional var should warn, not fail)", err)
	}
	if cfg.JWTExpiryHours != 3 {
		t.Errorf("JWTExpiryHours = %d, want fallback default %d", cfg.JWTExpiryHours, 3)
	}
	if !strings.Contains(buf.String(), "JWT_EXPIRY_HOURS") {
		t.Errorf("log output = %q, want a warning naming JWT_EXPIRY_HOURS", buf.String())
	}
}

// setRequiredEnv sets every required config var to a valid value so tests
// that only care about ONE specific var can isolate that var's behavior.
func setRequiredEnv(t *testing.T) {
	t.Helper()
	t.Setenv("DB_HOST", "db")
	t.Setenv("DB_NAME", "filas")
	t.Setenv("DB_USER", "filas")
	t.Setenv("DB_PASSWORD", "filas_local_dev")
	t.Setenv("JWT_SECRET", "super-secret")
}
