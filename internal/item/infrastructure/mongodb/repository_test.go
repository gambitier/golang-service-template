package item_test

import (
	"context"
	"os"
	"testing"
	"time"

	domainitem "github.com/gambitier/golang-service-template/internal/item/domain"
	mongoitem "github.com/gambitier/golang-service-template/internal/item/infrastructure/mongodb"
	"github.com/gambitier/golang-service-template/internal/shared/infrastructure/persistence/mongodb"
	"github.com/gambitier/golang-service-template/internal/shared/infrastructure/persistence/persistopts"
)

func TestMongoItemRepository_CRUD(t *testing.T) {
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		uri = "mongodb://127.0.0.1:27017"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, db, err := mongodb.Connect(ctx, uri, "golang-service-template-test")
	if err != nil {
		t.Skipf("mongo not available: %v", err)
	}
	defer func() {
		_ = db.Collection(mongoitem.CollectionName).Drop(context.Background())
		_ = client.Disconnect(context.Background())
	}()

	repo, err := mongoitem.NewItemRepository(db, persistopts.Options{})
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}

	it := &domainitem.Item{Name: "Pen", Description: "blue"}
	if err := repo.Create(ctx, it); err != nil {
		t.Fatalf("create: %v", err)
	}
	if it.ID.IsZero() {
		t.Fatal("expected id assigned")
	}

	got, err := repo.GetByID(ctx, it.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Name != "Pen" {
		t.Fatalf("name = %q", got.Name)
	}

	got.Description = "black"
	if err := repo.Update(ctx, got); err != nil {
		t.Fatalf("update: %v", err)
	}

	list, err := repo.List(ctx, 10, 0)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) == 0 {
		t.Fatal("expected at least one item")
	}

	if err := repo.Delete(ctx, it.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
}
