package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDefaultCORSOptions(t *testing.T) {
	opts := DefaultCORSOptions()

	if len(opts.AllowedOrigins) != 1 || opts.AllowedOrigins[0] != "*" {
		t.Errorf("AllowedOrigins = %v, want [*]", opts.AllowedOrigins)
	}

	expectedMethods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	if len(opts.AllowedMethods) != len(expectedMethods) {
		t.Errorf("AllowedMethods length = %d, want %d", len(opts.AllowedMethods), len(expectedMethods))
	}

	if opts.MaxAge != 10*time.Minute {
		t.Errorf("MaxAge = %v, want 10m", opts.MaxAge)
	}
}

func TestCORSMiddlewareNoOrigin(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := CORS(DefaultCORSOptions())
	wrapped := middleware(handler)

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusOK)
	}

	if rec.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Error("Should not set CORS headers when no Origin header")
	}
}

func TestCORSMiddlewareWithOrigin(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := CORS(DefaultCORSOptions())
	wrapped := middleware(handler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "http://example.com")
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusOK)
	}

	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("Access-Control-Allow-Origin = %s, want *", rec.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORSMiddlewarePreflightRequest(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := CORS(DefaultCORSOptions())
	wrapped := middleware(handler)

	req := httptest.NewRequest("OPTIONS", "/", nil)
	req.Header.Set("Origin", "http://example.com")
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestCORSMiddlewareForbiddenOrigin(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	opts := CORSOptions{
		AllowedOrigins: []string{"http://allowed.com"},
	}
	middleware := CORS(opts)
	wrapped := middleware(handler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "http://forbidden.com")
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestCORSMiddlewareSpecificOrigin(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	opts := CORSOptions{
		AllowedOrigins: []string{"http://allowed.com"},
		AllowedMethods: []string{"GET", "POST"},
	}
	middleware := CORS(opts)
	wrapped := middleware(handler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "http://allowed.com")
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", rec.Code, http.StatusOK)
	}

	if rec.Header().Get("Access-Control-Allow-Origin") != "http://allowed.com" {
		t.Errorf("Access-Control-Allow-Origin = %s, want http://allowed.com", rec.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORSMiddlewareWithCredentials(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	opts := CORSOptions{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
	}
	middleware := CORS(opts)
	wrapped := middleware(handler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "http://example.com")
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Error("Access-Control-Allow-Credentials should be true")
	}

	// When credentials are allowed and origin is *, should echo the request origin
	if rec.Header().Get("Access-Control-Allow-Origin") != "http://example.com" {
		t.Errorf("Access-Control-Allow-Origin = %s, want http://example.com", rec.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORSMiddlewareMaxAge(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	opts := CORSOptions{
		AllowedOrigins: []string{"*"},
		MaxAge:         30 * time.Minute,
	}
	middleware := CORS(opts)
	wrapped := middleware(handler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "http://example.com")
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.Header().Get("Access-Control-Max-Age") != "1800" {
		t.Errorf("Access-Control-Max-Age = %s, want 1800", rec.Header().Get("Access-Control-Max-Age"))
	}
}

func TestCORSMiddlewareExposedHeaders(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	opts := CORSOptions{
		AllowedOrigins: []string{"*"},
		ExposedHeaders: []string{"X-Custom-Header", "X-Another-Header"},
	}
	middleware := CORS(opts)
	wrapped := middleware(handler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "http://example.com")
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	expected := "X-Custom-Header, X-Another-Header"
	if rec.Header().Get("Access-Control-Expose-Headers") != expected {
		t.Errorf("Access-Control-Expose-Headers = %s, want %s", rec.Header().Get("Access-Control-Expose-Headers"), expected)
	}
}

func TestOriginAllowed(t *testing.T) {
	tests := []struct {
		name    string
		origin  string
		allowed []string
		want    bool
	}{
		{"wildcard", "http://example.com", []string{"*"}, true},
		{"exact match", "http://example.com", []string{"http://example.com"}, true},
		{"case insensitive", "HTTP://EXAMPLE.COM", []string{"http://example.com"}, true},
		{"not allowed", "http://forbidden.com", []string{"http://allowed.com"}, false},
		{"empty allowed", "http://example.com", []string{}, false},
		{"multiple allowed", "http://example.com", []string{"http://other.com", "http://example.com"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := originAllowed(tt.origin, tt.allowed)
			if got != tt.want {
				t.Errorf("originAllowed(%s, %v) = %v, want %v", tt.origin, tt.allowed, got, tt.want)
			}
		})
	}
}

func TestSetVaryHeaders(t *testing.T) {
	h := http.Header{}
	setVaryHeaders(h)

	vary := h.Values("Vary")
	if len(vary) != 3 {
		t.Errorf("Vary headers count = %d, want 3", len(vary))
	}
}

func TestSetVaryHeadersExisting(t *testing.T) {
	h := http.Header{}
	h.Set("Vary", "Origin")
	setVaryHeaders(h)

	// Should only add missing headers
	vary := h.Values("Vary")
	hasOrigin := false
	for _, v := range vary {
		if v == "Origin" {
			hasOrigin = true
		}
	}
	if !hasOrigin {
		t.Error("Vary should contain Origin")
	}
}

func TestOriginHeaderValue(t *testing.T) {
	tests := []struct {
		name    string
		origin  string
		allowed []string
		want    string
	}{
		{"wildcard", "http://example.com", []string{"*"}, "*"},
		{"specific origin", "http://example.com", []string{"http://example.com"}, "http://example.com"},
		{"case insensitive", "HTTP://EXAMPLE.COM", []string{"http://example.com"}, "HTTP://EXAMPLE.COM"},
		{"no match returns origin", "http://other.com", []string{"http://example.com"}, "http://other.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := originHeaderValue(tt.origin, tt.allowed)
			if got != tt.want {
				t.Errorf("originHeaderValue(%s, %v) = %s, want %s", tt.origin, tt.allowed, got, tt.want)
			}
		})
	}
}

func TestCORSOptionsFields(t *testing.T) {
	opts := CORSOptions{
		AllowedOrigins:   []string{"http://example.com"},
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"Content-Type"},
		ExposedHeaders:   []string{"X-Custom"},
		AllowCredentials: true,
		MaxAge:           5 * time.Minute,
	}

	if len(opts.AllowedOrigins) != 1 {
		t.Error("AllowedOrigins not set correctly")
	}
	if len(opts.AllowedMethods) != 2 {
		t.Error("AllowedMethods not set correctly")
	}
	if len(opts.AllowedHeaders) != 1 {
		t.Error("AllowedHeaders not set correctly")
	}
	if len(opts.ExposedHeaders) != 1 {
		t.Error("ExposedHeaders not set correctly")
	}
	if !opts.AllowCredentials {
		t.Error("AllowCredentials not set correctly")
	}
	if opts.MaxAge != 5*time.Minute {
		t.Error("MaxAge not set correctly")
	}
}
