package aqm

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

// HTTPModule exposes a route registration entrypoint for HTTP transports.
type HTTPModule interface {
	RegisterRoutes(router chi.Router)
}

// HTTPModuleFactory constructs an HTTPModule from the shared dependency container.
type HTTPModuleFactory func(*Deps) (HTTPModule, error)

// WithHTTPServerModules is a convenience helper for the common case where
// modules do not need to access the shared dependency container during
// construction. It wraps the provided modules into factories and delegates to
// WithHTTPServer.
func WithHTTPServerModules(addrKey string, modules ...HTTPModule) Option {
	factories := make([]HTTPModuleFactory, len(modules))
	for i, module := range modules {
		mod := module
		factories[i] = func(*Deps) (HTTPModule, error) {
			if mod == nil {
				return nil, errors.New("nil http module provided")
			}
			return mod, nil
		}
	}
	return WithHTTPServer(addrKey, factories...)
}

// WithHTTPServer wires a chi-based HTTP server runner. It instantiates the
// provided module factories, registers their routes, and mounts the resulting
// server as a lifecycle-managed runner.
func WithHTTPServer(addrKey string, factories ...HTTPModuleFactory) Option {
	return func(ms *Micro) error {
		if addrKey == "" {
			return errors.New("http addr property key required")
		}

		ms.mu.Lock()
		defer ms.mu.Unlock()
		if ms.httpConfigured {
			return errors.New("http server already configured")
		}
		ms.httpConfigured = true

		router := chi.NewRouter()
		for _, mw := range ms.httpMiddlewares {
			if mw == nil {
				continue
			}
			router.Use(mw)
		}

		healthRegistry := NewHealthRegistry()
		RegisterHealthEndpoints(router, healthRegistry)
		healthRegistry.RegisterLiveness("core", HealthStatusOK)
		healthRegistry.RegisterReadiness("core", HealthStatusOK)
		RegisterDebugRoutes(router, ms.debugRoutes)
		for _, configurer := range ms.routerConfig {
			if configurer != nil {
				configurer(router)
			}
		}

		for _, reg := range ms.healthChecks {
			if reg.liveness != nil {
				healthRegistry.RegisterLiveness(reg.name, reg.liveness)
			}
			if reg.readiness != nil {
				healthRegistry.RegisterReadiness(reg.name, reg.readiness)
			}
		}

		for _, factory := range factories {
			if factory == nil {
				return errors.New("nil http module factory")
			}
			module, err := factory(ms.deps)
			if err != nil {
				return fmt.Errorf("building http module: %w", err)
			}
			if module == nil {
				return errors.New("http module factory returned nil module")
			}
			module.RegisterRoutes(router)
			if reporter, ok := module.(HealthReporter); ok {
				healthRegistry.RegisterChecks(reporter.HealthChecks())
			}
			if startable, ok := module.(Startable); ok {
				ms.addStart(startable.Start)
			}
			if stoppable, ok := module.(Stoppable); ok {
				ms.addStop(stoppable.Stop)
			}
		}

		addr := ms.deps.Config.GetPort(addrKey, ":8080")

		server := &http.Server{
			Addr:    addr,
			Handler: router,
		}

		ms.runners = append(ms.runners, newHTTPServerRunner(server))
		return nil
	}
}

type httpServerRunner struct {
	server *http.Server
	errCh  chan error
}

func newHTTPServerRunner(server *http.Server) Runner {
	return &httpServerRunner{server: server, errCh: make(chan error, 1)}
}

func (r *httpServerRunner) Start(_ context.Context) error {
	go func() {
		if err := r.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			r.errCh <- err
		}
		close(r.errCh)
	}()
	return nil
}

func (r *httpServerRunner) Stop(ctx context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	err := r.server.Shutdown(shutdownCtx)
	select {
	case srvErr, ok := <-r.errCh:
		if ok && srvErr != nil {
			err = errors.Join(err, srvErr)
		}
	default:
	}
	return err
}
