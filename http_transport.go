package client

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
)

type HTTPTransport[T any] struct {
	httpClient *http.Client
	config     *HTTPConfig
}

type APIError struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	ErrorMsg   string `json:"error,omitempty"`
	Body       []byte `json:"-"`
}

func (e *APIError) Error() string {
	if e.ErrorMsg != "" {
		return e.ErrorMsg
	}
	if e.Message != "" {
		return e.Message
	}
	return http.StatusText(e.StatusCode)
}

func NewHTTPTransport[T any](config *HTTPConfig) (*HTTPTransport[T], error) {
	if config == nil {
		return nil, fmt.Errorf("http config is required")
	}
	if err := config.Validate(); err != nil {
		return nil, err
	}
	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: config.Timeout}
	}
	return &HTTPTransport[T]{httpClient: httpClient, config: config}, nil
}

func (t *HTTPTransport[T]) Type() TransportType {
	return TransportTypeHTTP
}

func (t *HTTPTransport[T]) Name() string {
	return "http"
}

func (t *HTTPTransport[T]) buildURL(parts ...string) string {
	base := strings.TrimRight(t.config.BaseURL, "/") + "/" + strings.Trim(t.config.EntityPath, "/")
	for _, part := range parts {
		if strings.TrimSpace(part) == "" {
			continue
		}
		base += "/" + url.PathEscape(part)
	}
	return base
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

	requestURL, err := t.withStaticQueryParams(rawURL)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, method, requestURL, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "pk-client/0.1")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
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
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("decode response: %w", err)
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

func (t *HTTPTransport[T]) GetByID(ctx context.Context, id string) (*ItemResponse[T], error) {
	req, err := t.newRequest(ctx, http.MethodGet, t.buildURL(id), nil)
	if err != nil {
		return nil, err
	}
	var result ItemResponse[T]
	if err := t.do(req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

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

func (t *HTTPTransport[T]) Update(ctx context.Context, id string, input *UpdateInput[T]) (*ItemResponse[T], error) {
	req, err := t.newRequest(ctx, http.MethodPut, t.buildURL(id), input)
	if err != nil {
		return nil, err
	}
	var result ItemResponse[T]
	if err := t.do(req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (t *HTTPTransport[T]) PartialUpdate(ctx context.Context, id string, input *PartialUpdateInput) (*ItemResponse[T], error) {
	req, err := t.newRequest(ctx, http.MethodPatch, t.buildURL(id), input)
	if err != nil {
		return nil, err
	}
	var result ItemResponse[T]
	if err := t.do(req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (t *HTTPTransport[T]) Delete(ctx context.Context, id string) error {
	req, err := t.newRequest(ctx, http.MethodDelete, t.buildURL(id), nil)
	if err != nil {
		return err
	}
	return t.do(req, nil)
}

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

func (t *HTTPTransport[T]) BulkDelete(ctx context.Context, ids []string) error {
	req, err := t.newRequest(ctx, http.MethodDelete, t.buildURL("bulk"), BulkDeleteInput{IDs: ids})
	if err != nil {
		return err
	}
	return t.do(req, nil)
}

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
	req, err := t.newRequest(ctx, http.MethodPost, rawURL, json.RawMessage(data))
	if err != nil {
		return nil, err
	}
	var result ImportResponse
	if err := t.do(req, &result); err != nil {
		return nil, err
	}
	return &result, nil
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
