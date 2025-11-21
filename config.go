package aqm

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-viper/mapstructure/v2"
	koanfyaml "github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	yamlv3 "gopkg.in/yaml.v3"
)

// Config stores configuration values keyed by hierarchical property names.
type Config struct {
	mu     sync.RWMutex
	values map[string]any
}

// NewConfig constructs an empty property store.
func NewConfig() *Config {
	return &Config{values: make(map[string]any)}
}

// Clone returns a copy of the stored properties.
func (p *Config) Clone() *Config {
	p.mu.RLock()
	defer p.mu.RUnlock()

	cloned := make(map[string]any, len(p.values))
	for k, v := range p.values {
		cloned[k] = v
	}
	return &Config{values: cloned}
}

// Set persists a value under the provided property path.
func (p *Config) Set(path string, value any) {
	p.mu.Lock()
	p.values[normalise(path)] = value
	p.mu.Unlock()
}

// MergeFlat stores a batch of already flattened properties.
func (p *Config) MergeFlat(values map[string]any) {
	if len(values) == 0 {
		return
	}
	p.mu.Lock()
	for k, v := range values {
		p.values[normalise(k)] = v
	}
	p.mu.Unlock()
}

// MergeNested accepts nested maps (like those decoded from YAML) and flattens them.
func (p *Config) MergeNested(values map[string]any) {
	flattenInto(p, "", values)
}

// MergeYAML unmarshals the provided YAML payload and merges it into the
// property store.
func (p *Config) MergeYAML(data []byte) error {
	var raw map[string]any
	if err := yamlv3.Unmarshal(data, &raw); err != nil {
		return err
	}
	p.MergeNested(raw)
	return nil
}

// MergeYAMLFile reads a YAML file from disk and merges it into the property
// store.
func (p *Config) MergeYAMLFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return p.MergeYAML(data)
}

// MergeYAMLFileWithEnv loads a YAML file using koanf, then overlays environment
// variables matching the provided prefix. Environment variables are transformed
// to property keys by trimming the prefix, lowering case, and replacing
// underscores with dots (e.g. TODO_HTTP_PORT -> http.port).
func (p *Config) MergeYAMLFileWithEnv(path, envPrefix string) error {
	k := koanf.New(".")
	if err := k.Load(file.Provider(path), koanfyaml.Parser()); err != nil {
		return err
	}
	if envPrefix != "" {
		tf := func(s string) string {
			s = strings.TrimPrefix(s, envPrefix)
			s = strings.ReplaceAll(s, "_", ".")
			return strings.ToLower(s)
		}
		if err := k.Load(env.Provider(envPrefix, ".", tf), nil); err != nil {
			return err
		}
	}
	var raw map[string]any
	if err := k.Unmarshal("", &raw); err != nil {
		return err
	}
	p.MergeNested(raw)
	return nil
}

// Get retrieves a raw value by property path.
func (p *Config) Get(path string) (any, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	v, ok := p.values[normalise(path)]
	return v, ok
}

// GetString retrieves the value as a string.
func (p *Config) GetString(path string) (string, bool) {
	raw, ok := p.Get(path)
	if !ok {
		return "", false
	}
	switch v := raw.(type) {
	case string:
		return v, true
	case fmt.Stringer:
		return v.String(), true
	case []byte:
		return string(v), true
	default:
		return fmt.Sprintf("%v", raw), true
	}
}

// GetInt retrieves the value as an int.
func (p *Config) GetInt(path string) (int, bool, error) {
	raw, ok := p.Get(path)
	if !ok {
		return 0, false, nil
	}
	switch v := raw.(type) {
	case int:
		return v, true, nil
	case int64:
		return int(v), true, nil
	case uint64:
		return int(v), true, nil
	case float64:
		return int(v), true, nil
	case string:
		parsed, err := strconv.Atoi(v)
		return parsed, true, err
	default:
		return 0, true, fmt.Errorf("config: cannot convert %T to int", raw)
	}
}

