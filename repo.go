package aqm

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var ErrRepoNotFound = errors.New("repository: aggregate not found")

// Repo is the minimum contract services depend on for aggregate storage.
type Repo[T Identifiable] interface {
	Save(ctx context.Context, aggregate T) error
	FindByID(ctx context.Context, id uuid.UUID) (T, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter any) ([]T, error)
}
