package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Env              string
	Port             string
	DatabaseURL      string
	JWTAccessSecret  string
	JWTRefreshSecret string
	JWTAccessTTL     time.Duration
	JWTRefreshTTL    time.Duration
}

func Load() (Config, error) {
	accessTTL, err := parseDurationMinutes("JWT_ACCESS_TTL_MINUTES", 15)
	if err != nil {
		return Config{}, err
	}

	refreshTTL, err := parseDurationDays("JWT_REFRESH_TTL_DAYS", 30)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		Env:              getEnv("ENV", "development"),
		Port:             getEnv("PORT", "8080"),
		DatabaseURL:      getEnv("DATABASE_URL", ""),
		JWTAccessSecret:  getEnv("JWT_ACCESS_SECRET", ""),
		JWTRefreshSecret: getEnv("JWT_REFRESH_SECRET", ""),
		JWTAccessTTL:     accessTTL,
		JWTRefreshTTL:    refreshTTL,
	}

	if cfg.Port == "" {
		return Config{}, errors.New("PORT is required")
	}
	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	if cfg.JWTAccessSecret == "" {
		return Config{}, errors.New("JWT_ACCESS_SECRET is required")
	}
	if cfg.JWTRefreshSecret == "" {
		return Config{}, errors.New("JWT_REFRESH_SECRET is required")
	}

	return cfg, nil
}

func (c Config) Addr() string {
	return fmt.Sprintf(":%s", c.Port)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}

// parseDurationMinutes membaca env variable sebagai menit lalu ubah ke time.Duration.
// Contoh: JWT_ACCESS_TTL_MINUTES=15 → 15 * time.Minute
func parseDurationMinutes(key string, fallback int) (time.Duration, error) {
	v := os.Getenv(key)
	if v == "" {
		return time.Duration(fallback) * time.Minute, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("%s must be a number, got %q", key, v)
	}
	return time.Duration(n) * time.Minute, nil
}

// parseDurationDays membaca env variable sebagai hari lalu ubah ke time.Duration.
// Contoh: JWT_REFRESH_TTL_DAYS=30 → 30 * 24 * time.Hour
func parseDurationDays(key string, fallback int) (time.Duration, error) {
	v := os.Getenv(key)
	if v == "" {
		return time.Duration(fallback) * 24 * time.Hour, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("%s must be a number, got %q", key, v)
	}
	return time.Duration(n) * 24 * time.Hour, nil
}
