# 14 — Redis Cache

Cache-aside notes API using [ss-keel-redis](https://github.com/slice-soft/ss-keel-redis).

## What This Demonstrates

- Connecting to Redis with `ssredis.New(...)`
- Registering `ssredis.NewHealthChecker(...)` for `/health`
- Injecting `*ssredis.Client` into a module
- Using `contracts.Cache` in the service layer
- Cache-aside reads on `GET /api/v1/notes/:id`
- Cache invalidation on writes and deletes

## Requirements

- Go 1.25+
- Redis 7+ (local or Docker)

## How to Run

**1. Start Redis:**
```bash
docker run -d \
  --name keel-redis \
  -p 6379:6379 \
  redis:7-alpine
```

**2. Run the example:**
```bash
cp .env.example .env
go mod download
go run main.go
```

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/notes/:id` | Read a note using cache-aside (`source` is `cache` or `store`) |
| POST | `/api/v1/notes` | Create or replace a note and invalidate its cache entry |
| DELETE | `/api/v1/notes/:id` | Delete a note and invalidate its cache entry |
| GET | `/health` | Health check (includes Redis ping) |
| GET | `/docs` | OpenAPI UI |

## Examples

```bash
# First read: comes from the in-memory store and fills Redis
curl http://localhost:7331/api/v1/notes/note-1

# Second read: same note now comes from Redis
curl http://localhost:7331/api/v1/notes/note-1

# Update a note and invalidate the cache
curl -X POST http://localhost:7331/api/v1/notes \
  -H "Content-Type: application/json" \
  -d '{"id":"note-1","title":"Updated title","body":"Redis cache invalidated"}'

# Delete a note and invalidate the cache again
curl -X DELETE http://localhost:7331/api/v1/notes/note-1
```

## Concepts Covered

- `contracts.Cache` in the service layer instead of a Redis-specific interface
- `ssredis.Client` injection at module bootstrap time
- Cache-aside reads with `Get` + `Set`
- Cache invalidation with `Delete`
- `core.NotFound()` and `core.Internal()` for infrastructure-backed handlers
