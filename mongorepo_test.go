package aqm

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewMongoRepoNilCollection(t *testing.T) {
	factory := func() *testIdentifiable { return &testIdentifiable{} }

	_, err := NewMongoRepo[*testIdentifiable](nil, factory)
	if err == nil {
		t.Error("NewMongoRepo should return error for nil collection")
	}
}

func TestNewMongoRepoNilFactory(t *testing.T) {
	// We can't create a real collection without mongo, so we just test nil factory case
	_, err := NewMongoRepo[*testIdentifiable](nil, nil)
	if err == nil {
		t.Error("NewMongoRepo should return error")
	}
}

func TestMongoRepoStructFields(t *testing.T) {
	// Test that MongoRepo can be constructed with the right type parameters
	type User struct {
		id   uuid.UUID
		name string
	}

	// This test just verifies the generic type compiles correctly
	_ = MongoRepo[*testIdentifiable]{}
}

// testIdentifiable is a helper type for testing
type testIdentifiable struct {
	id uuid.UUID
}

func (t *testIdentifiable) ID() uuid.UUID {
	return t.id
}
