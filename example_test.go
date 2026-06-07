package client_test

// example_test.go provides runnable godoc examples for the OSS client package.
//
// ADR: ADR-0029 (file purpose declaration).
// Convention: C-14 (every Go file declares its purpose).

import (
	"context"
	"fmt"

	client "github.com/septagon-oss/pk-client"
)

// Widget is a small entity type used by the examples.
type Widget struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// stubTransport is an in-memory CRUDTransport that returns canned responses, so
// the examples run without a network. Only the methods the examples exercise
// return meaningful data; the rest satisfy the interface.
type stubTransport struct{}

func (stubTransport) Type() client.TransportType { return client.TransportTypeHTTP }
func (stubTransport) Name() string               { return "stub" }

func (stubTransport) Create(_ context.Context, input *client.CreateInput[Widget]) (*client.ItemResponse[Widget], error) {
	created := input.Body
	created.ID = "widget-1"
	return &client.ItemResponse[Widget]{Data: created}, nil
}

func (stubTransport) GetByID(_ context.Context, id string) (*client.ItemResponse[Widget], error) {
	return &client.ItemResponse[Widget]{Data: Widget{ID: id, Name: "gadget"}}, nil
}

func (stubTransport) List(_ context.Context, _ *client.ListParams) (*client.ListResponse[Widget], error) {
	return &client.ListResponse[Widget]{
		Data:     []Widget{{ID: "widget-1", Name: "gadget"}},
		Metadata: &client.ListMetadata{Page: 1, PageSize: 25, TotalCount: 1, TotalPages: 1},
	}, nil
}

func (stubTransport) Update(_ context.Context, _ string, input *client.UpdateInput[Widget]) (*client.ItemResponse[Widget], error) {
	return &client.ItemResponse[Widget]{Data: input.Body}, nil
}

func (stubTransport) PartialUpdate(_ context.Context, id string, _ *client.PartialUpdateInput) (*client.ItemResponse[Widget], error) {
	return &client.ItemResponse[Widget]{Data: Widget{ID: id}}, nil
}

func (stubTransport) Delete(_ context.Context, _ string) error { return nil }

func (stubTransport) BulkCreate(_ context.Context, input *client.BulkCreateInput[Widget]) (*client.BulkResponse[Widget], error) {
	return &client.BulkResponse[Widget]{Succeeded: input.Items}, nil
}

func (stubTransport) BulkUpdate(_ context.Context, input *client.BulkUpdateInput[Widget]) (*client.BulkResponse[Widget], error) {
	return &client.BulkResponse[Widget]{Succeeded: input.Items}, nil
}

func (stubTransport) BulkDelete(_ context.Context, _ []string) error { return nil }

func (stubTransport) Export(_ context.Context, _ client.ExportParams) ([]byte, error) {
	return []byte("id,name\nwidget-1,gadget\n"), nil
}

func (stubTransport) Import(_ context.Context, _ []byte, _ string) (*client.ImportResponse, error) {
	return &client.ImportResponse{Imported: 1}, nil
}

// Example shows constructing a Client with New over a stub transport and
// creating an entity.
func Example() {
	c := client.New[Widget](stubTransport{})

	created, err := c.Create(context.Background(), client.CreateInput[Widget]{
		Body: Widget{Name: "gadget"},
	})
	if err != nil {
		fmt.Println("create failed:", err)
		return
	}
	fmt.Printf("created %s named %s\n", created.Data.ID, created.Data.Name)
	// Output: created widget-1 named gadget
}

// ExampleClient_List shows listing entities and reading pagination metadata.
func ExampleClient_List() {
	c := client.New[Widget](stubTransport{})

	page, err := c.List(context.Background(), client.ListParams{Page: 1, PageSize: 25})
	if err != nil {
		fmt.Println("list failed:", err)
		return
	}
	fmt.Printf("%d of %d items\n", len(page.Data), page.Metadata.TotalCount)
	// Output: 1 of 1 items
}
