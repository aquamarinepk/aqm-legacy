package aqm

import (
	"context"
	"time"
)

// Deps aggregates cross-cutting concerns shared across transports and modules.
type Deps struct {
	Logger    Logger
	Config    *Config
	Metrics   Metrics
	Tracer    Tracer
	Errors    ErrorReporter
	Validator Validator
	PubSub    PubSub
}

// DefaultDeps returns a container filled with no-op implementations.
func DefaultDeps() *Deps {
	return &Deps{
		Metrics:   NoopMetrics{},
		Tracer:    NoopTracer{},
		Errors:    NoopErrorReporter{},
		Validator: NoopValidator{},
		PubSub:    NoopPubSub{},
	}
}

// Metrics models a minimal counter/measure emission interface with HTTP-specific observations.
type Metrics interface {
	Counter(ctx context.Context, name string, value float64, labels map[string]string)
	ObserveHTTPRequest(path, method string, status int, duration time.Duration)
}

// Tracer models an instrumentation provider capable of creating spans.
type Tracer interface {
	Start(ctx context.Context, name string, attrs map[string]any) (context.Context, Span)
}

// Span is the handle returned by Tracer.Start.
type Span interface {
	End(err error)
}

// PubSub captures a minimal publish/subscribe abstraction.
type PubSub interface {
	Publish(ctx context.Context, subject string, payload []byte) error
}

type NoopMetrics struct{}

type NoopTracer struct{}

type NoopSpan struct{}

type NoopPubSub struct{}

func (NoopMetrics) Counter(context.Context, string, float64, map[string]string) {}
func (NoopMetrics) ObserveHTTPRequest(string, string, int, time.Duration)       {}

func (NoopTracer) Start(ctx context.Context, _ string, _ map[string]any) (context.Context, Span) {
	return ctx, NoopSpan{}
}

func (NoopSpan) End(error) {}

func (NoopPubSub) Publish(context.Context, string, []byte) error { return nil }
