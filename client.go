// Package client provides typed clients for PlatformKit-style CRUD APIs.
package client

// client.go owns the transport-neutral client facade used by OSS and
// downstream PlatformKit applications.
//
// ADR: ADR-0029 (file purpose declaration).
// Convention: C-10 (shared builders return errors), C-14 (every Go file declares its purpose).

import (
	"context"
	"errors"
)

// ErrNoTransport is returned by Client methods invoked before a transport has
// been configured. A zero-value Client[T] (for example one declared with var
// rather than constructed via New or NewHTTP) has no transport, so every
// operation fails with this sentinel. Match it with errors.Is:
//
//	if errors.Is(err, client.ErrNoTransport) {
//		// the client was never given a transport
//	}
var ErrNoTransport = errors.New("client: transport not configured")

// Client provides transport-agnostic CRUD operations for PlatformKit APIs.
//
// The type parameter T is the entity type carried by request and response
// bodies. A Client delegates every operation to its CRUDTransport, so the same
// facade works over HTTP today and other transports in the future. Construct
// one with New (any transport) or NewHTTP (the bundled HTTP transport); a
// zero-value Client has no transport and returns ErrNoTransport from every
// method.
type Client[T any] struct {
	transport CRUDTransport[T]
}

// New returns a Client backed by the supplied transport. It performs no
// validation: callers that need configuration checks should build their
// transport (for example via NewHTTPTransport) before passing it in. Use
// NewHTTP for the common case of an HTTP client built from an HTTPConfig.
func New[T any](transport CRUDTransport[T]) *Client[T] {
	return &Client[T]{transport: transport}
}

// NewHTTP builds a Client backed by an HTTP transport derived from config. The
// config is normalized and validated (see HTTPConfig.Normalize); NewHTTP
// returns the resulting error and a nil Client if the config is invalid.
func NewHTTP[T any](config *HTTPConfig) (*Client[T], error) {
	transport, err := NewHTTPTransport[T](config)
	if err != nil {
		return nil, err
	}
	return New[T](transport), nil
}

// WithTransport returns a copy of the Client that uses the supplied transport,
// leaving the receiver unchanged. It is a convenience for swapping transports
// (for example a test stub) without mutating a shared client.
func (c *Client[T]) WithTransport(transport CRUDTransport[T]) *Client[T] {
	return &Client[T]{transport: transport}
}

// Transport returns the CRUDTransport currently backing the Client, or nil if
// none has been configured.
func (c *Client[T]) Transport() CRUDTransport[T] {
	return c.transport
}

// Create sends input to the transport to create a single entity and returns the
// created item. It returns ErrNoTransport if the Client has no transport.
func (c *Client[T]) Create(ctx context.Context, input CreateInput[T]) (*ItemResponse[T], error) {
	if c.transport == nil {
		return nil, ErrNoTransport
	}
	return c.transport.Create(ctx, &input)
}

// GetByID fetches a single entity by its id. It returns ErrNoTransport if the
// Client has no transport.
func (c *Client[T]) GetByID(ctx context.Context, id string) (*ItemResponse[T], error) {
	if c.transport == nil {
		return nil, ErrNoTransport
	}
	return c.transport.GetByID(ctx, id)
}

// List fetches a page of entities matching params and returns the items
// together with pagination metadata. It returns ErrNoTransport if the Client
// has no transport.
func (c *Client[T]) List(ctx context.Context, params ListParams) (*ListResponse[T], error) {
	if c.transport == nil {
		return nil, ErrNoTransport
	}
	return c.transport.List(ctx, &params)
}

// Update replaces the entity identified by id with input and returns the
// updated item. If input.ID is empty it is set to id before the call. It
// returns ErrNoTransport if the Client has no transport.
func (c *Client[T]) Update(ctx context.Context, id string, input UpdateInput[T]) (*ItemResponse[T], error) {
	if c.transport == nil {
		return nil, ErrNoTransport
	}
	if input.ID == "" {
		input.ID = id
	}
	return c.transport.Update(ctx, id, &input)
}

// PartialUpdate applies a partial set of field updates to the entity identified
// by id and returns the updated item. If input.ID is empty it is set to id
// before the call. It returns ErrNoTransport if the Client has no transport.
func (c *Client[T]) PartialUpdate(ctx context.Context, id string, input PartialUpdateInput) (*ItemResponse[T], error) {
	if c.transport == nil {
		return nil, ErrNoTransport
	}
	if input.ID == "" {
		input.ID = id
	}
	return c.transport.PartialUpdate(ctx, id, &input)
}

// Delete removes the entity identified by id. It returns ErrNoTransport if the
// Client has no transport.
func (c *Client[T]) Delete(ctx context.Context, id string) error {
	if c.transport == nil {
		return ErrNoTransport
	}
	return c.transport.Delete(ctx, id)
}

// BulkCreate creates multiple entities in a single request and returns the
// per-item outcomes. It returns ErrNoTransport if the Client has no transport.
func (c *Client[T]) BulkCreate(ctx context.Context, input BulkCreateInput[T]) (*BulkResponse[T], error) {
	if c.transport == nil {
		return nil, ErrNoTransport
	}
	return c.transport.BulkCreate(ctx, &input)
}

// BulkUpdate updates multiple entities in a single request and returns the
// per-item outcomes. It returns ErrNoTransport if the Client has no transport.
func (c *Client[T]) BulkUpdate(ctx context.Context, input BulkUpdateInput[T]) (*BulkResponse[T], error) {
	if c.transport == nil {
		return nil, ErrNoTransport
	}
	return c.transport.BulkUpdate(ctx, &input)
}

// BulkDelete removes multiple entities identified by ids in a single request.
// It returns ErrNoTransport if the Client has no transport.
func (c *Client[T]) BulkDelete(ctx context.Context, ids []string) error {
	if c.transport == nil {
		return ErrNoTransport
	}
	return c.transport.BulkDelete(ctx, ids)
}

// Export retrieves a serialized export of the collection described by params
// (for example as CSV or JSON) and returns the raw bytes. It returns
// ErrNoTransport if the Client has no transport.
func (c *Client[T]) Export(ctx context.Context, params ExportParams) ([]byte, error) {
	if c.transport == nil {
		return nil, ErrNoTransport
	}
	return c.transport.Export(ctx, params)
}

// Import uploads data in the named format (for example "json" or "csv") and
// returns a summary of how many records were imported or failed. It returns
// ErrNoTransport if the Client has no transport.
func (c *Client[T]) Import(ctx context.Context, data []byte, format string) (*ImportResponse, error) {
	if c.transport == nil {
		return nil, ErrNoTransport
	}
	return c.transport.Import(ctx, data, format)
}
