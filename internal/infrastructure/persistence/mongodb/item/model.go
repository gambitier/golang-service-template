package item

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// ItemFields holds BSON field name constants for the Items collection.
type ItemFields struct {
	ID          string
	Name        string
	Description string
	CreatedAt   string
	UpdatedAt   string
}

// ItemField holds all field constants — use this to access item fields.
var ItemField = ItemFields{
	ID:          "_id",
	Name:        "name",
	Description: "description",
	CreatedAt:   "createdAt",
	UpdatedAt:   "updatedAt",
}

// CollectionName is the MongoDB collection for items.
const CollectionName = "Items"

// mongoItem represents an item document in MongoDB.
type mongoItem struct {
	ID          bson.ObjectID `bson:"_id,omitempty"`
	Name        string        `bson:"name"`
	Description string        `bson:"description,omitempty"`
	CreatedAt   time.Time     `bson:"createdAt"`
	UpdatedAt   time.Time     `bson:"updatedAt"`
}
