# 07 — JWT Auth

JWT-based authentication with role-based access control.

## What This Demonstrates

- Issuing HS256 JWTs with `golang-jwt/jwt`
- A reusable `JWTMiddleware` that validates the `Authorization: Bearer` header
- Storing the authenticated principal in Fiber locals via `c.Locals("_keel_user")`
- Retrieving the principal in handlers with `core.UserAs[Principal](ctx)`
- Role-based `RequireRole` middleware
- Protecting a route group with `app.Group("/api", JWTMiddleware(...))`

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
| POST | `/auth/login` | Exchange credentials for a JWT |
| GET | `/api/me` | Current user info (requires JWT) |
| GET | `/api/admin/dashboard` | Admin only (requires JWT + role=admin) |
| GET | `/health` | Health check |
| GET | `/docs` | OpenAPI UI |

## Demo Users

| Email | Password | Role |
|-------|----------|------|
| `alice@example.com` | `password123` | admin |
| `bob@example.com` | `pass456` | user |

## Examples

```bash
# Login and capture the token
TOKEN=$(curl -s -X POST http://localhost:7331/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@example.com","password":"password123"}' | \
  grep -o '"token":"[^"]*"' | cut -d'"' -f4)

# Get current user
curl -H "Authorization: Bearer $TOKEN" http://localhost:7331/api/me

# Access admin dashboard (alice is admin → 200)
curl -H "Authorization: Bearer $TOKEN" http://localhost:7331/api/admin/dashboard

# Login as bob and try admin dashboard (bob is user → 403)
BOB_TOKEN=$(curl -s -X POST http://localhost:7331/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"bob@example.com","password":"pass456"}' | \
  grep -o '"token":"[^"]*"' | cut -d'"' -f4)
curl -H "Authorization: Bearer $BOB_TOKEN" http://localhost:7331/api/admin/dashboard
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `JWT_SECRET` | `change-me-in-production` | HS256 signing key |
| `TOKEN_TTL_MINUTES` | `60` | Token expiry in minutes |

## Concepts Covered

- `golang-jwt/jwt` — token issuance and validation
- `fiber.Handler` — guard middleware pattern
- `c.Locals()` — authenticated user propagation
- `core.UserAs[T]()` — type-safe principal extraction
- Group middleware vs route middleware
- `core.Unauthorized()` / `core.Forbidden()` — structured error responses
