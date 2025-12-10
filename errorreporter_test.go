package aqm

import (
	"context"
	"errors"
	"testing"
)

func TestErrorReporterFuncReport(t *testing.T) {
	tests := []struct {
		name     string
		fn       ErrorReporterFunc
		err      error
		fields   map[string]any
		expectOK bool
	}{
		{
			name:     "nilFunc",
			fn:       nil,
			err:      errors.New("test error"),
			fields:   nil,
			expectOK: true,
		},
		{
			name: "validFunc",
			fn: func(ctx context.Context, err error, fields map[string]any) {
				// do nothing
			},
			err:      errors.New("test error"),
			fields:   map[string]any{"key": "value"},
			expectOK: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// should not panic
			tt.fn.Report(context.Background(), tt.err, tt.fields)
		})
	}
}

func TestErrorReporterFuncCalled(t *testing.T) {
	var called bool
	var capturedErr error
	var capturedFields map[string]any

	fn := ErrorReporterFunc(func(ctx context.Context, err error, fields map[string]any) {
		called = true
		capturedErr = err
		capturedFields = fields
	})

	testErr := errors.New("test error")
	testFields := map[string]any{"key": "value"}

	fn.Report(context.Background(), testErr, testFields)

	if !called {
		t.Error("expected function to be called")
	}
	if capturedErr != testErr {
		t.Errorf("expected error %v, got %v", testErr, capturedErr)
	}
	if capturedFields["key"] != "value" {
		t.Error("expected fields to be passed")
	}
}

func TestNoopErrorReporterReport(t *testing.T) {
	reporter := NoopErrorReporter{}
	// should not panic
	reporter.Report(context.Background(), errors.New("test"), nil)
	reporter.Report(nil, nil, nil)
}
