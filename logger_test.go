package aqm

import (
	"log/slog"
	"os"
	"testing"
)

func TestNewLogger(t *testing.T) {
	tests := []string{"debug", "info", "error", "unknown"}

	for _, level := range tests {
		t.Run(level, func(t *testing.T) {
			logger := NewLogger(level)
			if logger == nil {
				t.Error("NewLogger returned nil")
			}
		})
	}
}

func TestNewStandardLogger(t *testing.T) {
	tests := []string{"debug", "info", "error", "unknown"}

	for _, level := range tests {
		t.Run(level, func(t *testing.T) {
			logger := NewStandardLogger(level)
			if logger == nil {
				t.Error("NewStandardLogger returned nil")
			}
		})
	}
}

func TestNewNoopLogger(t *testing.T) {
	logger := NewNoopLogger()
	if logger == nil {
		t.Fatal("NewNoopLogger returned nil")
	}

	// All methods should not panic
	logger.Debug("test")
	logger.Debugf("test %s", "arg")
	logger.Info("test")
	logger.Infof("test %s", "arg")
	logger.Error("test")
	logger.Errorf("test %s", "arg")
	logger.SetLogLevel(DebugLevel)

	child := logger.With("key", "value")
	if child == nil {
		t.Error("With returned nil")
	}
}

func TestSlogLoggerDebug(t *testing.T) {
	logger := NewLogger("debug").(*slogLogger)
	// should not panic
	logger.Debug("test message")
	logger.Debug("test message", "key", "value")
	logger.Debugf("test %s", "message")
}

func TestSlogLoggerDebugFiltered(t *testing.T) {
	logger := NewLogger("error").(*slogLogger)
	// debug should be filtered out at error level
	logger.Debug("test message")
	logger.Debugf("test %s", "message")
}

func TestSlogLoggerInfo(t *testing.T) {
	logger := NewLogger("info").(*slogLogger)
	// should not panic
	logger.Info("test message")
	logger.Info("test message", "key", "value")
	logger.Infof("test %s", "message")
}

func TestSlogLoggerInfoFiltered(t *testing.T) {
	logger := NewLogger("error").(*slogLogger)
	// info should be filtered out at error level
	logger.Info("test message")
	logger.Infof("test %s", "message")
}

func TestSlogLoggerError(t *testing.T) {
	logger := NewLogger("error").(*slogLogger)
	// should not panic
	logger.Error("test message")
	logger.Error("test message", "key", "value")
	logger.Errorf("test %s", "message")
}

func TestSlogLoggerSetLogLevel(t *testing.T) {
	logger := NewLogger("info").(*slogLogger)
	logger.SetLogLevel(DebugLevel)

	if logger.logLevel != DebugLevel {
		t.Errorf("expected DebugLevel, got %v", logger.logLevel)
	}
}

func TestSlogLoggerWith(t *testing.T) {
	logger := NewLogger("info").(*slogLogger)
	child := logger.With("key", "value")

	if child == nil {
		t.Error("With returned nil")
	}

	childLogger, ok := child.(*slogLogger)
	if !ok {
		t.Error("With should return *slogLogger")
	}
	if childLogger.logLevel != logger.logLevel {
		t.Error("child should inherit log level")
	}
}

func TestStandardLoggerDebug(t *testing.T) {
	logger := NewStandardLogger("debug").(*standardLogger)
	// should not panic
	logger.Debug("test message")
	logger.Debug("test message", "key", "value")
	logger.Debugf("test %s", "message")
}

func TestStandardLoggerDebugFiltered(t *testing.T) {
	logger := NewStandardLogger("error").(*standardLogger)
	// debug should be filtered out at error level
	logger.Debug("test message")
	logger.Debugf("test %s", "message")
}

func TestStandardLoggerInfo(t *testing.T) {
	logger := NewStandardLogger("info").(*standardLogger)
	// should not panic
	logger.Info("test message")
	logger.Info("test message", "key", "value")
	logger.Infof("test %s", "message")
}

func TestStandardLoggerInfoFiltered(t *testing.T) {
	logger := NewStandardLogger("error").(*standardLogger)
	// info should be filtered out at error level
	logger.Info("test message")
	logger.Infof("test %s", "message")
}

func TestStandardLoggerError(t *testing.T) {
	logger := NewStandardLogger("error").(*standardLogger)
	// should not panic
	logger.Error("test message")
	logger.Error("test message", "key", "value")
	logger.Errorf("test %s", "message")
}

func TestStandardLoggerSetLogLevel(t *testing.T) {
	logger := NewStandardLogger("info").(*standardLogger)
	logger.SetLogLevel(DebugLevel)

	if logger.logLevel != DebugLevel {
		t.Errorf("expected DebugLevel, got %v", logger.logLevel)
	}
}

func TestStandardLoggerWith(t *testing.T) {
	logger := NewStandardLogger("info").(*standardLogger)

	child := logger.With("key", "value")
	if child == nil {
		t.Error("With returned nil")
	}

	childLogger := child.(*standardLogger)
	if childLogger.prefix == "" {
		t.Error("child should have prefix")
	}

	// Test with empty args
	child2 := logger.With()
	if child2 == nil {
		t.Error("With empty args returned nil")
	}
}

