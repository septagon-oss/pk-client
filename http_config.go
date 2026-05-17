package client

// http_config.go owns normalized HTTP transport configuration for generated and
// hand-written PlatformKit clients.
//
// ADR: ADR-0029 (file purpose declaration).
// Convention: C-10 (shared builders return errors), C-14 (every Go file declares its purpose).

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type HTTPConfig struct {
	BaseURL     string            `json:"base_url"`
	EntityPath  string            `json:"entity_path"`
	APIKey      string            `json:"api_key,omitempty"`
	Timeout     time.Duration     `json:"timeout"`
	Headers     map[string]string `json:"headers,omitempty"`
	QueryParams map[string]string `json:"query_params,omitempty"`
	HTTPClient  *http.Client      `json:"-"`
}

func NewHTTPConfig(baseURL, entityPath string, options ...Option) *HTTPConfig {
	config := &HTTPConfig{
		BaseURL:     baseURL,
		EntityPath:  entityPath,
		Timeout:     30 * time.Second,
		Headers:     map[string]string{},
		QueryParams: map[string]string{},
	}
	for _, option := range options {
		if option != nil {
			option(config)
		}
	}
	return config
}

func (c *HTTPConfig) Validate() error {
	normalized, err := normalizeHTTPConfig(c)
	if err != nil {
		return err
	}
	*c = normalized
	return nil
}

func normalizeHTTPConfig(config *HTTPConfig) (HTTPConfig, error) {
	if config == nil {
		return HTTPConfig{}, fmt.Errorf("http config is required")
	}
	normalized := *config
	normalized.BaseURL = strings.TrimRight(strings.TrimSpace(normalized.BaseURL), "/")
	normalized.EntityPath = strings.Trim(strings.TrimSpace(normalized.EntityPath), "/")
	normalized.APIKey = strings.TrimSpace(normalized.APIKey)
	if normalized.BaseURL == "" {
		return HTTPConfig{}, fmt.Errorf("base URL is required")
	}
	parsed, err := url.Parse(normalized.BaseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return HTTPConfig{}, fmt.Errorf("base URL must be an absolute URL")
	}
	if normalized.EntityPath == "" {
		return HTTPConfig{}, fmt.Errorf("entity path is required")
	}
	if strings.ContainsAny(normalized.EntityPath, "?#") {
		return HTTPConfig{}, fmt.Errorf("entity path must not include query or fragment")
	}
	if normalized.Timeout <= 0 {
		normalized.Timeout = 30 * time.Second
	}
	headers, err := normalizeStringMap(normalized.Headers, "header")
	if err != nil {
		return HTTPConfig{}, err
	}
	queryParams, err := normalizeStringMap(normalized.QueryParams, "query parameter")
	if err != nil {
		return HTTPConfig{}, err
	}
	normalized.Headers = headers
	normalized.QueryParams = queryParams
	return normalized, nil
}

func normalizeStringMap(values map[string]string, label string) (map[string]string, error) {
	out := map[string]string{}
	for key, value := range values {
		key = strings.TrimSpace(key)
		if key == "" {
			return nil, fmt.Errorf("%s key is required", label)
		}
		out[key] = strings.TrimSpace(value)
	}
	return out, nil
}

func (c *HTTPConfig) GetType() TransportType {
	return TransportTypeHTTP
}

func (c *HTTPConfig) WithHeader(key, value string) *HTTPConfig {
	key = strings.TrimSpace(key)
	if key == "" {
		return c
	}
	if c.Headers == nil {
		c.Headers = map[string]string{}
	}
	c.Headers[key] = strings.TrimSpace(value)
	return c
}

func (c *HTTPConfig) WithQueryParam(key, value string) *HTTPConfig {
	key = strings.TrimSpace(key)
	if key == "" {
		return c
	}
	if c.QueryParams == nil {
		c.QueryParams = map[string]string{}
	}
	c.QueryParams[key] = strings.TrimSpace(value)
	return c
}

func (c *HTTPConfig) WithBearerToken(token string) *HTTPConfig {
	token = strings.TrimSpace(token)
	if token == "" {
		return c
	}
	return c.WithHeader("Authorization", "Bearer "+token)
}

func (c *HTTPConfig) WithHTTPClient(client *http.Client) *HTTPConfig {
	c.HTTPClient = client
	return c
}
