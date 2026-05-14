package client

import "time"

type Option func(*HTTPConfig)

func WithHeader(key, value string) Option {
	return func(config *HTTPConfig) {
		config.WithHeader(key, value)
	}
}

func WithQueryParam(key, value string) Option {
	return func(config *HTTPConfig) {
		config.WithQueryParam(key, value)
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(config *HTTPConfig) {
		config.Timeout = timeout
	}
}

func WithBearerToken(token string) Option {
	return func(config *HTTPConfig) {
		config.WithBearerToken(token)
	}
}

func WithAPIKey(apiKey string) Option {
	return func(config *HTTPConfig) {
		config.APIKey = apiKey
	}
}
