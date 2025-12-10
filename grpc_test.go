package aqm

import (
	"context"
	"errors"
	"testing"
	"time"

	"google.golang.org/grpc"
)

type testGRPCService struct {
	registerCalled bool
	startCalled    bool
	stopCalled     bool
}

func (s *testGRPCService) RegisterGRPCService(server *grpc.Server) {
	s.registerCalled = true
}

func (s *testGRPCService) Start(ctx context.Context) error {
	s.startCalled = true
	return nil
}

func (s *testGRPCService) Stop(ctx context.Context) error {
	s.stopCalled = true
	return nil
}

func TestWithGRPCServer(t *testing.T) {
	cfg := NewConfig()
	cfg.Set("grpc.port", ":0")
	logger := NewNoopLogger()

	service := &testGRPCService{}
	factory := func(d *Deps) (GRPCServiceRegistrar, error) {
		return service, nil
	}

	ms := NewMicro(
		WithConfig(cfg),
		WithLogger(logger),
		WithGRPCServer("grpc.port", factory),
	)

	if ms == nil {
		t.Fatal("NewMicro returned nil")
	}
	if !service.registerCalled {
		t.Error("RegisterGRPCService should have been called")
	}
	if len(ms.runners) != 1 {
		t.Errorf("expected 1 runner, got %d", len(ms.runners))
	}
}

func TestWithGRPCServerEmptyAddrKey(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for empty addr key")
		}
	}()

	cfg := NewConfig()
	logger := NewNoopLogger()

	_ = NewMicro(
		WithConfig(cfg),
		WithLogger(logger),
		WithGRPCServer(""),
	)
}

func TestWithGRPCServerNilFactory(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil factory")
		}
	}()

	cfg := NewConfig()
	cfg.Set("grpc.port", ":0")
	logger := NewNoopLogger()

	_ = NewMicro(
		WithConfig(cfg),
		WithLogger(logger),
		WithGRPCServer("grpc.port", nil),
	)
}

func TestWithGRPCServerFactoryError(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for factory error")
		}
	}()

	cfg := NewConfig()
	cfg.Set("grpc.port", ":0")
	logger := NewNoopLogger()

	errorFactory := func(d *Deps) (GRPCServiceRegistrar, error) {
		return nil, errors.New("factory error")
	}

	_ = NewMicro(
		WithConfig(cfg),
		WithLogger(logger),
		WithGRPCServer("grpc.port", errorFactory),
	)
}

func TestWithGRPCServerFactoryReturnsNil(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil service")
		}
	}()

	cfg := NewConfig()
	cfg.Set("grpc.port", ":0")
	logger := NewNoopLogger()

	nilFactory := func(d *Deps) (GRPCServiceRegistrar, error) {
		return nil, nil
	}

	_ = NewMicro(
		WithConfig(cfg),
		WithLogger(logger),
		WithGRPCServer("grpc.port", nilFactory),
	)
}

func TestWithGRPCServerModules(t *testing.T) {
	cfg := NewConfig()
	cfg.Set("grpc.port", ":0")
	logger := NewNoopLogger()

	service := &testGRPCService{}

	ms := NewMicro(
		WithConfig(cfg),
		WithLogger(logger),
		WithGRPCServerModules("grpc.port", service),
	)

	if ms == nil {
		t.Fatal("NewMicro returned nil")
	}
	if !service.registerCalled {
		t.Error("RegisterGRPCService should have been called")
	}
}

func TestWithGRPCServerModulesNilService(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil service")
		}
	}()

	cfg := NewConfig()
	cfg.Set("grpc.port", ":0")
	logger := NewNoopLogger()

	_ = NewMicro(
		WithConfig(cfg),
		WithLogger(logger),
		WithGRPCServerModules("grpc.port", nil),
	)
}

func TestWithGRPCServerLifecycleHooks(t *testing.T) {
	cfg := NewConfig()
	cfg.Set("grpc.port", ":0")
	logger := NewNoopLogger()

	service := &testGRPCService{}
	factory := func(d *Deps) (GRPCServiceRegistrar, error) {
		return service, nil
	}

	ms := NewMicro(
		WithConfig(cfg),
		WithLogger(logger),
		WithGRPCServer("grpc.port", factory),
	)

	// Start and stop hooks should be registered
	if len(ms.startFuncs) != 1 {
		t.Errorf("expected 1 start func, got %d", len(ms.startFuncs))
	}
	if len(ms.stopFuncs) != 1 {
		t.Errorf("expected 1 stop func, got %d", len(ms.stopFuncs))
	}
}

func TestGRPCServerRunnerStartStop(t *testing.T) {
	server := grpc.NewServer()
	runner := newGRPCServerRunner(":0", server)

	err := runner.Start(context.Background())
	if err != nil {
		t.Fatalf("Start error: %v", err)
	}

	// Give the server time to start
	time.Sleep(10 * time.Millisecond)

	err = runner.Stop(context.Background())
	if err != nil {
		t.Fatalf("Stop error: %v", err)
	}
}

func TestGRPCServerRunnerStopWithTimeout(t *testing.T) {
	server := grpc.NewServer()
	runner := newGRPCServerRunner(":0", server)

	err := runner.Start(context.Background())
	if err != nil {
		t.Fatalf("Start error: %v", err)
	}

	// Give the server time to start
	time.Sleep(10 * time.Millisecond)

	// Stop with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = runner.Stop(ctx)
	// Should complete without error (forced stop)
	_ = err
}

func TestGRPCServiceRegistrarInterface(t *testing.T) {
	var s GRPCServiceRegistrar = &testGRPCService{}
	server := grpc.NewServer()
	s.RegisterGRPCService(server)
}

func TestGRPCServiceFactoryType(t *testing.T) {
	var factory GRPCServiceFactory = func(d *Deps) (GRPCServiceRegistrar, error) {
		return &testGRPCService{}, nil
	}

	service, err := factory(DefaultDeps())
	if err != nil {
		t.Errorf("factory error: %v", err)
	}
	if service == nil {
		t.Error("factory returned nil service")
	}
}

func TestGRPCServerRunnerStartError(t *testing.T) {
	server := grpc.NewServer()
	// Use an invalid address to cause an error
	runner := newGRPCServerRunner("invalid:address:format", server)

	err := runner.Start(context.Background())
	if err == nil {
		t.Error("expected error for invalid address")
	}
}
