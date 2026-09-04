package app

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Metadata struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Commit      string `json:"commit"`
	BuildDate   string `json:"build_date"`
	Environment string `json:"environment"`
}

type Service struct {
	metadata Metadata
	logger   *slog.Logger
	ready    atomic.Bool
	metrics  *metrics
}

type metrics struct {
	mu       sync.RWMutex
	requests map[string]uint64
	duration map[string]float64
}

func New(metadata Metadata, logger *slog.Logger) *Service {
	s := &Service{
		metadata: metadata,
		logger:   logger,
		metrics: &metrics{
			requests: make(map[string]uint64),
			duration: make(map[string]float64),
		},
	}
	s.ready.Store(true)
	return s
}

func (s *Service) SetReady(ready bool) { s.ready.Store(ready) }

func (s *Service) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /", s.instrument("root", http.HandlerFunc(s.root)))
	mux.Handle("GET /api/v1/info", s.instrument("info", http.HandlerFunc(s.info)))
	mux.Handle("GET /healthz", s.instrument("healthz", http.HandlerFunc(s.health)))
	mux.Handle("GET /readyz", s.instrument("readyz", http.HandlerFunc(s.readiness)))
	mux.Handle("GET /metrics", s.instrument("metrics", http.HandlerFunc(s.prometheus)))
	return s.securityHeaders(mux)
}

func (s *Service) root(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"service": s.metadata.Name,
		"message": "reliability engineering demo API",
	})
}

func (s *Service) info(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.metadata)
}

func (s *Service) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "alive"})
}

func (s *Service) readiness(w http.ResponseWriter, _ *http.Request) {
	if !s.ready.Load() {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not_ready"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func (s *Service) prometheus(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	fmt.Fprintln(w, "# HELP app_info Build and runtime metadata for the service.")
	fmt.Fprintln(w, "# TYPE app_info gauge")
	fmt.Fprintf(w, "app_info{version=%q,environment=%q} 1\n", escapeMetric(s.metadata.Version), escapeMetric(s.metadata.Environment))
	fmt.Fprintln(w, "# HELP http_requests_total Total HTTP requests handled.")
	fmt.Fprintln(w, "# TYPE http_requests_total counter")
	fmt.Fprintln(w, "# HELP http_request_duration_seconds_total Total request handling time.")
	fmt.Fprintln(w, "# TYPE http_request_duration_seconds_total counter")

	s.metrics.mu.RLock()
	defer s.metrics.mu.RUnlock()
	for key, count := range s.metrics.requests {
		parts := strings.Split(key, "|")
		fmt.Fprintf(w, "http_requests_total{route=%q,status=%q} %d\n", parts[0], parts[1], count)
		fmt.Fprintf(w, "http_request_duration_seconds_total{route=%q,status=%q} %.6f\n", parts[0], parts[1], s.metrics.duration[key])
	}
}

func (s *Service) instrument(route string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		requestID := requestID(r.Header.Get("X-Request-ID"))
		w.Header().Set("X-Request-ID", requestID)
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		duration := time.Since(started)
		key := route + "|" + strconv.Itoa(rw.status)
		s.metrics.mu.Lock()
		s.metrics.requests[key]++
		s.metrics.duration[key] += duration.Seconds()
		s.metrics.mu.Unlock()
		s.logger.Info("http_request",
			"request_id", requestID,
			"method", r.Method,
			"route", route,
			"status", rw.status,
			"duration_ms", duration.Milliseconds(),
		)
	})
}

func (s *Service) securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (w *responseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func requestID(candidate string) string {
	if len(candidate) > 0 && len(candidate) <= 64 {
		valid := true
		for _, r := range candidate {
			if !(r == '-' || r == '_' || r == '.' || r >= '0' && r <= '9' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z') {
				valid = false
				break
			}
		}
		if valid {
			return candidate
		}
	}
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 36)
	}
	return hex.EncodeToString(b)
}

func escapeMetric(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `"`, `\"`)
	return strings.ReplaceAll(value, "\n", `\n`)
}
