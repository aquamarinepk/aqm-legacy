package aqm

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestWithLogger(t *testing.T) {
	tests := []struct {
		name      string
		logger    Logger
		expectErr bool
	}{
		{
			name:      "validLogger",
			logger:    NewNoopLogger(),
			expectErr: false,
		},
		{
			name:      "nilLogger",
			logger:    nil,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &Micro{deps: DefaultDeps()}
			opt := WithLogger(tt.logger)
			err := opt(ms)

			if tt.expectErr && err == nil {
				t.Error("expected error")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.expectErr && ms.deps.Logger != tt.logger {
				t.Error("logger not set")
			}
		})
	}
}

func TestWithTracer(t *testing.T) {
	tests := []struct {
		name   string
		tracer Tracer
	}{
		{
			name:   "validTracer",
			tracer: NoopTracer{},
		},
		{
			name:   "nilTracer",
			tracer: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &Micro{deps: DefaultDeps()}
			opt := WithTracer(tt.tracer)
			err := opt(ms)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if ms.deps.Tracer == nil {
				t.Error("tracer should not be nil")
			}
		})
	}
}

func TestWithMetrics(t *testing.T) {
	tests := []struct {
		name    string
		metrics Metrics
	}{
		{
			name:    "validMetrics",
			metrics: NoopMetrics{},
		},
		{
			name:    "nilMetrics",
			metrics: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &Micro{deps: DefaultDeps()}
			opt := WithMetrics(tt.metrics)
			err := opt(ms)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if ms.deps.Metrics == nil {
				t.Error("metrics should not be nil")
			}
		})
	}
}

func TestWithErrorReporter(t *testing.T) {
	tests := []struct {
		name     string
		reporter ErrorReporter
	}{
		{
			name:     "validReporter",
			reporter: NoopErrorReporter{},
		},
		{
			name:     "nilReporter",
			reporter: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &Micro{deps: DefaultDeps()}
			opt := WithErrorReporter(tt.reporter)
			err := opt(ms)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if ms.deps.Errors == nil {
				t.Error("error reporter should not be nil")
			}
		})
	}
}

func TestWithHealthChecks(t *testing.T) {
	tests := []struct {
		name      string
		checkName string
		checks    []HealthCheck
		expectErr bool
	}{
		{
			name:      "validCheck",
			checkName: "test",
			checks:    []HealthCheck{HealthStatusOK},
			expectErr: false,
		},
		{
			name:      "emptyName",
			checkName: "",
			checks:    []HealthCheck{HealthStatusOK},
			expectErr: true,
		},
		{
			name:      "nilChecks",
			checkName: "test",
			checks:    nil,
			expectErr: false,
		},
		{
			name:      "twoChecks",
			checkName: "test",
			checks:    []HealthCheck{HealthStatusOK, HealthStatusOK},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &Micro{deps: DefaultDeps()}
			opt := WithHealthChecks(tt.checkName, tt.checks...)
			err := opt(ms)

			if tt.expectErr && err == nil {
				t.Error("expected error")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestWithDebugRoutes(t *testing.T) {
	ms := &Micro{deps: DefaultDeps()}
	opt := WithDebugRoutes()
	err := opt(ms)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !ms.debugRoutes {
		t.Error("debugRoutes should be true")
	}
}

func TestWithLifecycle(t *testing.T) {
	ms := &Micro{deps: DefaultDeps()}

	startCalled := false
	stopCalled := false

	comp := LifecycleHooks{
		OnStart: func(ctx context.Context) error {
			startCalled = true
			return nil
		},
		OnStop: func(ctx context.Context) error {
			stopCalled = true
			return nil
		},
	}

	opt := WithLifecycle(comp)
	err := opt(ms)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Execute start functions
	for _, fn := range ms.startFuncs {
		fn(context.Background())
	}
	if !startCalled {
		t.Error("start should have been called")
	}

	// Execute stop functions
	for _, fn := range ms.stopFuncs {
		fn(context.Background())
	}
	if !stopCalled {
		t.Error("stop should have been called")
	}
}

func TestWithLifecycleNilComponent(t *testing.T) {
	ms := &Micro{deps: DefaultDeps()}
	opt := WithLifecycle(nil)
	err := opt(ms)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestWithConfig(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		expectErr bool
	}{
		{
			name:      "validConfig",
			config:    NewConfig(),
			expectErr: false,
		},
		{
			name:      "nilConfig",
			config:    nil,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &Micro{deps: DefaultDeps()}
			opt := WithConfig(tt.config)
			err := opt(ms)

			if tt.expectErr && err == nil {
				t.Error("expected error")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

type mockRunner struct {
	startCalled bool
	stopCalled  bool
}

func (m *mockRunner) Start(ctx context.Context) error {
	m.startCalled = true
	return nil
}

func (m *mockRunner) Stop(ctx context.Context) error {
	m.stopCalled = true
	return nil
}

func TestWithRunner(t *testing.T) {
	tests := []struct {
		name      string
		runner    Runner
		expectErr bool
	}{
		{
			name:      "validRunner",
			runner:    &mockRunner{},
			expectErr: false,
		},
		{
			name:      "nilRunner",
			runner:    nil,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &Micro{deps: DefaultDeps()}
			opt := WithRunner(tt.runner)
			err := opt(ms)

			if tt.expectErr && err == nil {
				t.Error("expected error")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestWithHTTPMiddleware(t *testing.T) {
	ms := &Micro{deps: DefaultDeps()}

	mw := func(next http.Handler) http.Handler {
		return next
	}

	opt := WithHTTPMiddleware(mw)
	err := opt(ms)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(ms.httpMiddlewares) != 1 {
		t.Error("middleware should be added")
	}
}

func TestWithRouterConfigurator(t *testing.T) {
	tests := []struct {
		name       string
		configurer func(*chi.Mux)
		expectErr  bool
	}{
		{
			name: "validConfigurer",
			configurer: func(r *chi.Mux) {
				r.Get("/custom", func(w http.ResponseWriter, req *http.Request) {})
			},
			expectErr: false,
		},
		{
			name:       "nilConfigurer",
			configurer: nil,
			expectErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &Micro{deps: DefaultDeps()}
			opt := WithRouterConfigurator(tt.configurer)
			err := opt(ms)

			if tt.expectErr && err == nil {
				t.Error("expected error")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestWithShutdown(t *testing.T) {
	tests := []struct {
		name      string
		fn        ShutdownFunc
		expectErr bool
	}{
		{
			name:      "validShutdown",
			fn:        func(ctx context.Context) error { return nil },
			expectErr: false,
		},
		{
			name:      "nilShutdown",
			fn:        nil,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &Micro{deps: DefaultDeps()}
			opt := WithShutdown(tt.fn)
			err := opt(ms)

			if tt.expectErr && err == nil {
				t.Error("expected error")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestWithDeps(t *testing.T) {
	tests := []struct {
		name       string
		configurer func(*Deps) error
		expectErr  bool
	}{
		{
			name: "validConfigurer",
			configurer: func(d *Deps) error {
				d.Metrics = NoopMetrics{}
				return nil
			},
			expectErr: false,
		},
		{
			name:       "nilConfigurer",
			configurer: nil,
			expectErr:  true,
		},
		{
			name: "errorConfigurer",
			configurer: func(d *Deps) error {
				return errors.New("config error")
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &Micro{deps: DefaultDeps()}
			opt := WithDeps(tt.configurer)
			err := opt(ms)

			if tt.expectErr && err == nil {
				t.Error("expected error")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
