# pk-client

> Part of [PlatformKit](https://github.com/septagon-oss/platformkit) — the open-source Go backend for multi-tenant SaaS.

**Depends on.** `pk-shared` only (for canonical opaque-ID path segments). Nothing else in PlatformKit; no third-party dependencies.

[![Go Reference](https://pkg.go.dev/badge/github.com/septagon-oss/pk-client.svg)](https://pkg.go.dev/github.com/septagon-oss/pk-client)
[![CI](https://github.com/septagon-oss/pk-client/actions/workflows/go.yml/badge.svg)](https://github.com/septagon-oss/pk-client/actions/workflows/go.yml)
[![License: Apache-2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)

pk-client is a small, dependency-free Go client for PlatformKit-style CRUD APIs. It pairs a generic, transport-agnostic `Client[T]` facade with a batteries-included standard-library HTTP transport, covering single-item CRUD, partial updates, bulk operations, export, and import over typed request and response envelopes. It is part of the OSS PlatformKit family and intentionally stays transport-focused and free of private upstream imports, so downstream SDKs can wrap it with auth providers, telemetry, or hosted defaults.

## Install

```bash
go get github.com/septagon-oss/pk-client@v0.1.0
```

## Usage

```go
package main

import (
	"context"
	"fmt"

	client "github.com/septagon-oss/pk-client"
)

type Widget struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func main() {
	c, err := client.NewHTTP[Widget](client.NewHTTPConfig(
		"https://api.example.com",
		"/api/widgets",
		client.WithBearerToken("token"),
	))
	if err != nil {
		panic(err)
	}

	created, err := c.Create(context.Background(), client.CreateInput[Widget]{
		Body: Widget{Name: "gadget"},
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("created", created.Data.ID)
}
```

## Current Surface

- generic `Client[T]` facade over a pluggable `CRUDTransport[T]` (`New`, `NewHTTP`, `WithTransport`)
- single-item operations: `Create`, `GetByID`, `List`, `Update`, `PartialUpdate`, `Delete`
- bulk and data operations: `BulkCreate`, `BulkUpdate`, `BulkDelete`, `Export`, `Import`
- standard-library HTTP transport with headers, static query params, bearer/API-key auth, timeouts, and a caller-supplied `*http.Client`
- `HTTPConfig` builder plus composable `Option` values (`WithHeader`, `WithQueryParam`, `WithTimeout`, `WithBearerToken`, `WithAPIKey`)
- typed request/response envelopes (`CreateInput`, `ListParams`, `Filter`, `ItemResponse`, `ListResponse`, `BulkResponse`, `ImportResponse`)
- matchable errors: `ErrNoTransport`, config sentinels (`ErrBaseURLRequired`, …), and a structured `*APIError` recoverable with `errors.As`

## Verify

```bash
make verify   # go test + go vet + staticcheck + race
```

## License

Apache-2.0. See [LICENSE](LICENSE).
