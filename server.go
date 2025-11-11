package aqm

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
)

// Serve starts an HTTP server and handles graceful shutdown.
// It also calls the provided stops functions during shutdown.
func Serve(router *chi.Mux, opts ServerOpts, stops []func(context.Context) error, log Logger) {
	srv := &http.Server{
		Addr:    opts.Port,
		Handler: router,
	}

	// Initializing the server in a goroutine so that it won't block the graceful shutdown handling below
	go func() {
		log.Info(fmt.Sprintf("Starting server on %s...", opts.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error(fmt.Sprintf("could not listen and serve: %v", err))
		}
	}()

	// Listen for OS signals to perform a graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	Shutdown(context.Background(), srv, log, stops)
}

// ServerOpts holds server-related options.
type ServerOpts struct {
	Port string
}

// NormalizePort ensures ports always include the leading colon and fall back to
// a sensible default when unset.
func NormalizePort(port, fallback string) string {
	p := port
	if p == "" {
		p = fallback
	}
	if p == "" {
		return ":8080"
	}
	if len(p) > 0 && p[0] == ':' {
		return p
	}
	// Check if it already contains a colon (like "0.0.0.0:8080")
	for i := 0; i < len(p); i++ {
		if p[i] == ':' {
			return p
		}
	}
	return ":" + p
}
