package repository

import "github.com/gambitier/golang-service-template/internal/domain/item"

// Repositories holds write-side repository interfaces.
type Repositories struct {
	Item item.Repository
}

// Persistence combines write-side repositories (and later read models).
type Persistence struct {
	Repositories Repositories
}
