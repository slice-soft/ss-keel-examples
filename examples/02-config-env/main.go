package main

import (
	"github.com/slice-soft/ss-keel-core/config"
	"github.com/slice-soft/ss-keel-core/core"
	"github.com/slice-soft/ss-keel-core/logger"
)

func main() {
	port := config.GetEnvIntOrDefault("PORT", 7331)
	env := config.GetEnvOrDefault("APP_ENV", "development")
	serviceName := config.GetEnvOrDefault("SERVICE_NAME", "config-env")
	apiVersion := config.GetEnvOrDefault("API_VERSION", "v1")
	logLevel := config.GetEnvOrDefault("LOG_LEVEL", "info")
	maxPageSize := config.GetEnvIntOrDefault("MAX_PAGE_SIZE", 50)
	featureFlag := config.GetEnvBoolOrDefault("FEATURE_FLAG", false)

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
	app.RegisterController(core.ControllerFunc(func() []core.Route {
		return []core.Route{
			core.GET("/config", func(c *core.Ctx) error {
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
				WithResponse(core.WithResponse[map[string]any](200)),
		}
	}))

	if err := app.Listen(); err != nil {
		app.Logger().Error("server error: %v", err)
	}
}
