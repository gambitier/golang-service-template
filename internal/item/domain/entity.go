package item

import "time"

// Item is the domain entity for a simple catalog/todo-style resource.
type Item struct {
	ID          ID
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
