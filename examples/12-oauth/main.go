package main

import (
	"github.com/slice-soft/ss-keel-core/config"
	"github.com/slice-soft/ss-keel-core/contracts"
	"github.com/slice-soft/ss-keel-core/core"
	"github.com/slice-soft/ss-keel-core/core/httpx"
	"github.com/slice-soft/ss-keel-core/logger"
	"github.com/slice-soft/ss-keel-jwt/jwt"
	"github.com/slice-soft/ss-keel-oauth/oauth"
)

func main() {
	port := config.GetEnvIntOrDefault("PORT", 7331)
	env := config.GetEnvOrDefault("APP_ENV", "development")
	serviceName := config.GetEnvOrDefault("SERVICE_NAME", "oauth-example")

	// JWT config — tokens are issued after a successful OAuth flow.
	jwtSecret := config.GetEnvOrDefault("JWT_SECRET", "change-me-in-production")

	// GitHub OAuth app credentials — create one at https://github.com/settings/developers.
	// Set the callback URL to: http://localhost:7331/auth/github/callback
	githubClientID := config.GetEnvOrDefault("GITHUB_CLIENT_ID", "")
	githubClientSecret := config.GetEnvOrDefault("GITHUB_CLIENT_SECRET", "")
	githubCallback := config.GetEnvOrDefault("GITHUB_CALLBACK_URL", "http://localhost:7331/auth/github/callback")

	// Google OAuth app credentials — create one at https://console.cloud.google.com/apis/credentials.
	// Set the callback URL to: http://localhost:7331/auth/google/callback
	googleClientID := config.GetEnvOrDefault("GOOGLE_CLIENT_ID", "")
	googleClientSecret := config.GetEnvOrDefault("GOOGLE_CLIENT_SECRET", "")
	googleCallback := config.GetEnvOrDefault("GOOGLE_CALLBACK_URL", "http://localhost:7331/auth/google/callback")

	log := logger.NewLogger(env == "production")

	// Initialize the ss-keel-jwt addon.
	// Run:  keel add jwt
	jwtProvider, err := jwt.New(jwt.Config{
		SecretKey:     jwtSecret,
		Issuer:        serviceName,
		TokenTTLHours: 24,
		Logger:        log,
	})
	if err != nil {
		log.Error("failed to initialize JWT provider: %v", err)
	}

	// Initialize the ss-keel-oauth addon.
	// Run:  keel add oauth
	// Providers are optional — omit any ProviderConfig to disable that provider.
	oauthManager := oauth.New(oauth.Config{
		Signer: jwtProvider, // JWT addon signs the token after successful OAuth
		Logger: log,
		GitHub: &oauth.ProviderConfig{
			ClientID:     githubClientID,
			ClientSecret: githubClientSecret,
			RedirectURL:  githubCallback,
		},
		Google: &oauth.ProviderConfig{
			ClientID:     googleClientID,
			ClientSecret: googleClientSecret,
			RedirectURL:  googleCallback,
		},
	})

	app := core.New(core.KConfig{
		ServiceName: serviceName,
		Port:        port,
		Env:         env,
		Docs: core.DocsConfig{
			Title:       "OAuth2 API",
			Version:     "1.0.0",
			Description: "OAuth2 authentication using ss-keel-oauth (GitHub + Google). After the OAuth flow completes, a signed JWT is returned.",
			Tags: []core.DocsTag{
				{Name: "auth", Description: "OAuth2 flows"},
				{Name: "protected", Description: "JWT-protected endpoints"},
			},
		},
	})

	// RegisterController auto-generates routes for all configured providers:
	//   GET /auth/github          → redirect to GitHub authorization page
	//   GET /auth/github/callback → exchange code, return JWT
	//   GET /auth/google          → redirect to Google authorization page
	//   GET /auth/google/callback → exchange code, return JWT
	app.RegisterController(oauth.NewController(oauthManager))

	// Protected routes — require a valid JWT (issued by the OAuth callback above).
	api := app.Group("/api", jwtProvider.Middleware())
	api.RegisterController(contracts.ControllerFunc[httpx.Route](func() []httpx.Route {
		return []httpx.Route{
			// GET /api/me — returns the OAuth user info stored in the JWT.
			// The "sub" claim has the format "<provider>:<user-id>" (e.g. "github:12345").
			// The "data" claim contains email, name, avatar_url, and provider.
			httpx.GET("/me", func(c *httpx.Ctx) error {
				claims, ok := jwt.ClaimsFromCtx(c.Ctx)
				if !ok {
					return core.Unauthorized("missing claims")
				}
				return c.OK(map[string]any{
					"subject": claims["sub"],
					"data":    claims["data"],
				})
			}).
				Tag("protected").
				Describe("Current user", "Returns the OAuth provider user info from the JWT claims.").
				WithSecured("bearerAuth").
				WithResponse(httpx.WithResponse[map[string]any](200)),

			// GET /api/debug — example of how to extract a typed field from claims.
			httpx.GET("/debug/provider", func(c *httpx.Ctx) error {
				claims, _ := jwt.ClaimsFromCtx(c.Ctx)
				data, _ := claims["data"].(map[string]any)
				return c.OK(map[string]any{
					"provider":   data["provider"],
					"name":       data["name"],
					"avatar_url": data["avatar_url"],
				})
			}).
				Tag("protected").
				Describe("Provider debug", "Returns provider-specific fields from the JWT.").
				WithSecured("bearerAuth").
				WithResponse(httpx.WithResponse[map[string]any](200)),
		}
	}))

	// Verify-token helper — useful during development to inspect a raw JWT.
	app.RegisterController(contracts.ControllerFunc[httpx.Route](func() []httpx.Route {
		return []httpx.Route{
			httpx.POST("/auth/verify", func(c *httpx.Ctx) error {
				var body struct {
					Token string `json:"token" validate:"required"`
				}
				if err := c.ParseBody(&body); err != nil {
					return err
				}
				claims, err := jwtProvider.ValidateToken(body.Token)
				if err != nil {
					return core.Unauthorized("invalid token: " + err.Error())
				}
				return c.OK(map[string]any{
					"valid":  true,
					"claims": claims,
				})
			}).
				Tag("auth").
				Describe("Verify token", "Validates a JWT and returns its decoded claims.").
				WithBody(httpx.WithBody[struct {
					Token string `json:"token"`
				}]()).
				WithResponse(httpx.WithResponse[map[string]any](200)),
		}
	}))

	// Print OAuth login URLs to console for easy testing.
	log.Info("GitHub login: http://localhost:%d/auth/github", port)
	log.Info("Google login: http://localhost:%d/auth/google", port)
	log.Info("Docs:         http://localhost:%d/docs", port)

	log.Info("starting %s on port %d (env=%s)", serviceName, port, env)

	if err := app.Listen(); err != nil {
		app.Logger().Error("server error: %v", err)
	}
}
