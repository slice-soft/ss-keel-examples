# 08 — GORM + PostgreSQL

Database-backed CRUD using [ss-keel-gorm](https://github.com/slice-soft/ss-keel-gorm) and PostgreSQL.

## What This Demonstrates

- Connecting to PostgreSQL with `database.Connect()`
- Auto-migrating a GORM model with `db.AutoMigrate()`
- Soft deletes via `gorm.DeletedAt`
- Implementing `core.HealthChecker` that pings the database
- Repository operations: `Find`, `First`, `Create`, `Model.Updates`, `Delete`

## Requirements

- Go 1.21+
- Docker (for PostgreSQL)

## How to Run

**1. Start PostgreSQL:**
```bash
docker run -d \
  --name keel-postgres \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=keelexamples \
  -p 5432:5432 \
  postgres:16-alpine
```

Or use the shared Docker Compose file:
```bash
docker compose -f ../../shared/docker/postgres.yml up -d
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
| GET | `/api/v1/products` | List all products |
| GET | `/api/v1/products/:id` | Get product by ID |
| POST | `/api/v1/products` | Create a product |
| PATCH | `/api/v1/products/:id` | Update a product |
| DELETE | `/api/v1/products/:id` | Soft delete a product |
| GET | `/health` | Health check (includes DB ping) |
| GET | `/docs` | OpenAPI UI |

## Examples

```bash
# Create a product
curl -X POST http://localhost:7331/api/v1/products \
  -H "Content-Type: application/json" \
  -d '{"name":"Mechanical Keyboard","price":129.99,"stock":50}'

# List products
curl http://localhost:7331/api/v1/products

# Health check with database status
curl http://localhost:7331/health
```

## Concepts Covered

- `database.Connect()` from `ss-keel-gorm`
- `db.AutoMigrate()` — schema migrations on startup
- `gorm.DeletedAt` — soft deletes
- `core.HealthChecker` with database ping
- `db.Model(&entity).Updates(map)` — partial updates
- `core.Internal()` — wrapping database errors
