package aqm

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	cfg, err := LoadConfig("", nil)
	if err != nil {
		t.Fatalf("LoadConfig error: %v", err)
	}
	if cfg == nil {
		t.Fatal("LoadConfig returned nil")
	}
}

func TestLoadConfigWithArgs(t *testing.T) {
	args := []string{"--http.port=9090", "--debug"}
	cfg, err := LoadConfig("", args)
	if err != nil {
		t.Fatalf("LoadConfig error: %v", err)
	}

	val, ok := cfg.GetString("http.port")
	if !ok || val != "9090" {
		t.Errorf("expected http.port=9090, got %q", val)
	}

	val, ok = cfg.GetString("debug")
	if !ok || val != "true" {
		t.Errorf("expected debug=true, got %q", val)
	}
}

func TestLoadConfigWithEnv(t *testing.T) {
	os.Setenv("TEST_HTTP_PORT", "7070")
	defer os.Unsetenv("TEST_HTTP_PORT")

	cfg, err := LoadConfig("TEST_", nil)
	if err != nil {
		t.Fatalf("LoadConfig error: %v", err)
	}

	val, ok := cfg.GetString("http.port")
	if !ok || val != "7070" {
		t.Errorf("expected http.port=7070, got %q", val)
	}
}

func TestParseArgsToMap(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want map[string]any
	}{
		{
			name: "empty",
			args: nil,
			want: nil,
		},
		{
			name: "equalsSyntax",
			args: []string{"--key=value"},
			want: map[string]any{"key": "value"},
		},
		{
			name: "spaceSyntax",
			args: []string{"--key", "value"},
			want: map[string]any{"key": "value"},
		},
		{
			name: "boolFlag",
			args: []string{"--enabled"},
			want: map[string]any{"enabled": "true"},
		},
		{
			name: "underscoreToDotsEquals",
			args: []string{"--http_port=8080"},
			want: map[string]any{"http.port": "8080"},
		},
		{
			name: "underscoreToDotsSpace",
			args: []string{"--http_port", "8080"},
			want: map[string]any{"http.port": "8080"},
		},
		{
			name: "skipNonDash",
			args: []string{"notaflag", "--valid=true"},
			want: map[string]any{"valid": "true"},
		},
		{
			name: "skipShortDash",
			args: []string{"--", "--valid=true"},
			want: map[string]any{"valid": "true"},
		},
		{
			name: "multipleBoolFlags",
			args: []string{"--debug", "--verbose", "--help"},
			want: map[string]any{"debug": "true", "verbose": "true", "help": "true"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseArgsToMap(tt.args)
			if tt.want == nil {
				if got != nil {
					t.Errorf("expected nil, got %v", got)
				}
				return
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("key %q: got %v, want %v", k, got[k], v)
				}
			}
		})
	}
}

func TestFindConfigFile(t *testing.T) {
	// Test with no config file present (normal case)
	path, ok := findConfigFile()
	// This may or may not find a file depending on test environment
	// We just ensure it doesn't panic
	_ = path
	_ = ok
}

func TestLoadSourcesWithYAMLFile(t *testing.T) {
	// Create a temporary config file
	content := `
server:
  port: 5050
  name: testserver
`
	tmpfile, err := os.CreateTemp("", "config*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpfile.Close()

	// Test MergeYAMLFile
	cfg := NewConfig()
	err = cfg.MergeYAMLFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("MergeYAMLFile error: %v", err)
	}

	val, ok := cfg.Get("server.port")
	if !ok {
		t.Error("expected server.port to exist")
	}
	if val != 5050 {
		t.Errorf("expected 5050, got %v", val)
	}
}

func TestMergeYAMLFileNotFound(t *testing.T) {
	cfg := NewConfig()
	err := cfg.MergeYAMLFile("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}
