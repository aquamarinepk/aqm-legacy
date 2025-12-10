package aqm

import (
	"testing"
	"time"
)

func TestNewConfig(t *testing.T) {
	cfg := NewConfig()
	if cfg == nil {
		t.Fatal("NewConfig() returned nil")
	}
	if cfg.values == nil {
		t.Error("expected values map to be initialized")
	}
}

func TestConfigSetAndGet(t *testing.T) {
	cfg := NewConfig()
	cfg.Set("test.key", "value")

	val, ok := cfg.Get("test.key")
	if !ok {
		t.Error("expected key to exist")
	}
	if val != "value" {
		t.Errorf("expected 'value', got %v", val)
	}
}

func TestConfigGetNotFound(t *testing.T) {
	cfg := NewConfig()
	_, ok := cfg.Get("nonexistent")
	if ok {
		t.Error("expected key not to exist")
	}
}

func TestConfigClone(t *testing.T) {
	cfg := NewConfig()
	cfg.Set("key", "value")

	clone := cfg.Clone()
	if clone == cfg {
		t.Error("clone should be a different instance")
	}

	val, ok := clone.Get("key")
	if !ok || val != "value" {
		t.Error("clone should have same values")
	}

	// modify original
	cfg.Set("key", "modified")
	val, _ = clone.Get("key")
	if val != "value" {
		t.Error("clone should not be affected by original modifications")
	}
}

func TestConfigMergeFlat(t *testing.T) {
	cfg := NewConfig()
	cfg.MergeFlat(map[string]any{
		"a.b": "value1",
		"c.d": "value2",
	})

	val, ok := cfg.Get("a.b")
	if !ok || val != "value1" {
		t.Errorf("expected 'value1', got %v", val)
	}
}

func TestConfigMergeFlatEmpty(t *testing.T) {
	cfg := NewConfig()
	cfg.MergeFlat(nil)
	cfg.MergeFlat(map[string]any{})
	// should not panic
}

func TestConfigMergeNested(t *testing.T) {
	cfg := NewConfig()
	cfg.MergeNested(map[string]any{
		"server": map[string]any{
			"port": 8080,
			"host": "localhost",
		},
	})

	val, ok := cfg.Get("server.port")
	if !ok {
		t.Error("expected server.port to exist")
	}
	if val != 8080 {
		t.Errorf("expected 8080, got %v", val)
	}
}

func TestConfigMergeYAML(t *testing.T) {
	cfg := NewConfig()
	yaml := `
server:
  port: 9090
  host: example.com
`
	err := cfg.MergeYAML([]byte(yaml))
	if err != nil {
		t.Fatalf("MergeYAML error: %v", err)
	}

	val, ok := cfg.Get("server.port")
	if !ok {
		t.Error("expected server.port to exist")
	}
	if val != 9090 {
		t.Errorf("expected 9090, got %v", val)
	}
}

func TestConfigMergeYAMLInvalid(t *testing.T) {
	cfg := NewConfig()
	err := cfg.MergeYAML([]byte("invalid: yaml: content:"))
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestConfigGetString(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		expected string
		ok       bool
	}{
		{"string", "hello", "hello", true},
		{"bytes", []byte("world"), "world", true},
		{"int", 42, "42", true},
		{"notFound", nil, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewConfig()
			if tt.value != nil {
				cfg.Set("key", tt.value)
			}
			val, ok := cfg.GetString("key")
			if ok != tt.ok {
				t.Errorf("GetString ok = %v, want %v", ok, tt.ok)
			}
			if ok && val != tt.expected {
				t.Errorf("GetString = %q, want %q", val, tt.expected)
			}
		})
	}
}

func TestConfigGetInt(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		expected  int
		ok        bool
		expectErr bool
	}{
		{"int", 42, 42, true, false},
		{"int64", int64(100), 100, true, false},
		{"uint64", uint64(200), 200, true, false},
		{"float64", float64(300.0), 300, true, false},
		{"string", "500", 500, true, false},
		{"invalidString", "notanumber", 0, true, true},
		{"notFound", nil, 0, false, false},
		{"invalidType", []int{1, 2}, 0, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewConfig()
			if tt.value != nil {
				cfg.Set("key", tt.value)
			}
			val, ok, err := cfg.GetInt("key")
			if ok != tt.ok {
				t.Errorf("GetInt ok = %v, want %v", ok, tt.ok)
			}
			if tt.expectErr && err == nil {
				t.Error("expected error")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if ok && err == nil && val != tt.expected {
				t.Errorf("GetInt = %d, want %d", val, tt.expected)
			}
		})
	}
}

