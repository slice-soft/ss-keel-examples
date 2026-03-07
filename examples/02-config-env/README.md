# 02 — Config & Env

Structured configuration loading from environment variables with typed defaults.

## What This Demonstrates

- Defining a typed `Config` struct
- Loading environment variables with `os.Getenv` and safe defaults
- Type-converting strings to `int` and `bool`
- Exposing config values via an endpoint for debugging

## Requirements

- Go 1.21+

## How to Run

```bash
cp .env.example .env
# Edit .env with your values
go mod download
go run main.go
```

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/config` | Returns the active configuration |
| GET | `/health` | Built-in health check |
| GET | `/docs` | Interactive OpenAPI UI |

## Examples

```bash
# View active config
curl http://localhost:7331/config

# Override values at runtime
PORT=8080 ENV=production go run main.go
```

## Environment Variables

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `PORT` | int | `7331` | HTTP server port |
| `ENV` | string | `development` | Runtime environment |
| `SERVICE_NAME` | string | `config-env` | Service identifier |
| `API_VERSION` | string | `v1` | API version string |
| `LOG_LEVEL` | string | `info` | Logging verbosity |
| `MAX_PAGE_SIZE` | int | `50` | Pagination limit |
| `FEATURE_FLAG` | bool | `false` | Feature toggle |

## Concepts Covered

- Typed configuration struct with a `Load()` factory
- `os.Getenv` with fallback helper
- `strconv.Atoi` / `strconv.ParseBool` for safe type conversion
- Config injection via closure into route handlers
