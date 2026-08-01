package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

// Component manages the MongoDB client lifecycle.
type Component struct {
	uri      string
	database string
	client   *mongo.Client
	db       *mongo.Database
}

// NewComponent builds a Mongo lifecycle component (not yet connected).
func NewComponent(uri, database string) *Component {
	return &Component{uri: uri, database: database}
}

func (c *Component) Name() string { return "mongo" }

func (c *Component) Start(ctx context.Context) error {
	client, db, err := Connect(ctx, c.uri, c.database)
	if err != nil {
		return err
	}
	c.client = client
	c.db = db
	return nil
}

func (c *Component) Stop(ctx context.Context) error {
	if c.client == nil {
		return nil
	}
	err := c.client.Disconnect(ctx)
	c.client = nil
	c.db = nil
	return err
}

// DB returns the database after Start. Panics if not started.
func (c *Component) DB() *mongo.Database {
	if c.db == nil {
		panic(fmt.Sprintf("%s: database not started", c.Name()))
	}
	return c.db
}
