package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aquamarinepk/aqm"
)

func TestDefaultStack(t *testing.T) {
	opts := StackOptions{
		Logger:  aqm.NewNoopLogger(),
		Metrics: aqm.NoopMetrics{},
		Errors:  aqm.NoopErrorReporter{},
	}

	stack := DefaultStack(opts)

	if len(stack) == 0 {
		t.Error("DefaultStack should return non-empty stack")
	}
}

func TestDefaultStackWithTimeout(t *testing.T) {
	opts := StackOptions{
		Logger:          aqm.NewNoopLogger(),
		TimeoutDuration: 30 * time.Second,
	}

	stack := DefaultStack(opts)

	if len(stack) == 0 {
		t.Error("DefaultStack should return non-empty stack")
	}
}

func TestDefaultStackDisableTimeout(t *testing.T) {
	opts := StackOptions{
		Logger:         aqm.NewNoopLogger(),
		DisableTimeout: true,
	}

	stack := DefaultStack(opts)

	if len(stack) == 0 {
		t.Error("DefaultStack should return non-empty stack")
	}
}

func TestDefaultStackDisableCORS(t *testing.T) {
	opts := StackOptions{
		Logger:      aqm.NewNoopLogger(),
		DisableCORS: true,
	}

	stack := DefaultStack(opts)

	if len(stack) == 0 {
		t.Error("DefaultStack should return non-empty stack")
	}
}

func TestDefaultStackCustomCORS(t *testing.T) {
	corsOpts := CORSOptions{
		AllowedOrigins: []string{"http://example.com"},
	}
	opts := StackOptions{
		Logger:      aqm.NewNoopLogger(),
		CORSOptions: &corsOpts,
	}

	stack := DefaultStack(opts)

	if len(stack) == 0 {
		t.Error("DefaultStack should return non-empty stack")
	}
}

func TestRequestID(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequestID()
	wrapped := middleware(handler)

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestRealIP(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RealIP()
	wrapped := middleware(handler)

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestCompress(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test response"))
	})

	middleware := Compress(5)
	wrapped := middleware(handler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestCompressDefaultLevel(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test response"))
	})

	middleware := Compress(0) // Should default to 5
	wrapped := middleware(handler)

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestCompressNegativeLevel(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test response"))
	})

	middleware := Compress(-1) // Should default to 5
	wrapped := middleware(handler)

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestRecoverer(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	middleware := Recoverer()
	wrapped := middleware(handler)

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	// Should not panic
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestTimeoutZero(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := Timeout(0) // Should be passthrough
	wrapped := middleware(handler)

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestTimeoutNonZero(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := Timeout(5 * time.Second)
	wrapped := middleware(handler)

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestRequestLogger(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequestLogger(aqm.NewNoopLogger())
	wrapped := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestRequestLoggerNilLogger(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequestLogger(nil)
	wrapped := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestMetrics(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := Metrics(aqm.NoopMetrics{})
	wrapped := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestMetricsNilMetrics(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := Metrics(nil)
	wrapped := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusOK)
	}
}

type testErrorReporter struct {
	reported bool
}

func (r *testErrorReporter) Report(ctx context.Context, err error, fields map[string]any) {
	r.reported = true
}

func TestErrorReporter(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	reporter := &testErrorReporter{}
	middleware := ErrorReporter(reporter)
	wrapped := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if !reporter.reported {
		t.Error("ErrorReporter should report 5xx errors")
	}
}

func TestErrorReporterNilReporter(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	middleware := ErrorReporter(nil)
	wrapped := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	// Should not panic
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestErrorReporterPanic(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	reporter := &testErrorReporter{}
	middleware := ErrorReporter(reporter)
	wrapped := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Error("panic should propagate")
		}
		if !reporter.reported {
			t.Error("ErrorReporter should report panics")
		}
	}()

	wrapped.ServeHTTP(rec, req)
}

func TestAllowContentType(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := AllowContentType("application/json")
	wrapped := middleware(handler)

	req := httptest.NewRequest("POST", "/test", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestAllowContentTypeDefault(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := AllowContentType() // Uses defaults
	wrapped := middleware(handler)

	req := httptest.NewRequest("POST", "/test", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestNormalizeLogger(t *testing.T) {
	logger := normalizeLogger(nil)
	if logger == nil {
		t.Error("normalizeLogger should not return nil")
	}

	original := aqm.NewNoopLogger()
	logger = normalizeLogger(original)
	if logger != original {
		t.Error("normalizeLogger should return the same logger if not nil")
	}
}

func TestErrorFields(t *testing.T) {
	req := httptest.NewRequest("GET", "/test/path", nil)
	ctx := aqm.WithRequestID(req.Context(), "test-req-id")
	req = req.WithContext(ctx)

	fields := errorFields(req, 500)

	if fields["path"] != "/test/path" {
		t.Errorf("path = %v, want /test/path", fields["path"])
	}
	if fields["method"] != "GET" {
		t.Errorf("method = %v, want GET", fields["method"])
	}
	if fields["status"] != 500 {
		t.Errorf("status = %v, want 500", fields["status"])
	}
}

func TestErrorFieldsZeroStatus(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)

	fields := errorFields(req, 0)

	if _, ok := fields["status"]; ok {
		t.Error("status should not be present when 0")
	}
}

func TestToError(t *testing.T) {
	err := errors.New("test error")
	result := toError(err)
	if result != err {
		t.Errorf("toError should return same error")
	}

	result = toError("string panic")
	if result == nil {
		t.Error("toError should convert string to error")
	}
	if result.Error() != "panic: string panic" {
		t.Errorf("Error() = %s, want panic: string panic", result.Error())
	}

	result = toError(42)
	if result == nil {
		t.Error("toError should convert int to error")
	}
}

func TestStackOptionsFields(t *testing.T) {
	opts := StackOptions{
		Logger:              aqm.NewNoopLogger(),
		Metrics:             aqm.NoopMetrics{},
		Errors:              aqm.NoopErrorReporter{},
		TimeoutDuration:     30 * time.Second,
		DisableTimeout:      true,
		CompressLevel:       6,
		AllowedContentTypes: []string{"application/json"},
		DisableCORS:         true,
		CORSOptions:         &CORSOptions{AllowedOrigins: []string{"*"}},
	}

	if opts.Logger == nil {
		t.Error("Logger not set")
	}
	if opts.TimeoutDuration != 30*time.Second {
		t.Error("TimeoutDuration not set correctly")
	}
	if !opts.DisableTimeout {
		t.Error("DisableTimeout not set correctly")
	}
	if opts.CompressLevel != 6 {
		t.Error("CompressLevel not set correctly")
	}
	if !opts.DisableCORS {
		t.Error("DisableCORS not set correctly")
	}
}
