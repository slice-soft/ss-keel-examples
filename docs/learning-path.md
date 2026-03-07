# Learning Path

A guided progression through the Keel examples — from zero to production-ready.

---

## Phase 1 — Foundation (examples 01–03)

Start here if you have never used Keel before.

### 01 — Hello World
**Goal:** Run your first Keel service.

You will learn:
- `core.New(core.KConfig{})` — application bootstrap
- `core.ControllerFunc` — inline route registration without a module
- `core.GET()` — the route builder
- `ctx.OK()` — JSON 200 response

After this example you should be able to start a server and hit an endpoint.

---

### 02 — Config & Env
**Goal:** Manage configuration properly.

You will learn:
- Defining a typed `Config` struct
- Loading every value from environment variables with safe defaults
- Type-converting strings to `int` and `bool` with `strconv`

After this example you should never use `os.Getenv` raw again.

---

### 03 — Health Check
**Goal:** Make your service observable.

You will learn:
- The `core.HealthChecker` interface
- `app.RegisterHealthChecker()` — contributing to `/health`
- How Keel aggregates checkers and returns HTTP 503 on failure

After this example your services will be ready for load balancers and container orchestrators.

---

## Phase 2 — API Building (examples 04–06)

Build real HTTP APIs.

### 04 — REST CRUD
**Goal:** Build a complete CRUD API.

You will learn:
- `app.Group("/api/v1")` — path prefix and shared middleware
- All five HTTP verbs with Keel route builders
- `ctx.ParseBody()` — parsing and validating request bodies
- `ctx.Created()` / `ctx.NoContent()` — typed HTTP responses
- Thread-safe in-memory storage with `sync.RWMutex`

---

### 05 — Validation
**Goal:** Return structured errors automatically.

You will learn:
- All common `validate` struct tags: `required`, `email`, `min`, `max`, `gt`, `gte`, `url`, `e164`
- `omitempty` for optional fields
- `dive` for validating slice elements
- The 422 Unprocessable Entity response format Keel returns

---

### 06 — Middleware
**Goal:** Intercept and enrich requests.

You will learn:
- Writing `fiber.Handler` functions
- Group-level vs route-level middleware
- `c.Locals()` — passing values between middleware and handlers
- Response header injection
- Returning Keel errors from middleware

---

## Phase 3 — Security & Persistence (examples 07–08)

Production essentials.

### 07 — JWT Auth
**Goal:** Protect routes with authentication.

You will learn:
- Issuing HS256 JWTs with `golang-jwt/jwt`
- `Authorization: Bearer` header validation
- `c.Locals("_keel_user")` — authenticated user propagation
- `core.UserAs[T]()` — type-safe principal extraction
- Role-based `RequireRole` middleware

---

### 08 — GORM + PostgreSQL
**Goal:** Persist data in a real database.

You will learn:
- `database.New()` from `ss-keel-gorm`
- `db.AutoMigrate()` — schema migration on startup
- GORM model patterns: `gorm.DeletedAt` soft deletes
- `database.NewHealthChecker()` — built-in DB health check
- Partial updates with `db.Model(&entity).Updates(map)`

---

## Phase 4 — Background Work & Addons (examples 09–10)

Beyond request-response.

### 09 — Scheduler / Cron
**Goal:** Run recurring background jobs.

You will learn:
- `core.Scheduler` interface: `Add(core.Job)`, `Start()`, `Stop(ctx)`
- `core.Job` — name, schedule, handler
- `app.RegisterScheduler()` — lifecycle integration
- Implementing a ticker-based scheduler

---

### 10 — Addon Example
**Goal:** Understand the Keel addon ecosystem.

You will learn:
- What a `keel-addon.json` manifest looks like
- How `keel add <name>` installs and integrates an addon
- Consuming an addon package in `main.go`
- The rate limiter addon pattern: sliding window, response headers

---

## What Comes Next

After completing all examples:

1. Scaffold a new project: `keel new my-service`
2. Add a database: `keel add gorm`
3. Read the [official docs](https://docs.keel-go.dev)
4. Explore [ss-keel-core](https://github.com/slice-soft/ss-keel-core) source code
