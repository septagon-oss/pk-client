package client

// types.go owns the stable generic transport and response contracts exposed by
// the OSS client package.
//
// ADR: ADR-0029 (file purpose declaration).
// Convention: C-14 (every Go file declares its purpose).

import "context"

// TransportType identifies a client transport implementation.
type TransportType string

const (
	// TransportTypeHTTP identifies the bundled standard-library HTTP transport.
	TransportTypeHTTP TransportType = "http"
)

// Transport is the base contract shared by all client transports.
type Transport interface {
	Type() TransportType
	Name() string
}

// CRUDTransport defines the transport-safe CRUD surface exposed by PlatformKit.
type CRUDTransport[T any] interface {
	Transport

	Create(ctx context.Context, input *CreateInput[T]) (*ItemResponse[T], error)
	GetByID(ctx context.Context, id string) (*ItemResponse[T], error)
	List(ctx context.Context, params *ListParams) (*ListResponse[T], error)
	Update(ctx context.Context, id string, input *UpdateInput[T]) (*ItemResponse[T], error)
	PartialUpdate(ctx context.Context, id string, input *PartialUpdateInput) (*ItemResponse[T], error)
	Delete(ctx context.Context, id string) error

	BulkCreate(ctx context.Context, input *BulkCreateInput[T]) (*BulkResponse[T], error)
	BulkUpdate(ctx context.Context, input *BulkUpdateInput[T]) (*BulkResponse[T], error)
	BulkDelete(ctx context.Context, ids []string) error

	Export(ctx context.Context, params ExportParams) ([]byte, error)
	Import(ctx context.Context, data []byte, format string) (*ImportResponse, error)
}

// Config validates transport-specific configuration.
type Config interface {
	Validate() error
	GetType() TransportType
}

// CreateInput wraps the entity body for a create request.
type CreateInput[T any] struct {
	Body T `json:"body"`
}

// UpdateInput wraps the entity body for a full replacement (PUT) update. ID is
// optional and is populated from the request path id when empty.
type UpdateInput[T any] struct {
	ID   string `json:"id,omitempty"`
	Body T      `json:"body"`
}

// PartialUpdateInput carries a sparse set of field updates for a partial
// (PATCH) update. ID is optional and is populated from the request path id when
// empty. Updates maps field names to their new values.
type PartialUpdateInput struct {
	ID      string         `json:"id,omitempty"`
	Updates map[string]any `json:"updates"`
}

// BulkCreateInput carries the entities to create in a single bulk request.
type BulkCreateInput[T any] struct {
	Items []T `json:"items"`
}

// BulkUpdateInput carries the entities to update in a single bulk request.
type BulkUpdateInput[T any] struct {
	Items []T `json:"items"`
}

// BulkDeleteInput carries the ids to delete in a single bulk request.
type BulkDeleteInput struct {
	IDs []string `json:"ids"`
}

// ExportParams selects the serialization Format, an optional Filter to restrict
// the exported rows, and the optional Fields to include in the export.
type ExportParams struct {
	Format string   `json:"format,omitempty"`
	Filter *Filter  `json:"filter,omitempty"`
	Fields []string `json:"fields,omitempty"`
}

// ListParams describes pagination, ordering, filtering, and projection for a
// List request. All fields are optional; zero values are omitted from the
// request. Page/PageSize and Offset are alternative pagination styles, Sort and
// Order set ordering, Filter restricts results, Fields and Embed project the
// response, and IncludeDeleted opts in to soft-deleted rows.
type ListParams struct {
	Page           int      `json:"page,omitempty"`
	PageSize       int      `json:"page_size,omitempty"`
	Offset         int      `json:"offset,omitempty"`
	Search         string   `json:"search,omitempty"`
	Sort           string   `json:"sort,omitempty"`
	Order          string   `json:"order,omitempty"`
	Filter         *Filter  `json:"filter,omitempty"`
	Fields         []string `json:"fields,omitempty"`
	Embed          []string `json:"embed,omitempty"`
	IncludeDeleted bool     `json:"include_deleted,omitempty"`
}

// Filter expresses a query predicate. A leaf filter uses Field, Operator, and
// Value; a composite filter combines sub-filters with All (logical AND) or Any
// (logical OR), and Not negates a single sub-filter. The forms are mutually
// exclusive in practice — set either the leaf fields or one of the boolean
// combinators.
type Filter struct {
	Field    string   `json:"field,omitempty"`
	Operator string   `json:"operator,omitempty"`
	Value    any      `json:"value,omitempty"`
	All      []Filter `json:"all,omitempty"`
	Any      []Filter `json:"any,omitempty"`
	Not      *Filter  `json:"not,omitempty"`
}

// ItemResponse is the envelope for a single entity. Data holds the entity and
// Metadata carries optional server-supplied annotations.
type ItemResponse[T any] struct {
	Data     T              `json:"data"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// ListResponse is the envelope for a page of entities. Data holds the items and
// Metadata carries optional pagination details.
type ListResponse[T any] struct {
	Data     []T           `json:"data"`
	Metadata *ListMetadata `json:"metadata,omitempty"`
}

// ListMetadata describes the pagination state of a ListResponse: the current
// Page, the PageSize, the TotalCount of matching rows, and the TotalPages
// available.
type ListMetadata struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalCount int64 `json:"total_count"`
	TotalPages int   `json:"total_pages"`
}

// BulkResponse reports the outcome of a bulk operation. Succeeded holds the
// entities that were processed, Failed holds the per-item errors, and Metadata
// carries optional aggregate counts.
type BulkResponse[T any] struct {
	Succeeded []T           `json:"succeeded"`
	Failed    []BulkError   `json:"failed"`
	Metadata  *BulkMetadata `json:"metadata,omitempty"`
}

// BulkError identifies a single item that failed in a bulk operation by its ID
// and the associated Error message.
type BulkError struct {
	ID    string `json:"id"`
	Error string `json:"error"`
}

// BulkMetadata summarizes a bulk operation with the TotalCount of items
// submitted and the SucceededCount and FailedCount outcomes.
type BulkMetadata struct {
	TotalCount     int `json:"total_count"`
	SucceededCount int `json:"succeeded_count"`
	FailedCount    int `json:"failed_count"`
}

// ImportResponse summarizes an import: the number of records Imported, the
// number that Failed, and any per-record Errors.
type ImportResponse struct {
	Imported int      `json:"imported"`
	Failed   int      `json:"failed"`
	Errors   []string `json:"errors,omitempty"`
}
