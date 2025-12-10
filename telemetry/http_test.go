package telemetry

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aquamarinepk/aqm"
)

func TestNewHTTP(t *testing.T) {
	h := NewHTTP()

	if h == nil {
		t.Fatal("NewHTTP returned nil")
	}
	if h.tracer == nil {
		t.Error("tracer should not be nil")
	}
	if h.metrics == nil {
		t.Error("metrics should not be nil")
	}
}

func TestNewHTTPWithOptions(t *testing.T) {
	tracer := aqm.NoopTracer{}
	metrics := aqm.NoopMetrics{}

	h := NewHTTP(
		WithTracer(tracer),
		WithMetrics(metrics),
	)

	if h == nil {
		t.Fatal("NewHTTP returned nil")
	}
}

func TestNewHTTPWithNilOption(t *testing.T) {
	// Should not panic
	h := NewHTTP(nil)

	if h == nil {
		t.Fatal("NewHTTP returned nil")
	}
}

func TestWithTracer(t *testing.T) {
	tracer := aqm.NoopTracer{}
	h := NewHTTP(WithTracer(tracer))

	if h.tracer == nil {
		t.Error("tracer should not be nil")
	}
}

func TestWithTracerNil(t *testing.T) {
	h := NewHTTP(WithTracer(nil))

	if h.tracer == nil {
		t.Error("tracer should fall back to NoopTracer")
	}
}

func TestWithMetrics(t *testing.T) {
	metrics := aqm.NoopMetrics{}
	h := NewHTTP(WithMetrics(metrics))

	if h.metrics == nil {
		t.Error("metrics should not be nil")
	}
}

func TestWithMetricsNil(t *testing.T) {
	h := NewHTTP(WithMetrics(nil))

	if h.metrics == nil {
		t.Error("metrics should fall back to NoopMetrics")
	}
}

func TestHTTPStart(t *testing.T) {
	h := NewHTTP()

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	rw, r, finish := h.Start(rec, req, "test-span")

	if rw == nil {
		t.Error("response writer should not be nil")
	}
	if r == nil {
		t.Error("request should not be nil")
	}
	if finish == nil {
		t.Error("finish func should not be nil")
	}

	// Call finish - should not panic
	finish()
}

func TestHTTPStartWithNilTracer(t *testing.T) {
	h := &HTTP{
		tracer:  nil,
		metrics: nil,
	}

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	rw, r, finish := h.Start(rec, req, "test-span")

	if rw == nil {
		t.Error("response writer should not be nil")
	}
	if r == nil {
		t.Error("request should not be nil")
	}

	// Call finish - should not panic
	finish()
}

func TestHTTPStartWritesResponse(t *testing.T) {
	h := NewHTTP()

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	rw, _, finish := h.Start(rec, req, "test-span")
	defer finish()

	rw.WriteHeader(http.StatusCreated)
	rw.Write([]byte("test response"))

	if rec.Code != http.StatusCreated {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusCreated)
	}
	if rec.Body.String() != "test response" {
		t.Errorf("Body = %s, want test response", rec.Body.String())
	}
}

func TestNewMetricsMiddleware(t *testing.T) {
	metrics := aqm.NoopMetrics{}
	middleware := NewMetricsMiddleware(metrics)

	if middleware == nil {
		t.Fatal("NewMetricsMiddleware returned nil")
	}
}

func TestNewMetricsMiddlewareNilMetrics(t *testing.T) {
	middleware := NewMetricsMiddleware(nil)

	if middleware == nil {
		t.Fatal("NewMetricsMiddleware returned nil")
	}
}

func TestMetricsMiddlewareExecutesHandler(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	middleware := NewMetricsMiddleware(aqm.NoopMetrics{})
	wrapped := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Body.String() != "OK" {
		t.Errorf("Body = %s, want OK", rec.Body.String())
	}
}

func TestMetricsMiddlewareRecordsDuration(t *testing.T) {
	recorder := &metricsRecorder{}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	middleware := NewMetricsMiddleware(recorder)
	wrapped := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if !recorder.observed {
		t.Error("metrics should have been observed")
	}
	if recorder.path != "/test" {
		t.Errorf("path = %s, want /test", recorder.path)
	}
	if recorder.method != "GET" {
		t.Errorf("method = %s, want GET", recorder.method)
	}
	if recorder.status != http.StatusOK {
		t.Errorf("status = %d, want %d", recorder.status, http.StatusOK)
	}
}

type metricsRecorder struct {
	observed bool
	path     string
	method   string
	status   int
	duration time.Duration
}

func (m *metricsRecorder) Counter(ctx context.Context, name string, value float64, labels map[string]string) {
}

func (m *metricsRecorder) Gauge(ctx context.Context, name string, value float64, labels map[string]string) {
}

func (m *metricsRecorder) Histogram(ctx context.Context, name string, value float64, labels map[string]string) {
}

func (m *metricsRecorder) ObserveHTTPRequest(path, method string, status int, duration time.Duration) {
	m.observed = true
	m.path = path
	m.method = method
	m.status = status
	m.duration = duration
}

func TestOptionType(t *testing.T) {
	var opt Option = func(h *HTTP) {
		h.tracer = aqm.NoopTracer{}
	}

	h := &HTTP{}
	opt(h)

	if h.tracer == nil {
		t.Error("tracer should have been set")
	}
}

func TestHTTPStructFields(t *testing.T) {
	h := &HTTP{
		tracer:  aqm.NoopTracer{},
		metrics: aqm.NoopMetrics{},
	}

	if h.tracer == nil {
		t.Error("tracer should not be nil")
	}
	if h.metrics == nil {
		t.Error("metrics should not be nil")
	}
}
