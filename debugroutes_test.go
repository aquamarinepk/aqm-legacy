package aqm

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestRegisterDebugRoutesEnabled(t *testing.T) {
	r := chi.NewRouter()
	r.Get("/test", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	RegisterDebugRoutes(r, true)

	req := httptest.NewRequest(http.MethodGet, "/debug/routes", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var routes []RouteInfo
	if err := json.NewDecoder(rec.Body).Decode(&routes); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(routes) == 0 {
		t.Error("expected at least one route")
	}
}

func TestRegisterDebugRoutesDisabled(t *testing.T) {
	r := chi.NewRouter()
	RegisterDebugRoutes(r, false)

	req := httptest.NewRequest(http.MethodGet, "/debug/routes", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestRegisterDebugRoutesNilRouter(t *testing.T) {
	// should not panic
	RegisterDebugRoutes(nil, true)
}

func TestEnumerateRoutes(t *testing.T) {
	r := chi.NewRouter()
	r.Get("/users", func(w http.ResponseWriter, req *http.Request) {})
	r.Post("/users", func(w http.ResponseWriter, req *http.Request) {})
	r.Get("/users/{id}", func(w http.ResponseWriter, req *http.Request) {})

	routes := enumerateRoutes(r)

	if len(routes) < 3 {
		t.Errorf("expected at least 3 routes, got %d", len(routes))
	}

	var foundGet, foundPost bool
	for _, route := range routes {
		if route.Method == "GET" && route.Pattern == "/users" {
			foundGet = true
		}
		if route.Method == "POST" && route.Pattern == "/users" {
			foundPost = true
		}
	}

	if !foundGet {
		t.Error("expected GET /users route")
	}
	if !foundPost {
		t.Error("expected POST /users route")
	}
}

func TestMiddlewareName(t *testing.T) {
	name := middlewareName(nil)
	if name != "<nil>" {
		t.Errorf("expected '<nil>', got %q", name)
	}

	mw := func(next http.Handler) http.Handler { return next }
	name = middlewareName(mw)
	if name == "" || name == "<nil>" {
		t.Error("expected non-empty middleware name")
	}
}

func TestRuntimeFuncName(t *testing.T) {
	name := runtimeFuncName(nil)
	if name != "<nil>" {
		t.Errorf("expected '<nil>', got %q", name)
	}

	fn := func() {}
	name = runtimeFuncName(fn)
	if name == "" || name == "<nil>" {
		t.Error("expected non-empty function name")
	}
}

func TestRouteInfoFields(t *testing.T) {
	info := RouteInfo{
		Method:      "GET",
		Pattern:     "/test",
		Middlewares: []string{"mw1", "mw2"},
	}

	if info.Method != "GET" {
		t.Errorf("expected GET, got %s", info.Method)
	}
	if info.Pattern != "/test" {
		t.Errorf("expected /test, got %s", info.Pattern)
	}
	if len(info.Middlewares) != 2 {
		t.Errorf("expected 2 middlewares, got %d", len(info.Middlewares))
	}
}
