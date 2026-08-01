package item

import (
	"context"
	"strings"

	domainerr "github.com/gambitier/go-pkgs/errors"
	domainitem "github.com/gambitier/golang-service-template/internal/domain/item"
)

// Service orchestrates item use cases against the persistence port.
type Service struct {
	repo domainitem.Repository
}

// NewService constructs an item application service.
func NewService(repo domainitem.Repository) *Service {
	return &Service{repo: repo}
}

// CreateInput is the command for creating an item.
type CreateInput struct {
	Name        string
	Description string
}

// UpdateInput is the command for updating an item.
type UpdateInput struct {
	ID          domainitem.ID
	Name        string
	Description string
}

// Create validates input and persists a new item.
func (s *Service) Create(ctx context.Context, in CreateInput) (*domainitem.Item, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, domainerr.InvalidArgumentWithFields("name is required", map[string]any{"field": "name"})
	}

	it := &domainitem.Item{
		Name:        name,
		Description: strings.TrimSpace(in.Description),
	}
	if err := s.repo.Create(ctx, it); err != nil {
		return nil, err
	}
	return it, nil
}

// GetByID returns an item by domain ID.
func (s *Service) GetByID(ctx context.Context, id domainitem.ID) (*domainitem.Item, error) {
	if id.IsZero() {
		return nil, domainerr.InvalidArgumentWithFields("id is required", map[string]any{"field": "id"})
	}
	return s.repo.GetByID(ctx, id)
}

// List returns a page of items.
func (s *Service) List(ctx context.Context, limit, offset int) ([]*domainitem.Item, error) {
	return s.repo.List(ctx, limit, offset)
}

// Update updates an existing item.
func (s *Service) Update(ctx context.Context, in UpdateInput) (*domainitem.Item, error) {
	if in.ID.IsZero() {
		return nil, domainerr.InvalidArgumentWithFields("id is required", map[string]any{"field": "id"})
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, domainerr.InvalidArgumentWithFields("name is required", map[string]any{"field": "name"})
	}

	existing, err := s.repo.GetByID(ctx, in.ID)
	if err != nil {
		return nil, err
	}

	existing.Name = name
	existing.Description = strings.TrimSpace(in.Description)

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

// Delete removes an item by ID.
func (s *Service) Delete(ctx context.Context, id domainitem.ID) error {
	if id.IsZero() {
		return domainerr.InvalidArgumentWithFields("id is required", map[string]any{"field": "id"})
	}
	return s.repo.Delete(ctx, id)
}
