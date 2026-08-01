package item

import (
	domainitem "github.com/gambitier/golang-service-template/internal/domain/item"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func toMongoItem(it *domainitem.Item) (*mongoItem, error) {
	mc := &mongoItem{
		Name:        it.Name,
		Description: it.Description,
		CreatedAt:   it.CreatedAt,
		UpdatedAt:   it.UpdatedAt,
	}

	if !it.ID.IsZero() {
		objectID, err := bson.ObjectIDFromHex(string(it.ID))
		if err != nil {
			return nil, err
		}
		mc.ID = objectID
	}

	return mc, nil
}

func fromMongoItem(mc *mongoItem) *domainitem.Item {
	return &domainitem.Item{
		ID:          domainitem.ID(mc.ID.Hex()),
		Name:        mc.Name,
		Description: mc.Description,
		CreatedAt:   mc.CreatedAt,
		UpdatedAt:   mc.UpdatedAt,
	}
}
