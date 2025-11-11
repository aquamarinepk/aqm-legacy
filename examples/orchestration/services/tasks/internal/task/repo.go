package task

import (
	"context"
	"errors"

	"github.com/aquamarinepk/aqm"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
)

// Repo abstracts persistence for tasks.
type Repo interface {
	Save(ctx context.Context, task *Task) error
	FindByID(ctx context.Context, id uuid.UUID) (*Task, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context) ([]*Task, error)
}

var (
	ErrNotFound              = errors.New("task: not found")
	ErrValidationTitleNeeded = errors.New("task: title is required")
)

type mongoRepo struct {
	repo *aqm.MongoRepo[*Task]
}

// NewMongoRepo builds a Mongo-backed repository using aqm primitives.
func NewMongoRepo(collection *mongo.Collection) (Repo, error) {
	base, err := aqm.NewMongoRepo[*Task](collection, func() *Task { return &Task{} })
	if err != nil {
		return nil, err
	}
	return &mongoRepo{repo: base}, nil
}

func (r *mongoRepo) Save(ctx context.Context, task *Task) error {
	return r.repo.Save(ctx, task)
}

func (r *mongoRepo) FindByID(ctx context.Context, id uuid.UUID) (*Task, error) {
	task, err := r.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, aqm.ErrRepoNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return task, nil
}

func (r *mongoRepo) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.repo.Delete(ctx, id)
	if errors.Is(err, aqm.ErrRepoNotFound) {
		return ErrNotFound
	}
	return err
}

func (r *mongoRepo) List(ctx context.Context) ([]*Task, error) {
	return r.repo.List(ctx, nil)
}
