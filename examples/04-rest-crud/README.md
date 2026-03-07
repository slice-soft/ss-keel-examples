# 04 — REST CRUD

Full CRUD API for a `Task` resource using in-memory storage and route groups.

## What This Demonstrates

- All five HTTP verbs: GET, POST, PATCH, DELETE
- Route groups with `app.Group("/api/v1")`
- Request body parsing and validation with `ctx.ParseBody()`
- `core.NotFound()` error responses
- `ctx.Created()` for 201 responses
- `ctx.NoContent()` for 204 responses
- Thread-safe in-memory storage with `sync.RWMutex`

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
| GET | `/api/v1/tasks` | List all tasks |
| GET | `/api/v1/tasks/:id` | Get task by ID |
| POST | `/api/v1/tasks` | Create a task |
| PATCH | `/api/v1/tasks/:id` | Update a task |
| DELETE | `/api/v1/tasks/:id` | Delete a task |
| GET | `/health` | Health check |
| GET | `/docs` | OpenAPI UI |

## Examples

```bash
# Create a task
curl -X POST http://localhost:7331/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{"title":"Buy groceries","description":"Milk, eggs, bread"}'

# List tasks
curl http://localhost:7331/api/v1/tasks

# Mark as done
curl -X PATCH http://localhost:7331/api/v1/tasks/task_1 \
  -H "Content-Type: application/json" \
  -d '{"done":true}'

# Delete
curl -X DELETE http://localhost:7331/api/v1/tasks/task_1
```

## Concepts Covered

- `app.Group()` — route prefix + shared middlewares
- `core.GET/POST/PATCH/DELETE()` — route builder
- `ctx.ParseBody()` — JSON parsing + validation
- `ctx.OK()` / `ctx.Created()` / `ctx.NoContent()` — typed responses
- `core.NotFound()` — structured 404 error
- `sync.RWMutex` — concurrent-safe in-memory store
