# 11 — JWT Addon

Login, token refresh, and route protection using [ss-keel-jwt](https://github.com/slice-soft/ss-keel-jwt).

> **Note:** This example uses a compact flat layout (`main.go` at the root) to keep all the relevant code in one place and let you focus on the API surface — no CLI scaffolding required. Projects created with `keel new` place the entry point at `cmd/main.go` and generate a module-per-folder structure instead.

## What This Demonstrates

- Initializing `*jwt.JWT` manually with `jwt.New(...)`
- Generating a signed token with `jwtProvider.GenerateToken(...)`
- Protecting a route group with `jwtProvider.Middleware()`
- Refreshing an existing token with `jwtProvider.RefreshToken(...)`
- Reading JWT claims inside a handler with `jwt.ClaimsFromCtx(...)`
- Role-based access control as a plain Fiber middleware

## Requirements

- Go 1.25+

## How to Run

```bash
cp application.properties application.properties.local  # optional, defaults work out of the box
go mod download
go run main.go
```

The server starts on port **7331**.

## Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/auth/login` | — | Exchange credentials for a JWT |
| POST | `/auth/refresh` | — | Refresh a valid token |
| GET | `/api/me` | Bearer | Return the decoded JWT payload |
| GET | `/api/admin` | Bearer + role=admin | Admin-only route |
| GET | `/health` | — | Health check |
| GET | `/docs` | — | OpenAPI UI |

## Test Credentials

| Email | Password | Role |
|-------|----------|------|
| `alice@example.com` | `password123` | `admin` |
| `bob@example.com` | `pass456` | `member` |

## Examples

```bash
# Login
curl -s -X POST http://localhost:7331/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@example.com","password":"password123"}' | jq .

# Protected route
TOKEN=<token from above>
curl -s http://localhost:7331/api/me \
  -H "Authorization: Bearer $TOKEN" | jq .

# Admin route
curl -s http://localhost:7331/api/admin \
  -H "Authorization: Bearer $TOKEN" | jq .
```
