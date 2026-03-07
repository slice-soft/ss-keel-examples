package main

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/slice-soft/ss-keel-core/config"
	"github.com/slice-soft/ss-keel-core/core"
	"github.com/slice-soft/ss-keel-core/logger"
)

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
	ExpiresIn int    `json:"expires_in"`
	TokenType string `json:"token_type"`
}

// Claims holds the JWT payload.
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// Principal is the authenticated user stored in Fiber locals.
type Principal struct {
	ID    string
	Email string
	Role  string
}

// staticUsers simulates a user store.
var staticUsers = map[string]struct {
	user     User
	password string
}{
	"alice@example.com": {User{"usr_1", "alice@example.com", "admin"}, "password123"},
	"bob@example.com":   {User{"usr_2", "bob@example.com", "user"}, "pass456"},
}

func issueToken(user User, secret string, ttlMinutes int) (string, error) {
	claims := Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(ttlMinutes) * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "keel-jwt-example",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// JWTMiddleware validates the Bearer token and stores the principal in locals.
func JWTMiddleware(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return core.Unauthorized("missing or malformed Authorization header")
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, core.Unauthorized("unexpected signing method")
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			return core.Unauthorized("invalid or expired token")
		}

		c.Locals("_keel_user", Principal{
			ID:    claims.UserID,
			Email: claims.Email,
			Role:  claims.Role,
		})
		return c.Next()
	}
}

// RequireRole is a middleware that restricts access to a specific role.
func RequireRole(role string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		principal, ok := c.Locals("_keel_user").(Principal)
		if !ok {
			return core.Unauthorized("not authenticated")
		}
		if principal.Role != role {
			return core.Forbidden("requires role: " + role)
		}
		return c.Next()
	}
}

func main() {
	port := config.GetEnvIntOrDefault("PORT", 7331)
	env := config.GetEnvOrDefault("APP_ENV", "development")
	serviceName := config.GetEnvOrDefault("SERVICE_NAME", "jwt-auth")
	jwtSecret := config.GetEnvOrDefault("JWT_SECRET", "change-me-in-production")
	tokenTTL := config.GetEnvIntOrDefault("TOKEN_TTL_MINUTES", 60)

	log := logger.NewLogger(env == "production")

	app := core.New(core.KConfig{
		ServiceName: serviceName,
		Port:        port,
		Env:         env,
		Docs: core.DocsConfig{
			Title:       "JWT Auth API",
			Version:     "1.0.0",
			Description: "JWT authentication example. Login to get a token, then use it to access protected routes.",
			Tags: []core.DocsTag{
				{Name: "auth", Description: "Authentication"},
				{Name: "protected", Description: "Protected resources"},
			},
		},
	})

	// Public routes
	app.RegisterController(core.ControllerFunc(func() []core.Route {
		return []core.Route{
			core.POST("/auth/login", func(c *core.Ctx) error {
				var req LoginRequest
				if err := c.ParseBody(&req); err != nil {
					return err
				}

				entry, ok := staticUsers[req.Email]
				if !ok || entry.password != req.Password {
					return core.Unauthorized("invalid email or password")
				}

				token, err := issueToken(entry.user, jwtSecret, tokenTTL)
				if err != nil {
					return core.Internal("could not issue token", err)
				}

				return c.OK(LoginResponse{
					Token:     token,
					ExpiresIn: tokenTTL * 60,
					TokenType: "Bearer",
				})
			}).
				Tag("auth").
				Describe("Login", "Exchange credentials for a JWT.").
				WithBody(core.WithBody[LoginRequest]()).
				WithResponse(core.WithResponse[LoginResponse](200)),
		}
	}))

	// Protected routes — all require a valid JWT
	protected := app.Group("/api", JWTMiddleware(jwtSecret))
	protected.RegisterController(core.ControllerFunc(func() []core.Route {
		return []core.Route{
			core.GET("/me", func(c *core.Ctx) error {
				p, _ := core.UserAs[Principal](c)
				return c.OK(map[string]string{
					"user_id": p.ID,
					"email":   p.Email,
					"role":    p.Role,
				})
			}).
				Tag("protected").
				Describe("Current user", "Returns the authenticated user's info.").
				WithSecured("bearerAuth").
				WithResponse(core.WithResponse[map[string]string](200)),

			core.GET("/admin/dashboard", func(c *core.Ctx) error {
				return c.OK(map[string]string{
					"message": "welcome to the admin dashboard",
				})
			}).
				Tag("protected").
				Describe("Admin dashboard", "Only accessible by users with role=admin.").
				Use(RequireRole("admin")).
				WithSecured("bearerAuth").
				WithResponse(core.WithResponse[map[string]string](200)),
		}
	}))

	log.Info("starting %s on port %d (env=%s)", serviceName, port, env)

	if err := app.Listen(); err != nil {
		app.Logger().Error("server error: %v", err)
	}
}
