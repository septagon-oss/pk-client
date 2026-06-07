package client

// options.go owns composable HTTP client options for typed PlatformKit clients.
//
// ADR: ADR-0029 (file purpose declaration).
// Convention: C-14 (every Go file declares its purpose).

import "time"

// Option configures an HTTPConfig during construction. Options are applied in
// order by NewHTTPConfig, so a later Option overrides an earlier one. The With*
// constructors in this package return ready-to-use Option values.
type Option func(*HTTPConfig)

// WithHeader returns an Option that sets a request header sent on every request.
// An empty key is ignored.
func WithHeader(key, value string) Option {
	return func(config *HTTPConfig) {
		config.WithHeader(key, value)
	}
}

// WithQueryParam returns an Option that sets a static query parameter appended
// to every request URL. An empty key is ignored.
func WithQueryParam(key, value string) Option {
	return func(config *HTTPConfig) {
		config.WithQueryParam(key, value)
	}
}

// WithTimeout returns an Option that sets the request timeout. A non-positive
// timeout is replaced with the 30-second default during normalization. The
// timeout is ignored when a custom *http.Client is supplied via
// HTTPConfig.WithHTTPClient.
func WithTimeout(timeout time.Duration) Option {
	return func(config *HTTPConfig) {
		config.Timeout = timeout
	}
}

// WithBearerToken returns an Option that sets an "Authorization: Bearer
// <token>" header. An empty token is ignored.
func WithBearerToken(token string) Option {
	return func(config *HTTPConfig) {
		config.WithBearerToken(token)
	}
}

// WithAPIKey returns an Option that sets the API key sent as a bearer token in
// the Authorization header when no explicit Authorization header is present.
func WithAPIKey(apiKey string) Option {
	return func(config *HTTPConfig) {
		config.APIKey = apiKey
	}
}
