package main

import (
	"github.com/slice-soft/ss-keel-core/config"
	"github.com/slice-soft/ss-keel-core/core"
	"github.com/slice-soft/ss-keel-core/logger"
)

// RegisterRequest demonstrates a rich set of validation rules.
type RegisterRequest struct {
	Name     string `json:"name"     validate:"required,min=2,max=80"`
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
	Age      int    `json:"age"      validate:"required,gte=18,lte=120"`
	Website  string `json:"website"  validate:"omitempty,url"`
	Phone    string `json:"phone"    validate:"omitempty,e164"`
}

// RegisterResponse is returned on successful registration.
type RegisterResponse struct {
	Message string `json:"message"`
	Email   string `json:"email"`
	Name    string `json:"name"`
}

// OrderRequest demonstrates nested struct and slice validation.
type OrderRequest struct {
	CustomerID string      `json:"customer_id" validate:"required,min=1"`
	Items      []OrderItem `json:"items"       validate:"required,min=1,dive"`
	Note       string      `json:"note"        validate:"omitempty,max=200"`
}

// OrderItem is a line item within an order.
type OrderItem struct {
	SKU      string  `json:"sku"      validate:"required"`
	Quantity int     `json:"quantity" validate:"required,gt=0"`
	Price    float64 `json:"price"    validate:"required,gt=0"`
}

func main() {
	port := config.GetEnvIntOrDefault("PORT", 7331)
	env := config.GetEnvOrDefault("APP_ENV", "development")
	serviceName := config.GetEnvOrDefault("SERVICE_NAME", "validation-example")

	log := logger.NewLogger(env == "production")

	app := core.New(core.KConfig{
		ServiceName: serviceName,
		Port:        port,
		Env:         env,
		Docs: core.DocsConfig{
			Title:       "Validation API",
			Version:     "1.0.0",
			Description: "Demonstrates request validation with struct tags. Send invalid payloads to see structured 422 errors.",
			Tags: []core.DocsTag{
				{Name: "users", Description: "User endpoints"},
				{Name: "orders", Description: "Order endpoints"},
			},
		},
	})

	app.RegisterController(core.ControllerFunc(func() []core.Route {
		return []core.Route{
			// Registration — shows field-level validation errors on 422.
			core.POST("/users/register", func(c *core.Ctx) error {
				var req RegisterRequest
				// ParseBody returns 400 for malformed JSON, 422 for validation errors.
				if err := c.ParseBody(&req); err != nil {
					return err
				}
				return c.Created(RegisterResponse{
					Message: "registration successful",
					Email:   req.Email,
					Name:    req.Name,
				})
			}).
				Tag("users").
				Describe("Register a user", "Returns 422 with per-field errors on invalid input.").
				WithBody(core.WithBody[RegisterRequest]()).
				WithResponse(core.WithResponse[RegisterResponse](201)),

			// Order — shows nested struct and slice validation.
			core.POST("/orders", func(c *core.Ctx) error {
				var req OrderRequest
				if err := c.ParseBody(&req); err != nil {
					return err
				}
				total := 0.0
				for _, item := range req.Items {
					total += item.Price * float64(item.Quantity)
				}
				return c.Created(map[string]any{
					"message":     "order created",
					"customer_id": req.CustomerID,
					"item_count":  len(req.Items),
					"total":       total,
				})
			}).
				Tag("orders").
				Describe("Place an order", "Validates nested items slice with `dive`.").
				WithBody(core.WithBody[OrderRequest]()).
				WithResponse(core.WithResponse[map[string]any](201)),
		}
	}))

	log.Info("starting %s on port %d (env=%s)", serviceName, port, env)

	if err := app.Listen(); err != nil {
		app.Logger().Error("server error: %v", err)
	}
}
