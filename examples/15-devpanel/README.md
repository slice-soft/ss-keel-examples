# 15 — DevPanel

> **Note:** This example uses a compact flat layout (`main.go` at the root) to keep all the relevant code in one place and let you focus on the API surface — no CLI scaffolding required. Projects created with `keel new` + `keel add devpanel` place the entry point at `cmd/main.go` and generate a dedicated `cmd/setup_devpanel.go` provider file.

Real-time observability UI for your Keel application, powered by the [ss-keel-devpanel](https://github.com/slice-soft/ss-keel-devpanel) addon.

## What this example demonstrates

- Installing the DevPanel addon with `keel add devpanel`
- Loading typed config with `config.MustLoadConfig` and `application.properties` — no lookup helpers needed
- Mounting `panel.RequestMiddleware()` to capture every HTTP request into a ring buffer
- Using `panel.GlobalGuard()` to disable the panel in production via `KEEL_PANEL_ENABLED=false`
- Logging structured messages with `panel.Logger()` — visible in the Logs tab
- Protecting the panel with an optional bearer token via `KEEL_PANEL_SECRET`

## How to run

```bash
cp .env.example .env
go mod download
go run .
```

The server starts on port **7331**. Open the panel at:

```
http://localhost:7331/keel/panel
```

Hit the API routes a few times first — `GET /api/events` and `POST /api/events` — to populate the **Requests** and **Logs** tabs.

## Panel tabs

| Tab | URL | What it shows |
|-----|-----|---------------|
| Requests | `/keel/panel/requests` | Captured HTTP requests with method, path, status, and latency |
| Logs | `/keel/panel/logs` | Structured log entries from `panel.Logger()` |
| Routes | `/keel/panel/routes` | All registered Fiber routes |
| Addons | `/keel/panel/addons` | Debuggable addons registered with `panel.RegisterAddon()` |
| Config | `/keel/panel/config` | Resolved `application.properties` values |

## Protecting the panel in production

Set `KEEL_PANEL_ENABLED=false` to return 404 on all panel routes — the `GlobalGuard` middleware enforces this before requests reach the route group.

To require a bearer token instead of disabling the panel entirely:

```env
KEEL_PANEL_SECRET=my-secret-token
```

Then access the panel with `Authorization: Bearer my-secret-token`.
