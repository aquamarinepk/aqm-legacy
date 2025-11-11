package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/aquamarinepk/aqm"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

// StackOptions configures the default middleware bundle.
type StackOptions struct {
	Logger              aqm.Logger
	Metrics             aqm.Metrics
	Errors              aqm.ErrorReporter
	Timeout             time.Duration
	CompressLevel       int
	AllowedContentTypes []string
}

// DefaultStack wires the recommended middleware order for aqm services.
func DefaultStack(opts StackOptions) []func(http.Handler) http.Handler {
	stack := []func(http.Handler) http.Handler{
		RequestID(),
		RealIP(),
		Compress(opts.CompressLevel),
		Recoverer(),
		ErrorReporter(opts.Errors),
		Timeout(opts.Timeout),
		RequestLogger(opts.Logger),
		Metrics(opts.Metrics),
		AllowContentType(opts.AllowedContentTypes...),
	}
	return stack
}

// RequestID ensures every request carries a correlation identifier.
func RequestID() func(http.Handler) http.Handler {
	return aqm.RequestIDMiddleware
}

// RealIP resolves the actual remote IP when behind proxies/load balancers.
func RealIP() func(http.Handler) http.Handler {
	return chimiddleware.RealIP
}

// Compress enables gzip compression.
func Compress(level int) func(http.Handler) http.Handler {
	if level <= 0 {
		level = 5
	}
	return chimiddleware.Compress(level)
}

// Recoverer prevents panics from tearing down the server.
func Recoverer() func(http.Handler) http.Handler {
	return chimiddleware.Recoverer
}

// Timeout aborts requests that exceed the configured duration.
func Timeout(duration time.Duration) func(http.Handler) http.Handler {
	if duration <= 0 {
		duration = 60 * time.Second
	}
	return chimiddleware.Timeout(duration)
}

// RequestLogger emits structured request lifecycle logs.
func RequestLogger(logger aqm.Logger) func(http.Handler) http.Handler {
	return aqm.NewRequestLogger(normalizeLogger(logger))
}

// Metrics publishes request counters and latencies using the shared Metrics.
func Metrics(metrics aqm.Metrics) func(http.Handler) http.Handler {
	if metrics == nil {
		metrics = aqm.NoopMetrics{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			recorder := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
			start := time.Now()
			next.ServeHTTP(recorder, r)

			labels := map[string]string{
				"method": r.Method,
				"path":   r.URL.Path,
				"status": strconv.Itoa(recorder.Status()),
			}
			metrics.Counter(r.Context(), "http_requests_total", 1, labels)
			metrics.Counter(r.Context(), "http_request_duration_ms", float64(time.Since(start).Milliseconds()), labels)
		})
	}
}

// ErrorReporter forwards 5xx responses and panics to the configured reporter.
func ErrorReporter(reporter aqm.ErrorReporter) func(http.Handler) http.Handler {
	if reporter == nil {
		reporter = aqm.NoopErrorReporter{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			recorder := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
			defer func() {
				if rec := recover(); rec != nil {
					reporter.Report(r.Context(), toError(rec), errorFields(r, 0))
					panic(rec)
				}
			}()

			next.ServeHTTP(recorder, r)

			status := recorder.Status()
			if status >= http.StatusInternalServerError {
				reporter.Report(r.Context(), fmt.Errorf("http %d", status), errorFields(r, status))
			}
		})
	}
}

// AllowContentType gate-keeps supported media types.
func AllowContentType(types ...string) func(http.Handler) http.Handler {
	if len(types) == 0 {
		types = []string{"application/json", "application/x-www-form-urlencoded", "multipart/form-data"}
	}
	return chimiddleware.AllowContentType(types...)
}

func normalizeLogger(logger aqm.Logger) aqm.Logger {
	if logger == nil {
		return aqm.NewNoopLogger()
	}
	return logger
}

func errorFields(r *http.Request, status int) map[string]any {
	fields := map[string]any{
		"request_id": aqm.RequestIDFrom(r.Context()),
		"path":       r.URL.Path,
		"method":     r.Method,
	}
	if status != 0 {
		fields["status"] = status
	}
	return fields
}

func toError(v any) error {
	switch err := v.(type) {
	case error:
		return err
	default:
		return fmt.Errorf("panic: %v", err)
	}
}
