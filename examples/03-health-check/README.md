# 03 — Health Check

Custom health checkers plugged into Keel's built-in `/health` endpoint.

## What This Demonstrates

- Implementing the `core.HealthChecker` interface (`Name()`, `Check()`)
- Registering multiple health checkers with `app.RegisterHealthChecker()`
- Keel aggregates all checkers: if any returns an error, the response is HTTP 503
- Using `sync/atomic` for thread-safe state without a mutex

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
| GET | `/health` | Aggregated health status |
| POST | `/demo/cache/down` | Simulate cache failure |
| POST | `/demo/cache/up` | Restore cache |
| POST | `/demo/database/down` | Simulate database failure |
| POST | `/demo/database/up` | Restore database |
| GET | `/docs` | Interactive OpenAPI UI |

## Examples

```bash
# All healthy → 200
curl http://localhost:7331/health
# {"status":"UP","service":"health-check","version":"1.0.0","checks":{"cache":"UP","database":"UP"}}

# Simulate a failure
curl -X POST http://localhost:7331/demo/cache/down

# Now health returns 503
curl http://localhost:7331/health
# {"status":"DOWN","service":"health-check",...,"checks":{"cache":"DOWN: cache is not ready","database":"UP"}}
```

## Concepts Covered

- `core.HealthChecker` interface
- `app.RegisterHealthChecker()` — pluggable health contributions
- Automatic HTTP 503 on `DOWN` status
- `sync/atomic.Bool` for safe concurrent state
