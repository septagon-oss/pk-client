# pk-client

Public Go client primitives for PlatformKit APIs.

This repo intentionally stays transport-focused and free of private
`septagon-dev` imports. Pro clients can wrap these types with generated SDKs,
authentication providers, telemetry, or hosted defaults.

## Current Surface

- generic CRUD client wrapper
- HTTP transport with headers, query params, bearer/API-key auth, and timeouts
- transport-safe request and response DTOs

## Verify

```bash
make test
```
