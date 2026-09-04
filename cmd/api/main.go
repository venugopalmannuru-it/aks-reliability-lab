package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/venugopalmannuru-it/aks-reliability-lab/internal/app"
	"github.com/venugopalmannuru-it/aks-reliability-lab/internal/config"
)

var (
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

func main() {
	healthcheck := flag.Bool("healthcheck", false, "check the running service health endpoint")
	flag.Parse()

	if *healthcheck {
		if err := checkHealth(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	cfg, err := config.FromEnv()
	if err != nil {
		fmt.Fprintln(os.Stderr, "invalid configuration:", err)
		os.Exit(2)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))
	service := app.New(app.Metadata{
		Name:        "aks-reliability-lab",
		Version:     version,
		Commit:      commit,
		BuildDate:   buildDate,
		Environment: cfg.Environment,
	}, logger)

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           service.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("server_started", "address", server.Addr, "environment", cfg.Environment, "version", version)
		errCh <- server.ListenAndServe()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		logger.Info("shutdown_requested", "signal", sig.String())
	case serveErr := <-errCh:
		if serveErr != nil && serveErr != http.ErrServerClosed {
			logger.Error("server_failed", "error", serveErr)
			os.Exit(1)
		}
		return
	}

	service.SetReady(false)
	time.Sleep(cfg.ShutdownDelay)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("shutdown_failed", "error", err)
		os.Exit(1)
	}
	logger.Info("server_stopped")
}

func checkHealth() error {
	url := os.Getenv("HEALTHCHECK_URL")
	if url == "" {
		url = "http://127.0.0.1:8080/healthz"
	}
	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("health request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health endpoint returned %s", resp.Status)
	}
	return nil
}
