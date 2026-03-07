package main

import (
	"github.com/slice-soft/ss-keel-core/config"
	"github.com/slice-soft/ss-keel-core/core"
	"github.com/slice-soft/ss-keel-core/logger"
)

func main() {
	port := config.GetEnvIntOrDefault("PORT", 7331)
	env := config.GetEnvOrDefault("APP_ENV", "development")
	serviceName := config.GetEnvOrDefault("SERVICE_NAME", "hello-world")

	log := logger.NewLogger(env == "production")

	app := core.New(core.KConfig{
		ServiceName: serviceName,
		Port:        port,
		Env:         env,
		Docs: core.DocsConfig{
			Title:       "Hello World API",
			Version:     "1.0.0",
			Description: "Minimal Keel example with a single route.",
			Tags: []core.DocsTag{
				{Name: "hello", Description: "Hello endpoints"},
			},
		},
	})

	app.RegisterController(core.ControllerFunc(func() []core.Route {
		return []core.Route{
			core.GET("/hello", func(c *core.Ctx) error {
				name := c.Query("name")
				if name == "" {
					name = "world"
				}
				return c.OK(map[string]string{
					"message": "Hello, " + name + "!",
				})
			}).
				Tag("hello").
				Describe("Say hello", "Returns a greeting message.").
				WithQueryParam("name", "string", false, "Name to greet").
				WithResponse(core.WithResponse[map[string]string](200)),
		}
	}))

	log.Info("starting %s on port %d (env=%s)", serviceName, port, env)

	if err := app.Listen(); err != nil {
		app.Logger().Error("server error: %v", err)
	}
}
