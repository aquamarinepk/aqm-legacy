package seed

import (
	"context"
	"errors"
	"testing"
)

type fakeTracker struct {
	ran      map[string]bool
	errQuery error
	errMark  error
}

func newFakeTracker() *fakeTracker {
	return &fakeTracker{ran: make(map[string]bool)}
}

func (f *fakeTracker) HasRun(_ context.Context, id string) (bool, error) {
	if f.errQuery != nil {
		return false, f.errQuery
	}
	return f.ran[id], nil
}

func (f *fakeTracker) MarkRun(_ context.Context, record Record) error {
	if f.errMark != nil {
		return f.errMark
	}
	f.ran[record.ID] = true
	return nil
}

func TestApplyExecutesSeedsOnce(t *testing.T) {
	tracker := newFakeTracker()
	var calls []string

	seeds := []Seed{
		{
			ID: "2024-01-alpha",
			Run: func(ctx context.Context) error {
				calls = append(calls, "alpha")
				return nil
			},
		},
		{
			ID: "2024-01-beta",
			Run: func(ctx context.Context) error {
				calls = append(calls, "beta")
				return nil
			},
		},
	}

	if err := Apply(context.Background(), tracker, seeds, "test-app"); err != nil {
		t.Fatalf("first apply returned error: %v", err)
	}
	if len(calls) != 2 {
		t.Fatalf("expected 2 runs, got %d", len(calls))
	}

	if err := Apply(context.Background(), tracker, seeds, "test-app"); err != nil {
		t.Fatalf("second apply returned error: %v", err)
	}
	if len(calls) != 2 {
		t.Fatalf("expected second apply to skip seeds, got %d runs", len(calls))
	}
}

func TestApplyPropagatesErrors(t *testing.T) {
	boom := errors.New("boom")
	tracker := newFakeTracker()

	seeds := []Seed{
		{
			ID: "bad",
			Run: func(ctx context.Context) error {
				return boom
			},
		},
	}

	err := Apply(context.Background(), tracker, seeds, "test-app")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, boom) {
		t.Fatalf("expected wrapped boom error, got %v", err)
	}

	if tracker.ran["bad"] {
		t.Fatalf("seed should not be marked as run when execution fails")
	}
}

func TestApplyValidatesSeeds(t *testing.T) {
	tracker := newFakeTracker()

	tests := []struct {
		name  string
		seeds []Seed
	}{
		{name: "missing id", seeds: []Seed{{Run: func(ctx context.Context) error { return nil }}}},
		{name: "missing run", seeds: []Seed{{ID: "x"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Apply(context.Background(), tracker, tt.seeds, "app"); err == nil {
				t.Fatalf("expected validation error")
			}
		})
	}
}
