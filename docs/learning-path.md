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
**Goal:** Protect routes with the JWT addon in a minimal auth flow.

You will learn:
- `jwt.New(jwt.Config{})` — initialize the JWT provider
- `jwtProvider.GenerateToken()` — issue signed tokens from a login endpoint
- `jwtProvider.Middleware()` — validate `Authorization: Bearer` headers
- `jwt.ClaimsFromCtx()` — read JWT claims inside handlers
- Role-based `RequireRole` middleware on top of addon claims

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

## Phase 5 — Addon Ecosystem (examples 11–14)

First-class addons for authentication and persistence.

### 11 — JWT Addon
**Goal:** Go deeper with the official JWT addon in a dedicated addon example.

You will learn:
- Installing an addon with `keel add jwt`
- `jwt.New(jwt.Config{})` — provider initialization
- `jwtProvider.GenerateToken()` / `jwtProvider.RefreshToken()` — token lifecycle
- `jwtProvider.Middleware()` — drop-in route protection
- `jwt.ClaimsFromCtx()` — reading JWT payload in handlers
- Role-based `RequireRole` guard on top of the addon

---

### 12 — OAuth2
**Goal:** Add social login without managing passwords.

You will learn:
- Installing the OAuth addon with `keel add oauth`
- `oauth.New(oauth.Config{})` — configuring GitHub and Google providers
- `oauth.NewController()` — auto-generated redirect and callback routes
- How the addon issues a signed JWT after a successful OAuth flow
- Combining `ss-keel-oauth` + `ss-keel-jwt` in a single service
- Extracting provider-specific claims (`sub`, `data.provider`, `data.avatar_url`)

---

### 13 — MongoDB
**Goal:** Persist documents in MongoDB using the official addon.

You will learn:
- Installing the addon with `keel add mongo`
- `mongo.New(mongo.Config{})` — client initialization and connection
- `mongo.EntityBase` — UUID ID, `CreatedAt`, `UpdatedAt` in milliseconds
- `mongo.NewRepository[T, ID]()` — generic typed repository
- `repo.FindAll()` with built-in pagination via `ctx.ParsePagination()`
- `OnCreate()` / `OnUpdate()` hooks for timestamp management
- `mongo.NewHealthChecker()` — MongoDB wired into `/health`

---

### 14 — Redis Cache
**Goal:** Add cache-aside reads and invalidation with the official Redis addon.

You will learn:
- Installing the addon with `keel add redis`
- `ssredis.New(ssredis.Config{})` — client initialization and connection
- `ssredis.NewHealthChecker()` — Redis wired into `/health`
- Injecting `*ssredis.Client` into a module
- Accepting `contracts.Cache` in the service layer
- Cache-aside reads with `Get` + `Set` and invalidation with `Delete`

---

## What Comes Next

After completing all examples:

1. Scaffold a new project: `keel new my-service`
2. Add a database: `keel add gorm` or `keel add mongo`
3. Add authentication: `keel add jwt` or `keel add oauth`
4. Add caching when needed: `keel add redis`
5. Read the [official docs](https://docs.keel-go.dev)
6. Explore [ss-keel-core](https://github.com/slice-soft/ss-keel-core) source code
