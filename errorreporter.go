package aqm

import "context"

// ErrorReporter captures unexpected failures so services can forward them to
// alerting/observability systems.
type ErrorReporter interface {
	Report(ctx context.Context, err error, fields map[string]any)
}

// ErrorReporterFunc adapts a function into an ErrorReporter.
type ErrorReporterFunc func(ctx context.Context, err error, fields map[string]any)

func (f ErrorReporterFunc) Report(ctx context.Context, err error, fields map[string]any) {
	if f == nil {
		return
	}
	f(ctx, err, fields)
}

// NoopErrorReporter drops all reports.
type NoopErrorReporter struct{}

func (NoopErrorReporter) Report(context.Context, error, map[string]any) {}
