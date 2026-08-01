package item

import "context"

// Repository is the persistence port for Item aggregates.
// Adapters (Mongo, Postgres, etc.) implement this interface.
type Repository interface {
	Create(ctx context.Context, item *Item) error
	GetByID(ctx context.Context, id ID) (*Item, error)
	List(ctx context.Context, limit, offset int) ([]*Item, error)
	Update(ctx context.Context, item *Item) error
	Delete(ctx context.Context, id ID) error
}
