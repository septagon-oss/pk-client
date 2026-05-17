package client

// client_test.go validates the transport-neutral client facade.
//
// ADR: ADR-0029 (file purpose declaration).
// Convention: C-14 (every Go file declares its purpose).

import (
	"context"
	"strings"
	"testing"
)

type memoryTransport[T any] struct {
	item T
}

func (m memoryTransport[T]) Type() TransportType { return TransportTypeHTTP }
func (m memoryTransport[T]) Name() string        { return "memory" }
func (m memoryTransport[T]) Create(context.Context, *CreateInput[T]) (*ItemResponse[T], error) {
	return &ItemResponse[T]{Data: m.item}, nil
}
func (m memoryTransport[T]) GetByID(context.Context, string) (*ItemResponse[T], error) {
	return &ItemResponse[T]{Data: m.item}, nil
}
func (m memoryTransport[T]) List(context.Context, *ListParams) (*ListResponse[T], error) {
	return &ListResponse[T]{Data: []T{m.item}}, nil
}
func (m memoryTransport[T]) Update(context.Context, string, *UpdateInput[T]) (*ItemResponse[T], error) {
	return &ItemResponse[T]{Data: m.item}, nil
}
func (m memoryTransport[T]) PartialUpdate(context.Context, string, *PartialUpdateInput) (*ItemResponse[T], error) {
	return &ItemResponse[T]{Data: m.item}, nil
}
func (m memoryTransport[T]) Delete(context.Context, string) error { return nil }
func (m memoryTransport[T]) BulkCreate(context.Context, *BulkCreateInput[T]) (*BulkResponse[T], error) {
	return &BulkResponse[T]{Succeeded: []T{m.item}}, nil
}
func (m memoryTransport[T]) BulkUpdate(context.Context, *BulkUpdateInput[T]) (*BulkResponse[T], error) {
	return &BulkResponse[T]{Succeeded: []T{m.item}}, nil
}
func (m memoryTransport[T]) BulkDelete(context.Context, []string) error { return nil }
func (m memoryTransport[T]) Export(context.Context, ExportParams) ([]byte, error) {
	return []byte("ok"), nil
}
func (m memoryTransport[T]) Import(context.Context, []byte, string) (*ImportResponse, error) {
	return &ImportResponse{Imported: 1}, nil
}

func TestClientRequiresTransport(t *testing.T) {
	var c Client[map[string]string]
	_, err := c.GetByID(context.Background(), "missing")
	if err == nil || !strings.Contains(err.Error(), "transport not configured") {
		t.Fatalf("expected transport error, got %v", err)
	}
}

func TestClientDelegatesToTransport(t *testing.T) {
	want := map[string]string{"id": "one"}
	c := New[map[string]string](memoryTransport[map[string]string]{item: want})

	got, err := c.GetByID(context.Background(), "one")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got.Data["id"] != "one" {
		t.Fatalf("got %q, want one", got.Data["id"])
	}
}
