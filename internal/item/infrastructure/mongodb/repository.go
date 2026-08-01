package item

import (
	"context"
	"errors"
	"time"

	domainerr "github.com/gambitier/go-pkgs/errors"
	domainitem "github.com/gambitier/golang-service-template/internal/item/domain"
	"github.com/gambitier/golang-service-template/internal/shared/infrastructure/persistence/persistopts"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type mongoItemRepository struct {
	db *mongo.Database
}

// NewItemRepository creates a MongoDB-backed item.Repository.
func NewItemRepository(db *mongo.Database, opts persistopts.Options) (domainitem.Repository, error) {
	repo := &mongoItemRepository{db: db}

	if !opts.SkipIndexes {
		if err := repo.createIndexes(context.Background()); err != nil {
			return nil, domainerr.Internal(ErrMsgFailedToCreateIndexes, err, nil)
		}
	}

	return repo, nil
}

func (r *mongoItemRepository) createIndexes(ctx context.Context) error {
	collection := r.db.Collection(CollectionName)

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: ItemField.Name, Value: 1}},
		},
		{
			Keys: bson.D{{Key: ItemField.CreatedAt, Value: -1}},
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	return err
}

func (r *mongoItemRepository) Create(ctx context.Context, it *domainitem.Item) error {
	collection := r.db.Collection(CollectionName)

	now := time.Now().UTC()
	it.CreatedAt = now
	it.UpdatedAt = now

	mc, err := toMongoItem(it)
	if err != nil {
		return domainerr.Internal(ErrMsgFailedToConvertItem, err, nil)
	}

	result, err := collection.InsertOne(ctx, mc)
	if err != nil {
		return domainerr.Internal(ErrMsgFailedToInsertItem, err, nil)
	}

	it.ID = domainitem.ID(result.InsertedID.(bson.ObjectID).Hex())
	return nil
}

func (r *mongoItemRepository) GetByID(ctx context.Context, id domainitem.ID) (*domainitem.Item, error) {
	collection := r.db.Collection(CollectionName)

	objectID, err := bson.ObjectIDFromHex(string(id))
	if err != nil {
		return nil, domainerr.InvalidArgumentWithFields(ErrMsgInvalidItemID, map[string]any{
			"field": ItemField.ID,
			"value": string(id),
		})
	}

	var mc mongoItem
	err = collection.FindOne(ctx, bson.M{ItemField.ID: objectID}).Decode(&mc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, domainerr.NotFound(ErrMsgItemNotFound, nil, map[string]any{
				"id": string(id),
			})
		}
		return nil, domainerr.Internal(ErrMsgFailedToFindItem, err, nil)
	}

	return fromMongoItem(&mc), nil
}

func (r *mongoItemRepository) List(ctx context.Context, limit, offset int) ([]*domainitem.Item, error) {
	collection := r.db.Collection(CollectionName)

	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: ItemField.CreatedAt, Value: -1}})

	cursor, err := collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, domainerr.Internal(ErrMsgFailedToListItems, err, nil)
	}
	defer func() { _ = cursor.Close(ctx) }()

	var items []*domainitem.Item
	for cursor.Next(ctx) {
		var mc mongoItem
		if err := cursor.Decode(&mc); err != nil {
			return nil, domainerr.Internal(ErrMsgFailedToListItems, err, nil)
		}
		items = append(items, fromMongoItem(&mc))
	}
	if err := cursor.Err(); err != nil {
		return nil, domainerr.Internal(ErrMsgFailedToListItems, err, nil)
	}

	if items == nil {
		items = []*domainitem.Item{}
	}
	return items, nil
}

func (r *mongoItemRepository) Update(ctx context.Context, it *domainitem.Item) error {
	collection := r.db.Collection(CollectionName)

	objectID, err := bson.ObjectIDFromHex(string(it.ID))
	if err != nil {
		return domainerr.InvalidArgumentWithFields(ErrMsgInvalidItemID, map[string]any{
			"field": ItemField.ID,
			"value": string(it.ID),
		})
	}

	it.UpdatedAt = time.Now().UTC()

	mc, err := toMongoItem(it)
	if err != nil {
		return domainerr.Internal(ErrMsgFailedToConvertItem, err, nil)
	}

	result, err := collection.ReplaceOne(ctx, bson.M{ItemField.ID: objectID}, mc)
	if err != nil {
		return domainerr.Internal(ErrMsgFailedToUpdateItem, err, nil)
	}
	if result.MatchedCount == 0 {
		return domainerr.NotFound(ErrMsgItemNotFound, nil, map[string]any{
			"id": string(it.ID),
		})
	}
	return nil
}

func (r *mongoItemRepository) Delete(ctx context.Context, id domainitem.ID) error {
	collection := r.db.Collection(CollectionName)

	objectID, err := bson.ObjectIDFromHex(string(id))
	if err != nil {
		return domainerr.InvalidArgumentWithFields(ErrMsgInvalidItemID, map[string]any{
			"field": ItemField.ID,
			"value": string(id),
		})
	}

	result, err := collection.DeleteOne(ctx, bson.M{ItemField.ID: objectID})
	if err != nil {
		return domainerr.Internal(ErrMsgFailedToDeleteItem, err, nil)
	}
	if result.DeletedCount == 0 {
		return domainerr.NotFound(ErrMsgItemNotFound, nil, map[string]any{
			"id": string(id),
		})
	}
	return nil
}
