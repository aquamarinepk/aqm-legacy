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

func TestApplyNilTracker(t *testing.T) {
	seeds := []Seed{{ID: "test", Run: func(ctx context.Context) error { return nil }}}

	err := Apply(context.Background(), nil, seeds, "app")
	if err == nil {
		t.Fatal("expected error for nil tracker")
	}
}

func TestApplyTrackerHasRunError(t *testing.T) {
	tracker := newFakeTracker()
	tracker.errQuery = errors.New("query failed")

	seeds := []Seed{{ID: "test", Run: func(ctx context.Context) error { return nil }}}

	err := Apply(context.Background(), tracker, seeds, "app")
	if err == nil {
		t.Fatal("expected error from HasRun")
	}
}

func TestApplyTrackerMarkRunError(t *testing.T) {
	tracker := newFakeTracker()
	tracker.errMark = errors.New("mark failed")

	seeds := []Seed{{ID: "test", Run: func(ctx context.Context) error { return nil }}}

	err := Apply(context.Background(), tracker, seeds, "app")
	if err == nil {
		t.Fatal("expected error from MarkRun")
	}
}

func TestApplyContextCancelled(t *testing.T) {
	tracker := newFakeTracker()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	seeds := []Seed{{ID: "test", Run: func(ctx context.Context) error { return nil }}}

	err := Apply(ctx, tracker, seeds, "app")
	if err == nil {
		t.Fatal("expected context error")
	}
}

func TestApplyEmptySeeds(t *testing.T) {
	tracker := newFakeTracker()

	err := Apply(context.Background(), tracker, []Seed{}, "app")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSeedStruct(t *testing.T) {
	s := Seed{
		ID:          "test-seed",
		Description: "Test seed description",
		Run:         func(ctx context.Context) error { return nil },
	}

	if s.ID != "test-seed" {
		t.Errorf("ID = %s, want test-seed", s.ID)
	}
	if s.Description != "Test seed description" {
		t.Errorf("Description = %s, want Test seed description", s.Description)
	}
	if s.Run == nil {
		t.Error("Run should not be nil")
	}
}

func TestRecordStruct(t *testing.T) {
	r := Record{
		ID:          "record-id",
		Application: "test-app",
		Description: "test description",
	}

	if r.ID != "record-id" {
		t.Errorf("ID = %s, want record-id", r.ID)
	}
	if r.Application != "test-app" {
		t.Errorf("Application = %s, want test-app", r.Application)
	}
}

func TestWithCollectionName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"normal name", "custom_seeds", "custom_seeds"},
		{"with spaces", "  trimmed  ", "trimmed"},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &mongoTrackerConfig{collectionName: "_seeds"}
			opt := WithCollectionName(tt.input)
			opt(cfg)

			if tt.input == "" || tt.input == "  " {
				// Empty or whitespace-only should not change default
				if cfg.collectionName != "_seeds" {
					t.Errorf("collectionName = %s, want _seeds", cfg.collectionName)
				}
			} else {
				if cfg.collectionName != tt.expected {
					t.Errorf("collectionName = %s, want %s", cfg.collectionName, tt.expected)
				}
			}
		})
	}
}

func TestMongoTrackerHasRunNilCollection(t *testing.T) {
	tracker := &MongoTracker{collection: nil}

	_, err := tracker.HasRun(context.Background(), "test")
	if err == nil {
		t.Fatal("expected error for nil collection")
	}
}

func TestMongoTrackerHasRunNilTracker(t *testing.T) {
	var tracker *MongoTracker

	_, err := tracker.HasRun(context.Background(), "test")
	if err == nil {
		t.Fatal("expected error for nil tracker")
	}
}

func TestMongoTrackerMarkRunNilCollection(t *testing.T) {
	tracker := &MongoTracker{collection: nil}

	err := tracker.MarkRun(context.Background(), Record{ID: "test"})
	if err == nil {
		t.Fatal("expected error for nil collection")
	}
}

func TestMongoTrackerMarkRunNilTracker(t *testing.T) {
	var tracker *MongoTracker

	err := tracker.MarkRun(context.Background(), Record{ID: "test"})
	if err == nil {
		t.Fatal("expected error for nil tracker")
	}
}

func TestMongoTrackerMarkRunEmptyID(t *testing.T) {
	tracker := &MongoTracker{collection: nil}

	err := tracker.MarkRun(context.Background(), Record{ID: ""})
	if err == nil {
		t.Fatal("expected error for empty ID")
	}
}

func TestUpsertOnceNilCollection(t *testing.T) {
	err := UpsertOnce(context.Background(), nil, map[string]string{"_id": "1"}, map[string]string{"name": "test"})
	if err == nil {
		t.Fatal("expected error for nil collection")
	}
}

func TestUpsertOnceNilFilter(t *testing.T) {
	// We can't create a real mongo.Collection without a connection,
	// but the nil filter check happens first
	err := UpsertOnce(context.Background(), nil, nil, map[string]string{"name": "test"})
	if err == nil {
		t.Fatal("expected error for nil filter")
	}
}

func TestUpsertOnceNilDocument(t *testing.T) {
	err := UpsertOnce(context.Background(), nil, map[string]string{"_id": "1"}, nil)
	if err == nil {
		t.Fatal("expected error for nil document")
	}
}

func TestTrackerInterface(t *testing.T) {
	var _ Tracker = &fakeTracker{}
	var _ Tracker = &MongoTracker{}
}

func TestDefaultCollectionName(t *testing.T) {
	if defaultCollectionName != "_seeds" {
		t.Errorf("defaultCollectionName = %s, want _seeds", defaultCollectionName)
	}
}
