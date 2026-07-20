package client

// Implements: REQ-013.
// Per: ADR-0028.
// Discipline: C-14.
// http_transport.go owns the standard-library HTTP CRUD transport.
//
// ADR: ADR-0029 (file purpose declaration).
// Convention: C-10 (shared builders return errors), C-14 (every Go file declares its purpose).

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/septagon-oss/pk-shared/pkg/pathsegment"
)

// HTTPTransport is the standard-library CRUDTransport that talks to a
// PlatformKit HTTP API. The type parameter T is the entity type carried by
// request and response bodies. It owns a normalized copy of its HTTPConfig, so
// mutating the original config after construction does not affect the
// transport. Construct one with NewHTTPTransport, or use NewHTTP to wrap one in
// a Client.
type HTTPTransport[T any] struct {
	httpClient *http.Client
	config     *HTTPConfig
}

// APIError is returned for any HTTP response with a status code of 400 or
// higher. StatusCode holds the HTTP status code, Message and ErrorMsg carry the
// server-supplied detail (parsed from the JSON body when present), and Body
// holds the raw response body for callers that need the unparsed payload. The
// transport returns it as an error value, so recover the typed form with
// errors.As:
//
//	var apiErr *client.APIError
//	if errors.As(err, &apiErr) {
//		log.Printf("status %d: %s", apiErr.StatusCode, apiErr.Body)
//	}
type APIError struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	ErrorMsg   string `json:"error,omitempty"`
	Body       []byte `json:"-"`
}

// Error implements the error interface, preferring the server-supplied error
// message, then the message field, then the standard text for the status code.
func (e *APIError) Error() string {
	if e.ErrorMsg != "" {
		return e.ErrorMsg
	}
	if e.Message != "" {
		return e.Message
	}
	return http.StatusText(e.StatusCode)
}

// NewHTTPTransport builds an HTTPTransport from config. The config is
// normalized and validated (see HTTPConfig.Normalize), and a copy is retained
// so later mutation of the caller's config has no effect. If config.HTTPClient
// is nil, a default *http.Client using config.Timeout is created. It returns
// the normalization error and a nil transport when config is invalid.
func NewHTTPTransport[T any](config *HTTPConfig) (*HTTPTransport[T], error) {
	normalized, err := normalizeHTTPConfig(config)
	if err != nil {
		return nil, err
	}
	httpClient := normalized.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: normalized.Timeout}
	}
	return &HTTPTransport[T]{httpClient: httpClient, config: &normalized}, nil
}

// Type reports the transport kind, always TransportTypeHTTP. It satisfies the
// Transport interface.
func (t *HTTPTransport[T]) Type() TransportType {
	return TransportTypeHTTP
}

// Name returns the human-readable transport name, "http". It satisfies the
// Transport interface.
func (t *HTTPTransport[T]) Name() string {
	return "http"
}

func (t *HTTPTransport[T]) buildURL(parts ...string) string {
	var b strings.Builder
	b.WriteString(strings.TrimRight(t.config.BaseURL, "/"))
	b.WriteByte('/')
	b.WriteString(strings.Trim(t.config.EntityPath, "/"))
	for _, part := range parts {
		if strings.TrimSpace(part) == "" {
			continue
		}
		b.WriteByte('/')
		b.WriteString(url.PathEscape(part))
	}
	return b.String()
}

func (t *HTTPTransport[T]) buildEntityURL(id string) (string, error) {
	segment, ok := pathsegment.EncodeOpaqueID(id)
	if !ok {
		return "", fmt.Errorf("entity ID cannot be represented as a canonical opaque path segment")
	}
	return t.buildURL(segment), nil
}

func (t *HTTPTransport[T]) newRequest(ctx context.Context, method, rawURL string, body any) (*http.Request, error) {
	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		reader = bytes.NewReader(payload)
	}

	return t.newRequestWithReader(ctx, method, rawURL, reader, body != nil, "application/json")
}

func (t *HTTPTransport[T]) newRawRequest(ctx context.Context, method, rawURL string, body []byte, contentType string) (*http.Request, error) {
	return t.newRequestWithReader(ctx, method, rawURL, bytes.NewReader(body), true, contentType)
}

func (t *HTTPTransport[T]) newRequestWithReader(ctx context.Context, method, rawURL string, body io.Reader, hasBody bool, contentType string) (*http.Request, error) {
	requestURL, err := t.withStaticQueryParams(rawURL)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, method, requestURL, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "pk-client/0.1")
	if hasBody && strings.TrimSpace(contentType) != "" {
		req.Header.Set("Content-Type", strings.TrimSpace(contentType))
	}
	if t.config.APIKey != "" && req.Header.Get("Authorization") == "" {
		req.Header.Set("Authorization", "Bearer "+t.config.APIKey)
	}
	for key, value := range t.config.Headers {
		req.Header.Set(key, value)
	}
	return req, nil
}

