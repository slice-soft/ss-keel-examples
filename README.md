# Keel Examples

A curated collection of example projects demonstrating how to build production-style services with the [Keel](https://keel-go.dev) framework and the [Keel CLI](https://github.com/slice-soft/ss-keel-cli).

Each example is small, focused, and self-contained — you can run any of them independently in under a minute.

---

## What is Keel?

[Keel](https://keel-go.dev) is a modular Go framework for building web services. It provides:

- Structured modules, controllers, and services
- Built-in health checks, request logging, and OpenAPI docs
- Composable middleware and guards
- Scheduler, event bus, and tracing hooks
- First-class validation via struct tags

---

## Learning Path

If you are new to Keel, work through the examples in order:

| Step | Example | Concept |
|------|---------|---------|
| 1 | [01-hello-world](./examples/01-hello-world) | Bootstrap a Keel app and define a route |
| 2 | [02-config-env](./examples/02-config-env) | Load configuration from environment variables |
| 3 | [03-health-check](./examples/03-health-check) | Add custom health checkers |
| 4 | [04-rest-crud](./examples/04-rest-crud) | Build a complete CRUD REST service |
| 5 | [05-validation](./examples/05-validation) | Validate request bodies with struct tags |
| 6 | [06-middleware](./examples/06-middleware) | Write and apply custom middleware |
| 7 | [07-jwt-auth](./examples/07-jwt-auth) | Protect routes with the ss-keel-jwt addon |
| 8 | [08-gorm-postgres](./examples/08-gorm-postgres) | Persist data with GORM and PostgreSQL |
| 9 | [09-scheduler-cron](./examples/09-scheduler-cron) | Schedule background jobs with cron |
| 10 | [10-addon-example](./examples/10-addon-example) | Integrate a Keel addon |
| 11 | [11-jwt-addon](./examples/11-jwt-addon) | JWT addon flow with refresh and RBAC |
| 12 | [12-oauth](./examples/12-oauth) | OAuth2 login with GitHub and Google |
| 13 | [13-mongo](./examples/13-mongo) | Document CRUD with MongoDB |

---

## Examples

### 01 — Hello World
Minimal Keel app with a single `GET /hello` route. The starting point for every Keel service.

### 02 — Config & Env
Structured configuration loader using environment variables with typed defaults.

### 03 — Health Check
Built-in `/health` endpoint plus a custom `HealthChecker` that inspects an in-memory dependency.

### 04 — REST CRUD
Full CRUD for a `Task` resource: list, get, create, update, and delete — all in-memory, no database required.

### 05 — Validation
Shows how Keel uses `validate` struct tags together with `ctx.ParseBody()` to return structured 422 errors automatically.

### 06 — Middleware
Custom request middleware: correlation ID injection, response timing header, and a simple IP blocklist.

### 07 — JWT Auth
JWT authentication using the [ss-keel-jwt](https://github.com/slice-soft/ss-keel-jwt) addon in a minimal login flow. Issues tokens on `POST /auth/login` and protects routes with the addon's reusable middleware.

### 08 — GORM + PostgreSQL
Database-backed CRUD using [ss-keel-gorm](https://github.com/slice-soft/ss-keel-gorm) with migrations, a repository pattern, and connection health checks.

### 09 — Scheduler / Cron
Register recurring background jobs with the Keel scheduler. Includes a simple in-memory job that runs on a configurable cron expression.

### 10 — Addon Example
Demonstrates how to consume a Keel addon installed via the Keel CLI (`keel add`).

### 11 — JWT Addon
A deeper [ss-keel-jwt](https://github.com/slice-soft/ss-keel-jwt) addon example focused on token refresh, claims inspection, and role-based access control on top of `jwtProvider.Middleware()`.

### 12 — OAuth2
Social login with GitHub and Google via the [ss-keel-oauth](https://github.com/slice-soft/ss-keel-oauth) addon. After the OAuth flow the addon issues a signed JWT so protected routes work identically to the JWT addon example.

### 13 — MongoDB
Document CRUD using the [ss-keel-mongo](https://github.com/slice-soft/ss-keel-mongo) addon: `EntityBase`, a generic typed repository, pagination, partial updates, and a built-in MongoDB health checker.

---

## How to Run an Example

Each example is an independent Go module. To run any example:

```bash
# 1. Enter the example directory
cd examples/01-hello-world

# 2. Copy the environment file
cp .env.example .env

# 3. Download dependencies
go mod download

# 4. Run the service
go run main.go
```

The server starts on port **7331** by default.

Open the interactive API docs at [http://localhost:7331/docs](http://localhost:7331/docs).

> **Note:** Some examples (08-gorm-postgres, 12-oauth, 13-mongo) require Docker or external credentials. See each example's README for details.

---

## Repository Structure

```
ss-keel-examples/
├── examples/          # One subdirectory per example
│   ├── 01-hello-world/
│   ├── 02-config-env/
│   └── ...
├── shared/
│   ├── docker/        # Shared Docker Compose files for dependencies
│   └── scripts/       # Helper shell scripts
└── docs/
    ├── learning-path.md
    └── example-matrix.md
```

---

## Related Projects

| Project | Description |
|---------|-------------|
| [ss-keel-core](https://github.com/slice-soft/ss-keel-core) | The Keel framework |
| [ss-keel-cli](https://github.com/slice-soft/ss-keel-cli) | CLI for scaffolding Keel projects |
| [ss-keel-docs](https://github.com/slice-soft/ss-keel-docs) | Official documentation |
| [ss-keel-addon-template](https://github.com/slice-soft/ss-keel-addon-template) | Template for creating Keel addons |

---

## Contributing

Examples that fix bugs, improve clarity, or add new concepts are welcome. Please open an issue first to discuss the change.

## License

[MIT](./LICENSE) — SliceSoft
