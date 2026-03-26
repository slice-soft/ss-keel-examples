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

// User is a static in-memory user for demo purposes.
type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// LoginRequest is the payload for the login endpoint.
type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// TokenResponse carries the issued or refreshed JWT.
type TokenResponse struct {
	Token     string `json:"token"`
	TokenType string `json:"token_type"`
}

// staticUsers simulates a user store with hashed passwords.
var staticUsers = map[string]struct {
	user     User
	password string
}{
	"alice@example.com": {user: User{ID: "usr_1", Name: "Alice", Email: "alice@example.com", Role: "admin"}, password: "password123"},
	"bob@example.com":   {user: User{ID: "usr_2", Name: "Bob", Email: "bob@example.com", Role: "member"}, password: "pass456"},
}

// RequireRole is a middleware that checks the role stored inside the JWT "data" claim.
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

type AppConfig struct {
	Name        string `keel:"app.name"`
	Env         string `keel:"app.env"`
	Port        int    `keel:"server.port"`
	JWTSecret   string `keel:"jwt.secret"`
	JWTIssuer   string `keel:"jwt.issuer"`
	JWTTokenTTL uint   `keel:"jwt.token-ttl-hours"`
}

func main() {
	cfg := config.MustLoadConfig[AppConfig]()

	log := logger.NewLogger(cfg.Env == "production")

	// Initialize the ss-keel-jwt addon.
	// Run:  keel add jwt
	jwtProvider, err := jwt.New(jwt.Config{
		SecretKey:     cfg.JWTSecret,
		Issuer:        cfg.JWTIssuer,
		TokenTTLHours: cfg.JWTTokenTTL,
		Logger:        log,
	})
	if err != nil {
		log.Error("failed to initialize JWT provider: %v", err)
	}

	app := core.New(core.KConfig{
		ServiceName: cfg.Name,
		Port:        cfg.Port,
		Env:         cfg.Env,
		Docs: core.DocsConfig{
			Title:       "JWT Addon API",
			Version:     "1.0.0",
			Description: "Demonstrates the ss-keel-jwt addon: token generation, route protection, and refresh.",
			Tags: []core.DocsTag{
				{Name: "auth", Description: "Authentication"},
				{Name: "protected", Description: "JWT-protected endpoints"},
			},
		},
	})

	// Public routes — no middleware required.
	app.RegisterController(contracts.ControllerFunc[httpx.Route](func() []httpx.Route {
		return []httpx.Route{
			// POST /auth/login — exchange credentials for a JWT.
			httpx.POST("/auth/login", func(c *httpx.Ctx) error {
				var req LoginRequest
				if err := c.ParseBody(&req); err != nil {
					return err
				}
				entry, ok := staticUsers[req.Email]
				if !ok || entry.password != req.Password {
					return core.Unauthorized("invalid email or password")
				}

				// GenerateToken stores the payload inside the "data" JWT claim.
				token, err := jwtProvider.GenerateToken(map[string]any{
					"user_id": entry.user.ID,
					"name":    entry.user.Name,
					"email":   entry.user.Email,
					"role":    entry.user.Role,
				})
				if err != nil {
					return core.Internal("could not issue token", err)
				}
				return c.OK(TokenResponse{Token: token, TokenType: "Bearer"})
			}).
				Tag("auth").
				Describe("Login", "Returns a signed JWT on valid credentials.").
				WithBody(httpx.WithBody[LoginRequest]()).
				WithResponse(httpx.WithResponse[TokenResponse](200)),

			// POST /auth/refresh — exchange a valid token for a fresh one with reset expiry.
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
				return c.OK(TokenResponse{Token: newToken, TokenType: "Bearer"})
			}).
				Tag("auth").
				Describe("Refresh token", "Returns a new token with reset expiry. Original payload is preserved.").
				WithBody(httpx.WithBody[struct {
					Token string `json:"token"`
				}]()).
				WithResponse(httpx.WithResponse[TokenResponse](200)),
		}
	}))

	// Protected routes — jwtProvider.Middleware() validates the Bearer token.
	api := app.Group("/api", jwtProvider.Middleware())
	api.RegisterController(contracts.ControllerFunc[httpx.Route](func() []httpx.Route {
		return []httpx.Route{
			// GET /api/me — returns the decoded JWT payload.
			httpx.GET("/me", func(c *httpx.Ctx) error {
				claims, _ := jwt.ClaimsFromCtx(c.Ctx)
				data, _ := claims["data"].(map[string]any)
				return c.OK(data)
			}).
				Tag("protected").
				Describe("Current user", "Returns the payload stored in the JWT claims.").
				WithSecured("bearerAuth").
				WithResponse(httpx.WithResponse[map[string]any](200)),

			// GET /api/admin — only accessible by the "admin" role.
			httpx.GET("/admin", func(c *httpx.Ctx) error {
				return c.OK(map[string]string{"message": "welcome, admin"})
			}).
				Tag("protected").
				Describe("Admin area", "Requires role=admin in the JWT claims.").
				Use(RequireRole("admin")).
				WithSecured("bearerAuth").
				WithResponse(httpx.WithResponse[map[string]string](200)),
		}
	}))

	log.Info("starting %s on port %d (env=%s)", cfg.Name, cfg.Port, cfg.Env)

	if err := app.Listen(); err != nil {
		app.Logger().Error("server error: %v", err)
	}
}
