package aqm

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
)

// HealthCheck represents a liveness or readiness probe.
type HealthCheck func(context.Context) error

// HealthChecks aggregates liveness and readiness probes.
type HealthChecks struct {
	Liveness  map[string]HealthCheck
	Readiness map[string]HealthCheck
}

// HealthReporter allows components to expose their health probes.
type HealthReporter interface {
	HealthChecks() HealthChecks
}

// HealthRegistry stores registered probes and exposes HTTP handlers.
type HealthRegistry struct {
	mu        sync.RWMutex
	liveness  map[string]HealthCheck
	readiness map[string]HealthCheck
}

// NewHealthRegistry constructs an empty registry.
func NewHealthRegistry() *HealthRegistry {
	return &HealthRegistry{
		liveness:  map[string]HealthCheck{},
		readiness: map[string]HealthCheck{},
	}
}

// RegisterChecks installs both liveness and readiness checks from the reporter.
func (hr *HealthRegistry) RegisterChecks(checks HealthChecks) {
	for name, check := range checks.Liveness {
		hr.RegisterLiveness(name, check)
	}
	for name, check := range checks.Readiness {
		hr.RegisterReadiness(name, check)
	}
}

// RegisterLiveness adds a liveness probe under the provided name.
func (hr *HealthRegistry) RegisterLiveness(name string, check HealthCheck) {
	if check == nil || name == "" {
		return
	}
	hr.mu.Lock()
	hr.liveness[name] = check
	hr.mu.Unlock()
}

// RegisterReadiness adds a readiness probe under the provided name.
func (hr *HealthRegistry) RegisterReadiness(name string, check HealthCheck) {
	if check == nil || name == "" {
		return
	}
	hr.mu.Lock()
	hr.readiness[name] = check
	hr.mu.Unlock()
}

// RegisterHealthEndpoints mounts default health endpoints into the router.
func RegisterHealthEndpoints(r chi.Router, registry *HealthRegistry) {
	if registry == nil {
		registry = NewHealthRegistry()
	}

	r.Get("/healthz", makeHealthHandler(registry, registry.liveness))
	r.Get("/livez", makeHealthHandler(registry, registry.liveness))
	r.Get("/readyz", makeHealthHandler(registry, registry.readiness))
	r.Get("/ping", pingHandler)
	r.Get("/metrics", notImplementedHandler)
	r.Get("/version", notImplementedHandler)
}

func pingHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("pong"))
}

func makeHealthHandler(registry *HealthRegistry, checks map[string]HealthCheck) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		registry.mu.RLock()
		snapshot := make(map[string]HealthCheck, len(checks))
		for name, check := range checks {
			snapshot[name] = check
		}
		registry.mu.RUnlock()

		summary := runChecks(r.Context(), snapshot)
		status := http.StatusOK
		for _, res := range summary.Results {
			if res.Error != "" {
				status = http.StatusServiceUnavailable
				break
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(summary)
	}
}

func runChecks(ctx context.Context, checks map[string]HealthCheck) ProbeResponse {
	results := make([]HealthResult, 0, len(checks))
	for name, check := range checks {
		result := HealthResult{Name: name}
		if check != nil {
			if err := check(ctx); err != nil {
				result.Error = err.Error()
			}
		}
		results = append(results, result)
	}

	status := "ok"
	for _, res := range results {
		if res.Error != "" {
			status = "degraded"
			break
		}
	}

	return ProbeResponse{
		Status:    status,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Results:   results,
	}
}

func notImplementedHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

// HealthStatusOK is a helper that always reports a healthy state.
func HealthStatusOK(context.Context) error { return nil }

// HealthResult captures the outcome of a single probe.
type HealthResult struct {
	Name  string `json:"name"`
	Error string `json:"error,omitempty"`
}

// ProbeResponse wraps probe results in a standard JSON envelope.
type ProbeResponse struct {
	Status    string         `json:"status"`
	Timestamp string         `json:"timestamp"`
	Results   []HealthResult `json:"results,omitempty"`
}
