package aqm

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// RedirectNotFound configures the router to send unmatched routes to the provided target.
// Useful for web frontends where a custom fallback page is preferable to the default 404.
func RedirectNotFound(r chi.Router, target string) {
	if target == "" {
		target = "/"
	}

	r.NotFound(func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, target, http.StatusFound)
	})

	r.MethodNotAllowed(func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, target, http.StatusFound)
	})
}
