# 13 — MongoDB

Notes CRUD API using [ss-keel-mongo](https://github.com/slice-soft/ss-keel-mongo).

> **Note:** This example uses a compact flat layout (`main.go` at the root) to keep all the relevant code in one place and let you focus on the API surface — no CLI scaffolding required. Projects created with `keel new` + `keel generate module posts --mongo` produce a module-per-folder structure under `internal/modules/` with separate entity, repository, service, controller, and DTO files.

## What This Demonstrates

- Connecting to MongoDB with `mongo.New(...)`
- Registering `mongo.NewHealthChecker(...)` for `/health`
- Wrapping `mongo.MongoRepository[T, ID]` in a module-level repository
- Using `mongo.EntityBase` for `ID`, `CreatedAt`, `UpdatedAt` with BSON tags
- Calling `entity.OnCreate()` / `entity.OnUpdate()` for timestamp management
- Separating the domain entity from the internal BSON document type
- Pagination with `httpx.PageQuery` and `httpx.Page[T]`

## Requirements

- Go 1.25+
- MongoDB 6+ (local or Docker)

## How to Run

**1. Start MongoDB:**
```bash
docker run -d \
  --name keel-mongo \
  -p 27017:27017 \
  mongo:7
```

**2. Run the example:**
```bash
cp application.properties application.properties.local  # optional, defaults work out of the box
go mod download
go run main.go
```

The server starts on port **7331**.

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/notes` | List notes (paginated) |
| POST | `/api/v1/notes` | Create a note |
| GET | `/api/v1/notes/:id` | Get a note by ID |
| PUT | `/api/v1/notes/:id` | Replace all fields (PUT semantics) |
| PATCH | `/api/v1/notes/:id` | Update provided fields only (PATCH semantics) |
| DELETE | `/api/v1/notes/:id` | Delete a note |
| GET | `/health` | Health check (includes MongoDB ping) |
| GET | `/docs` | OpenAPI UI |

## Examples

```bash
# Create a note
curl -s -X POST http://localhost:7331/api/v1/notes \
  -H "Content-Type: application/json" \
  -d '{"title":"Buy groceries","body":"Milk, eggs, bread"}' | jq .

# List notes
curl -s "http://localhost:7331/api/v1/notes?page=1&limit=10" | jq .

# Partial update
curl -s -X PATCH http://localhost:7331/api/v1/notes/<id> \
  -H "Content-Type: application/json" \
  -d '{"body":"Updated body only"}' | jq .
```
