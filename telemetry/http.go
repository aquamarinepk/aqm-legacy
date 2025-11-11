package telemetry

import (
	"net/http"
	"time"

	"github.com/aquamarinepk/aqm"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

// HTTP instruments HTTP handlers with tracing and metrics.
type HTTP struct {
	tracer  aqm.Tracer
	metrics aqm.Metrics
}

// Option mutates HTTP configuration.
type Option func(*HTTP)

// NewHTTP builds an HTTP instrumentation helper with optional custom deps.
func NewHTTP(opts ...Option) *HTTP {
	h := &HTTP{
		tracer:  aqm.NoopTracer{},
		metrics: aqm.NoopMetrics{},
	}
	for _, opt := range opts {
		if opt != nil {
			opt(h)
		}
	}
	return h
}

// WithTracer overrides the tracer implementation used by HTTP spans.
func WithTracer(t aqm.Tracer) Option {
	return func(h *HTTP) {
		if t == nil {
			t = aqm.NoopTracer{}
		}
		h.tracer = t
	}
}

// WithMetrics overrides the metrics collector used by HTTP instrumentation.
func WithMetrics(m aqm.Metrics) Option {
	return func(h *HTTP) {
		if m == nil {
			m = aqm.NoopMetrics{}
		}
		h.metrics = m
	}
}

// Start wraps the request/response pair, starting a span and returning a finish
// function that records duration and status.
func (h *HTTP) Start(w http.ResponseWriter, r *http.Request, spanName string) (http.ResponseWriter, *http.Request, func()) {
	tracer := h.tracer
	if tracer == nil {
		tracer = aqm.NoopTracer{}
	}
	metrics := h.metrics
	if metrics == nil {
		metrics = aqm.NoopMetrics{}
	}

	ctx, span := tracer.Start(r.Context(), spanName, nil)
	rw := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
	start := time.Now()
	reqWithCtx := r.WithContext(ctx)

	finish := func() {
		span.End(nil)
		metrics.ObserveHTTPRequest(reqWithCtx.URL.Path, reqWithCtx.Method, rw.Status(), time.Since(start))
	}

	return rw, reqWithCtx, finish
}

// NewMetricsMiddleware measures request durations and reports them through the
// provided Metrics implementation.
func NewMetricsMiddleware(metrics aqm.Metrics) func(http.Handler) http.Handler {
	if metrics == nil {
		metrics = aqm.NoopMetrics{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rw := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(rw, r)

			duration := time.Since(start)
			metrics.ObserveHTTPRequest(r.URL.Path, r.Method, rw.Status(), duration)
		})
	}
}
