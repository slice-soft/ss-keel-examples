package main

import (
	"time"

	"addon-example/internal/addon/ratelimit"

	"github.com/slice-soft/ss-keel-core/config"
	"github.com/slice-soft/ss-keel-core/contracts"
	"github.com/slice-soft/ss-keel-core/core"
	"github.com/slice-soft/ss-keel-core/core/httpx"
	"github.com/slice-soft/ss-keel-core/logger"
)

type AppConfig struct {
	Name            string `keel:"app.name"`
	Env             string `keel:"app.env"`
	Port            int    `keel:"server.port"`
	RateLimitMax    int    `keel:"ratelimit.max"`
	RateLimitWindow string `keel:"ratelimit.window"`
}

func main() {
	cfg := config.MustLoadConfig[AppConfig]()
	port := cfg.Port
	env := cfg.Env
	serviceName := cfg.Name
	rateLimitMax := cfg.RateLimitMax
	rateLimitWindowStr := cfg.RateLimitWindow

	rateWindow, err := time.ParseDuration(rateLimitWindowStr)
	if err != nil {
		rateWindow = time.Minute
	}

	log := logger.NewLogger(env == "production")

	app := core.New(core.KConfig{
		ServiceName: serviceName,
		Port:        port,
		Env:         env,
		Docs: core.DocsConfig{
			Title:       "Addon Example API",
			Version:     "1.0.0",
			Description: "Demonstrates consuming a Keel addon. The ratelimit addon was installed via: keel add ratelimit",
			Tags: []core.DocsTag{
				{Name: "demo", Description: "Demo endpoints"},
				{Name: "admin", Description: "Admin endpoints"},
			},
		},
	})

	// Initialize the addon.
	// In a real project this package would come from an external module installed
	// via: keel add ratelimit  (or: keel add github.com/example/keel-addon-ratelimit)
	rl := ratelimit.New(ratelimit.Config{
		Max:    rateLimitMax,
		Window: rateWindow,
	})

	// Apply the rate limiter globally via a route group.
	api := app.Group("/api", rl.Middleware())
	api.RegisterController(contracts.ControllerFunc[httpx.Route](func() []httpx.Route {
		return []httpx.Route{
			httpx.GET("/ping", func(c *httpx.Ctx) error {
				return c.OK(map[string]string{
					"message": "pong",
					"note":    "check X-RateLimit-* headers in the response",
				})
			}).
				Tag("demo").
				Describe("Ping", "Rate-limited endpoint. Check X-RateLimit-Remaining header.").
				WithResponse(httpx.WithResponse[map[string]string](200)),

			httpx.GET("/data", func(c *httpx.Ctx) error {
				return c.OK(map[string]any{
					"items": []string{"alpha", "beta", "gamma"},
					"total": 3,
				})
			}).
				Tag("demo").
				Describe("Get data", "Returns sample data. Subject to rate limiting.").
				WithResponse(httpx.WithResponse[map[string]any](200)),
		}
	}))

	// Admin endpoint to inspect rate limiter stats (not rate-limited itself).
	app.RegisterController(contracts.ControllerFunc[httpx.Route](func() []httpx.Route {
		return []httpx.Route{
			httpx.GET("/admin/ratelimit/stats", func(c *httpx.Ctx) error {
				return c.OK(map[string]any{
					"config": map[string]any{
						"max":    rateLimitMax,
						"window": rateWindow.String(),
					},
					"stats": rl.Stats(),
				})
			}).
				Tag("admin").
				Describe("Rate limiter stats", "Shows current request counts per IP.").
				WithResponse(httpx.WithResponse[map[string]any](200)),
		}
	}))

	log.Info("starting %s on port %d (env=%s)", serviceName, port, env)

	if err := app.Listen(); err != nil {
		app.Logger().Error("server error: %v", err)
	}
}
