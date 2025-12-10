package aqm

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestErrRepoNotFound(t *testing.T) {
	if ErrRepoNotFound.Error() != "repository: aggregate not found" {
		t.Errorf("ErrRepoNotFound = %s, want repository: aggregate not found", ErrRepoNotFound.Error())
	}
}

func TestRepoInterface(t *testing.T) {
	// Test that mockRepo implements Repo interface
	var _ Repo[*mockAggregate] = &mockRepo{}
}

type mockAggregate struct {
	id uuid.UUID
}

func (m *mockAggregate) ID() uuid.UUID {
	return m.id
}

type mockRepo struct {
	saveErr   error
	findErr   error
	deleteErr error
	listErr   error
	items     []*mockAggregate
}

func (m *mockRepo) Save(ctx context.Context, aggregate *mockAggregate) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.items = append(m.items, aggregate)
	return nil
}

func (m *mockRepo) FindByID(ctx context.Context, id uuid.UUID) (*mockAggregate, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	for _, item := range m.items {
		if item.id == id {
			return item, nil
		}
	}
	return nil, ErrRepoNotFound
}

func (m *mockRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	for i, item := range m.items {
		if item.id == id {
			m.items = append(m.items[:i], m.items[i+1:]...)
			return nil
		}
	}
	return ErrRepoNotFound
}

func (m *mockRepo) List(ctx context.Context, filter any) ([]*mockAggregate, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.items, nil
}

func TestMockRepoSave(t *testing.T) {
	repo := &mockRepo{}
	aggregate := &mockAggregate{id: uuid.New()}

	err := repo.Save(context.Background(), aggregate)
	if err != nil {
		t.Errorf("Save error: %v", err)
	}

	if len(repo.items) != 1 {
		t.Errorf("items length = %d, want 1", len(repo.items))
	}
}

func TestMockRepoSaveError(t *testing.T) {
	expectedErr := errors.New("save failed")
	repo := &mockRepo{saveErr: expectedErr}
	aggregate := &mockAggregate{id: uuid.New()}

	err := repo.Save(context.Background(), aggregate)
	if err != expectedErr {
		t.Errorf("error = %v, want %v", err, expectedErr)
	}
}

func TestMockRepoFindByID(t *testing.T) {
	id := uuid.New()
	repo := &mockRepo{items: []*mockAggregate{{id: id}}}

	found, err := repo.FindByID(context.Background(), id)
	if err != nil {
		t.Errorf("FindByID error: %v", err)
	}
	if found.id != id {
		t.Errorf("found.id = %v, want %v", found.id, id)
	}
}

func TestMockRepoFindByIDNotFound(t *testing.T) {
	repo := &mockRepo{}

	_, err := repo.FindByID(context.Background(), uuid.New())
	if !errors.Is(err, ErrRepoNotFound) {
		t.Errorf("error = %v, want ErrRepoNotFound", err)
	}
}

func TestMockRepoDelete(t *testing.T) {
	id := uuid.New()
	repo := &mockRepo{items: []*mockAggregate{{id: id}}}

	err := repo.Delete(context.Background(), id)
	if err != nil {
		t.Errorf("Delete error: %v", err)
	}

	if len(repo.items) != 0 {
		t.Errorf("items length = %d, want 0", len(repo.items))
	}
}

func TestMockRepoDeleteNotFound(t *testing.T) {
	repo := &mockRepo{}

	err := repo.Delete(context.Background(), uuid.New())
	if !errors.Is(err, ErrRepoNotFound) {
		t.Errorf("error = %v, want ErrRepoNotFound", err)
	}
}

func TestMockRepoList(t *testing.T) {
	repo := &mockRepo{items: []*mockAggregate{
		{id: uuid.New()},
		{id: uuid.New()},
	}}

	items, err := repo.List(context.Background(), nil)
	if err != nil {
		t.Errorf("List error: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("items length = %d, want 2", len(items))
	}
}

func TestMockRepoListError(t *testing.T) {
	expectedErr := errors.New("list failed")
	repo := &mockRepo{listErr: expectedErr}

	_, err := repo.List(context.Background(), nil)
	if err != expectedErr {
		t.Errorf("error = %v, want %v", err, expectedErr)
	}
}
