package item_test

import (
	"context"
	"sync"
	"testing"
	"time"

	appitem "github.com/gambitier/golang-service-template/internal/application/item"
	"github.com/gambitier/golang-service-template/internal/domain/domainerr"
	domainitem "github.com/gambitier/golang-service-template/internal/domain/item"
)

type fakeRepo struct {
	mu    sync.Mutex
	items map[domainitem.ID]*domainitem.Item
	seq   int
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{items: make(map[domainitem.ID]*domainitem.Item)}
}

func (r *fakeRepo) Create(_ context.Context, item *domainitem.Item) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seq++
	id := domainitem.ID("item-" + time.Now().Format("150405") + "-" + string(rune('a'+r.seq%26)))
	now := time.Now().UTC()
	item.ID = id
	item.CreatedAt = now
	item.UpdatedAt = now
	cp := *item
	r.items[id] = &cp
	return nil
}

func (r *fakeRepo) GetByID(_ context.Context, id domainitem.ID) (*domainitem.Item, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	it, ok := r.items[id]
	if !ok {
		return nil, domainerr.NotFound("item not found", map[string]any{"id": string(id)})
	}
	cp := *it
	return &cp, nil
}

func (r *fakeRepo) List(_ context.Context, limit, offset int) ([]*domainitem.Item, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]*domainitem.Item, 0, len(r.items))
	for _, it := range r.items {
		cp := *it
		out = append(out, &cp)
	}
	if offset >= len(out) {
		return []*domainitem.Item{}, nil
	}
	out = out[offset:]
	if limit > 0 && limit < len(out) {
		out = out[:limit]
	}
	return out, nil
}

func (r *fakeRepo) Update(_ context.Context, item *domainitem.Item) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[item.ID]; !ok {
		return domainerr.NotFound("item not found", map[string]any{"id": string(item.ID)})
	}
	item.UpdatedAt = time.Now().UTC()
	cp := *item
	r.items[item.ID] = &cp
	return nil
}

func (r *fakeRepo) Delete(_ context.Context, id domainitem.ID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[id]; !ok {
		return domainerr.NotFound("item not found", map[string]any{"id": string(id)})
	}
	delete(r.items, id)
	return nil
}

func TestServiceCreateRequiresName(t *testing.T) {
	svc := appitem.NewService(newFakeRepo())
	_, err := svc.Create(context.Background(), appitem.CreateInput{Name: "  "})
	if err == nil {
		t.Fatal("expected error")
	}
	de, ok := domainerr.As(err)
	if !ok || de.Code != domainerr.CodeInvalidArgument {
		t.Fatalf("expected invalid_argument, got %v", err)
	}
}

func TestServiceCreateGetUpdateDelete(t *testing.T) {
	svc := appitem.NewService(newFakeRepo())
	ctx := context.Background()

	created, err := svc.Create(ctx, appitem.CreateInput{Name: "Notebook", Description: "ruled"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.ID.IsZero() {
		t.Fatal("expected id")
	}

	got, err := svc.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Name != "Notebook" {
		t.Fatalf("name = %q", got.Name)
	}

	updated, err := svc.Update(ctx, appitem.UpdateInput{
		ID:          created.ID,
		Name:        "Sketchbook",
		Description: "blank",
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Name != "Sketchbook" {
		t.Fatalf("updated name = %q", updated.Name)
	}

	if err := svc.Delete(ctx, created.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	_, err = svc.GetByID(ctx, created.ID)
	if err == nil {
		t.Fatal("expected not found after delete")
	}
}
