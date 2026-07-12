// Package config loads runtime configuration from environment variables
// into a typed struct. All values are env-driven per the local-only
// deployment decision — no hardcoded hosts or secrets.
package config

import (
	"fmt"
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

// Load reads Config from environment variables, applying sane local
// defaults for optional values so `go run` works without a full .env.
func Load() Config {
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
	}
}

// DSN builds the MySQL data source name consumed by the sqlx/mysql driver.
func (c Config) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=Local",
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
