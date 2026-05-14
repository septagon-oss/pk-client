package client

import (
	"fmt"
	"net/http"
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
		option(config)
	}
	return config
}

func (c *HTTPConfig) Validate() error {
	if strings.TrimSpace(c.BaseURL) == "" {
		return fmt.Errorf("base URL is required")
	}
	if strings.TrimSpace(c.EntityPath) == "" {
		return fmt.Errorf("entity path is required")
	}
	if c.Timeout <= 0 {
		c.Timeout = 30 * time.Second
	}
	if c.Headers == nil {
		c.Headers = map[string]string{}
	}
	if c.QueryParams == nil {
		c.QueryParams = map[string]string{}
	}
	return nil
}

func (c *HTTPConfig) GetType() TransportType {
	return TransportTypeHTTP
}

func (c *HTTPConfig) WithHeader(key, value string) *HTTPConfig {
	if c.Headers == nil {
		c.Headers = map[string]string{}
	}
	c.Headers[key] = value
	return c
}

func (c *HTTPConfig) WithQueryParam(key, value string) *HTTPConfig {
	if c.QueryParams == nil {
		c.QueryParams = map[string]string{}
	}
	c.QueryParams[key] = value
	return c
}

func (c *HTTPConfig) WithBearerToken(token string) *HTTPConfig {
	return c.WithHeader("Authorization", "Bearer "+token)
}

func (c *HTTPConfig) WithHTTPClient(client *http.Client) *HTTPConfig {
	c.HTTPClient = client
	return c
}
