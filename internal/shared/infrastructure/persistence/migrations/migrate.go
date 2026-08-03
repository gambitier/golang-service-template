package migrations

import (
	"embed"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mongodb"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed *.json
var files embed.FS

// Up applies all pending MongoDB migrations. Safe to call on every process start.
func Up(mongoURI, database string) error {
	database = strings.TrimSpace(database)
	if database == "" {
		return fmt.Errorf("mongo database name is required")
	}

	src, err := iofs.New(files, ".")
	if err != nil {
		return fmt.Errorf("migrations source: %w", err)
	}

	dbURL, err := databaseURL(mongoURI, database)
	if err != nil {
		return err
	}

	m, err := migrate.NewWithSourceInstance("iofs", src, dbURL)
	if err != nil {
		return fmt.Errorf("migrate init: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}

func databaseURL(mongoURI, database string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(mongoURI))
	if err != nil {
		return "", fmt.Errorf("parse mongo uri: %w", err)
	}
	if u.Scheme == "" {
		return "", fmt.Errorf("mongo uri missing scheme")
	}
	u.Path = "/" + database
	return u.String(), nil
}
