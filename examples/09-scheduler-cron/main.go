package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/slice-soft/ss-keel-core/config"
	"github.com/slice-soft/ss-keel-core/contracts"
	"github.com/slice-soft/ss-keel-core/core"
	"github.com/slice-soft/ss-keel-core/core/httpx"
	"github.com/slice-soft/ss-keel-core/logger"
)

// JobRecord tracks a single execution of a scheduled job.
type JobRecord struct {
	JobName   string    `json:"job_name"`
	RunAt     time.Time `json:"run_at"`
	Message   string    `json:"message"`
	RunNumber int       `json:"run_number"`
}

// History stores recent job executions in memory.
type History struct {
	mu      sync.RWMutex
	records []JobRecord
	max     int
}

func NewHistory(max int) *History {
	return &History{max: max}
}

func (h *History) Add(r JobRecord) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.records = append(h.records, r)
	if len(h.records) > h.max {
		h.records = h.records[len(h.records)-h.max:]
	}
}

func (h *History) All() []JobRecord {
	h.mu.RLock()
	defer h.mu.RUnlock()
	out := make([]JobRecord, len(h.records))
	copy(out, h.records)
	return out
}

// IntervalScheduler is a simple interval-based scheduler that implements contracts.Scheduler.
// Schedules are specified as Go duration strings (e.g. "10s", "1m").
type IntervalScheduler struct {
	mu     sync.Mutex
	jobs   []contracts.Job
	cancel context.CancelFunc
	wg     sync.WaitGroup
	log    *logger.Logger
}

func NewIntervalScheduler(log *logger.Logger) *IntervalScheduler {
	return &IntervalScheduler{log: log}
}

// Add registers a job. Schedule must be a valid Go duration string (e.g. "10s").
func (s *IntervalScheduler) Add(job contracts.Job) error {
	if _, err := time.ParseDuration(job.Schedule); err != nil {
		return fmt.Errorf("invalid schedule %q: must be a Go duration string (e.g. 10s, 1m)", job.Schedule)
	}
	s.mu.Lock()
	s.jobs = append(s.jobs, job)
	s.mu.Unlock()
	s.log.Info("scheduler: registered job=%q schedule=%s", job.Name, job.Schedule)
	return nil
}

// Start begins all registered jobs. Called automatically by app.Listen().
func (s *IntervalScheduler) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	s.mu.Lock()
	jobs := append([]contracts.Job(nil), s.jobs...)
	s.mu.Unlock()

	for _, job := range jobs {
		interval, _ := time.ParseDuration(job.Schedule)
		s.wg.Add(1)
		go s.run(ctx, job, interval)
	}
	s.log.Info("scheduler: started with %d job(s)", len(jobs))
}

func (s *IntervalScheduler) run(ctx context.Context, job contracts.Job, interval time.Duration) {
	defer s.wg.Done()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			runCtx, cancel := context.WithTimeout(ctx, interval)
			if err := job.Handler(runCtx); err != nil {
				s.log.Warn("scheduler: job=%q failed: %v", job.Name, err)
			}
			cancel()
		}
	}
}

// Stop signals all jobs to stop and waits for them to finish.
func (s *IntervalScheduler) Stop(ctx context.Context) {
	if s.cancel != nil {
		s.cancel()
	}
	done := make(chan struct{})
	go func() { s.wg.Wait(); close(done) }()
	select {
	case <-done:
		s.log.Info("scheduler: stopped gracefully")
	case <-ctx.Done():
		s.log.Warn("scheduler: stop timed out")
	}
}

func main() {
	port := config.GetEnvIntOrDefault("PORT", 7331)
	env := config.GetEnvOrDefault("APP_ENV", "development")
	serviceName := config.GetEnvOrDefault("SERVICE_NAME", "scheduler-cron")
	heartbeatInterval := config.GetEnvOrDefault("HEARTBEAT_INTERVAL", "10s")
	cleanupInterval := config.GetEnvOrDefault("CLEANUP_INTERVAL", "1m")

	log := logger.NewLogger(env == "production")

	app := core.New(core.KConfig{
		ServiceName: serviceName,
		Port:        port,
		Env:         env,
		Docs: core.DocsConfig{
			Title:       "Scheduler / Cron API",
			Version:     "1.0.0",
			Description: "Background job scheduling with interval-based execution.",
			Tags: []core.DocsTag{
				{Name: "jobs", Description: "Job history"},
			},
		},
	})

	history := NewHistory(100)
	sched := NewIntervalScheduler(app.Logger())

	heartbeatCount := 0

	// Job 1: heartbeat on a configurable interval.
	if err := sched.Add(contracts.Job{
		Name:     "heartbeat",
		Schedule: heartbeatInterval,
		Handler: func(_ context.Context) error {
			heartbeatCount++
			msg := fmt.Sprintf("heartbeat #%d", heartbeatCount)
			app.Logger().Info("job:heartbeat — %s", msg)
			history.Add(JobRecord{
				JobName:   "heartbeat",
				RunAt:     time.Now(),
				Message:   msg,
				RunNumber: heartbeatCount,
			})
			return nil
		},
	}); err != nil {
		log.Error("invalid heartbeat schedule: %v", err)
	}

	cleanupCount := 0

	// Job 2: simulated cleanup on a configurable interval.
	if err := sched.Add(contracts.Job{
		Name:     "cleanup",
		Schedule: cleanupInterval,
		Handler: func(_ context.Context) error {
			cleanupCount++
			msg := fmt.Sprintf("cleanup #%d — 0 stale records removed", cleanupCount)
			app.Logger().Info("job:cleanup — %s", msg)
			history.Add(JobRecord{
				JobName:   "cleanup",
				RunAt:     time.Now(),
				Message:   msg,
				RunNumber: cleanupCount,
			})
			return nil
		},
	}); err != nil {
		log.Error("invalid cleanup schedule: %v", err)
	}

	// Register scheduler — Keel calls Start() on Listen() and Stop() on shutdown.
	app.RegisterScheduler(sched)

	// Expose job history via HTTP.
	app.RegisterController(contracts.ControllerFunc[httpx.Route](func() []httpx.Route {
		return []httpx.Route{
			httpx.GET("/jobs/history", func(c *httpx.Ctx) error {
				all := history.All()
				return c.OK(map[string]any{
					"history": all,
					"count":   len(all),
				})
			}).
				Tag("jobs").
				Describe("Job execution history", "Returns the last 100 job executions.").
				WithResponse(httpx.WithResponse[map[string]any](200)),
		}
	}))

	log.Info("starting %s on port %d (env=%s)", serviceName, port, env)

	if err := app.Listen(); err != nil {
		app.Logger().Error("server error: %v", err)
	}
}
