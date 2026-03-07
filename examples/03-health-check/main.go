package main

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/slice-soft/ss-keel-core/config"
	"github.com/slice-soft/ss-keel-core/core"
	"github.com/slice-soft/ss-keel-core/logger"
)

// CacheChecker verifies that the in-memory cache is responding.
// Implements core.HealthChecker.
type CacheChecker struct {
	ready atomic.Bool
}

func NewCacheChecker() *CacheChecker {
	c := &CacheChecker{}
	c.ready.Store(true) // starts healthy
	return c
}

func (c *CacheChecker) Name() string { return "cache" }

func (c *CacheChecker) Check(_ context.Context) error {
	if !c.ready.Load() {
		return errors.New("cache is not ready")
	}
	return nil
}

// DatabaseChecker simulates a database ping.
// Implements core.HealthChecker.
type DatabaseChecker struct {
	connected atomic.Bool
}

func NewDatabaseChecker() *DatabaseChecker {
	d := &DatabaseChecker{}
	d.connected.Store(true)
	return d
}

func (d *DatabaseChecker) Name() string { return "database" }

func (d *DatabaseChecker) Check(_ context.Context) error {
	if !d.connected.Load() {
		return errors.New("database connection lost")
	}
	return nil
}

func main() {
	port := config.GetEnvIntOrDefault("PORT", 7331)
	env := config.GetEnvOrDefault("APP_ENV", "development")
	serviceName := config.GetEnvOrDefault("SERVICE_NAME", "health-check")

	log := logger.NewLogger(env == "production")

	app := core.New(core.KConfig{
		ServiceName: serviceName,
		Port:        port,
		Env:         env,
		Docs: core.DocsConfig{
			Title:       "Health Check API",
			Version:     "1.0.0",
			Description: "Demonstrates custom health checkers with the built-in /health endpoint.",
			Tags: []core.DocsTag{
				{Name: "system", Description: "System endpoints"},
				{Name: "demo", Description: "Demo control endpoints"},
			},
		},
	})

	cache := NewCacheChecker()
	db := NewDatabaseChecker()

	// Register custom health checkers — Keel aggregates them into /health.
	app.RegisterHealthChecker(cache)
	app.RegisterHealthChecker(db)

	// Control endpoints to toggle dependency health (for demo purposes).
	app.RegisterController(core.ControllerFunc(func() []core.Route {
		return []core.Route{
			core.POST("/demo/cache/down", func(c *core.Ctx) error {
				cache.ready.Store(false)
				return c.OK(map[string]string{"cache": "marked as DOWN"})
			}).Tag("demo").Describe("Mark cache as DOWN"),

			core.POST("/demo/cache/up", func(c *core.Ctx) error {
				cache.ready.Store(true)
				return c.OK(map[string]string{"cache": "marked as UP"})
			}).Tag("demo").Describe("Mark cache as UP"),

			core.POST("/demo/database/down", func(c *core.Ctx) error {
				db.connected.Store(false)
				return c.OK(map[string]string{"database": "marked as DOWN"})
			}).Tag("demo").Describe("Mark database as DOWN"),

			core.POST("/demo/database/up", func(c *core.Ctx) error {
				db.connected.Store(true)
				return c.OK(map[string]string{"database": "marked as UP"})
			}).Tag("demo").Describe("Mark database as UP"),
		}
	}))

	log.Info("starting %s on port %d (env=%s)", serviceName, port, env)

	if err := app.Listen(); err != nil {
		app.Logger().Error("server error: %v", err)
	}
}
