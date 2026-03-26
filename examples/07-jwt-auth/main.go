package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/slice-soft/ss-keel-core/config"
	"github.com/slice-soft/ss-keel-core/contracts"
	"github.com/slice-soft/ss-keel-core/core"
	"github.com/slice-soft/ss-keel-core/core/httpx"
	"github.com/slice-soft/ss-keel-core/logger"
	"github.com/slice-soft/ss-keel-jwt/jwt"
)

type AppConfig struct {
	Name      string `keel:"app.name"`
	Env       string `keel:"app.env"`
	Port      int    `keel:"server.port"`
	JWTSecret string `keel:"jwt.secret"`
}

// User is a static in-memory user for demo purposes.
type User struct {
	ID    string
	Email string
	Role  string
}

// LoginRequest is the payload for the login endpoint.
type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse carries the issued JWT.
type LoginResponse struct {
	Token     string `json:"token"`
	TokenType string `json:"token_type"`
}

// staticUsers simulates a user store.
var staticUsers = map[string]struct {
	user     User
	password string
}{
	"alice@example.com": {User{"usr_1", "alice@example.com", "admin"}, "password123"},
	"bob@example.com":   {User{"usr_2", "bob@example.com", "user"}, "pass456"},
}

// RequireRole is a middleware that restricts access to a specific role.
func RequireRole(role string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims, ok := jwt.ClaimsFromCtx(c)
		if !ok {
			return core.Unauthorized("not authenticated")
		}
		data, _ := claims["data"].(map[string]any)
		if data["role"] != role {
			return core.Forbidden("requires role: " + role)
		}
		return c.Next()
	}
}

func main() {
	cfg := config.MustLoadConfig[AppConfig]()
	port := cfg.Port
	env := cfg.Env
	serviceName := cfg.Name
	jwtSecret := cfg.JWTSecret

	log := logger.NewLogger(env == "production")

	// Initialize the ss-keel-jwt addon.
	jwtProvider, err := jwt.New(jwt.Config{
		SecretKey:     jwtSecret,
		Issuer:        serviceName,
		TokenTTLHours: 24,
		Logger:        log,
	})
	if err != nil {
		log.Error("failed to initialize JWT provider: %v", err)
	}

	app := core.New(core.KConfig{
		ServiceName: serviceName,
		Port:        port,
		Env:         env,
		Docs: core.DocsConfig{
			Title:       "JWT Auth API",
			Version:     "1.0.0",
			Description: "JWT authentication using the ss-keel-jwt addon. Login to get a token, then use it to access protected routes.",
			Tags: []core.DocsTag{
				{Name: "auth", Description: "Authentication"},
				{Name: "protected", Description: "Protected resources"},
			},
		},
	})

	// Public routes
	app.RegisterController(contracts.ControllerFunc[httpx.Route](func() []httpx.Route {
		return []httpx.Route{
			httpx.POST("/auth/login", func(c *httpx.Ctx) error {
				var req LoginRequest
				if err := c.ParseBody(&req); err != nil {
					return err
				}

				entry, ok := staticUsers[req.Email]
				if !ok || entry.password != req.Password {
					return core.Unauthorized("invalid email or password")
				}

				// GenerateToken embeds the payload in the "data" claim.
				token, err := jwtProvider.GenerateToken(map[string]any{
					"user_id": entry.user.ID,
					"email":   entry.user.Email,
					"role":    entry.user.Role,
				})
				if err != nil {
					return core.Internal("could not issue token", err)
				}

				return c.OK(LoginResponse{Token: token, TokenType: "Bearer"})
			}).
				Tag("auth").
				Describe("Login", "Exchange credentials for a JWT.").
				WithBody(httpx.WithBody[LoginRequest]()).
				WithResponse(httpx.WithResponse[LoginResponse](200)),

			httpx.POST("/auth/refresh", func(c *httpx.Ctx) error {
				var body struct {
					Token string `json:"token" validate:"required"`
				}
				if err := c.ParseBody(&body); err != nil {
					return err
				}
				newToken, err := jwtProvider.RefreshToken(body.Token)
				if err != nil {
					return core.Unauthorized("invalid or expired token")
				}
				return c.OK(LoginResponse{Token: newToken, TokenType: "Bearer"})
			}).
				Tag("auth").
				Describe("Refresh token", "Exchange a valid token for a fresh one.").
				WithBody(httpx.WithBody[struct {
					Token string `json:"token"`
				}]()).
				WithResponse(httpx.WithResponse[LoginResponse](200)),
		}
	}))

	// Protected routes — all require a valid JWT via jwtProvider.Middleware().
	protected := app.Group("/api", jwtProvider.Middleware())
	protected.RegisterController(contracts.ControllerFunc[httpx.Route](func() []httpx.Route {
		return []httpx.Route{
			httpx.GET("/me", func(c *httpx.Ctx) error {
				claims, _ := jwt.ClaimsFromCtx(c.Ctx)
				data, _ := claims["data"].(map[string]any)
				return c.OK(data)
			}).
				Tag("protected").
				Describe("Current user", "Returns the authenticated user's info from the JWT claims.").
				WithSecured("bearerAuth").
				WithResponse(httpx.WithResponse[map[string]any](200)),

			httpx.GET("/admin/dashboard", func(c *httpx.Ctx) error {
				return c.OK(map[string]string{
					"message": "welcome to the admin dashboard",
				})
			}).
				Tag("protected").
				Describe("Admin dashboard", "Only accessible by users with role=admin.").
				Use(RequireRole("admin")).
				WithSecured("bearerAuth").
				WithResponse(httpx.WithResponse[map[string]string](200)),
		}
	}))

	log.Info("starting %s on port %d (env=%s)", serviceName, port, env)

	if err := app.Listen(); err != nil {
		app.Logger().Error("server error: %v", err)
	}
}
