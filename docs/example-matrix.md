# Example Matrix

Quick reference: which example covers which Keel concept.

## API Concepts

| Concept | 01 | 02 | 03 | 04 | 05 | 06 | 07 | 08 | 09 | 10 |
|---------|----|----|----|----|----|----|----|----|----|----|
| `core.New()` | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| `core.ControllerFunc` | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| `app.Group()` | | | | ✓ | | ✓ | ✓ | ✓ | | ✓ |
| Module pattern | | | | | | | | | | |
| `ctx.OK()` | ✓ | ✓ | | ✓ | | ✓ | ✓ | ✓ | ✓ | ✓ |
| `ctx.Created()` | | | | ✓ | ✓ | | | ✓ | | |
| `ctx.NoContent()` | | | | ✓ | | | | ✓ | | |
| `ctx.ParseBody()` | | | | ✓ | ✓ | | ✓ | ✓ | | |
| `ctx.Params()` | | | | ✓ | | | | ✓ | | |
| `ctx.Query()` | ✓ | | | | | | | | | |
| Route `.Tag()` | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| Route `.Describe()` | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| Route `.Use()` | | | | | | ✓ | ✓ | | | ✓ |
| Route `.WithBody()` | | | | ✓ | ✓ | | ✓ | ✓ | | |
| Route `.WithResponse()` | ✓ | ✓ | | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| Route `.WithSecured()` | | | | | | | ✓ | | | |
| DELETE route | | | | ✓ | | | | ✓ | | |
| PATCH route | | | | ✓ | | | | ✓ | | |

## Error Handling

| Error type | 01 | 02 | 03 | 04 | 05 | 06 | 07 | 08 | 09 | 10 |
|-----------|----|----|----|----|----|----|----|----|----|----|
| `core.NotFound()` | | | | ✓ | | | | ✓ | | |
| `core.Unauthorized()` | | | | | | ✓ | ✓ | | | |
| `core.Forbidden()` | | | | | | ✓ | ✓ | | | |
| `core.BadRequest()` | | | | | | | | | | |
| `core.Internal()` | | | | | | | ✓ | ✓ | | |
| `core.Conflict()` | | | | | | | | | | |
| `&core.KError{}` (custom) | | | | | | | | | | ✓ |
| 422 via `ctx.ParseBody()` | | | | ✓ | ✓ | | ✓ | ✓ | | |

## Infrastructure

| Feature | 01 | 02 | 03 | 04 | 05 | 06 | 07 | 08 | 09 | 10 |
|---------|----|----|----|----|----|----|----|----|----|----|
| Config loading | | ✓ | | | | | | | | |
| Health checker | | | ✓ | | | | | ✓ | | |
| Scheduler | | | | | | | | | ✓ | |
| JWT auth | | | | | | | ✓ | | | |
| Middleware (custom) | | | | | | ✓ | ✓ | | | |
| GORM + PostgreSQL | | | | | | | | ✓ | | |
| Addon integration | | | | | | | | | | ✓ |

## Validation Tags Used

| Tag | Example |
|-----|---------|
| `required` | 04, 05, 07 |
| `min` | 04, 05, 07 |
| `max` | 04, 05 |
| `email` | 05, 07 |
| `url` | 05 |
| `gt=0` | 04, 05 |
| `gte=0` | 05 |
| `gte=18` | 05 |
| `lte=120` | 05 |
| `e164` | 05 |
| `omitempty` | 04, 05 |
| `dive` | 05 |

## External Dependencies

| Dependency | Example |
|-----------|---------|
| `github.com/golang-jwt/jwt/v5` | 07 |
| `github.com/slice-soft/ss-keel-gorm` | 08 |
| `gorm.io/driver/postgres` | 08 |
| Docker / PostgreSQL | 08 |

All other examples use only `github.com/slice-soft/ss-keel-core` and the Go standard library.
