# 16 — OpenTelemetry Addon

Demonstrates `ss-keel-otel`: automatic HTTP spans, manual child spans, span attributes, and error recording via OTLP export to Jaeger.

## What this example teaches

- How to initialize `ss-keel-otel` using the `setupOtel` provider pattern
- Automatic HTTP spans for every Fiber request (zero code in handlers)
- Manual child spans with `app.Tracer().Start(ctx, "OperationName")`
- Span attributes: `span.SetAttribute("key", value)`
- Error recording: `span.RecordError(err)`
- Safe default: `OTEL_ENABLED=false` — the service runs normally without any collector

## How to run

### Without telemetry (default)

```bash
cp .env.example .env
go run .
```

Open the API docs at <http://localhost:7331/docs>.

### With telemetry (Jaeger)

```bash
docker compose up -d          # Start Jaeger (UI at http://localhost:16686)
cp .env.example .env
```

Edit `.env` and set:

```
OTEL_ENABLED=true
```

Then run:

```bash
go run .
```

Hit a few endpoints via `requests.http`, then open Jaeger at <http://localhost:16686> and select `otel-example` from the **Service** dropdown.

## Key files

| File | Purpose |
|------|---------|
| `main.go` | App setup, `setupOtel` provider function, orders module |
| `application.properties` | Typed config keys for the OTel addon |
| `.env.example` | All environment variables with safe defaults |
| `docker-compose.yml` | Jaeger all-in-one as local OTLP collector |
| `requests.http` | HTTP requests to trigger spans |
