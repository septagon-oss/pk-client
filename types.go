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

type CreateInput[T any] struct {
	Body T `json:"body"`
}

type UpdateInput[T any] struct {
	ID   string `json:"id,omitempty"`
	Body T      `json:"body"`
}

type PartialUpdateInput struct {
	ID      string         `json:"id,omitempty"`
	Updates map[string]any `json:"updates"`
}

type BulkCreateInput[T any] struct {
	Items []T `json:"items"`
}

type BulkUpdateInput[T any] struct {
	Items []T `json:"items"`
}

type BulkDeleteInput struct {
	IDs []string `json:"ids"`
}

type ExportParams struct {
	Format string   `json:"format,omitempty"`
	Filter *Filter  `json:"filter,omitempty"`
	Fields []string `json:"fields,omitempty"`
}

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

type Filter struct {
	Field    string   `json:"field,omitempty"`
	Operator string   `json:"operator,omitempty"`
	Value    any      `json:"value,omitempty"`
	All      []Filter `json:"all,omitempty"`
	Any      []Filter `json:"any,omitempty"`
	Not      *Filter  `json:"not,omitempty"`
}

type ItemResponse[T any] struct {
	Data     T              `json:"data"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

type ListResponse[T any] struct {
	Data     []T           `json:"data"`
	Metadata *ListMetadata `json:"metadata,omitempty"`
}

type ListMetadata struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalCount int64 `json:"total_count"`
	TotalPages int   `json:"total_pages"`
}

type BulkResponse[T any] struct {
	Succeeded []T           `json:"succeeded"`
	Failed    []BulkError   `json:"failed"`
	Metadata  *BulkMetadata `json:"metadata,omitempty"`
}

type BulkError struct {
	ID    string `json:"id"`
	Error string `json:"error"`
}

type BulkMetadata struct {
	TotalCount     int `json:"total_count"`
	SucceededCount int `json:"succeeded_count"`
	FailedCount    int `json:"failed_count"`
}

type ImportResponse struct {
	Imported int      `json:"imported"`
	Failed   int      `json:"failed"`
	Errors   []string `json:"errors,omitempty"`
}
