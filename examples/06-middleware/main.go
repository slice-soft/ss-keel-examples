package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/slice-soft/ss-keel-core/config"
	"github.com/slice-soft/ss-keel-core/core"
	"github.com/slice-soft/ss-keel-core/logger"
)

// CorrelationID injects an X-Correlation-ID header into every response.
// If the request already carries one, it is forwarded as-is.
func CorrelationID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Get("X-Correlation-ID")
		if id == "" {
			// Re-use the request ID that Keel sets automatically.
			if rid, ok := c.Locals("requestid").(string); ok {
				id = rid
			} else {
				id = fmt.Sprintf("gen-%d", time.Now().UnixNano())
			}
		}
		c.Locals("correlation_id", id)
		c.Set("X-Correlation-ID", id)
		return c.Next()
	}
}

// ResponseTimer adds an X-Response-Time header with the request duration.
func ResponseTimer() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		elapsed := time.Since(start)
		c.Set("X-Response-Time", fmt.Sprintf("%dms", elapsed.Milliseconds()))
		return err
	}
}

// IPBlocklist rejects requests from a set of blocked IP addresses.
func IPBlocklist(blocked ...string) fiber.Handler {
	set := make(map[string]struct{}, len(blocked))
	for _, ip := range blocked {
		set[strings.TrimSpace(ip)] = struct{}{}
	}
	return func(c *fiber.Ctx) error {
		if _, blocked := set[c.IP()]; blocked {
			return core.Forbidden("your IP address is not allowed")
		}
		return c.Next()
	}
}

// requireInternalKey is a route-level middleware that checks for a shared secret header.
func requireInternalKey(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Get("X-Internal-Key") != secret {
			return core.Unauthorized("missing or invalid X-Internal-Key header")
		}
		return c.Next()
	}
}

func main() {
	port := config.GetEnvIntOrDefault("PORT", 7331)
	env := config.GetEnvOrDefault("APP_ENV", "development")
	serviceName := config.GetEnvOrDefault("SERVICE_NAME", "middleware-example")
	internalKey := config.GetEnvOrDefault("INTERNAL_KEY", "dev-internal-key")
	blockedIPsRaw := config.GetEnvOrDefault("BLOCKED_IPS", "")

	// Parse blocked IPs from the environment (comma-separated).
	var blockedIPs []string
	if blockedIPsRaw != "" {
		blockedIPs = strings.Split(blockedIPsRaw, ",")
	}

	log := logger.NewLogger(env == "production")

	app := core.New(core.KConfig{
		ServiceName: serviceName,
		Port:        port,
		Env:         env,
		Docs: core.DocsConfig{
			Title:       "Middleware API",
			Version:     "1.0.0",
			Description: "Custom middleware: correlation ID, response timer, and IP blocklist.",
			Tags: []core.DocsTag{
				{Name: "demo", Description: "Demo endpoints"},
			},
		},
	})

	// Apply global middleware via a route group.
	api := app.Group("/api",
		CorrelationID(),
		ResponseTimer(),
		IPBlocklist(blockedIPs...),
	)

	api.RegisterController(core.ControllerFunc(func() []core.Route {
		return []core.Route{
			core.GET("/ping", func(c *core.Ctx) error {
				correlationID, _ := c.Locals("correlation_id").(string)
				return c.OK(map[string]string{
					"message":        "pong",
					"correlation_id": correlationID,
				})
			}).
				Tag("demo").
				Describe("Ping", "Check response headers for X-Correlation-ID and X-Response-Time.").
				WithResponse(core.WithResponse[map[string]string](200)),

			// Route-level middleware: only this route requires the X-Internal-Key header.
			core.GET("/internal", func(c *core.Ctx) error {
				return c.OK(map[string]string{
					"message": "welcome to the internal endpoint",
				})
			}).
				Tag("demo").
				Describe("Internal endpoint", "Protected by a route-level middleware.").
				Use(requireInternalKey(internalKey)).
				WithResponse(core.WithResponse[map[string]string](200)),
		}
	}))

	log.Info("starting %s on port %d (env=%s)", serviceName, port, env)

	if err := app.Listen(); err != nil {
		app.Logger().Error("server error: %v", err)
	}
}
