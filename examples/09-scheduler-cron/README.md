# 09 — Scheduler / Cron

Background job scheduling using Keel's `core.Scheduler` interface.

## What This Demonstrates

- Implementing the `core.Scheduler` interface (`Add`, `Start`, `Stop`)
- Defining jobs with `core.Job{Name, Schedule, Handler}`
- Registering the scheduler with `app.RegisterScheduler()` — Keel calls `Start()` on `Listen()` and `Stop()` on graceful shutdown
- Running interval-based background jobs concurrently
- Exposing job history via an HTTP endpoint

## Requirements

- Go 1.21+

## How to Run

```bash
cp .env.example .env
go mod download
go run main.go
```

After a few seconds, you will see heartbeat log entries. Check the history endpoint to see execution records.

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/jobs/history` | Last 100 job executions |
| GET | `/health` | Health check |
| GET | `/docs` | OpenAPI UI |

## Example

```bash
# Wait a few seconds, then check history
curl http://localhost:7331/jobs/history
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `HEARTBEAT_INTERVAL` | `10s` | How often the heartbeat job runs |
| `CLEANUP_INTERVAL` | `1m` | How often the cleanup job runs |

Schedules are standard Go duration strings: `10s`, `30s`, `1m`, `5m`, etc.

## Concepts Covered

- `core.Scheduler` interface — `Add(core.Job)`, `Start()`, `Stop(ctx)`
- `core.Job` — `Name`, `Schedule`, `Handler`
- `app.RegisterScheduler()` — lifecycle integration
- Graceful shutdown waits for running jobs to finish
- `sync.WaitGroup` for concurrent job coordination
