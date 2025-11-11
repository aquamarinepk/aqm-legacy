package middleware

import (
	"net"
	"net/http"
	"strings"
)

// InternalOnly returns a middleware that restricts access to requests from
// localhost and RFC1918 private networks (10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16).
//
// This middleware is NOT part of the default stack and must be explicitly added.
// Use it for internal API services that should only accept requests from:
// - localhost (development)
// - other services in the same cluster/network (production)
//
// This provides defense-in-depth alongside network policies (Nomad/K8s/firewall).
//
// Example usage:
//
//	stack := middleware.DefaultStack(middleware.StackOptions{Logger: logger})
//	stack = append(stack, middleware.InternalOnly())
func InternalOnly() func(http.Handler) http.Handler {
	allowedNetworks := []*net.IPNet{
		parseCIDR("127.0.0.0/8"),    // localhost
		parseCIDR("10.0.0.0/8"),     // RFC1918 private
		parseCIDR("172.16.0.0/12"),  // RFC1918 private
		parseCIDR("192.168.0.0/16"), // RFC1918 private
		parseCIDR("::1/128"),        // IPv6 localhost
		parseCIDR("fc00::/7"),       // IPv6 unique local
	}
	return AllowFromNetworks(allowedNetworks...)
}

// AllowFromNetworks returns a middleware that restricts access to requests
// originating from the specified CIDR networks.
//
// This middleware respects X-Forwarded-For and X-Real-IP headers when present,
// checking the originating client IP rather than just the immediate connection.
//
// Example usage:
//
//	stack = append(stack, middleware.AllowFromNetworks(
//		parseCIDR("10.1.0.0/16"),
//		parseCIDR("192.168.1.0/24"),
//	))
func AllowFromNetworks(networks ...*net.IPNet) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := extractClientIP(r)
			if clientIP == nil {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			allowed := false
			for _, network := range networks {
				if network.Contains(clientIP) {
					allowed = true
					break
				}
			}

			if !allowed {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// extractClientIP extracts the real client IP from the request, checking
// X-Forwarded-For and X-Real-IP headers before falling back to RemoteAddr.
func extractClientIP(r *http.Request) net.IP {
	// Check X-Forwarded-For (can be a comma-separated list)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		// Take the first (leftmost) IP as the original client
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if parsed := net.ParseIP(ip); parsed != nil {
				return parsed
			}
		}
	}

	// Check X-Real-IP
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		if parsed := net.ParseIP(xri); parsed != nil {
			return parsed
		}
	}

	// Fall back to RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// RemoteAddr might not have a port
		host = r.RemoteAddr
	}
	return net.ParseIP(host)
}

// parseCIDR is a helper that panics on invalid CIDR (for compile-time constants).
func parseCIDR(cidr string) *net.IPNet {
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		panic("invalid CIDR: " + cidr)
	}
	return network
}
