package todo

import (
	"context"
	"strings"
	"time"

	"github.com/aquamarinepk/aqm"
)

// Service captures the domain logic for todo operations.
type Service struct {
	repo   Repo
	logger aqm.Logger
	cfg    *aqm.Config
}

// NewService builds a Service with the given repository, logger, and config.
func NewService(repo Repo, logger aqm.Logger, cfg *aqm.Config) *Service {
	if repo == nil {
		repo = NewInMemoryRepo(logger, cfg)
	}
	if logger == nil {
		logger = aqm.NewNoopLogger()
	}
	return &Service{repo: repo, logger: logger, cfg: cfg}
}

func (s *Service) ListItems(ctx context.Context) ([]Item, error) {
	return s.repo.List(ctx)
}

func (s *Service) GetItem(ctx context.Context, id string) (Item, error) {
	return s.repo.Get(ctx, id)
}

func (s *Service) CreateItem(ctx context.Context, title, description string) (Item, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return Item{}, ErrValidationTitleRequired
	}

	now := time.Now().UTC()
	item := Item{
		ID:          aqm.GenerateNewID().String(),
		Title:       title,
		Description: strings.TrimSpace(description),
		Completed:   false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.repo.Create(ctx, item); err != nil {
		return Item{}, err
	}
	s.logger.Info("todo item created", "id", item.ID)
	return item, nil
}

func (s *Service) UpdateItem(ctx context.Context, id string, update UpdateItem) (Item, error) {
	if update.Title != nil {
		trimmed := strings.TrimSpace(*update.Title)
		if trimmed == "" {
			return Item{}, ErrValidationTitleRequired
		}
		update.Title = &trimmed
	}
	return s.repo.Update(ctx, id, update)
}

func (s *Service) DeleteItem(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

var ErrValidationTitleRequired = ErrValidation("title is required")

type ErrValidation string

func (e ErrValidation) Error() string { return string(e) }

func (s *Service) Start(context.Context) error { return nil }

func (s *Service) Stop(context.Context) error { return nil }
