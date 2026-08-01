// Package bsonutil converts between domain hex IDs and MongoDB ObjectIDs.
package bsonutil

import (
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// ObjectIDFromHex converts a required hex string to ObjectID.
func ObjectIDFromHex(hex string) (bson.ObjectID, error) {
	if hex == "" {
		return bson.ObjectID{}, fmt.Errorf("object id hex is empty")
	}
	return bson.ObjectIDFromHex(hex)
}

// HexFromObjectID returns the hex form of id, or empty when zero.
func HexFromObjectID(id bson.ObjectID) string {
	if id.IsZero() {
		return ""
	}
	return id.Hex()
}