func (t *HTTPTransport[T]) do(req *http.Request, result any) error {
	resp, err := t.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return parseAPIError(body, resp.StatusCode)
	}
	if result == nil || resp.StatusCode == http.StatusNoContent {
		return nil
	}
	if err := decodeJSONBody(resp.Body, result); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

func decodeJSONBody(reader io.Reader, result any) error {
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(result); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return fmt.Errorf("response body must contain exactly one JSON document")
		}
		return err
	}
	return nil
}

func parseAPIError(body []byte, statusCode int) error {
	var apiErr APIError
	if err := json.Unmarshal(body, &apiErr); err != nil {
		return &APIError{StatusCode: statusCode, Message: string(body), Body: body}
	}
	apiErr.StatusCode = statusCode
	apiErr.Body = body
	return &apiErr
}

func (t *HTTPTransport[T]) withStaticQueryParams(rawURL string) (string, error) {
	if len(t.config.QueryParams) == 0 {
		return rawURL, nil
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid request URL: %w", err)
	}
	query := parsed.Query()
	for key, value := range t.config.QueryParams {
		if strings.TrimSpace(key) != "" && query.Get(key) == "" {
			query.Set(key, value)
		}
	}
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}

// Create issues a POST to the collection endpoint to create a single entity
// and decodes the created item. It returns an *APIError for non-2xx responses.
func (t *HTTPTransport[T]) Create(ctx context.Context, input *CreateInput[T]) (*ItemResponse[T], error) {
	req, err := t.newRequest(ctx, http.MethodPost, t.buildURL(), input)
	if err != nil {
		return nil, err
	}
	var result ItemResponse[T]
	if err := t.do(req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetByID issues a GET to the entity endpoint for id and decodes the item. It
// returns an *APIError for non-2xx responses.
func (t *HTTPTransport[T]) GetByID(ctx context.Context, id string) (*ItemResponse[T], error) {
	rawURL, err := t.buildEntityURL(id)
	if err != nil {
		return nil, err
	}
	req, err := t.newRequest(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	var result ItemResponse[T]
	if err := t.do(req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// List issues a GET to the collection endpoint with params encoded as query
// parameters and decodes the page of items and metadata. It returns an
// *APIError for non-2xx responses.
func (t *HTTPTransport[T]) List(ctx context.Context, params *ListParams) (*ListResponse[T], error) {
	rawURL, err := t.listURL(params)
	if err != nil {
		return nil, err
	}
	req, err := t.newRequest(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	var result ListResponse[T]
	if err := t.do(req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update issues a PUT to the entity endpoint for id to replace the entity and
// decodes the updated item. It returns an *APIError for non-2xx responses.
func (t *HTTPTransport[T]) Update(ctx context.Context, id string, input *UpdateInput[T]) (*ItemResponse[T], error) {
	rawURL, err := t.buildEntityURL(id)
	if err != nil {
		return nil, err
	}
	req, err := t.newRequest(ctx, http.MethodPut, rawURL, input)
	if err != nil {
		return nil, err
	}
	var result ItemResponse[T]
	if err := t.do(req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// PartialUpdate issues a PATCH to the entity endpoint for id to apply a partial
// set of field updates and decodes the updated item. It returns an *APIError
// for non-2xx responses.
func (t *HTTPTransport[T]) PartialUpdate(ctx context.Context, id string, input *PartialUpdateInput) (*ItemResponse[T], error) {
	rawURL, err := t.buildEntityURL(id)
	if err != nil {
		return nil, err
	}
	req, err := t.newRequest(ctx, http.MethodPatch, rawURL, input)
	if err != nil {
		return nil, err
	}
	var result ItemResponse[T]
	if err := t.do(req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete issues a DELETE to the entity endpoint for id. It returns an *APIError
// for non-2xx responses.
func (t *HTTPTransport[T]) Delete(ctx context.Context, id string) error {
	rawURL, err := t.buildEntityURL(id)
	if err != nil {
		return err
	}
	req, err := t.newRequest(ctx, http.MethodDelete, rawURL, nil)
	if err != nil {
		return err
	}
	return t.do(req, nil)
}

// BulkCreate issues a POST to the collection's "bulk" endpoint to create many
// entities at once and decodes the per-item outcomes. It returns an *APIError
// for non-2xx responses.
func (t *HTTPTransport[T]) BulkCreate(ctx context.Context, input *BulkCreateInput[T]) (*BulkResponse[T], error) {
	req, err := t.newRequest(ctx, http.MethodPost, t.buildURL("bulk"), input)
	if err != nil {
		return nil, err
	}
	var result BulkResponse[T]
	if err := t.do(req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// BulkUpdate issues a PUT to the collection's "bulk" endpoint to update many
// entities at once and decodes the per-item outcomes. It returns an *APIError
// for non-2xx responses.
func (t *HTTPTransport[T]) BulkUpdate(ctx context.Context, input *BulkUpdateInput[T]) (*BulkResponse[T], error) {
	req, err := t.newRequest(ctx, http.MethodPut, t.buildURL("bulk"), input)
	if err != nil {
		return nil, err
	}
	var result BulkResponse[T]
	if err := t.do(req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// BulkDelete issues a DELETE to the collection's "bulk" endpoint with the
// supplied ids in the request body. It returns an *APIError for non-2xx
// responses.
func (t *HTTPTransport[T]) BulkDelete(ctx context.Context, ids []string) error {
	req, err := t.newRequest(ctx, http.MethodDelete, t.buildURL("bulk"), BulkDeleteInput{IDs: ids})
	if err != nil {
		return err
	}
	return t.do(req, nil)
}

// Export issues a GET to the collection's "export" endpoint with params encoded
// as query parameters and returns the raw response body unparsed. It returns an
// *APIError for non-2xx responses.
func (t *HTTPTransport[T]) Export(ctx context.Context, params ExportParams) ([]byte, error) {
	rawURL, err := t.exportURL(params)
	if err != nil {
		return nil, err
	}
	req, err := t.newRequest(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read export response: %w", err)
	}
	if resp.StatusCode >= 400 {
		return nil, parseAPIError(body, resp.StatusCode)
	}
	return body, nil
}

// Import issues a POST to the collection's "import" endpoint, sending data as
// the raw request body with a Content-Type derived from format (for example
// "json" or "csv") and a matching format query parameter. It decodes and
// returns the import summary, and returns an *APIError for non-2xx responses.
func (t *HTTPTransport[T]) Import(ctx context.Context, data []byte, format string) (*ImportResponse, error) {
	rawURL := t.buildURL("import")
	if format != "" {
		parsed, err := url.Parse(rawURL)
		if err != nil {
			return nil, err
		}
		query := parsed.Query()
		query.Set("format", format)
		parsed.RawQuery = query.Encode()
		rawURL = parsed.String()
	}
	req, err := t.newRawRequest(ctx, http.MethodPost, rawURL, data, importContentType(format))
	if err != nil {
		return nil, err
	}
	var result ImportResponse
	if err := t.do(req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func importContentType(format string) string {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "json":
		return "application/json"
	case "csv":
		return "text/csv"
	case "ndjson", "jsonl":
		return "application/x-ndjson"
	default:
		return "application/octet-stream"
	}
}

func (t *HTTPTransport[T]) listURL(params *ListParams) (string, error) {
	parsed, err := url.Parse(t.buildURL())
	if err != nil {
		return "", err
	}
	query := parsed.Query()
	if params != nil {
		setInt(query, "page", params.Page)
		setInt(query, "page_size", params.PageSize)
		setInt(query, "offset", params.Offset)
		setString(query, "search", params.Search)
		setString(query, "sort", params.Sort)
		setString(query, "order", params.Order)
		if params.Filter != nil {
			filter, err := json.Marshal(params.Filter)
			if err != nil {
				return "", fmt.Errorf("marshal filter: %w", err)
			}
			query.Set("filter", string(filter))
		}
		for _, field := range params.Fields {
			query.Add("fields", field)
		}
		for _, embed := range params.Embed {
			query.Add("embed", embed)
		}
		if params.IncludeDeleted {
			query.Set("include_deleted", "true")
		}
	}
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}

func (t *HTTPTransport[T]) exportURL(params ExportParams) (string, error) {
	parsed, err := url.Parse(t.buildURL("export"))
	if err != nil {
		return "", err
	}
	query := parsed.Query()
	setString(query, "format", params.Format)
	if params.Filter != nil {
		filter, err := json.Marshal(params.Filter)
		if err != nil {
			return "", fmt.Errorf("marshal filter: %w", err)
		}
		query.Set("filter", string(filter))
	}
	for _, field := range params.Fields {
		query.Add("fields", field)
	}
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}

func setString(query url.Values, key, value string) {
	if value != "" {
		query.Set(key, value)
	}
}

func setInt(query url.Values, key string, value int) {
	if value > 0 {
		query.Set(key, strconv.Itoa(value))
	}
}
