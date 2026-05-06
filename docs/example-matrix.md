# Example Matrix

Quick reference: which example covers which Keel concept.

## API Concepts

| Concept | 01 | 02 | 03 | 04 | 05 | 06 | 07 | 08 | 09 | 10 | 11 | 12 | 13 | 14 | 15 | 16 |
|---------|----|----|----|----|----|----|----|----|----|----|----|----|-----|-----|-----|-----|
| `core.New()` | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| `core.ControllerFunc` | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| `app.Group()` | | | | ✓ | | ✓ | ✓ | ✓ | | ✓ | ✓ | ✓ | ✓ | ✓ | | ✓ |
| Module pattern | | | | | | | | | | | | | | | | |
| `ctx.OK()` | ✓ | ✓ | | ✓ | | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| `ctx.Created()` | | | | ✓ | ✓ | | | ✓ | | | | | ✓ | ✓ | ✓ | ✓ |
| `ctx.NoContent()` | | | | ✓ | | | | ✓ | | | | | ✓ | ✓ | | |
| `ctx.ParseBody()` | | | | ✓ | ✓ | | ✓ | ✓ | | | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| `ctx.Params()` | | | | ✓ | | | | ✓ | | | | | ✓ | ✓ | | ✓ |
| `ctx.Query()` | ✓ | | | | | | | | | | | | | | | |
| Route `.Tag()` | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| Route `.Describe()` | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| Route `.Use()` | | | | | | ✓ | ✓ | | | ✓ | ✓ | | | | | |
| Route `.WithBody()` | | | | ✓ | ✓ | | ✓ | ✓ | | | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| Route `.WithResponse()` | ✓ | ✓ | | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| Route `.WithSecured()` | | | | | | | ✓ | | | | ✓ | ✓ | | | | |
| DELETE route | | | | ✓ | | | | ✓ | | | | | ✓ | ✓ | | |
| PATCH route | | | | ✓ | | | | ✓ | | | | | ✓ | | | |

## Error Handling

| Error type | 01 | 02 | 03 | 04 | 05 | 06 | 07 | 08 | 09 | 10 | 11 | 12 | 13 | 14 | 15 | 16 |
|-----------|----|----|----|----|----|----|----|----|----|----|----|----|-----|-----|-----|-----|
| `core.NotFound()` | | | | ✓ | | | | ✓ | | | | | ✓ | ✓ | | ✓ |
| `core.Unauthorized()` | | | | | | ✓ | ✓ | | | | ✓ | ✓ | | | | |
| `core.Forbidden()` | | | | | | ✓ | ✓ | | | | ✓ | | | | | |
| `core.BadRequest()` | | | | | | | | | | | | | | | | |
| `core.Internal()` | | | | | | | ✓ | ✓ | | | ✓ | | ✓ | ✓ | | |
| `core.Conflict()` | | | | | | | | | | | | | | | | |
| `&core.KError{}` (custom) | | | | | | | | | | ✓ | | | | | | |
| 422 via `ctx.ParseBody()` | | | | ✓ | ✓ | | ✓ | ✓ | | | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |

## Infrastructure

| Feature | 01 | 02 | 03 | 04 | 05 | 06 | 07 | 08 | 09 | 10 | 11 | 12 | 13 | 14 | 15 | 16 |
|---------|----|----|----|----|----|----|----|----|----|----|----|----|-----|-----|-----|-----|
| `config.MustLoadConfig` | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| `application.properties` | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| Health checker | | | ✓ | | | | | ✓ | | | | | ✓ | ✓ | | |
| Scheduler | | | | | | | | | ✓ | | | | | | | |
| JWT auth | | | | | | | ✓ | | | | ✓ | ✓ | | | | |
| Middleware (custom) | | | | | | ✓ | ✓ | | | | ✓ | | | | | |
| GORM + PostgreSQL | | | | | | | | ✓ | | | | | | | | |
| MongoDB | | | | | | | | | | | | | ✓ | | | |
| OAuth2 | | | | | | | | | | | | ✓ | | | | |
| Redis cache | | | | | | | | | | | | | | ✓ | | |
| DevPanel observability | | | | | | | | | | | | | | | ✓ | |
| `panel.RequestMiddleware()` | | | | | | | | | | | | | | | ✓ | |
| `panel.Logger()` | | | | | | | | | | | | | | | ✓ | |
| OpenTelemetry (OTel) | | | | | | | | | | | | | | | | ✓ |
| `app.Tracer().Start()` | | | | | | | | | | | | | | | | ✓ |
| `span.SetAttribute()` | | | | | | | | | | | | | | | | ✓ |
| `span.RecordError()` | | | | | | | | | | | | | | | | ✓ |
| OTel HTTP middleware | | | | | | | | | | | | | | | | ✓ |
| Addon integration | | | | | | | ✓ | ✓ | | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |

## Validation Tags Used

| Tag | Example |
|-----|---------|
| `required` | 04, 05, 07, 11, 13, 14 |
| `min` | 04, 05, 07, 11, 13, 14 |
| `max` | 04, 05, 11, 13, 14 |
| `email` | 05, 07, 11 |
| `url` | 05 |
| `gt=0` | 04, 05 |
| `gte=0` | 05 |
| `gte=18` | 05 |
| `lte=120` | 05 |
| `e164` | 05 |
| `omitempty` | 04, 05, 13 |
| `dive` | 05 |

## External Dependencies

| Dependency | Example |
|-----------|---------|
| `github.com/golang-jwt/jwt/v5` | 07, 11, 12 |
| `github.com/slice-soft/ss-keel-gorm` | 08 |
| `github.com/slice-soft/ss-keel-jwt` | 07, 11, 12 |
| `github.com/slice-soft/ss-keel-oauth` | 12 |
| `github.com/slice-soft/ss-keel-mongo` | 13 |
| `github.com/slice-soft/ss-keel-redis` | 14 |
| `github.com/slice-soft/ss-keel-devpanel` | 15 |
| `github.com/slice-soft/ss-keel-otel` | 16 |
| `gorm.io/driver/postgres` | 08 |
| Docker / PostgreSQL | 08 |
| Docker / MongoDB | 13 |
| Docker / Redis | 14 |
| Docker / Jaeger (OTLP collector) | 16 |
| GitHub / Google OAuth app credentials | 12 |

All other examples use only `github.com/slice-soft/ss-keel-core` and the Go standard library.
