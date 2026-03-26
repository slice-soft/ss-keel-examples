package main

import (
	"github.com/slice-soft/ss-keel-core/config"
	"github.com/slice-soft/ss-keel-core/contracts"
	"github.com/slice-soft/ss-keel-core/core"
	"github.com/slice-soft/ss-keel-core/core/httpx"
	"github.com/slice-soft/ss-keel-core/logger"
)

type AppConfig struct {
	Name        string `keel:"app.name"`
	Env         string `keel:"app.env"`
	Port        int    `keel:"server.port"`
	APIVersion  string `keel:"app.api-version"`
	LogLevel    string `keel:"app.log-level"`
	MaxPageSize int    `keel:"app.max-page-size"`
	FeatureFlag bool   `keel:"app.feature-flag"`
}

func main() {
	cfg := config.MustLoadConfig[AppConfig]()
	port := cfg.Port
	env := cfg.Env
	serviceName := cfg.Name
	apiVersion := cfg.APIVersion
	logLevel := cfg.LogLevel
	maxPageSize := cfg.MaxPageSize
	featureFlag := cfg.FeatureFlag

	log := logger.NewLogger(env == "production")
	log.Info("loaded config: env=%s api_version=%s log_level=%s", env, apiVersion, logLevel)

	app := core.New(core.KConfig{
		ServiceName: serviceName,
		Port:        port,
		Env:         env,
		Docs: core.DocsConfig{
			Title:       "Config & Env API",
			Version:     "1.0.0",
			Description: "Demonstrates structured configuration loading using the Keel config package.",
			Tags: []core.DocsTag{
				{Name: "config", Description: "Configuration endpoints"},
			},
		},
	})

	// Expose the loaded config via an endpoint (useful during development).
	app.RegisterController(contracts.ControllerFunc[httpx.Route](func() []httpx.Route {
		return []httpx.Route{
			httpx.GET("/config", func(c *httpx.Ctx) error {
				return c.OK(map[string]any{
					"service_name":  serviceName,
					"env":           env,
					"api_version":   apiVersion,
					"log_level":     logLevel,
					"max_page_size": maxPageSize,
					"feature_flag":  featureFlag,
				})
			}).
				Tag("config").
				Describe("Get active configuration", "Returns the current runtime configuration.").
				WithResponse(httpx.WithResponse[map[string]any](200)),
		}
	}))

	if err := app.Listen(); err != nil {
		app.Logger().Error("server error: %v", err)
	}
}
