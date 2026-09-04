package app

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testService() *Service {
	return New(Metadata{Name: "test-service", Version: "1.2.3", Environment: "test"}, slog.New(slog.NewJSONHandler(io.Discard, nil)))
}

func TestHealthAndReadiness(t *testing.T) {
	s := testService()
	tests := []struct {
		path string
		code int
		body string
	}{
		{"/healthz", http.StatusOK, `"alive"`},
		{"/readyz", http.StatusOK, `"ready"`},
		{"/api/v1/info", http.StatusOK, `"version":"1.2.3"`},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			rr := httptest.NewRecorder()
			s.Handler().ServeHTTP(rr, httptest.NewRequest(http.MethodGet, tt.path, nil))
			if rr.Code != tt.code || !strings.Contains(rr.Body.String(), tt.body) {
				t.Fatalf("response = %d %q", rr.Code, rr.Body.String())
			}
		})
	}

	s.SetReady(false)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("not-ready response = %d", rr.Code)
	}
}

func TestMetricsAndRequestID(t *testing.T) {
	s := testService()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("X-Request-ID", "trace-123")
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)
	if rr.Header().Get("X-Request-ID") != "trace-123" {
		t.Fatalf("request ID was not propagated")
	}

	rr = httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if !strings.Contains(rr.Body.String(), `http_requests_total{route="healthz",status="200"} 1`) {
		t.Fatalf("metrics missing request counter:\n%s", rr.Body.String())
	}
	if got := rr.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/plain") {
		t.Fatalf("metrics content type = %q", got)
	}
}

func TestInvalidRequestIDIsReplaced(t *testing.T) {
	s := New(Metadata{Name: "test"}, slog.New(slog.NewJSONHandler(&bytes.Buffer{}, nil)))
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("X-Request-ID", "bad id with spaces")
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, req)
	if got := rr.Header().Get("X-Request-ID"); got == "" || got == "bad id with spaces" {
		t.Fatalf("unexpected generated request ID %q", got)
	}
}
