package main

import (
	"github.com/slice-soft/ss-keel-core/config"
	"github.com/slice-soft/ss-keel-core/contracts"
	"github.com/slice-soft/ss-keel-core/core"
	"github.com/slice-soft/ss-keel-core/core/httpx"
	"github.com/slice-soft/ss-keel-core/logger"
)

type AppConfig struct {
	Name string `keel:"app.name"`
	Env  string `keel:"app.env"`
	Port int    `keel:"server.port"`
}

func main() {
	cfg := config.MustLoadConfig[AppConfig]()
	port := cfg.Port
	env := cfg.Env
	serviceName := cfg.Name

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

	app.RegisterController(contracts.ControllerFunc[httpx.Route](func() []httpx.Route {
		return []httpx.Route{
			httpx.GET("/hello", func(c *httpx.Ctx) error {
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
				WithResponse(httpx.WithResponse[map[string]string](200)),
		}
	}))

	log.Info("starting %s on port %d (env=%s)", serviceName, port, env)

	if err := app.Listen(); err != nil {
		app.Logger().Error("server error: %v", err)
	}
}
