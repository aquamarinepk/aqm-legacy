package aqm

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type Startable interface {
	Start(context.Context) error
}

type Stoppable interface {
	Stop(context.Context) error
}

type RouteRegistrar interface {
	RegisterRoutes(chi.Router)
}

// LifecycleHooks adapts plain functions to the Startable/Stoppable interfaces so
// callers can wire arbitrary logic into the service lifecycle.
type LifecycleHooks struct {
	OnStart func(context.Context) error
	OnStop  func(context.Context) error
}

func (h LifecycleHooks) Start(ctx context.Context) error {
	if h.OnStart == nil {
		return nil
	}
	return h.OnStart(ctx)
}

func (h LifecycleHooks) Stop(ctx context.Context) error {
	if h.OnStop == nil {
		return nil
	}
	return h.OnStop(ctx)
}

func Setup(ctx context.Context, r chi.Router, comps ...any) (
	starts []func(context.Context) error,
	stops []func(context.Context) error,
	health *HealthRegistry,
) {
	health = NewHealthRegistry()
	RegisterHealthEndpoints(r, health)

	health.RegisterLiveness("core", func(context.Context) error { return nil })
	health.RegisterReadiness("core", func(context.Context) error { return nil })

	for _, c := range comps {
		if rr, ok := c.(RouteRegistrar); ok {
			rr.RegisterRoutes(r)
		}
		if s, ok := c.(Startable); ok {
			starts = append(starts, s.Start)
		}
		if st, ok := c.(Stoppable); ok {
			stops = append(stops, st.Stop)
		}
		if hp, ok := c.(HealthReporter); ok {
			health.RegisterChecks(hp.HealthChecks())
		}
	}
	return
}

func Start(ctx context.Context, logger Logger, starts []func(context.Context) error, stops []func(context.Context) error) error {
	if logger == nil {
		logger = NewNoopLogger()
	}
	for i, start := range starts {
		if err := start(ctx); err != nil {
			logger.Error("component start failed", "index", i, "error", err)
			rollbackCtx := context.Background()
			for j := i - 1; j >= 0; j-- {
				if stopErr := stops[j](rollbackCtx); stopErr != nil {
					logger.Error("component rollback failed", "index", j, "error", stopErr)
				}
			}
			return err
		}
	}
	return nil
}

func Shutdown(ctx context.Context, srv *http.Server, logger Logger, stops []func(context.Context) error) {
	if logger == nil {
		logger = NewNoopLogger()
	}
	if ctx == nil {
		ctx = context.Background()
	}

	logger.Info("Shutting down gracefully, press Ctrl+C again to force")

	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if srv != nil {
		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Error("server shutdown failed", "error", err)
		}
	}

	stopCtx := context.Background()
	for i := len(stops) - 1; i >= 0; i-- {
		if err := stops[i](stopCtx); err != nil {
			logger.Error("error stopping component", "index", i, "error", err)
		}
	}
}
