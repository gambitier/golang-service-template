package item

// ID is a database-agnostic identifier for an item (hex string when using Mongo ObjectID).
type ID string

func (id ID) String() string {
	return string(id)
}

func (id ID) IsZero() bool {
	return id == ""
}
