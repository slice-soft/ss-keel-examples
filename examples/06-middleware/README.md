# 06 — Middleware

Custom HTTP middleware: correlation ID, response timer, and IP blocklist.

## What This Demonstrates

- Writing `fiber.Handler` middleware functions
- Applying group-level middleware with `app.Group(prefix, middlewares...)`
- Applying route-level middleware with `.Use(middleware)`
- Reading and writing request/response headers
- Using `c.Locals()` to pass values between middleware and handlers
- Returning Keel errors (`core.Forbidden`, `core.Unauthorized`) from middleware

## Requirements

- Go 1.21+

## How to Run

```bash
cp .env.example .env
go mod download
go run main.go
```

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/ping` | Returns correlation ID; adds response time header |
| GET | `/api/internal` | Requires `X-Internal-Key` header |
| GET | `/health` | Health check |
| GET | `/docs` | OpenAPI UI |

## Middleware Included

| Middleware | Scope | Behavior |
|-----------|-------|---------|
| `CorrelationID` | Group | Injects `X-Correlation-ID` into response; forwards request value if present |
| `ResponseTimer` | Group | Adds `X-Response-Time` header |
| `IPBlocklist` | Group | Returns 403 for blocked IPs (set `BLOCKED_IPS` env var) |
| `requireInternalKey` | Route | Returns 401 unless `X-Internal-Key` header matches |

## Examples

```bash
# Check correlation and timing headers
curl -i http://localhost:7331/api/ping

# Forward your own correlation ID
curl -H "X-Correlation-ID: my-trace-abc" http://localhost:7331/api/ping

# Blocked without key
curl http://localhost:7331/api/internal
# {"status_code":401,"code":"UNAUTHORIZED",...}

# Access with key
curl -H "X-Internal-Key: dev-internal-key" http://localhost:7331/api/internal
```

## Concepts Covered

- `fiber.Handler` — Keel middleware signature
- `c.Next()` — chain to the next handler
- `c.Locals()` — share values across the request lifecycle
- `app.Group()` — attach middleware to a path prefix
- `.Use()` on a route — per-route middleware
- `core.Forbidden()` / `core.Unauthorized()` — structured error helpers
