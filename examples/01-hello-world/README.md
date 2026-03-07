# 01 — Hello World

The minimal Keel application. One file, one route, zero boilerplate.

## What This Demonstrates

- Bootstrapping a Keel app with `core.New()`
- Registering an inline route with `core.ControllerFunc`
- Reading query parameters with `ctx.Query()`
- Responding with `ctx.OK()`
- The built-in `/health` and `/docs` endpoints

## Requirements

- Go 1.21+

## How to Run

```bash
cp .env.example .env
go mod download
go run main.go
```

The server starts on [http://localhost:7331](http://localhost:7331).

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/hello` | Returns a greeting |
| GET | `/health` | Built-in health check |
| GET | `/docs` | Interactive OpenAPI UI |

## Examples

```bash
# Basic greeting
curl http://localhost:7331/hello
# {"message":"Hello, world!"}

# Greeting with name
curl "http://localhost:7331/hello?name=Keel"
# {"message":"Hello, Keel!"}
```

## Concepts Covered

- `core.New(core.KConfig{})` — application bootstrap
- `core.ControllerFunc` — inline route registration
- `core.GET()` — route builder
- `ctx.Query()` — read query string parameters
- `ctx.OK()` — JSON 200 response
- `.Tag()`, `.Describe()`, `.WithQueryParam()`, `.WithResponse()` — OpenAPI documentation
