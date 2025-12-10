package aqm

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestNewHealthRegistry(t *testing.T) {
	hr := NewHealthRegistry()
	if hr == nil {
		t.Fatal("NewHealthRegistry returned nil")
	}
	if hr.liveness == nil {
		t.Error("liveness map should be initialized")
	}
	if hr.readiness == nil {
		t.Error("readiness map should be initialized")
	}
}

func TestHealthRegistryRegisterLiveness(t *testing.T) {
	hr := NewHealthRegistry()

	check := func(ctx context.Context) error { return nil }
	hr.RegisterLiveness("test", check)

	if len(hr.liveness) != 1 {
		t.Errorf("expected 1 liveness check, got %d", len(hr.liveness))
	}
}

func TestHealthRegistryRegisterLivenessEmptyName(t *testing.T) {
	hr := NewHealthRegistry()

	check := func(ctx context.Context) error { return nil }
	hr.RegisterLiveness("", check)

	if len(hr.liveness) != 0 {
		t.Error("should not register check with empty name")
	}
}

func TestHealthRegistryRegisterLivenessNilCheck(t *testing.T) {
	hr := NewHealthRegistry()
	hr.RegisterLiveness("test", nil)

	if len(hr.liveness) != 0 {
		t.Error("should not register nil check")
	}
}

func TestHealthRegistryRegisterReadiness(t *testing.T) {
	hr := NewHealthRegistry()

	check := func(ctx context.Context) error { return nil }
	hr.RegisterReadiness("test", check)

	if len(hr.readiness) != 1 {
		t.Errorf("expected 1 readiness check, got %d", len(hr.readiness))
	}
}

func TestHealthRegistryRegisterReadinessEmptyName(t *testing.T) {
	hr := NewHealthRegistry()
	hr.RegisterReadiness("", func(ctx context.Context) error { return nil })

	if len(hr.readiness) != 0 {
		t.Error("should not register check with empty name")
	}
}

func TestHealthRegistryRegisterReadinessNilCheck(t *testing.T) {
	hr := NewHealthRegistry()
	hr.RegisterReadiness("test", nil)

	if len(hr.readiness) != 0 {
		t.Error("should not register nil check")
	}
}

func TestHealthRegistryRegisterChecks(t *testing.T) {
	hr := NewHealthRegistry()

	checks := HealthChecks{
		Liveness: map[string]HealthCheck{
			"live1": func(ctx context.Context) error { return nil },
			"live2": func(ctx context.Context) error { return nil },
		},
		Readiness: map[string]HealthCheck{
			"ready1": func(ctx context.Context) error { return nil },
		},
	}

	hr.RegisterChecks(checks)

	if len(hr.liveness) != 2 {
		t.Errorf("expected 2 liveness checks, got %d", len(hr.liveness))
	}
	if len(hr.readiness) != 1 {
		t.Errorf("expected 1 readiness check, got %d", len(hr.readiness))
	}
}

func TestRegisterHealthEndpoints(t *testing.T) {
	r := chi.NewRouter()
	hr := NewHealthRegistry()
	hr.RegisterLiveness("core", HealthStatusOK)
	hr.RegisterReadiness("core", HealthStatusOK)

	RegisterHealthEndpoints(r, hr)

	endpoints := []string{"/healthz", "/livez", "/readyz", "/ping", "/metrics", "/version"}

	for _, ep := range endpoints {
		req := httptest.NewRequest(http.MethodGet, ep, nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code == http.StatusNotFound {
			t.Errorf("endpoint %s not registered", ep)
		}
	}
}

func TestRegisterHealthEndpointsNilRegistry(t *testing.T) {
	r := chi.NewRouter()
	RegisterHealthEndpoints(r, nil)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestHealthzEndpointOK(t *testing.T) {
	r := chi.NewRouter()
	hr := NewHealthRegistry()
	hr.RegisterLiveness("test", func(ctx context.Context) error { return nil })

	RegisterHealthEndpoints(r, hr)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp ProbeResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Status != "ok" {
		t.Errorf("expected status 'ok', got %q", resp.Status)
	}
}

func TestHealthzEndpointDegraded(t *testing.T) {
	r := chi.NewRouter()
	hr := NewHealthRegistry()
	hr.RegisterLiveness("failing", func(ctx context.Context) error {
		return errors.New("service unavailable")
	})

	RegisterHealthEndpoints(r, hr)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status %d, got %d", http.StatusServiceUnavailable, rec.Code)
	}

	var resp ProbeResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Status != "degraded" {
		t.Errorf("expected status 'degraded', got %q", resp.Status)
	}
}

func TestLivezEndpoint(t *testing.T) {
	r := chi.NewRouter()
	hr := NewHealthRegistry()
	hr.RegisterLiveness("core", HealthStatusOK)

	RegisterHealthEndpoints(r, hr)

	req := httptest.NewRequest(http.MethodGet, "/livez", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestReadyzEndpoint(t *testing.T) {
	r := chi.NewRouter()
	hr := NewHealthRegistry()
	hr.RegisterReadiness("db", func(ctx context.Context) error { return nil })

	RegisterHealthEndpoints(r, hr)

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestPingEndpoint(t *testing.T) {
	r := chi.NewRouter()
	RegisterHealthEndpoints(r, nil)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if rec.Body.String() != "pong" {
		t.Errorf("expected 'pong', got %q", rec.Body.String())
	}
}

func TestMetricsEndpoint(t *testing.T) {
	r := chi.NewRouter()
	RegisterHealthEndpoints(r, nil)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotImplemented {
		t.Errorf("expected status %d, got %d", http.StatusNotImplemented, rec.Code)
	}
}

func TestVersionEndpoint(t *testing.T) {
	r := chi.NewRouter()
	RegisterHealthEndpoints(r, nil)

	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotImplemented {
		t.Errorf("expected status %d, got %d", http.StatusNotImplemented, rec.Code)
	}
}

func TestHealthStatusOK(t *testing.T) {
	err := HealthStatusOK(context.Background())
	if err != nil {
		t.Errorf("HealthStatusOK should return nil, got %v", err)
	}
}

func TestRunChecksWithNilCheck(t *testing.T) {
	checks := map[string]HealthCheck{
		"nilcheck": nil,
	}

	resp := runChecks(context.Background(), checks)

	if resp.Status != "ok" {
		t.Errorf("expected status 'ok' for nil check, got %q", resp.Status)
	}
}

func TestHealthResultFields(t *testing.T) {
	result := HealthResult{
		Name:  "test",
		Error: "test error",
	}

	if result.Name != "test" {
		t.Errorf("expected Name 'test', got %q", result.Name)
	}
	if result.Error != "test error" {
		t.Errorf("expected Error 'test error', got %q", result.Error)
	}
}

func TestProbeResponseFields(t *testing.T) {
	resp := ProbeResponse{
		Status:    "ok",
		Timestamp: "2024-01-01T00:00:00Z",
		Results:   []HealthResult{{Name: "test"}},
	}

	if resp.Status != "ok" {
		t.Errorf("expected Status 'ok', got %q", resp.Status)
	}
	if len(resp.Results) != 1 {
		t.Errorf("expected 1 result, got %d", len(resp.Results))
	}
}

func TestHealthChecksStruct(t *testing.T) {
	hc := HealthChecks{
		Liveness:  make(map[string]HealthCheck),
		Readiness: make(map[string]HealthCheck),
	}

	if hc.Liveness == nil {
		t.Error("Liveness should be initialized")
	}
	if hc.Readiness == nil {
		t.Error("Readiness should be initialized")
	}
}