// GetFloat64 retrieves the value as a float64.
func (p *Config) GetFloat64(path string) (float64, bool, error) {
	raw, ok := p.Get(path)
	if !ok {
		return 0, false, nil
	}
	switch v := raw.(type) {
	case float64:
		return v, true, nil
	case float32:
		return float64(v), true, nil
	case int:
		return float64(v), true, nil
	case int64:
		return float64(v), true, nil
	case string:
		parsed, err := strconv.ParseFloat(v, 64)
		return parsed, true, err
	default:
		return 0, true, fmt.Errorf("config: cannot convert %T to float64", raw)
	}
}

// GetDuration retrieves the value as a time.Duration.
func (p *Config) GetDuration(path string) (time.Duration, bool, error) {
	raw, ok := p.Get(path)
	if !ok {
		return 0, false, nil
	}
	switch v := raw.(type) {
	case time.Duration:
		return v, true, nil
	case string:
		d, err := time.ParseDuration(v)
		return d, true, err
	case int64:
		return time.Duration(v), true, nil
	default:
		return 0, true, fmt.Errorf("config: cannot convert %T to duration", raw)
	}
}

// GetStringSlice returns a slice of strings parsed from either []string or comma-separated string.
func (p *Config) GetStringSlice(path string) ([]string, bool) {
	raw, ok := p.Get(path)
	if !ok {
		return nil, false
	}
	switch v := raw.(type) {
	case []string:
		return append([]string(nil), v...), true
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			out = append(out, fmt.Sprint(item))
		}
		return out, true
	case string:
		if v == "" {
			return nil, true
		}
		parts := strings.Split(v, ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		return parts, true
	default:
		return []string{fmt.Sprint(v)}, true
 }
}

// GetStringOrDef retrieves the value as a string or returns def when not found.
func (p *Config) GetStringOrDef(path string, def string) string {
	if v, ok := p.GetString(path); ok {
		return v
	}
	return def
}

// GetIntOrDef retrieves the value as an int or returns def when not found or on conversion error.
func (p *Config) GetIntOrDef(path string, def int) int {
	if v, ok, err := p.GetInt(path); ok && err == nil {
		return v
	}
	return def
}

// GetFloat64OrDef retrieves the value as a float64 or returns def when not found or on conversion error.
func (p *Config) GetFloat64OrDef(path string, def float64) float64 {
	if v, ok, err := p.GetFloat64(path); ok && err == nil {
		return v
	}
	return def
}

// GetDurationOrDef retrieves the value as a time.Duration or returns def when not found or on conversion error.
func (p *Config) GetDurationOrDef(path string, def time.Duration) time.Duration {
	if v, ok, err := p.GetDuration(path); ok && err == nil {
		return v
	}
	return def
}

// GetStringSliceOrDef retrieves the value as a []string or returns def when not found.
func (p *Config) GetStringSliceOrDef(path string, def []string) []string {
	if v, ok := p.GetStringSlice(path); ok {
		return v
	}
	return def
}

// GetPort retrieves a port configuration and normalizes it, applying a default if not found.
// The defaultPort should be in the format ":8080" or "8080".
func (p *Config) GetPort(path string, defaultPort string) string {
	port, ok := p.GetString(path)
	if !ok {
		port = ""
	}
	return NormalizePort(port, defaultPort)
}

func flattenInto(p *Config, prefix string, values map[string]any) {
	for k, v := range values {
		var path string
		if prefix == "" {
			path = k
		} else {
			path = prefix + "." + k
		}

		switch nested := v.(type) {
		case map[string]any:
			flattenInto(p, path, nested)
		default:
			p.Set(path, nested)
		}
	}
}

func normalise(path string) string {
	segments := strings.Split(path, ".")
	for i := range segments {
		segments[i] = strings.ToLower(strings.TrimSpace(segments[i]))
	}
	return strings.Join(segments, ".")
}