func TestStandardLoggerWithPrefix(t *testing.T) {
	logger := NewStandardLogger("info").(*standardLogger)
	logger.prefix = "existing"

	child := logger.With("key", "value")
	childLogger := child.(*standardLogger)

	if childLogger.prefix == "" {
		t.Error("child should have combined prefix")
	}
}

func TestToValidLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected LogLevel
	}{
		{"debug", DebugLevel},
		{"dbg", DebugLevel},
		{"DEBUG", DebugLevel},
		{"info", InfoLevel},
		{"inf", InfoLevel},
		{"INFO", InfoLevel},
		{"error", ErrorLevel},
		{"err", ErrorLevel},
		{"ERROR", ErrorLevel},
		{"unknown", InfoLevel},
		{"", InfoLevel},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := toValidLevel(tt.input)
			if got != tt.expected {
				t.Errorf("toValidLevel(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestSlogLevel(t *testing.T) {
	tests := []struct {
		input    LogLevel
		expected slog.Level
	}{
		{DebugLevel, slog.LevelDebug},
		{InfoLevel, slog.LevelInfo},
		{ErrorLevel, slog.LevelError},
		{LogLevel(99), slog.LevelInfo}, // unknown defaults to info
	}

	for _, tt := range tests {
		t.Run(string(rune('0'+tt.input)), func(t *testing.T) {
			got := slogLevel(tt.input)
			if got != tt.expected {
				t.Errorf("slogLevel(%v) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestIsTerminal(t *testing.T) {
	original := os.Getenv("LOG_FORMAT")
	defer os.Setenv("LOG_FORMAT", original)

	os.Setenv("LOG_FORMAT", "json")
	if isTerminal() {
		t.Error("expected false when LOG_FORMAT=json")
	}

	os.Unsetenv("LOG_FORMAT")
	if !isTerminal() {
		t.Error("expected true when LOG_FORMAT is not set")
	}
}

func TestNewRequestLogger(t *testing.T) {
	logger := NewLogger("info")
	middleware := NewRequestLogger(logger)

	if middleware == nil {
		t.Error("NewRequestLogger returned nil")
	}
}

func TestNewRequestLoggerNil(t *testing.T) {
	middleware := NewRequestLogger(nil)
	if middleware == nil {
		t.Error("NewRequestLogger with nil should return middleware")
	}
}

func TestFormatKeyValueAttrs(t *testing.T) {
	tests := []struct {
		name string
		args []any
		want string
	}{
		{
			name: "empty",
			args: nil,
			want: "",
		},
		{
			name: "slogAttrs",
			args: []any{slog.String("key", "value")},
			want: "key=value",
		},
		{
			name: "keyValuePairs",
			args: []any{"key1", "value1", "key2", "value2"},
			want: "key1=value1 key2=value2",
		},
		{
			name: "mixedTypes",
			args: []any{123, 456},
			want: "123 456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatKeyValueAttrs(tt.args)
			if got != tt.want {
				t.Errorf("formatKeyValueAttrs() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNormalizeArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     []any
		wantMsg  string
		wantLen  int
	}{
		{
			name:    "empty",
			args:    nil,
			wantMsg: "",
			wantLen: 0,
		},
		{
			name:    "singleMessage",
			args:    []any{"message"},
			wantMsg: "message",
			wantLen: 0,
		},
		{
			name:    "messageWithAttrs",
			args:    []any{"message", "key", "value"},
			wantMsg: "message",
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, attrs := normalizeArgs(tt.args...)
			if msg != tt.wantMsg {
				t.Errorf("normalizeArgs() msg = %q, want %q", msg, tt.wantMsg)
			}
			if len(attrs) != tt.wantLen {
				t.Errorf("normalizeArgs() attrs len = %d, want %d", len(attrs), tt.wantLen)
			}
		})
	}
}

func TestToAttrsIfPossible(t *testing.T) {
	tests := []struct {
		name     string
		args     []any
		wantNil  bool
	}{
		{
			name:    "allSlogAttrs",
			args:    []any{slog.String("key", "value")},
			wantNil: false,
		},
		{
			name:    "mixedTypes",
			args:    []any{slog.String("key", "value"), "string"},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toAttrsIfPossible(tt.args)
			if (got == nil) != tt.wantNil {
				t.Errorf("toAttrsIfPossible() nil = %v, want %v", got == nil, tt.wantNil)
			}
		})
	}
}

func TestLogLevelConstants(t *testing.T) {
	if DebugLevel != 0 {
		t.Errorf("DebugLevel = %v, want 0", DebugLevel)
	}
	if InfoLevel != 1 {
		t.Errorf("InfoLevel = %v, want 1", InfoLevel)
	}
	if ErrorLevel != 2 {
		t.Errorf("ErrorLevel = %v, want 2", ErrorLevel)
	}
}
