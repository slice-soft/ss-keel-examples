# 05 — Validation

Request body validation with struct tags. Keel returns structured 422 errors per field automatically.

## What This Demonstrates

- `validate` struct tags: `required`, `min`, `max`, `email`, `url`, `gte`, `lte`, `gt`, `e164`, `omitempty`
- Nested struct validation with `dive`
- `ctx.ParseBody()` returns 400 for malformed JSON, 422 for rule violations
- Per-field error objects `[{"field":"Email","message":"must be a valid email"}]`

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
| POST | `/users/register` | Register with rich field validation |
| POST | `/orders` | Place an order with nested item validation |
| GET | `/health` | Health check |
| GET | `/docs` | OpenAPI UI |

## Examples

```bash
# Valid request → 201
curl -X POST http://localhost:7331/users/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Jane","email":"jane@example.com","password":"secret123","age":30}'

# Invalid → 422 with field errors
curl -X POST http://localhost:7331/users/register \
  -H "Content-Type: application/json" \
  -d '{"name":"J","email":"bad","password":"short","age":15}'
```

## Validation Tags Used

| Tag | Meaning |
|-----|---------|
| `required` | Field must be present and non-zero |
| `min=N` | String: min length N |
| `max=N` | String: max length N |
| `email` | Must be a valid email address |
| `url` | Must be a valid URL |
| `gte=N` / `lte=N` | Number: >= N / <= N |
| `gt=0` | Number must be positive |
| `e164` | Phone number in E.164 format |
| `omitempty` | Skip validation if field is zero/empty |
| `dive` | Validate each element in a slice |

## Concepts Covered

- Struct tags with `go-playground/validator`
- `ctx.ParseBody()` — automatic 400/422 handling
- Nested validation with `dive`
- Structured error response format
