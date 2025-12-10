package middleware

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInternalOnly(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := InternalOnly()
	wrapped := middleware(handler)

	tests := []struct {
		name       string
		remoteAddr string
		wantStatus int
	}{
		{"localhost IPv4", "127.0.0.1:12345", http.StatusOK},
		{"localhost range", "127.0.0.5:12345", http.StatusOK},
		{"private 10.x", "10.0.0.1:12345", http.StatusOK},
		{"private 172.16.x", "172.16.0.1:12345", http.StatusOK},
		{"private 192.168.x", "192.168.1.1:12345", http.StatusOK},
		{"public IP", "8.8.8.8:12345", http.StatusForbidden},
		{"public IP 2", "203.0.113.1:12345", http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = tt.remoteAddr
			rec := httptest.NewRecorder()

			wrapped.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("Status = %d, want %d for %s", rec.Code, tt.wantStatus, tt.remoteAddr)
			}
		})
	}
}

func TestInternalOnlyIPv6(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := InternalOnly()
	wrapped := middleware(handler)

	tests := []struct {
		name       string
		remoteAddr string
		wantStatus int
	}{
		{"IPv6 localhost", "[::1]:12345", http.StatusOK},
		{"IPv6 unique local", "[fc00::1]:12345", http.StatusOK},
		{"IPv6 public", "[2001:db8::1]:12345", http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = tt.remoteAddr
			rec := httptest.NewRecorder()

			wrapped.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("Status = %d, want %d for %s", rec.Code, tt.wantStatus, tt.remoteAddr)
			}
		})
	}
}

func TestAllowFromNetworks(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	network := parseCIDR("10.1.0.0/16")
	middleware := AllowFromNetworks(network)
	wrapped := middleware(handler)

	tests := []struct {
		name       string
		remoteAddr string
		wantStatus int
	}{
		{"in network", "10.1.0.1:12345", http.StatusOK},
		{"in network 2", "10.1.255.255:12345", http.StatusOK},
		{"outside network", "10.2.0.1:12345", http.StatusForbidden},
		{"different network", "192.168.1.1:12345", http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = tt.remoteAddr
			rec := httptest.NewRecorder()

			wrapped.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("Status = %d, want %d for %s", rec.Code, tt.wantStatus, tt.remoteAddr)
			}
		})
	}
}

func TestAllowFromNetworksMultiple(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	network1 := parseCIDR("10.1.0.0/16")
	network2 := parseCIDR("192.168.0.0/24")
	middleware := AllowFromNetworks(network1, network2)
	wrapped := middleware(handler)

	tests := []struct {
		name       string
		remoteAddr string
		wantStatus int
	}{
		{"first network", "10.1.0.1:12345", http.StatusOK},
		{"second network", "192.168.0.1:12345", http.StatusOK},
		{"neither network", "172.16.0.1:12345", http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = tt.remoteAddr
			rec := httptest.NewRecorder()

			wrapped.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("Status = %d, want %d for %s", rec.Code, tt.wantStatus, tt.remoteAddr)
			}
		})
	}
}

func TestExtractClientIPXForwardedFor(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.1, 10.0.0.1")
	req.RemoteAddr = "127.0.0.1:12345"

	ip := extractClientIP(req)

	expected := net.ParseIP("203.0.113.1")
	if !ip.Equal(expected) {
		t.Errorf("extractClientIP = %v, want %v", ip, expected)
	}
}

func TestExtractClientIPXRealIP(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Real-IP", "203.0.113.2")
	req.RemoteAddr = "127.0.0.1:12345"

	ip := extractClientIP(req)

	expected := net.ParseIP("203.0.113.2")
	if !ip.Equal(expected) {
		t.Errorf("extractClientIP = %v, want %v", ip, expected)
	}
}

func TestExtractClientIPRemoteAddr(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.100:54321"

	ip := extractClientIP(req)

	expected := net.ParseIP("192.168.1.100")
	if !ip.Equal(expected) {
		t.Errorf("extractClientIP = %v, want %v", ip, expected)
	}
}

func TestExtractClientIPRemoteAddrNoPort(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.100"

	ip := extractClientIP(req)

	expected := net.ParseIP("192.168.1.100")
	if !ip.Equal(expected) {
		t.Errorf("extractClientIP = %v, want %v", ip, expected)
	}
}

func TestExtractClientIPInvalidXFF(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "invalid-ip")
	req.RemoteAddr = "192.168.1.100:12345"

	ip := extractClientIP(req)

	expected := net.ParseIP("192.168.1.100")
	if !ip.Equal(expected) {
		t.Errorf("extractClientIP = %v, want %v", ip, expected)
	}
}

func TestExtractClientIPInvalidXRealIP(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Real-IP", "not-an-ip")
	req.RemoteAddr = "192.168.1.100:12345"

	ip := extractClientIP(req)

	expected := net.ParseIP("192.168.1.100")
	if !ip.Equal(expected) {
		t.Errorf("extractClientIP = %v, want %v", ip, expected)
	}
}

func TestExtractClientIPNilResult(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "invalid"

	ip := extractClientIP(req)

	if ip != nil {
		t.Errorf("extractClientIP = %v, want nil", ip)
	}
}

func TestAllowFromNetworksNilIP(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	network := parseCIDR("10.0.0.0/8")
	middleware := AllowFromNetworks(network)
	wrapped := middleware(handler)

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "invalid-address"
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("Status = %d, want %d for invalid address", rec.Code, http.StatusForbidden)
	}
}

func TestParseCIDR(t *testing.T) {
	network := parseCIDR("10.0.0.0/8")
	if network == nil {
		t.Error("parseCIDR returned nil for valid CIDR")
	}
}

func TestParseCIDRPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for invalid CIDR")
		}
	}()

	parseCIDR("invalid")
}

func TestXForwardedForPriority(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.1")
	req.Header.Set("X-Real-IP", "203.0.113.2")
	req.RemoteAddr = "127.0.0.1:12345"

	ip := extractClientIP(req)

	// X-Forwarded-For should take priority
	expected := net.ParseIP("203.0.113.1")
	if !ip.Equal(expected) {
		t.Errorf("extractClientIP = %v, want %v (X-Forwarded-For should take priority)", ip, expected)
	}
}
