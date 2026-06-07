package client

// http_config.go owns normalized HTTP transport configuration for generated and
// hand-written PlatformKit clients.
//
// ADR: ADR-0029 (file purpose declaration).
// Convention: C-10 (shared builders return errors), C-14 (every Go file declares its purpose).

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Errors returned by HTTPConfig.Normalize (and therefore by NewHTTPConfig
// consumers such as NewHTTP and NewHTTPTransport) when configuration is
// invalid. Match them with errors.Is to distinguish validation failures:
//
//	if errors.Is(err, client.ErrBaseURLRequired) {
//		// the caller forgot to set BaseURL
//	}
var (
	// ErrNilConfig indicates that a nil *HTTPConfig was supplied where a
	// configuration value was required.
	ErrNilConfig = errors.New("client: http config is required")
	// ErrBaseURLRequired indicates that HTTPConfig.BaseURL was empty after
	// trimming.
	ErrBaseURLRequired = errors.New("client: base URL is required")
	// ErrBaseURLInvalid indicates that HTTPConfig.BaseURL was not a valid
	// absolute URL (it must include both a scheme and a host).
	ErrBaseURLInvalid = errors.New("client: base URL must be an absolute URL")
	// ErrEntityPathRequired indicates that HTTPConfig.EntityPath was empty
	// after trimming.
	ErrEntityPathRequired = errors.New("client: entity path is required")
	// ErrEntityPathInvalid indicates that HTTPConfig.EntityPath contained a
	// query ('?') or fragment ('#') marker.
	ErrEntityPathInvalid = errors.New("client: entity path must not include query or fragment")
	// ErrHeaderKeyRequired indicates that a header or query-parameter map
	// contained an empty key after trimming.
	ErrHeaderKeyRequired = errors.New("client: header or query parameter key is required")
)

// HTTPConfig describes how an HTTP CRUD transport reaches a PlatformKit API.
// BaseURL and EntityPath together locate the collection endpoint (for example
// "https://api.example.com" + "/api/widgets"); the remaining fields are
// optional. APIKey, Headers, QueryParams, and Timeout customize every request,
// and HTTPClient supplies a caller-owned *http.Client when the default is not
// suitable. Build one with NewHTTPConfig and refine it with the With* helpers
// or Option values.
type HTTPConfig struct {
	BaseURL     string            `json:"base_url"`
	EntityPath  string            `json:"entity_path"`
	APIKey      string            `json:"api_key,omitempty"`
	Timeout     time.Duration     `json:"timeout"`
	Headers     map[string]string `json:"headers,omitempty"`
	QueryParams map[string]string `json:"query_params,omitempty"`
	HTTPClient  *http.Client      `json:"-"`
}

// NewHTTPConfig returns an HTTPConfig for the given base URL and entity path
// with a 30-second default timeout and empty header and query-parameter maps.
// Each non-nil Option is applied in order, so later options override earlier
// ones. The returned config is not yet normalized; NewHTTP and
// NewHTTPTransport call Normalize for you.
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

// Normalize validates the config and rewrites the receiver in place with
// defaults applied (trimmed base URL, default timeout, copied header/query
// maps). It returns one of the package's sentinel errors (ErrBaseURLRequired,
// ErrBaseURLInvalid, ErrEntityPathRequired, ErrEntityPathInvalid, or
// ErrHeaderKeyRequired) if the config is invalid, matchable with errors.Is.
// The name makes the mutation explicit — unlike a plain Validate, calling this
// changes c.
func (c *HTTPConfig) Normalize() error {
	normalized, err := normalizeHTTPConfig(c)
	if err != nil {
		return err
	}
	*c = normalized
	return nil
}

func normalizeHTTPConfig(config *HTTPConfig) (HTTPConfig, error) {
	if config == nil {
		return HTTPConfig{}, ErrNilConfig
	}
	normalized := *config
	normalized.BaseURL = strings.TrimRight(strings.TrimSpace(normalized.BaseURL), "/")
	normalized.EntityPath = strings.Trim(strings.TrimSpace(normalized.EntityPath), "/")
	normalized.APIKey = strings.TrimSpace(normalized.APIKey)
	if normalized.BaseURL == "" {
		return HTTPConfig{}, ErrBaseURLRequired
	}
	parsed, err := url.Parse(normalized.BaseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return HTTPConfig{}, ErrBaseURLInvalid
	}
	if normalized.EntityPath == "" {
		return HTTPConfig{}, ErrEntityPathRequired
	}
	if strings.ContainsAny(normalized.EntityPath, "?#") {
		return HTTPConfig{}, ErrEntityPathInvalid
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
			return nil, fmt.Errorf("%w: %s", ErrHeaderKeyRequired, label)
		}
		out[key] = strings.TrimSpace(value)
	}
	return out, nil
}

// GetType reports the transport this config configures, always
// TransportTypeHTTP. It satisfies the Config interface.
func (c *HTTPConfig) GetType() TransportType {
	return TransportTypeHTTP
}

// WithHeader sets a request header to send on every request and returns the
// receiver for chaining. The key is trimmed; an empty key is ignored.
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

// WithQueryParam sets a static query parameter to append to every request URL
// and returns the receiver for chaining. The key is trimmed; an empty key is
// ignored.
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

// WithBearerToken sets an "Authorization: Bearer <token>" header and returns
// the receiver for chaining. The token is trimmed; an empty token is ignored.
func (c *HTTPConfig) WithBearerToken(token string) *HTTPConfig {
	token = strings.TrimSpace(token)
	if token == "" {
		return c
	}
	return c.WithHeader("Authorization", "Bearer "+token)
}

// WithHTTPClient sets a caller-owned *http.Client to use for requests and
// returns the receiver for chaining. When set, the config Timeout is not
// applied automatically; configure the timeout on the supplied client instead.
func (c *HTTPConfig) WithHTTPClient(client *http.Client) *HTTPConfig {
	c.HTTPClient = client
	return c
}
