package aqm

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestLifecycleHooksStart(t *testing.T) {
	tests := []struct {
		name      string
		onStart   func(context.Context) error
		expectErr bool
	}{
		{
			name:      "nilOnStart",
			onStart:   nil,
			expectErr: false,
		},
		{
			name:      "successfulOnStart",
			onStart:   func(ctx context.Context) error { return nil },
			expectErr: false,
		},
		{
			name:      "failingOnStart",
			onStart:   func(ctx context.Context) error { return errors.New("start failed") },
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hooks := LifecycleHooks{OnStart: tt.onStart}
			err := hooks.Start(context.Background())

			if tt.expectErr && err == nil {
				t.Error("expected error")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestLifecycleHooksStop(t *testing.T) {
	tests := []struct {
		name      string
		onStop    func(context.Context) error
		expectErr bool
	}{
		{
			name:      "nilOnStop",
			onStop:    nil,
			expectErr: false,
		},
		{
			name:      "successfulOnStop",
			onStop:    func(ctx context.Context) error { return nil },
			expectErr: false,
		},
		{
			name:      "failingOnStop",
			onStop:    func(ctx context.Context) error { return errors.New("stop failed") },
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hooks := LifecycleHooks{OnStop: tt.onStop}
			err := hooks.Stop(context.Background())

			if tt.expectErr && err == nil {
				t.Error("expected error")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

type mockComponent struct {
	startCalled bool
	stopCalled  bool
	startErr    error
	stopErr     error
}

func (m *mockComponent) Start(ctx context.Context) error {
	m.startCalled = true
	return m.startErr
}

func (m *mockComponent) Stop(ctx context.Context) error {
	m.stopCalled = true
	return m.stopErr
}

type mockRouteRegistrar struct {
	registerCalled bool
}

func (m *mockRouteRegistrar) RegisterRoutes(r chi.Router) {
	m.registerCalled = true
	r.Get("/mock", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

type mockHealthReporter struct{}

func (m *mockHealthReporter) HealthChecks() HealthChecks {
	return HealthChecks{
		Liveness: map[string]HealthCheck{
			"mock": func(ctx context.Context) error { return nil },
		},
	}
}

func TestSetup(t *testing.T) {
	r := chi.NewRouter()
	comp := &mockComponent{}
	rr := &mockRouteRegistrar{}

	starts, stops, health := Setup(context.Background(), r, comp, rr)

	if len(starts) != 1 {
		t.Errorf("expected 1 start func, got %d", len(starts))
	}
	if len(stops) != 1 {
		t.Errorf("expected 1 stop func, got %d", len(stops))
	}
	if health == nil {
		t.Error("expected health registry")
	}
	if !rr.registerCalled {
		t.Error("expected RegisterRoutes to be called")
	}
}

func TestSetupWithHealthReporter(t *testing.T) {
	r := chi.NewRouter()
	hr := &mockHealthReporter{}

	_, _, health := Setup(context.Background(), r, hr)

	if health == nil {
		t.Error("expected health registry")
	}
}

func TestStart(t *testing.T) {
	starts := []func(context.Context) error{
		func(ctx context.Context) error { return nil },
		func(ctx context.Context) error { return nil },
	}
	stops := []func(context.Context) error{
		func(ctx context.Context) error { return nil },
		func(ctx context.Context) error { return nil },
	}

	err := Start(context.Background(), nil, starts, stops)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestStartWithFailure(t *testing.T) {
	var rollbackCalled bool
	starts := []func(context.Context) error{
		func(ctx context.Context) error { return nil },
		func(ctx context.Context) error { return errors.New("start failed") },
	}
	stops := []func(context.Context) error{
		func(ctx context.Context) error {
			rollbackCalled = true
			return nil
		},
		func(ctx context.Context) error { return nil },
	}

	err := Start(context.Background(), nil, starts, stops)
	if err == nil {
		t.Error("expected error")
	}
	if !rollbackCalled {
		t.Error("expected rollback to be called")
	}
}

func TestStartWithRollbackError(t *testing.T) {
	starts := []func(context.Context) error{
		func(ctx context.Context) error { return nil },
		func(ctx context.Context) error { return errors.New("start failed") },
	}
	stops := []func(context.Context) error{
		func(ctx context.Context) error { return errors.New("rollback failed") },
		func(ctx context.Context) error { return nil },
	}

	logger := NewNoopLogger()
	err := Start(context.Background(), logger, starts, stops)
	if err == nil {
		t.Error("expected error")
	}
}

func TestShutdown(t *testing.T) {
	var stopCalled bool
	stops := []func(context.Context) error{
		func(ctx context.Context) error {
			stopCalled = true
			return nil
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()

	// Create a custom server with a specific listener
	testSrv := &http.Server{}

	logger := NewNoopLogger()
	Shutdown(context.Background(), testSrv, logger, stops)

	if !stopCalled {
		t.Error("expected stop to be called")
	}
}

func TestShutdownNilServer(t *testing.T) {
	var stopCalled bool
	stops := []func(context.Context) error{
		func(ctx context.Context) error {
			stopCalled = true
			return nil
		},
	}

	logger := NewNoopLogger()
	Shutdown(context.Background(), nil, logger, stops)

	if !stopCalled {
		t.Error("expected stop to be called")
	}
}

func TestShutdownNilContext(t *testing.T) {
	stops := []func(context.Context) error{}
	logger := NewNoopLogger()

	// should not panic with nil context
	Shutdown(nil, nil, logger, stops)
}

func TestShutdownWithStopError(t *testing.T) {
	stops := []func(context.Context) error{
		func(ctx context.Context) error { return errors.New("stop error") },
	}

	logger := NewNoopLogger()
	Shutdown(context.Background(), nil, logger, stops)
	// should not panic
}

func TestStartableInterface(t *testing.T) {
	var s Startable = &mockComponent{}
	err := s.Start(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestStoppableInterface(t *testing.T) {
	var s Stoppable = &mockComponent{}
	err := s.Stop(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRouteRegistrarInterface(t *testing.T) {
	var rr RouteRegistrar = &mockRouteRegistrar{}
	r := chi.NewRouter()
	rr.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/mock", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}
