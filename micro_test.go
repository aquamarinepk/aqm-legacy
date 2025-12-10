package aqm

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestNewMicro(t *testing.T) {
	cfg := NewConfig()
	logger := NewNoopLogger()

	ms := NewMicro(
		WithConfig(cfg),
		WithLogger(logger),
	)

	if ms == nil {
		t.Fatal("NewMicro returned nil")
	}
	if ms.deps == nil {
		t.Error("deps should not be nil")
	}
}

func TestNewMicroPanicsOnOptionError(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on option error")
		}
	}()

	_ = NewMicro(
		WithLogger(nil), // This should cause an error
	)
}

func TestNewMicroPanicsWithoutLogger(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic without logger")
		}
	}()

	_ = NewMicro(
		WithConfig(NewConfig()),
	)
}

func TestNewMicroPanicsWithoutConfig(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic without config")
		}
	}()

	_ = NewMicro(
		WithLogger(NewNoopLogger()),
	)
}

func TestMicroDeps(t *testing.T) {
	cfg := NewConfig()
	logger := NewNoopLogger()

	ms := NewMicro(
		WithConfig(cfg),
		WithLogger(logger),
	)

	deps := ms.Deps()
	if deps == nil {
		t.Error("Deps() returned nil")
	}
	if deps.Logger != logger {
		t.Error("Logger not set correctly")
	}
	if deps.Config != cfg {
		t.Error("Config not set correctly")
	}
}

func TestMicroRun(t *testing.T) {
	cfg := NewConfig()
	logger := NewNoopLogger()

	startCalled := false
	stopCalled := false
	shutdownCalled := false

	ms := NewMicro(
		WithConfig(cfg),
		WithLogger(logger),
		WithLifecycle(LifecycleHooks{
			OnStart: func(ctx context.Context) error {
				startCalled = true
				return nil
			},
			OnStop: func(ctx context.Context) error {
				stopCalled = true
				return nil
			},
		}),
		WithShutdown(func(ctx context.Context) error {
			shutdownCalled = true
			return nil
		}),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := ms.Run(ctx)
	// err will be nil since context times out gracefully

	if !startCalled {
		t.Error("start should have been called")
	}
	if !stopCalled {
		t.Error("stop should have been called")
	}
	if !shutdownCalled {
		t.Error("shutdown should have been called")
	}
	_ = err // context cancellation is expected
}

func TestMicroRunStartError(t *testing.T) {
	cfg := NewConfig()
	logger := NewNoopLogger()

	ms := NewMicro(
		WithConfig(cfg),
		WithLogger(logger),
		WithLifecycle(LifecycleHooks{
			OnStart: func(ctx context.Context) error {
				return errors.New("start failed")
			},
		}),
	)

	ctx := context.Background()
	err := ms.Run(ctx)

	if err == nil {
		t.Error("expected error from failed start")
	}
}

func TestMicroRunStartErrorWithRollback(t *testing.T) {
	cfg := NewConfig()
	logger := NewNoopLogger()

	rollbackCalled := false

	ms := NewMicro(
		WithConfig(cfg),
		WithLogger(logger),
	)

	// Manually add start and stop functions to test rollback
	ms.startFuncs = []func(context.Context) error{
		func(ctx context.Context) error { return nil },
		func(ctx context.Context) error { return errors.New("start failed") },
	}
	ms.stopFuncs = []func(context.Context) error{
		func(ctx context.Context) error {
			rollbackCalled = true
			return nil
		},
		func(ctx context.Context) error { return nil },
	}

	ctx := context.Background()
	err := ms.Run(ctx)

	if err == nil {
		t.Error("expected error from failed start")
	}
	if !rollbackCalled {
		t.Error("rollback should have been called")
	}
}

func TestMicroRunWithRunner(t *testing.T) {
	cfg := NewConfig()
	logger := NewNoopLogger()

	runner := &mockRunner{}

	ms := NewMicro(
		WithConfig(cfg),
		WithLogger(logger),
		WithRunner(runner),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_ = ms.Run(ctx)

	if !runner.startCalled {
		t.Error("runner.Start should have been called")
	}
	if !runner.stopCalled {
		t.Error("runner.Stop should have been called")
	}
}

func TestMicroRunnerStartError(t *testing.T) {
	cfg := NewConfig()
	logger := NewNoopLogger()

	ms := NewMicro(
		WithConfig(cfg),
		WithLogger(logger),
	)

	ms.runners = []Runner{
		&errorRunner{startErr: errors.New("runner start failed")},
	}

	ctx := context.Background()
	err := ms.Run(ctx)

	if err == nil {
		t.Error("expected error from failed runner start")
	}
}

type errorRunner struct {
	startErr error
	stopErr  error
}

func (r *errorRunner) Start(ctx context.Context) error { return r.startErr }
func (r *errorRunner) Stop(ctx context.Context) error  { return r.stopErr }

func TestMicroAddRunner(t *testing.T) {
	ms := &Micro{deps: DefaultDeps()}
	runner := &mockRunner{}

	ms.addRunner(runner)

	if len(ms.runners) != 1 {
		t.Errorf("expected 1 runner, got %d", len(ms.runners))
	}
}

func TestMicroAddShutdown(t *testing.T) {
	ms := &Micro{deps: DefaultDeps()}
	fn := func(ctx context.Context) error { return nil }

	ms.addShutdown(fn)

	if len(ms.shutdown) != 1 {
		t.Errorf("expected 1 shutdown hook, got %d", len(ms.shutdown))
	}
}

func TestMicroAddHealthCheck(t *testing.T) {
	ms := &Micro{deps: DefaultDeps()}
	reg := healthCheckRegistration{
		name:      "test",
		liveness:  HealthStatusOK,
		readiness: HealthStatusOK,
	}

	ms.addHealthCheck(reg)

	if len(ms.healthChecks) != 1 {
		t.Errorf("expected 1 health check, got %d", len(ms.healthChecks))
	}
}

func TestMicroAddStart(t *testing.T) {
	ms := &Micro{deps: DefaultDeps()}

	ms.addStart(func(ctx context.Context) error { return nil })
	ms.addStart(nil) // should be ignored

	if len(ms.startFuncs) != 1 {
		t.Errorf("expected 1 start func, got %d", len(ms.startFuncs))
	}
}

func TestMicroAddStop(t *testing.T) {
	ms := &Micro{deps: DefaultDeps()}

	ms.addStop(func(ctx context.Context) error { return nil })
	ms.addStop(nil) // should be ignored

	if len(ms.stopFuncs) != 1 {
		t.Errorf("expected 1 stop func, got %d", len(ms.stopFuncs))
	}
}

func TestShutdownFuncType(t *testing.T) {
	var fn ShutdownFunc = func(ctx context.Context) error {
		return nil
	}

	err := fn(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestHealthCheckRegistrationFields(t *testing.T) {
	reg := healthCheckRegistration{
		name:      "test",
		liveness:  HealthStatusOK,
		readiness: HealthStatusOK,
	}

	if reg.name != "test" {
		t.Error("name not set")
	}
	if reg.liveness == nil {
		t.Error("liveness not set")
	}
	if reg.readiness == nil {
		t.Error("readiness not set")
	}
}
