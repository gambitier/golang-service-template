package mongodb

import (
	"context"
	"fmt"

	"github.com/gambitier/golang-service-template/internal/domain/repository"
	itemRepo "github.com/gambitier/golang-service-template/internal/infrastructure/persistence/mongodb/item"
	"github.com/gambitier/golang-service-template/internal/infrastructure/persistence/mongodb/persistopts"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// InitializePersistence wires MongoDB repository adapters into the Persistence bag.
func InitializePersistence(ctx context.Context, db *mongo.Database, opts ...persistopts.Options) (repository.Persistence, error) {
	_ = ctx

	var o persistopts.Options
	for _, item := range opts {
		if item.SkipIndexes {
			o.SkipIndexes = true
		}
	}

	itemRepository, err := itemRepo.NewItemRepository(db, o)
	if err != nil {
		return repository.Persistence{}, fmt.Errorf("item repository: %w", err)
	}

	return repository.Persistence{
		Repositories: repository.Repositories{
			Item: itemRepository,
		},
	}, nil
}
