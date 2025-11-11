package runtime

import (
	"net/http"
	"time"

	"github.com/aquamarinepk/aqm"
	aqmmiddleware "github.com/aquamarinepk/aqm/middleware"
)

// MiddlewareConfig controls how the shared HTTP middleware stack behaves.
type MiddlewareConfig struct {
	Timeout             time.Duration
	CompressLevel       int
	AllowedContentTypes []string
}

// MiddlewareStack returns the default middleware bundle used across the
// orchestration services.
func MiddlewareStack(logger aqm.Logger, cfg MiddlewareConfig) []func(http.Handler) http.Handler {
	return aqmmiddleware.DefaultStack(aqmmiddleware.StackOptions{
		Logger:              logger,
		Timeout:             fallbackDuration(cfg.Timeout, 30*time.Second),
		CompressLevel:       fallbackInt(cfg.CompressLevel, 5),
		AllowedContentTypes: cfg.AllowedContentTypes,
	})
}

func fallbackDuration(value time.Duration, def time.Duration) time.Duration {
	if value <= 0 {
		return def
	}
	return value
}

func fallbackInt(value, def int) int {
	if value <= 0 {
		return def
	}
	return value
}
