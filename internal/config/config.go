package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port            string
	Environment     string
	LogLevel        slog.Level
	ShutdownDelay   time.Duration
	ShutdownTimeout time.Duration
}

func FromEnv() (Config, error) {
	cfg := Config{
		Port:            envOr("PORT", "8080"),
		Environment:     envOr("APP_ENV", "development"),
		ShutdownDelay:   2 * time.Second,
		ShutdownTimeout: 10 * time.Second,
	}

	port, err := strconv.Atoi(cfg.Port)
	if err != nil || port < 1 || port > 65535 {
		return Config{}, fmt.Errorf("PORT must be an integer from 1 to 65535")
	}

	if err := cfg.LogLevel.UnmarshalText([]byte(envOr("LOG_LEVEL", "INFO"))); err != nil {
		return Config{}, fmt.Errorf("LOG_LEVEL: %w", err)
	}

	if value := os.Getenv("SHUTDOWN_DELAY"); value != "" {
		cfg.ShutdownDelay, err = time.ParseDuration(value)
		if err != nil || cfg.ShutdownDelay < 0 {
			return Config{}, fmt.Errorf("SHUTDOWN_DELAY must be a non-negative duration")
		}
	}

	if value := os.Getenv("SHUTDOWN_TIMEOUT"); value != "" {
		cfg.ShutdownTimeout, err = time.ParseDuration(value)
		if err != nil || cfg.ShutdownTimeout <= 0 {
			return Config{}, fmt.Errorf("SHUTDOWN_TIMEOUT must be a positive duration")
		}
	}

	return cfg, nil
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
