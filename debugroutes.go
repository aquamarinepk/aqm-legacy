package aqm

import (
	"encoding/json"
	"net/http"
	"reflect"
	"runtime"

	"github.com/go-chi/chi/v5"
)

// RouteInfo represents a single registered route for debugging purposes.
type RouteInfo struct {
	Method      string   `json:"method"`
	Pattern     string   `json:"pattern"`
	Middlewares []string `json:"middlewares,omitempty"`
}

// RegisterDebugRoutes exposes GET /debug/routes when enabled. The endpoint
// lists every route currently registered on the router.
func RegisterDebugRoutes(r chi.Router, enabled bool) {
	if !enabled || r == nil {
		return
	}

	r.Get("/debug/routes", func(w http.ResponseWriter, req *http.Request) {
		routes := enumerateRoutes(r)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(routes)
	})
}

func enumerateRoutes(r chi.Router) []RouteInfo {
	routes := make([]RouteInfo, 0)
	_ = chi.Walk(r, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		info := RouteInfo{Method: method, Pattern: route}
		if len(middlewares) > 0 {
			info.Middlewares = make([]string, 0, len(middlewares))
			for _, mw := range middlewares {
				info.Middlewares = append(info.Middlewares, middlewareName(mw))
			}
		}
		routes = append(routes, info)
		return nil
	})
	return routes
}

func middlewareName(mw func(http.Handler) http.Handler) string {
	if mw == nil {
		return "<nil>"
	}
	return runtimeFuncName(mw)
}

func runtimeFuncName(fn interface{}) string {
	if fn == nil {
		return "<nil>"
	}
	return runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
}
