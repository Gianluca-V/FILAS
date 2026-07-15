// Package config loads runtime configuration from environment variables
// into a typed struct. All values are env-driven per the local-only
// deployment decision — no hardcoded hosts or secrets.
package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

// Config holds all environment-driven settings for the API service.
type Config struct {
	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string

	APIPort string

	JWTSecret      string
	JWTExpiryHours int

	CORSAllowedOrigins []string
}

// requiredVars is checked, IN ORDER, by Load. DB_PORT is intentionally
// excluded: it has a safe local default (3306) and is not a fatal startup
// condition when absent.
var requiredVars = []string{"DB_HOST", "DB_NAME", "DB_USER", "DB_PASSWORD", "JWT_SECRET"}

// MissingVarsError is returned by Load when one or more required
// environment variables are unset, naming every offender at once so a
// developer fixes them all in one pass instead of one-at-a-time.
type MissingVarsError struct {
	Vars []string
}

func (e *MissingVarsError) Error() string {
	return fmt.Sprintf("config: missing required environment variables: %s", strings.Join(e.Vars, ", "))
}

// Load reads Config from environment variables, applying sane local
// defaults for optional values so `go run` works without a full .env.
// It fails fast with a *MissingVarsError listing every missing required
// variable; malformed optional variables log a warning and fall back to
// their default instead of failing the whole process.
func Load() (Config, error) {
	var missing []string
	for _, name := range requiredVars {
		if os.Getenv(name) == "" {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		return Config{}, &MissingVarsError{Vars: missing}
	}

	return Config{
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     getEnvOrDefault("DB_PORT", "3306"),
		DBName:     os.Getenv("DB_NAME"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),

		APIPort: getEnvOrDefault("API_PORT", "8080"),

		JWTSecret:      os.Getenv("JWT_SECRET"),
		JWTExpiryHours: getEnvIntOrDefault("JWT_EXPIRY_HOURS", 3),

		CORSAllowedOrigins: splitAndTrim(os.Getenv("CORS_ALLOWED_ORIGINS")),
	}, nil
}

// DSN builds the MySQL data source name consumed by the sqlx/mysql driver.
// Modest local timeouts prevent a hung connection attempt from blocking
// the process indefinitely (fix #10 from the PR1 gate review).
func (c Config) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=Local&timeout=5s&readTimeout=5s&writeTimeout=5s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}

func getEnvOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvIntOrDefault(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		log.Printf("config: %s=%q is not a valid integer, falling back to default %d", key, v, fallback)
		return fallback
	}
	return n
}

func splitAndTrim(csv string) []string {
	if csv == "" {
		return nil
	}
	parts := strings.Split(csv, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
