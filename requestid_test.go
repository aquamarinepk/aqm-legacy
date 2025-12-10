package aqm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWithRequestID(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
		id   string
		want string
	}{
		{
			name: "validIDSet",
			ctx:  context.Background(),
			id:   "test-123",
			want: "test-123",
		},
		{
			name: "emptyID",
			ctx:  context.Background(),
			id:   "",
			want: "",
		},
		{
			name: "nilContext",
			ctx:  nil,
			id:   "test-123",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WithRequestID(tt.ctx, tt.id)
			if tt.ctx == nil {
				if result != nil {
					t.Error("expected nil context when input is nil")
				}
				return
			}
			got := RequestIDFrom(result)
			if got != tt.want {
				t.Errorf("RequestIDFrom() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRequestIDFrom(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
		want string
	}{
		{
			name: "nilContext",
			ctx:  nil,
			want: "",
		},
		{
			name: "contextWithoutID",
			ctx:  context.Background(),
			want: "",
		},
		{
			name: "contextWithID",
			ctx:  WithRequestID(context.Background(), "req-456"),
			want: "req-456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RequestIDFrom(tt.ctx)
			if got != tt.want {
				t.Errorf("RequestIDFrom() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRequestIDMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		existingHeader string
		expectNewID    bool
	}{
		{
			name:           "noExistingHeader",
			existingHeader: "",
			expectNewID:    true,
		},
		{
			name:           "existingHeader",
			existingHeader: "existing-id-789",
			expectNewID:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedID string
			handler := RequestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedID = RequestIDFrom(r.Context())
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.existingHeader != "" {
				req.Header.Set(RequestIDHeader, tt.existingHeader)
			}

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			respHeader := rec.Header().Get(RequestIDHeader)
			if respHeader == "" {
				t.Error("expected response header to be set")
			}

			if capturedID == "" {
				t.Error("expected request ID in context")
			}

			if !tt.expectNewID && capturedID != tt.existingHeader {
				t.Errorf("expected existing ID %q, got %q", tt.existingHeader, capturedID)
			}

			if tt.expectNewID && capturedID == "" {
				t.Error("expected new ID to be generated")
			}
		})
	}
}
