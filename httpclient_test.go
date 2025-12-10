package aqm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewHTTPClient(t *testing.T) {
	tests := []struct {
		name   string
		config HTTPClientConfig
	}{
		{
			name:   "defaultConfig",
			config: HTTPClientConfig{},
		},
		{
			name: "customConfig",
			config: HTTPClientConfig{
				BaseURL:    "http://localhost:8080",
				Timeout:    5 * time.Second,
				MaxRetries: 5,
				RetryDelay: 2 * time.Second,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewHTTPClient(tt.config)
			if client == nil {
				t.Fatal("NewHTTPClient returned nil")
			}
			if client.HTTPClient == nil {
				t.Error("HTTPClient should not be nil")
			}
		})
	}
}

func TestHTTPClientGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	client := NewHTTPClient(HTTPClientConfig{BaseURL: server.URL})
	var result map[string]string
	err := client.Get(context.Background(), "/test", &result)

	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if result["status"] != "ok" {
		t.Errorf("expected status ok, got %v", result)
	}
}

func TestHTTPClientPost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"created": "true"})
	}))
	defer server.Close()

	client := NewHTTPClient(HTTPClientConfig{BaseURL: server.URL})
	body := map[string]string{"name": "test"}
	var result map[string]string
	err := client.Post(context.Background(), "/test", body, &result)

	if err != nil {
		t.Fatalf("Post error: %v", err)
	}
}

func TestHTTPClientPut(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"updated": "true"})
	}))
	defer server.Close()

	client := NewHTTPClient(HTTPClientConfig{BaseURL: server.URL})
	var result map[string]string
	err := client.Put(context.Background(), "/test", map[string]string{}, &result)

	if err != nil {
		t.Fatalf("Put error: %v", err)
	}
}

func TestHTTPClientPatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"patched": "true"})
	}))
	defer server.Close()

	client := NewHTTPClient(HTTPClientConfig{BaseURL: server.URL})
	var result map[string]string
	err := client.Patch(context.Background(), "/test", map[string]string{}, &result)

	if err != nil {
		t.Fatalf("Patch error: %v", err)
	}
}

func TestHTTPClientDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewHTTPClient(HTTPClientConfig{BaseURL: server.URL})
	err := client.Delete(context.Background(), "/test")

	if err != nil {
		t.Fatalf("Delete error: %v", err)
	}
}

func TestHTTPClientErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("bad request"))
	}))
	defer server.Close()

	client := NewHTTPClient(HTTPClientConfig{BaseURL: server.URL, MaxRetries: 0})
	var result map[string]string
	err := client.Get(context.Background(), "/test", &result)

	if err == nil {
		t.Fatal("expected error")
	}

	httpErr, ok := err.(*HTTPError)
	if !ok {
		t.Fatalf("expected HTTPError, got %T", err)
	}
	if httpErr.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", httpErr.StatusCode)
	}
}

func TestHTTPClientPing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/healthz" {
			t.Errorf("expected /healthz, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewHTTPClient(HTTPClientConfig{BaseURL: server.URL})
	err := client.Ping(context.Background())

	if err != nil {
		t.Fatalf("Ping error: %v", err)
	}
}

func TestHTTPClientPingUnhealthy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	client := NewHTTPClient(HTTPClientConfig{BaseURL: server.URL})
	err := client.Ping(context.Background())

	if err == nil {
		t.Fatal("expected error for unhealthy service")
	}
}

func TestHTTPClientWithRequestID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get(RequestIDHeader)
		if reqID != "test-req-id" {
			t.Errorf("expected request ID 'test-req-id', got %s", reqID)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{})
	}))
	defer server.Close()

	client := NewHTTPClient(HTTPClientConfig{BaseURL: server.URL})
	ctx := WithRequestID(context.Background(), "test-req-id")
	var result map[string]string
	err := client.Get(ctx, "/test", &result)

	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
}

func TestHTTPError(t *testing.T) {
	err := &HTTPError{StatusCode: 404, Message: "not found"}

	if err.Error() != "HTTP 404: not found" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
	if !err.IsNotFound() {
		t.Error("IsNotFound should return true")
	}
	if err.IsUnauthorized() {
		t.Error("IsUnauthorized should return false")
	}
	if err.IsForbidden() {
		t.Error("IsForbidden should return false")
	}
}

func TestHTTPErrorIsUnauthorized(t *testing.T) {
	err := &HTTPError{StatusCode: 401, Message: "unauthorized"}
	if !err.IsUnauthorized() {
		t.Error("IsUnauthorized should return true")
	}
}

func TestHTTPErrorIsForbidden(t *testing.T) {
	err := &HTTPError{StatusCode: 403, Message: "forbidden"}
	if !err.IsForbidden() {
		t.Error("IsForbidden should return true")
	}
}

func TestHTTPClientShouldRetry(t *testing.T) {
	client := NewHTTPClient(HTTPClientConfig{})

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil", nil, false},
		{"nonHTTPError", context.DeadlineExceeded, true},
		{"badRequest", &HTTPError{StatusCode: 400}, false},
		{"unauthorized", &HTTPError{StatusCode: 401}, false},
		{"forbidden", &HTTPError{StatusCode: 403}, false},
		{"notFound", &HTTPError{StatusCode: 404}, false},
		{"conflict", &HTTPError{StatusCode: 409}, false},
		{"tooManyRequests", &HTTPError{StatusCode: 429}, true},
		{"internalError", &HTTPError{StatusCode: 500}, true},
		{"badGateway", &HTTPError{StatusCode: 502}, true},
		{"serviceUnavailable", &HTTPError{StatusCode: 503}, true},
		{"gatewayTimeout", &HTTPError{StatusCode: 504}, true},
		{"unknownServerError", &HTTPError{StatusCode: 599}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := client.shouldRetry(tt.err)
			if got != tt.expected {
				t.Errorf("shouldRetry() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestHTTPClientRetryBehavior(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	client := NewHTTPClient(HTTPClientConfig{
		BaseURL:    server.URL,
		MaxRetries: 3,
		RetryDelay: 1 * time.Millisecond,
	})

	var result map[string]string
	err := client.Get(context.Background(), "/test", &result)

	if err != nil {
		t.Fatalf("Get error after retries: %v", err)
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestHTTPClientContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewHTTPClient(HTTPClientConfig{
		BaseURL:    server.URL,
		MaxRetries: 5,
		RetryDelay: 10 * time.Millisecond,
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	var result map[string]string
	err := client.Get(ctx, "/test", &result)

	if err == nil {
		t.Fatal("expected error due to cancelled context")
	}
}

func TestHTTPClientNoContentResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewHTTPClient(HTTPClientConfig{BaseURL: server.URL})
	var result map[string]string
	err := client.Get(context.Background(), "/test", &result)

	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
}

func TestHTTPClientConfigFields(t *testing.T) {
	cfg := HTTPClientConfig{
		BaseURL:    "http://test",
		Timeout:    5 * time.Second,
		MaxRetries: 3,
		RetryDelay: time.Second,
	}

	if cfg.BaseURL != "http://test" {
		t.Error("BaseURL not set")
	}
	if cfg.Timeout != 5*time.Second {
		t.Error("Timeout not set")
	}
	if cfg.MaxRetries != 3 {
		t.Error("MaxRetries not set")
	}
	if cfg.RetryDelay != time.Second {
		t.Error("RetryDelay not set")
	}
}
