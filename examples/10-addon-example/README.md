# 10 — Addon Example

Demonstrates how to integrate a Keel addon into a service.

## What This Demonstrates

- The addon contract: `keel-addon.json`
- How `keel add <name>` installs an addon into your project
- Consuming an addon package in `main.go`
- The rate limiter addon: IP-based sliding window with response headers
- Per-route vs group-level addon usage

## About Keel Addons

A Keel addon is a Go module that provides reusable infrastructure functionality (rate limiting, caching, logging, database, etc.). It ships a `keel-addon.json` manifest that the Keel CLI uses to install and integrate it automatically.

**To install a verified addon:**
```bash
keel add ratelimit
```

**To install an unofficial addon:**
```bash
keel add github.com/example/keel-addon-ratelimit
```

In this example, the `ratelimit` addon lives in `internal/addon/ratelimit/` to keep the example self-contained. In a real project it would be an external Go module.

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
| GET | `/api/ping` | Rate-limited ping |
| GET | `/api/data` | Rate-limited data endpoint |
| GET | `/admin/ratelimit/stats` | Current request counts per IP |
| GET | `/health` | Health check |
| GET | `/docs` | OpenAPI UI |

## Rate Limit Headers

Every response from `/api/*` includes:

| Header | Description |
|--------|-------------|
| `X-RateLimit-Limit` | Max requests per window |
| `X-RateLimit-Remaining` | Remaining requests |
| `X-RateLimit-Reset` | Unix timestamp when the window resets |

## Examples

```bash
# Make requests — watch X-RateLimit-Remaining decrease
curl -i http://localhost:7331/api/ping

# After 10 requests → 429
curl http://localhost:7331/api/ping
# {"status_code":429,"code":"RATE_LIMITED","message":"too many requests — slow down"}

# Check live stats
curl http://localhost:7331/admin/ratelimit/stats
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `RATE_LIMIT_MAX` | `10` | Max requests per window |
| `RATE_LIMIT_WINDOW` | `1m` | Sliding window duration |

## Concepts Covered

- `keel-addon.json` — addon manifest (name, version, steps)
- `keel add` — CLI command to install and integrate addons
- Addon as a Go package with a `New(Config)` factory
- Group-level addon middleware via `app.Group(prefix, addon.Middleware())`
- Response header injection from middleware
- `&core.KError{}` — returning custom HTTP status codes (429)