// Unmarshal decodes the stored properties into the provided target struct using
// the "koanf" tag for field mapping. When path is non-empty, only that subtree
// is decoded.
func (p *Config) Unmarshal(path string, target any) error {
	if target == nil {
		return fmt.Errorf("config: nil target")
	}
	nested := p.snapshot()
	if path != "" {
		var ok bool
		nested, ok = walkNested(nested, path)
		if !ok {
			nested = map[string]any{}
		}
	}
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName:          "koanf",
		Result:           target,
		WeaklyTypedInput: true,
	})
	if err != nil {
		return fmt.Errorf("config: decoder: %w", err)
	}
	if err := decoder.Decode(nested); err != nil {
		return fmt.Errorf("config: decode: %w", err)
	}
	return nil
}

func (p *Config) snapshot() map[string]any {
	p.mu.RLock()
	defer p.mu.RUnlock()
	root := make(map[string]any)
	for key, value := range p.values {
		assignNested(root, strings.Split(key, "."), value)
	}
	return root
}

func assignNested(root map[string]any, parts []string, value any) {
	if len(parts) == 0 {
		return
	}
	head := parts[0]
	if len(parts) == 1 {
		root[head] = value
		return
	}
	next, ok := root[head].(map[string]any)
	if !ok {
		next = make(map[string]any)
		root[head] = next
	}
	assignNested(next, parts[1:], value)
}

func walkNested(root map[string]any, path string) (map[string]any, bool) {
	if len(root) == 0 {
		return map[string]any{}, false
	}
	current := root
	segments := strings.Split(strings.Trim(path, "."), ".")
	for _, segment := range segments {
		next, ok := current[segment]
		if !ok {
			return map[string]any{}, false
		}
		asMap, ok := next.(map[string]any)
		if !ok {
			return map[string]any{}, false
		}
		current = asMap
	}
	return current, true
}

func (p *Config) addAliasKeys() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for key, value := range p.values {
		if !strings.Contains(key, ".") {
			continue
		}
		parts := strings.Split(key, ".")
		current := append([]string(nil), parts...)
		for i := len(current) - 1; i > 0; i-- {
			current = mergeSegments(current, i-1)
			alias := strings.Join(current, ".")
			if _, exists := p.values[alias]; !exists {
				p.values[alias] = value
			}
		}
	}
}

func mergeSegments(parts []string, idx int) []string {
	merged := make([]string, 0, len(parts)-1)
	merged = append(merged, parts[:idx]...)
	merged = append(merged, parts[idx]+"_"+parts[idx+1])
	merged = append(merged, parts[idx+2:]...)
	return merged
}


// GetBool retrieves the value as a bool.
func (p *Config) GetBool(path string) (bool, bool, error) {
	raw, ok := p.Get(path)
	if !ok {
		return false, false, nil
	}
	switch v := raw.(type) {
	case bool:
		return v, true, nil
	case string:
		parsed, err := strconv.ParseBool(strings.TrimSpace(v))
		return parsed, true, err
	case int:
		return v != 0, true, nil
	case int64:
		return v != 0, true, nil
	case uint64:
		return v != 0, true, nil
	case float64:
		return v != 0, true, nil
	default:
		return false, true, fmt.Errorf("config: cannot convert %T to bool", raw)
	}
}

// GetBoolOrTrue retrieves the value as a bool or returns true when not found or on conversion error.
func (p *Config) GetBoolOrTrue(path string) bool {
	if v, ok, err := p.GetBool(path); ok && err == nil {
		return v
	}
	return true
}

// GetBoolOrFalse retrieves the value as a bool or returns false when not found or on conversion error.
func (p *Config) GetBoolOrFalse(path string) bool {
	if v, ok, err := p.GetBool(path); ok && err == nil {
		return v
	}
	return false
}
