package aqm

import (
	"context"
	"testing"
	"time"
)

func TestDefaultDeps(t *testing.T) {
	deps := DefaultDeps()

	if deps == nil {
		t.Fatal("DefaultDeps() returned nil")
	}
	if deps.Metrics == nil {
		t.Error("expected Metrics to be set")
	}
	if deps.Tracer == nil {
		t.Error("expected Tracer to be set")
	}
	if deps.Errors == nil {
		t.Error("expected Errors to be set")
	}
	if deps.Validator == nil {
		t.Error("expected Validator to be set")
	}
	if deps.PubSub == nil {
		t.Error("expected PubSub to be set")
	}
}

func TestNoopMetrics(t *testing.T) {
	m := NoopMetrics{}
	// should not panic
	m.Counter(context.Background(), "test", 1.0, nil)
	m.Counter(context.Background(), "test", 1.0, map[string]string{"key": "value"})
	m.ObserveHTTPRequest("/test", "GET", 200, time.Second)
}

func TestNoopTracer(t *testing.T) {
	tracer := NoopTracer{}
	ctx, span := tracer.Start(context.Background(), "test", nil)

	if ctx == nil {
		t.Error("expected context to be returned")
	}
	if span == nil {
		t.Error("expected span to be returned")
	}

	// should not panic
	span.End(nil)
}

func TestNoopSpan(t *testing.T) {
	span := NoopSpan{}
	// should not panic
	span.End(nil)
}

func TestNoopPubSub(t *testing.T) {
	ps := NoopPubSub{}
	err := ps.Publish(context.Background(), "subject", []byte("payload"))
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}
