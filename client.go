// Package client provides typed clients for PlatformKit-style CRUD APIs.
package client

// client.go owns the transport-neutral client facade used by OSS and
// downstream PlatformKit applications.
//
// ADR: ADR-0029 (file purpose declaration).
// Convention: C-10 (shared builders return errors), C-14 (every Go file declares its purpose).

import (
	"context"
	"fmt"
)

// Client provides transport-agnostic CRUD operations for PlatformKit APIs.
type Client[T any] struct {
	transport CRUDTransport[T]
}

func New[T any](transport CRUDTransport[T]) *Client[T] {
	return &Client[T]{transport: transport}
}

func NewHTTP[T any](config *HTTPConfig) (*Client[T], error) {
	transport, err := NewHTTPTransport[T](config)
	if err != nil {
		return nil, err
	}
	return New[T](transport), nil
}

func (c *Client[T]) WithTransport(transport CRUDTransport[T]) *Client[T] {
	return &Client[T]{transport: transport}
}

func (c *Client[T]) Transport() CRUDTransport[T] {
	return c.transport
}

func (c *Client[T]) Create(ctx context.Context, input CreateInput[T]) (*ItemResponse[T], error) {
	if c.transport == nil {
		return nil, fmt.Errorf("transport not configured")
	}
	return c.transport.Create(ctx, &input)
}

func (c *Client[T]) GetByID(ctx context.Context, id string) (*ItemResponse[T], error) {
	if c.transport == nil {
		return nil, fmt.Errorf("transport not configured")
	}
	return c.transport.GetByID(ctx, id)
}

func (c *Client[T]) List(ctx context.Context, params ListParams) (*ListResponse[T], error) {
	if c.transport == nil {
		return nil, fmt.Errorf("transport not configured")
	}
	return c.transport.List(ctx, &params)
}

func (c *Client[T]) Update(ctx context.Context, id string, input UpdateInput[T]) (*ItemResponse[T], error) {
	if c.transport == nil {
		return nil, fmt.Errorf("transport not configured")
	}
	if input.ID == "" {
		input.ID = id
	}
	return c.transport.Update(ctx, id, &input)
}

func (c *Client[T]) PartialUpdate(ctx context.Context, id string, input PartialUpdateInput) (*ItemResponse[T], error) {
	if c.transport == nil {
		return nil, fmt.Errorf("transport not configured")
	}
	if input.ID == "" {
		input.ID = id
	}
	return c.transport.PartialUpdate(ctx, id, &input)
}

func (c *Client[T]) Delete(ctx context.Context, id string) error {
	if c.transport == nil {
		return fmt.Errorf("transport not configured")
	}
	return c.transport.Delete(ctx, id)
}

func (c *Client[T]) BulkCreate(ctx context.Context, input BulkCreateInput[T]) (*BulkResponse[T], error) {
	if c.transport == nil {
		return nil, fmt.Errorf("transport not configured")
	}
	return c.transport.BulkCreate(ctx, &input)
}

func (c *Client[T]) BulkUpdate(ctx context.Context, input BulkUpdateInput[T]) (*BulkResponse[T], error) {
	if c.transport == nil {
		return nil, fmt.Errorf("transport not configured")
	}
	return c.transport.BulkUpdate(ctx, &input)
}

func (c *Client[T]) BulkDelete(ctx context.Context, ids []string) error {
	if c.transport == nil {
		return fmt.Errorf("transport not configured")
	}
	return c.transport.BulkDelete(ctx, ids)
}

func (c *Client[T]) Export(ctx context.Context, params ExportParams) ([]byte, error) {
	if c.transport == nil {
		return nil, fmt.Errorf("transport not configured")
	}
	return c.transport.Export(ctx, params)
}

func (c *Client[T]) Import(ctx context.Context, data []byte, format string) (*ImportResponse, error) {
	if c.transport == nil {
		return nil, fmt.Errorf("transport not configured")
	}
	return c.transport.Import(ctx, data, format)
}
