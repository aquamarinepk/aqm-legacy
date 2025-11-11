package aqm

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Option mutates the Micro instance during construction.
type Option func(*Micro) error

// WithLogger installs the shared logger instance.
func WithLogger(logger Logger) Option {
	return func(ms *Micro) error {
		if logger == nil {
			return errors.New("nil logger provided")
		}
		ms.mu.Lock()
		defer ms.mu.Unlock()
		ms.deps.Logger = logger
		return nil
	}
}

// WithTracer installs the shared tracer instance for distributed tracing.
func WithTracer(tracer Tracer) Option {
	return func(ms *Micro) error {
		if tracer == nil {
			tracer = NoopTracer{}
		}
		ms.mu.Lock()
		defer ms.mu.Unlock()
		ms.deps.Tracer = tracer
		return nil
	}
}

// WithMetrics installs the shared metrics collector.
func WithMetrics(metrics Metrics) Option {
	return func(ms *Micro) error {
		if metrics == nil {
			metrics = NoopMetrics{}
		}
		ms.mu.Lock()
		defer ms.mu.Unlock()
		ms.deps.Metrics = metrics
		return nil
	}
}

// WithErrorReporter installs the shared error reporting instance.
func WithErrorReporter(reporter ErrorReporter) Option {
	return func(ms *Micro) error {
		if reporter == nil {
			reporter = NoopErrorReporter{}
		}
		ms.mu.Lock()
		defer ms.mu.Unlock()
		ms.deps.Errors = reporter
		return nil
	}
}

// WithHealthChecks registers service-level liveness/readiness probes using the
// option pattern. Nil checks default to a pass-through implementation.
func WithHealthChecks(name string, checks ...HealthCheck) Option {
	return func(ms *Micro) error {
		if name == "" {
			return errors.New("health check name required")
		}
		liveness := HealthStatusOK
		readiness := HealthStatusOK
		if len(checks) > 0 && checks[0] != nil {
			liveness = checks[0]
		}
		if len(checks) > 1 && checks[1] != nil {
			readiness = checks[1]
		}
		ms.addHealthCheck(healthCheckRegistration{
			name:      name,
			liveness:  liveness,
			readiness: readiness,
		})
		return nil
	}
}

// WithDebugRoutes enables the /debug/routes endpoint on the HTTP server.
func WithDebugRoutes() Option {
	return func(ms *Micro) error {
		ms.mu.Lock()
		ms.debugRoutes = true
		ms.mu.Unlock()
		return nil
	}
}

// WithLifecycle registers components whose Start/Stop methods will be invoked
// by the orchestrator alongside other runners.
func WithLifecycle(components ...any) Option {
	return func(ms *Micro) error {
		for _, component := range components {
			if component == nil {
				continue
			}
			if startable, ok := component.(Startable); ok {
				ms.addStart(startable.Start)
			}
			if stoppable, ok := component.(Stoppable); ok {
				ms.addStop(stoppable.Stop)
			}
		}
		return nil
	}
}

// WithConfig wires a property-based configuration provider.
func WithConfig(cfg *Config) Option {
	return func(ms *Micro) error {
		if cfg == nil {
			return errors.New("nil config provided")
		}
		ms.mu.Lock()
		defer ms.mu.Unlock()
		ms.deps.Config = cfg
		return nil
	}
}

// WithRunner appends a lifecycle-managed component to the orchestrator.
func WithRunner(r Runner) Option {
	return func(ms *Micro) error {
		if r == nil {
			return errors.New("nil runner provided")
		}
		ms.addRunner(r)
		return nil
	}
}

// WithHTTPMiddleware registers middlewares applied to every HTTP server managed
// by Micro. Middlewares run in the order provided.
func WithHTTPMiddleware(middlewares ...func(http.Handler) http.Handler) Option {
	return func(ms *Micro) error {
		ms.mu.Lock()
		defer ms.mu.Unlock()
		ms.httpMiddlewares = append(ms.httpMiddlewares, middlewares...)
		return nil
	}
}

// WithRouterConfigurator allows callers to mutate the underlying *chi.Mux
// before HTTP modules register their routes.
func WithRouterConfigurator(configurer func(*chi.Mux)) Option {
	return func(ms *Micro) error {
		if configurer == nil {
			return errors.New("nil router configurator provided")
		}
		ms.mu.Lock()
		defer ms.mu.Unlock()
		ms.routerConfig = append(ms.routerConfig, configurer)
		return nil
	}
}

// WithShutdown registers a shutdown hook that runs after the runners stop.
func WithShutdown(fn ShutdownFunc) Option {
	return func(ms *Micro) error {
		if fn == nil {
			return errors.New("nil shutdown hook provided")
		}
		ms.addShutdown(fn)
		return nil
	}
}

// WithDeps allows bulk mutation of the dependency container.
func WithDeps(configurer func(*Deps) error) Option {
	return func(ms *Micro) error {
		if configurer == nil {
			return errors.New("nil dependency configurer provided")
		}
		ms.mu.Lock()
		defer ms.mu.Unlock()
		if err := configurer(ms.deps); err != nil {
			return fmt.Errorf("configuring dependencies: %w", err)
		}
		return nil
	}
}