func TestConfigGetFloat64(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		expected  float64
		ok        bool
		expectErr bool
	}{
		{"float64", 3.14, 3.14, true, false},
		{"float32", float32(2.5), 2.5, true, false},
		{"int", 10, 10.0, true, false},
		{"int64", int64(20), 20.0, true, false},
		{"string", "1.5", 1.5, true, false},
		{"invalidString", "notanumber", 0, true, true},
		{"notFound", nil, 0, false, false},
		{"invalidType", []int{1}, 0, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewConfig()
			if tt.value != nil {
				cfg.Set("key", tt.value)
			}
			val, ok, err := cfg.GetFloat64("key")
			if ok != tt.ok {
				t.Errorf("GetFloat64 ok = %v, want %v", ok, tt.ok)
			}
			if tt.expectErr && err == nil {
				t.Error("expected error")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if ok && err == nil && val != tt.expected {
				t.Errorf("GetFloat64 = %f, want %f", val, tt.expected)
			}
		})
	}
}

func TestConfigGetDuration(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		expected  time.Duration
		ok        bool
		expectErr bool
	}{
		{"duration", time.Second, time.Second, true, false},
		{"string", "5s", 5 * time.Second, true, false},
		{"int64", int64(1000000000), time.Second, true, false},
		{"invalidString", "invalid", 0, true, true},
		{"notFound", nil, 0, false, false},
		{"invalidType", []int{1}, 0, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewConfig()
			if tt.value != nil {
				cfg.Set("key", tt.value)
			}
			val, ok, err := cfg.GetDuration("key")
			if ok != tt.ok {
				t.Errorf("GetDuration ok = %v, want %v", ok, tt.ok)
			}
			if tt.expectErr && err == nil {
				t.Error("expected error")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if ok && err == nil && val != tt.expected {
				t.Errorf("GetDuration = %v, want %v", val, tt.expected)
			}
		})
	}
}

func TestConfigGetBool(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		expected  bool
		ok        bool
		expectErr bool
	}{
		{"true", true, true, true, false},
		{"false", false, false, true, false},
		{"stringTrue", "true", true, true, false},
		{"stringFalse", "false", false, true, false},
		{"int1", 1, true, true, false},
		{"int0", 0, false, true, false},
		{"int64Nonzero", int64(5), true, true, false},
		{"uint64Zero", uint64(0), false, true, false},
		{"float64Nonzero", float64(1.5), true, true, false},
		{"invalidString", "notbool", false, true, true},
		{"notFound", nil, false, false, false},
		{"invalidType", []int{1}, false, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewConfig()
			if tt.value != nil {
				cfg.Set("key", tt.value)
			}
			val, ok, err := cfg.GetBool("key")
			if ok != tt.ok {
				t.Errorf("GetBool ok = %v, want %v", ok, tt.ok)
			}
			if tt.expectErr && err == nil {
				t.Error("expected error")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if ok && err == nil && val != tt.expected {
				t.Errorf("GetBool = %v, want %v", val, tt.expected)
			}
		})
	}
}

func TestConfigGetStringSlice(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		expected []string
		ok       bool
	}{
		{"stringSlice", []string{"a", "b"}, []string{"a", "b"}, true},
		{"anySlice", []any{"x", "y"}, []string{"x", "y"}, true},
		{"commaSeparated", "a,b,c", []string{"a", "b", "c"}, true},
		{"emptyString", "", nil, true},
		{"singleValue", 123, []string{"123"}, true},
		{"notFound", nil, nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewConfig()
			if tt.value != nil {
				cfg.Set("key", tt.value)
			}
			val, ok := cfg.GetStringSlice("key")
			if ok != tt.ok {
				t.Errorf("GetStringSlice ok = %v, want %v", ok, tt.ok)
			}
			if ok && len(val) != len(tt.expected) {
				t.Errorf("GetStringSlice len = %d, want %d", len(val), len(tt.expected))
			}
		})
	}
}

