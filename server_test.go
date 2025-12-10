package aqm

import (
	"testing"
)

func TestNormalizePort(t *testing.T) {
	tests := []struct {
		name     string
		port     string
		fallback string
		want     string
	}{
		{
			name:     "emptyPortAndFallback",
			port:     "",
			fallback: "",
			want:     ":8080",
		},
		{
			name:     "emptyPortWithFallback",
			port:     "",
			fallback: ":9090",
			want:     ":9090",
		},
		{
			name:     "portWithColon",
			port:     ":3000",
			fallback: ":8080",
			want:     ":3000",
		},
		{
			name:     "portWithoutColon",
			port:     "4000",
			fallback: ":8080",
			want:     ":4000",
		},
		{
			name:     "portWithHost",
			port:     "0.0.0.0:5000",
			fallback: ":8080",
			want:     "0.0.0.0:5000",
		},
		{
			name:     "fallbackWithoutColon",
			port:     "",
			fallback: "6000",
			want:     ":6000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizePort(tt.port, tt.fallback)
			if got != tt.want {
				t.Errorf("NormalizePort(%q, %q) = %q, want %q", tt.port, tt.fallback, got, tt.want)
			}
		})
	}
}

func TestServerOptsFields(t *testing.T) {
	opts := ServerOpts{Port: ":8080"}
	if opts.Port != ":8080" {
		t.Errorf("expected Port :8080, got %s", opts.Port)
	}
}
