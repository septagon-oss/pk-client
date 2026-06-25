# pk-client Charter

## Purpose

Public HTTP client primitives for PlatformKit OSS. A generic, typed API client that works with any PlatformKit-compatible server.

## In Scope

- Generic CRUD client: typed `Get`, `List`, `Create`, `Update`, `Delete` operations
- HTTP transport: config, middleware, error handling, response envelopes
- Bulk operations: batch create, update, delete
- Typed request/response envelopes
- Matchable error types

## Out of Scope

- Server or handler implementations
- Authentication flows (OAuth, session management)
- WebSocket or streaming transports
- CLI wrappers (handled by pk-tools)

## Dependencies

None (zero-dependency module).
