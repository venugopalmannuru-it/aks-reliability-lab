package config

import (
	"testing"
	"time"
)

func TestFromEnvDefaults(t *testing.T) {
	for _, key := range []string{"PORT", "APP_ENV", "LOG_LEVEL", "SHUTDOWN_DELAY", "SHUTDOWN_TIMEOUT"} {
		t.Setenv(key, "")
	}
	cfg, err := FromEnv()
	if err != nil {
		t.Fatalf("FromEnv() error = %v", err)
	}
	if cfg.Port != "8080" || cfg.Environment != "development" {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
	if cfg.ShutdownDelay != 2*time.Second || cfg.ShutdownTimeout != 10*time.Second {
		t.Fatalf("unexpected shutdown defaults: %+v", cfg)
	}
}

func TestFromEnvRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		key, value string
	}{
		{"PORT", "70000"},
		{"LOG_LEVEL", "verbose"},
		{"SHUTDOWN_DELAY", "-1s"},
		{"SHUTDOWN_TIMEOUT", "0s"},
	}
	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			for _, key := range []string{"PORT", "APP_ENV", "LOG_LEVEL", "SHUTDOWN_DELAY", "SHUTDOWN_TIMEOUT"} {
				t.Setenv(key, "")
			}
			t.Setenv(tt.key, tt.value)
			if _, err := FromEnv(); err == nil {
				t.Fatalf("FromEnv() accepted %s=%q", tt.key, tt.value)
			}
		})
	}
}
