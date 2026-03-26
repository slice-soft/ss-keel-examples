package main

import (
	"strings"

	"github.com/slice-soft/ss-keel-core/config"
	"github.com/slice-soft/ss-keel-core/contracts"
	"github.com/slice-soft/ss-keel-core/core"
	"github.com/slice-soft/ss-keel-core/core/httpx"
	"github.com/slice-soft/ss-keel-core/logger"
	"github.com/slice-soft/ss-keel-jwt/jwt"
	"github.com/slice-soft/ss-keel-oauth/oauth"
)

type AppConfig struct {
	Name                  string `keel:"app.name"`
	Env                   string `keel:"app.env"`
	Port                  int    `keel:"server.port"`
	JWTSecret             string `keel:"jwt.secret"`
	OAuthRoutePrefix      string `keel:"oauth.route-prefix"`
	OAuthRedirectBase     string `keel:"oauth.redirect-base-url"`
	OAuthRedirectSuccess  string `keel:"oauth.redirect-on-success"`
	OAuthRedirectParam    string `keel:"oauth.redirect-token-param"`
	OAuthEnabledProviders string `keel:"oauth.enabled-providers"`
	OAuthGoogleClientID   string `keel:"oauth.google.client-id"`
	OAuthGoogleSecret     string `keel:"oauth.google.client-secret"`
	OAuthGitHubClientID   string `keel:"oauth.github.client-id"`
	OAuthGitHubSecret     string `keel:"oauth.github.client-secret"`
	OAuthGitLabClientID   string `keel:"oauth.gitlab.client-id"`
	OAuthGitLabSecret     string `keel:"oauth.gitlab.client-secret"`
}

func main() {
	cfg := config.MustLoadConfig[AppConfig]()
	routePrefix := normalizeOAuthRoutePrefix(cfg.OAuthRoutePrefix)
	redirectBase := normalizeOAuthRedirectBase(cfg.OAuthRedirectBase)
	redirectOnSuccess := normalizeOAuthSuccessRedirect(cfg.OAuthRedirectSuccess)
	redirectTokenParam := normalizeOAuthRedirectTokenParam(cfg.OAuthRedirectParam)
	enabledProviders := parseOAuthEnabledProviders(cfg.OAuthEnabledProviders)

	log := logger.NewLogger(cfg.Env == "production")

	// Initialize the ss-keel-jwt addon.
	// Run:  keel add jwt
	jwtProvider, err := jwt.New(jwt.Config{
		SecretKey:     cfg.JWTSecret,
		Issuer:        cfg.Name,
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
		Google: oauthProviderConfig(redirectBase, routePrefix, enabledProviders, oauth.ProviderGoogle,
			cfg.OAuthGoogleClientID,
			cfg.OAuthGoogleSecret,
		),
		GitHub: oauthProviderConfig(redirectBase, routePrefix, enabledProviders, oauth.ProviderGitHub,
			cfg.OAuthGitHubClientID,
			cfg.OAuthGitHubSecret,
		),
		GitLab: oauthProviderConfig(redirectBase, routePrefix, enabledProviders, oauth.ProviderGitLab,
			cfg.OAuthGitLabClientID,
			cfg.OAuthGitLabSecret,
		),
		RedirectOnSuccess:  redirectOnSuccess,
		RedirectTokenParam: redirectTokenParam,
	})

	app := core.New(core.KConfig{
		ServiceName: cfg.Name,
		Port:        cfg.Port,
		Env:         cfg.Env,
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

	// RegisterController auto-generates routes for all configured providers.
	// With the default routePrefix it exposes:
	//   GET /auth/github          → redirect to GitHub authorization page
	//   GET /auth/github/callback → exchange code, return JWT or redirect to the frontend
	//   GET /auth/google          → redirect to Google authorization page
	//   GET /auth/google/callback → exchange code, return JWT or redirect to the frontend
	app.RegisterController(oauth.NewController(oauthManager, routePrefix))

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
	log.Info("GitHub login: http://localhost:%d%s/github", cfg.Port, routePrefix)
	log.Info("Google login: http://localhost:%d%s/google", cfg.Port, routePrefix)
	log.Info("GitLab login: http://localhost:%d%s/gitlab", cfg.Port, routePrefix)
	log.Info("Docs:         http://localhost:%d/docs", cfg.Port)

	log.Info("starting %s on port %d (env=%s)", cfg.Name, cfg.Port, cfg.Env)

	if err := app.Listen(); err != nil {
		app.Logger().Error("server error: %v", err)
	}
}

func oauthProviderConfig(redirectBase, routePrefix string, enabledProviders map[oauth.ProviderName]struct{}, provider oauth.ProviderName, clientID, clientSecret string) *oauth.ProviderConfig {
	clientID = strings.TrimSpace(clientID)
	clientSecret = strings.TrimSpace(clientSecret)
	if clientID == "" || clientSecret == "" {
		return nil
	}
	if len(enabledProviders) > 0 {
		if _, ok := enabledProviders[provider]; !ok {
			return nil
		}
	}
	return &oauth.ProviderConfig{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectBase + routePrefix + "/" + string(provider) + "/callback",
	}
}

func parseOAuthEnabledProviders(raw string) map[oauth.ProviderName]struct{} {
	enabledProviders := make(map[oauth.ProviderName]struct{})
	for _, part := range strings.Split(raw, ",") {
		switch oauth.ProviderName(strings.ToLower(strings.TrimSpace(part))) {
		case oauth.ProviderGoogle, oauth.ProviderGitHub, oauth.ProviderGitLab:
			enabledProviders[oauth.ProviderName(strings.ToLower(strings.TrimSpace(part)))] = struct{}{}
		}
	}
	return enabledProviders
}

func normalizeOAuthRoutePrefix(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || trimmed == "/" {
		return "/auth"
	}
	if !strings.HasPrefix(trimmed, "/") {
		trimmed = "/" + trimmed
	}
	trimmed = strings.TrimRight(trimmed, "/")
	if trimmed == "" {
		return "/auth"
	}
	return trimmed
}

func normalizeOAuthRedirectBase(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		trimmed = "http://127.0.0.1:7331"
	}
	return strings.TrimRight(trimmed, "/")
}

func normalizeOAuthSuccessRedirect(raw string) string {
	return strings.TrimSpace(raw)
}

func normalizeOAuthRedirectTokenParam(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "token"
	}
	return trimmed
}
