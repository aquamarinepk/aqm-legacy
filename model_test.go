package aqm

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestGenerateNewID(t *testing.T) {
	id1 := GenerateNewID()
	id2 := GenerateNewID()

	if id1 == uuid.Nil {
		t.Error("GenerateNewID() returned nil UUID")
	}
	if id2 == uuid.Nil {
		t.Error("GenerateNewID() returned nil UUID")
	}
	if id1 == id2 {
		t.Error("GenerateNewID() returned duplicate IDs")
	}
}

func TestSetAuditFieldsBeforeCreate(t *testing.T) {
	var createdAt, updatedAt time.Time
	var createdBy, updatedBy uuid.UUID

	before := time.Now().UTC()
	SetAuditFieldsBeforeCreate(&createdAt, &updatedAt, &createdBy, &updatedBy)
	after := time.Now().UTC()

	if createdAt.IsZero() {
		t.Error("createdAt should be set")
	}
	if updatedAt.IsZero() {
		t.Error("updatedAt should be set")
	}
	if createdAt != updatedAt {
		t.Error("createdAt and updatedAt should be equal on create")
	}
	if createdAt.Before(before) || createdAt.After(after) {
		t.Error("createdAt should be within test bounds")
	}
}

func TestSetAuditFieldsBeforeUpdate(t *testing.T) {
	updatedAt := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	var updatedBy uuid.UUID

	before := time.Now().UTC()
	SetAuditFieldsBeforeUpdate(&updatedAt, &updatedBy)
	after := time.Now().UTC()

	if updatedAt.Before(before) || updatedAt.After(after) {
		t.Error("updatedAt should be within test bounds")
	}
}
