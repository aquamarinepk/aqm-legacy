package todo

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/aquamarinepk/aqm"
)

var ErrNotFound = errors.New("todo: item not found")

// Repo defines the storage contract for todo items.
type Repo interface {
	List(ctx context.Context) ([]Item, error)
	Create(ctx context.Context, item Item) error
	Get(ctx context.Context, id string) (Item, error)
	Update(ctx context.Context, id string, update UpdateItem) (Item, error)
	Delete(ctx context.Context, id string) error
}

type inMemoryRepo struct {
	mu     sync.RWMutex
	items  map[string]Item
	logger aqm.Logger
	cfg    *aqm.Config
}

// UpdateItem holds mutable fields for repository updates.
type UpdateItem struct {
	Title       *string
	Description *string
	Completed   *bool
}

// NewInMemoryRepo creates a repository backed by an in-memory map.
func NewInMemoryRepo(logger aqm.Logger, cfg *aqm.Config) Repo {
	if logger == nil {
		logger = aqm.NewNoopLogger()
	}
	return &inMemoryRepo{
		items:  make(map[string]Item),
		logger: logger,
		cfg:    cfg,
	}
}

func (r *inMemoryRepo) List(ctx context.Context) ([]Item, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]Item, 0, len(r.items))
	for _, item := range r.items {
		out = append(out, item)
	}
	return out, nil
}

func (r *inMemoryRepo) Create(ctx context.Context, item Item) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.items[item.ID] = item
	return nil
}

func (r *inMemoryRepo) Get(ctx context.Context, id string) (Item, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	item, ok := r.items[id]
	if !ok {
		return Item{}, ErrNotFound
	}
	return item, nil
}

func (r *inMemoryRepo) Update(ctx context.Context, id string, update UpdateItem) (Item, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	item, ok := r.items[id]
	if !ok {
		return Item{}, ErrNotFound
	}
	if update.Title != nil {
		item.Title = *update.Title
	}
	if update.Description != nil {
		item.Description = *update.Description
	}
	if update.Completed != nil {
		item.Completed = *update.Completed
	}
	item.UpdatedAt = time.Now().UTC()
	r.items[id] = item
	return item, nil
}

func (r *inMemoryRepo) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.items[id]; !ok {
		return ErrNotFound
	}
	delete(r.items, id)
	return nil
}

func (r *inMemoryRepo) Start(context.Context) error { return nil }

func (r *inMemoryRepo) Stop(context.Context) error { return nil }