func TestConfigGetOrDefFunctions(t *testing.T) {
	cfg := NewConfig()
	cfg.Set("existing", "value")
	cfg.Set("intKey", 42)
	cfg.Set("floatKey", 3.14)
	cfg.Set("durationKey", "5s")
	cfg.Set("sliceKey", "a,b")

	if cfg.GetStringOrDef("existing", "default") != "value" {
		t.Error("GetStringOrDef should return existing value")
	}
	if cfg.GetStringOrDef("missing", "default") != "default" {
		t.Error("GetStringOrDef should return default for missing key")
	}

	if cfg.GetIntOrDef("intKey", 0) != 42 {
		t.Error("GetIntOrDef should return existing value")
	}
	if cfg.GetIntOrDef("missing", 100) != 100 {
		t.Error("GetIntOrDef should return default for missing key")
	}

	if cfg.GetFloat64OrDef("floatKey", 0) != 3.14 {
		t.Error("GetFloat64OrDef should return existing value")
	}
	if cfg.GetFloat64OrDef("missing", 1.0) != 1.0 {
		t.Error("GetFloat64OrDef should return default for missing key")
	}

	if cfg.GetDurationOrDef("durationKey", 0) != 5*time.Second {
		t.Error("GetDurationOrDef should return existing value")
	}
	if cfg.GetDurationOrDef("missing", time.Minute) != time.Minute {
		t.Error("GetDurationOrDef should return default for missing key")
	}

	slice := cfg.GetStringSliceOrDef("sliceKey", nil)
	if len(slice) != 2 {
		t.Error("GetStringSliceOrDef should return existing value")
	}
	defSlice := cfg.GetStringSliceOrDef("missing", []string{"default"})
	if len(defSlice) != 1 || defSlice[0] != "default" {
		t.Error("GetStringSliceOrDef should return default for missing key")
	}
}

func TestConfigGetBoolOrTrue(t *testing.T) {
	cfg := NewConfig()
	cfg.Set("true", true)
	cfg.Set("false", false)

	if !cfg.GetBoolOrTrue("true") {
		t.Error("GetBoolOrTrue should return true")
	}
	if cfg.GetBoolOrTrue("false") {
		t.Error("GetBoolOrTrue should return false")
	}
	if !cfg.GetBoolOrTrue("missing") {
		t.Error("GetBoolOrTrue should return true for missing key")
	}
}

func TestConfigGetBoolOrFalse(t *testing.T) {
	cfg := NewConfig()
	cfg.Set("true", true)
	cfg.Set("false", false)

	if !cfg.GetBoolOrFalse("true") {
		t.Error("GetBoolOrFalse should return true")
	}
	if cfg.GetBoolOrFalse("false") {
		t.Error("GetBoolOrFalse should return false")
	}
	if cfg.GetBoolOrFalse("missing") {
		t.Error("GetBoolOrFalse should return false for missing key")
	}
}

func TestConfigGetPort(t *testing.T) {
	cfg := NewConfig()
	cfg.Set("port", "9090")

	port := cfg.GetPort("port", ":8080")
	if port != ":9090" {
		t.Errorf("GetPort = %q, want :9090", port)
	}

	port = cfg.GetPort("missing", ":8080")
	if port != ":8080" {
		t.Errorf("GetPort = %q, want :8080", port)
	}
}

func TestConfigUnmarshal(t *testing.T) {
	cfg := NewConfig()
	cfg.Set("server.port", 8080)
	cfg.Set("server.host", "localhost")

	type ServerConfig struct {
		Port int    `koanf:"port"`
		Host string `koanf:"host"`
	}

	var sc ServerConfig
	err := cfg.Unmarshal("server", &sc)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if sc.Port != 8080 {
		t.Errorf("Port = %d, want 8080", sc.Port)
	}
	if sc.Host != "localhost" {
		t.Errorf("Host = %q, want localhost", sc.Host)
	}
}

func TestConfigUnmarshalNilTarget(t *testing.T) {
	cfg := NewConfig()
	err := cfg.Unmarshal("", nil)
	if err == nil {
		t.Error("expected error for nil target")
	}
}

func TestConfigUnmarshalEmptyPath(t *testing.T) {
	cfg := NewConfig()
	cfg.Set("key", "value")

	var result map[string]any
	err := cfg.Unmarshal("", &result)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
}

func TestConfigUnmarshalNonexistentPath(t *testing.T) {
	cfg := NewConfig()
	cfg.Set("other.key", "value")

	var result map[string]any
	err := cfg.Unmarshal("nonexistent", &result)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
}

func TestConfigNormalise(t *testing.T) {
	cfg := NewConfig()
	cfg.Set("Test.Key", "value1")
	cfg.Set("  UPPER.case  ", "value2")

	val, ok := cfg.Get("test.key")
	if !ok || val != "value1" {
		t.Error("key normalization failed for Test.Key")
	}

	val, ok = cfg.Get("upper.case")
	if !ok || val != "value2" {
		t.Error("key normalization failed for UPPER.case")
	}
}
