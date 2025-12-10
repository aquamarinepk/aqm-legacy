package aqm

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestRedirectNotFound(t *testing.T) {
	tests := []struct {
		name           string
		target         string
		requestPath    string
		expectedTarget string
	}{
		{
			name:           "customTarget",
			target:         "/home",
			requestPath:    "/nonexistent",
			expectedTarget: "/home",
		},
		{
			name:           "emptyTargetDefaultsToRoot",
			target:         "",
			requestPath:    "/missing",
			expectedTarget: "/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()
			RedirectNotFound(r, tt.target)

			req := httptest.NewRequest(http.MethodGet, tt.requestPath, nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			if rec.Code != http.StatusFound {
				t.Errorf("expected status %d, got %d", http.StatusFound, rec.Code)
			}

			location := rec.Header().Get("Location")
			if location != tt.expectedTarget {
				t.Errorf("expected redirect to %q, got %q", tt.expectedTarget, location)
			}
		})
	}
}

func TestRedirectMethodNotAllowed(t *testing.T) {
	r := chi.NewRouter()
	r.Get("/only-get", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	RedirectNotFound(r, "/fallback")

	req := httptest.NewRequest(http.MethodPost, "/only-get", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("expected status %d, got %d", http.StatusFound, rec.Code)
	}

	location := rec.Header().Get("Location")
	if location != "/fallback" {
		t.Errorf("expected redirect to /fallback, got %q", location)
	}
}
