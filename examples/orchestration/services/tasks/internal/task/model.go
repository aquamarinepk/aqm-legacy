package task

import (
	"strings"
	"time"

	"github.com/aquamarinepk/aqm"
	"github.com/google/uuid"
)

// Task represents the aggregate managed by the Tasks service.
type Task struct {
	TaskID      uuid.UUID `json:"id" bson:"_id"`
	Title       string    `json:"title" bson:"title"`
	Description string    `json:"description,omitempty" bson:"description"`
	Completed   bool      `json:"completed" bson:"completed"`
	CreatedAt   time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" bson:"updated_at"`
}

// ID satisfies aqm.Identifiable.
func (t *Task) ID() uuid.UUID {
	return t.TaskID
}

// Touch updates audit fields on mutation.
func (t *Task) Touch() {
	t.UpdatedAt = time.Now().UTC()
}

// NewTask constructs a new aggregate enforcing defaults.
func NewTask(title, description string) *Task {
	now := time.Now().UTC()
	return &Task{
		TaskID:      aqm.GenerateNewID(),
		Title:       strings.TrimSpace(title),
		Description: strings.TrimSpace(description),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// UpdateTask groups mutable fields.
type UpdateTask struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Completed   *bool   `json:"completed"`
}
