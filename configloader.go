package aqm

import (
	"fmt"
	"os"
	"strings"

	koanfyaml "github.com/knadh/koanf/parsers/yaml"
	confmap "github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

var defaultConfigPaths = []string{
	"config.yaml",
	"config.yml",
	"config/config.yaml",
	"config/config.yml",
	".config/config.yaml",
	".config/config.yml",
}

// LoadConfig builds a Config instance merging, in order, defaults, YAML file
// (if present), environment variables and CLI arguments. Environment variables
// are matched using the provided prefix, replacing underscores with dots and
// lower-casing the remainder (e.g. TODO_HTTP_PORT -> http.port). CLI arguments
// use a simple --key=value or --key value syntax with flags taking precedence.
func LoadConfig(envNamespace string, args []string) (*Config, error) {
	cfg := NewConfig()
	if err := cfg.LoadSources(envNamespace, args); err != nil {
		return nil, err
	}
	return cfg, nil
}

// LoadSources merges configuration from the default sources into the receiver.
// Sources are applied in the following order (later overrides earlier):
//  1. YAML file (first match among config/config.{yaml,yml}, .config/...)
//  2. Environment variables with the given prefix
//  3. CLI arguments in --key=value or --key value form
func (p *Config) LoadSources(envNamespace string, args []string) error {
	k := koanf.New(".")

	if path, ok := findConfigFile(); ok {
		if err := k.Load(file.Provider(path), koanfyaml.Parser()); err != nil {
			return fmt.Errorf("config: loading %s: %w", path, err)
		}
	}

	if envNamespace != "" {
		envPrefix := strings.ToUpper(strings.TrimSuffix(envNamespace, "_")) + "_"
		transform := func(s string) string {
			s = strings.TrimPrefix(s, envPrefix)
			s = strings.ReplaceAll(s, "_", ".")
			return strings.ToLower(s)
		}
		if err := k.Load(env.Provider(envPrefix, ".", transform), nil); err != nil {
			return fmt.Errorf("config: loading env: %w", err)
		}
	}

	if kv := parseArgsToMap(args); len(kv) > 0 {
		if err := k.Load(confmap.Provider(kv, "."), nil); err != nil {
			return fmt.Errorf("config: loading args: %w", err)
		}
	}

	raw := map[string]any{}
	if err := k.Unmarshal("", &raw); err != nil {
		return fmt.Errorf("config: unmarshal: %w", err)
	}
	p.MergeNested(raw)
	p.addAliasKeys()
	return nil
}

func findConfigFile() (string, bool) {
	for _, path := range defaultConfigPaths {
		if _, err := os.Stat(path); err == nil {
			return path, true
		}
	}
	return "", false
}

func parseArgsToMap(args []string) map[string]any {
	out := make(map[string]any)
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "--") || len(arg) <= 2 {
			continue
		}
		key := strings.TrimPrefix(arg, "--")
		if key == "" {
			continue
		}
		if strings.Contains(key, "=") {
			parts := strings.SplitN(key, "=", 2)
			mappedKey := strings.ReplaceAll(parts[0], "_", ".")
			out[mappedKey] = parts[1]
			continue
		}
		value := "true"
		if i+1 < len(args) {
			next := args[i+1]
			if !strings.HasPrefix(next, "--") {
				value = next
				i++
			}
		}
		mappedKey := strings.ReplaceAll(key, "_", ".")
		out[mappedKey] = value
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
