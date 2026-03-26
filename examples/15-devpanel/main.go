package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/slice-soft/ss-keel-core/config"
	"github.com/slice-soft/ss-keel-core/contracts"
	"github.com/slice-soft/ss-keel-core/core"
	"github.com/slice-soft/ss-keel-core/core/httpx"
	"github.com/slice-soft/ss-keel-devpanel/devpanel"
)

// AppConfig maps application.properties keys to typed Go fields.
// config.MustLoadConfig reads env vars and application.properties automatically.
type AppConfig struct {
	Name string `keel:"app.name"`
	Env  string `keel:"app.env"`
	Port int    `keel:"server.port"`
}

// Event is a simple in-memory record used to generate visible panel traffic.
type Event struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	CreatedAt int64  `json:"created_at"`
}

// CreateEventRequest is the payload for POST /api/events.
type CreateEventRequest struct {
	Title string `json:"title" validate:"required,min=3,max=120"`
}

var (
	eventsMu sync.RWMutex
	events   = []Event{
		{ID: "evt-1", Title: "App started", CreatedAt: time.Now().UnixMilli()},
		{ID: "evt-2", Title: "Panel mounted", CreatedAt: time.Now().UnixMilli()},
	}
	eventCounter int
)

func main() {
	// Load typed config from application.properties + env vars.
	// No lookup helpers needed — MustLoadConfig handles resolution and defaults.
	cfg := config.MustLoadConfig[AppConfig]()
	panelCfg := config.MustLoadConfig[devpanel.Config]()

	// Initialize the DevPanel addon.
	// Run:  keel add devpanel
	panel := devpanel.New(panelCfg)
	panelLog := panel.Logger()

	app := core.New(core.KConfig{
		ServiceName: cfg.Name,
		Port:        cfg.Port,
		Env:         cfg.Env,
		Docs: core.DocsConfig{
			Title:       "DevPanel Example API",
			Version:     "1.0.0",
			Description: "Demonstrates the ss-keel-devpanel addon: real-time request capture, structured logs, and config inspection.",
			Tags: []core.DocsTag{
				{Name: "events", Description: "In-memory events — hit these to populate the panel's request log"},
				{Name: "panel", Description: "Dev panel access"},
			},
		},
	})

	// Mount the DevPanel middleware and UI before registering any routes.
	// RequestMiddleware captures every non-panel HTTP request into the ring buffer.
	// GlobalGuard blocks panel routes immediately when Enabled=false.
	fiberApp := app.Fiber()
	fiberApp.Use(panel.RequestMiddleware())
	fiberApp.Use(panel.GlobalGuard())
	panel.Mount(fiberApp)

	// App routes — these generate the traffic visible in the panel's request log.
	app.RegisterController(contracts.ControllerFunc[httpx.Route](func() []httpx.Route {
		return []httpx.Route{
			// GET /api/events — list in-memory events.
			httpx.GET("/api/events", func(c *httpx.Ctx) error {
				eventsMu.RLock()
				snapshot := make([]Event, len(events))
				copy(snapshot, events)
				eventsMu.RUnlock()

				panelLog.Info("listed %d events", len(snapshot))
				return c.OK(snapshot)
			}).
				Tag("events").
				Describe("List events", "Returns all in-memory events. Each call is captured by the panel's request log.").
				WithResponse(httpx.WithResponse[[]Event](200)),

			// POST /api/events — create an event and log it to the panel.
			httpx.POST("/api/events", func(c *httpx.Ctx) error {
				var req CreateEventRequest
				if err := c.ParseBody(&req); err != nil {
					return err
				}

				eventsMu.Lock()
				eventCounter++
				ev := Event{
					ID:        fmt.Sprintf("evt-%d", eventCounter+2),
					Title:     req.Title,
					CreatedAt: time.Now().UnixMilli(),
				}
				events = append(events, ev)
				eventsMu.Unlock()

				panelLog.Info("created event id=%s title=%q", ev.ID, ev.Title)
				return c.Created(ev)
			}).
				Tag("events").
				Describe("Create event", "Adds a new event to the in-memory list and logs it via panel.Logger() — visible in the panel's Logs tab.").
				WithBody(httpx.WithBody[CreateEventRequest]()).
				WithResponse(httpx.WithResponse[Event](201)),

			// GET /api/ping — lightweight endpoint for liveness checks.
			httpx.GET("/api/ping", func(c *httpx.Ctx) error {
				return c.OK(map[string]string{"status": "ok", "panel": panelCfg.Path})
			}).
				Tag("panel").
				Describe("Ping", "Returns ok and the panel path. Use this to confirm the service is running before opening the panel.").
				WithResponse(httpx.WithResponse[map[string]string](200)),
		}
	}))

	defer panel.Shutdown()

	app.Logger().Info("starting %s on :%d (env=%s, panel=%s)",
		cfg.Name, cfg.Port, cfg.Env, panelCfg.Path)

	if err := app.Listen(); err != nil {
		app.Logger().Error("server error: %v", err)
	}
}
