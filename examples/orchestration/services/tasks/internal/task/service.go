package task

import (
	"context"
	"strings"

	"github.com/aquamarinepk/aqm"
	"github.com/google/uuid"
)

// Service contains business logic for tasks.
type Service struct {
	repo   Repo
	log    aqm.Logger
	config *aqm.Config
}

// NewService wires the task service.
func NewService(repo Repo, logger aqm.Logger, cfg *aqm.Config) *Service {
	if logger == nil {
		logger = aqm.NewNoopLogger()
	}
	return &Service{repo: repo, log: logger, config: cfg}
}

func (s *Service) List(ctx context.Context) ([]*Task, error) {
	return s.repo.List(ctx)
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (*Task, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, title, description string) (*Task, error) {
	if strings.TrimSpace(title) == "" {
		return nil, ErrValidationTitleNeeded
	}
	task := NewTask(title, description)
	if err := s.repo.Save(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, update UpdateTask) (*Task, error) {
	task, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if update.Title != nil {
		trimmed := strings.TrimSpace(*update.Title)
		if trimmed == "" {
			return nil, ErrValidationTitleNeeded
		}
		task.Title = trimmed
	}
	if update.Description != nil {
		task.Description = strings.TrimSpace(*update.Description)
	}
	if update.Completed != nil {
		task.Completed = *update.Completed
	}

	task.Touch()

	if err := s.repo.Save(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) Complete(ctx context.Context, id uuid.UUID) (*Task, error) {
	done := true
	return s.Update(ctx, id, UpdateTask{Completed: &done})
}

func (s *Service) Uncomplete(ctx context.Context, id uuid.UUID) (*Task, error) {
	done := false
	return s.Update(ctx, id, UpdateTask{Completed: &done})
}
