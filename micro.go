package aqm

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
)

// Micro orchestrates dependency wiring, runner lifecycle management, and shutdown hooks.
type Micro struct {
	deps     *Deps
	runners  []Runner
	shutdown []ShutdownFunc

	mu              sync.RWMutex
	httpConfigured  bool
	httpMiddlewares []func(http.Handler) http.Handler
	routerConfig    []func(*chi.Mux)

	healthChecks []healthCheckRegistration
	debugRoutes  bool

	startFuncs []func(context.Context) error
	stopFuncs  []func(context.Context) error
}

type healthCheckRegistration struct {
	name      string
	liveness  HealthCheck
	readiness HealthCheck
}

// ShutdownFunc is executed when Run exits, giving modules a chance to release resources.
type ShutdownFunc func(context.Context) error

// NewMicro builds a new Micro instance, applying the provided options sequentially.
// It panics when an option returns an error or mandatory dependencies are missing.
func NewMicro(opts ...Option) *Micro {
	ms := &Micro{
		deps:        DefaultDeps(),
		debugRoutes: true,
	}
	for _, opt := range opts {
		if err := opt(ms); err != nil {
			panic(fmt.Errorf("applying option: %w", err))
		}
	}
	ms.ensureCoreDependencies()
	return ms
}

// Run starts all registered runners, blocks until the context is cancelled, and then stops
// runners in reverse order before executing shutdown hooks. Errors emitted while stopping
// or during shutdown are aggregated.
func (micro *Micro) Run(ctx context.Context) error {
	micro.mu.RLock()
	runners := append([]Runner(nil), micro.runners...)
	shutdown := append([]ShutdownFunc(nil), micro.shutdown...)
	startFns := append([]func(context.Context) error(nil), micro.startFuncs...)
	stopFns := append([]func(context.Context) error(nil), micro.stopFuncs...)
	micro.mu.RUnlock()

	for i, start := range startFns {
		if err := start(ctx); err != nil {
			// attempt rollback of previously started components
			for j := i - 1; j >= 0; j-- {
				if stopErr := stopFns[j](context.Background()); stopErr != nil {
					err = errors.Join(err, fmt.Errorf("lifecycle rollback: %w", stopErr))
				}
			}
			return fmt.Errorf("lifecycle start: %w", err)
		}
	}

	for _, runner := range runners {
		if err := runner.Start(ctx); err != nil {
			return fmt.Errorf("runner start: %w", err)
		}
	}

	<-ctx.Done()

	var aggErr error
	for i := len(runners) - 1; i >= 0; i-- {
		if err := runners[i].Stop(ctx); err != nil {
			aggErr = errors.Join(aggErr, fmt.Errorf("runner stop: %w", err))
		}
	}
	for i := len(stopFns) - 1; i >= 0; i-- {
		if err := stopFns[i](ctx); err != nil {
			aggErr = errors.Join(aggErr, fmt.Errorf("lifecycle stop: %w", err))
		}
	}
	for _, hook := range shutdown {
		if err := hook(ctx); err != nil {
			aggErr = errors.Join(aggErr, fmt.Errorf("shutdown hook: %w", err))
		}
	}
	return aggErr
}

// Deps exposes the wired dependency container.
func (micro *Micro) Deps() *Deps {
	micro.mu.RLock()
	defer micro.mu.RUnlock()
	return micro.deps
}

// addRunner installs a runner in a threadsafe manner.
func (micro *Micro) addRunner(r Runner) {
	micro.mu.Lock()
	defer micro.mu.Unlock()
	micro.runners = append(micro.runners, r)
}

// addShutdown registers a shutdown hook executed after runners stop.
func (micro *Micro) addShutdown(fn ShutdownFunc) {
	micro.mu.Lock()
	defer micro.mu.Unlock()
	micro.shutdown = append(micro.shutdown, fn)
}

func (micro *Micro) addHealthCheck(reg healthCheckRegistration) {
	micro.mu.Lock()
	defer micro.mu.Unlock()
	micro.healthChecks = append(micro.healthChecks, reg)
}

func (micro *Micro) ensureCoreDependencies() {
	micro.mu.RLock()
	logger := micro.deps.Logger
	config := micro.deps.Config
	micro.mu.RUnlock()

	if logger == nil {
		panic("logger dependency must be configured")
	}
	if config == nil {
		panic("config dependency must be configured")
	}
}

func (micro *Micro) addStart(fn func(context.Context) error) {
	if fn == nil {
		return
	}
	micro.mu.Lock()
	micro.startFuncs = append(micro.startFuncs, fn)
	micro.mu.Unlock()
}

func (micro *Micro) addStop(fn func(context.Context) error) {
	if fn == nil {
		return
	}
	micro.mu.Lock()
	micro.stopFuncs = append(micro.stopFuncs, fn)
	micro.mu.Unlock()
}
