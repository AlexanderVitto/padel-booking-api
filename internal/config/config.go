package config

import (
	"errors"
	"fmt"
	"os"
)

type Config struct {
	Env  string
	Port string
}

func Load() (Config, error) {
	cfg := Config{
		Env:  getEnv("ENV", "development"),
		Port: getEnv("PORT", "8080"),
	}
	if cfg.Port == "" {
		return Config{}, errors.New("PORT is required")
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
