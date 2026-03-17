package main

import (
	"os"
	"time"

	"github.com/slice-soft/ss-keel-core/config"
	"github.com/slice-soft/ss-keel-core/contracts"
	"github.com/slice-soft/ss-keel-core/core"
	"github.com/slice-soft/ss-keel-core/core/httpx"
	"github.com/slice-soft/ss-keel-core/logger"
	"github.com/slice-soft/ss-keel-gorm/database"
	"gorm.io/gorm"
)

// Product is the GORM model.
type Product struct {
	ID          uint           `json:"id"          gorm:"primarykey"`
	Name        string         `json:"name"        gorm:"not null"`
	Description string         `json:"description"`
	Price       float64        `json:"price"       gorm:"not null"`
	Stock       int            `json:"stock"       gorm:"default:0"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-"           gorm:"index"`
}

// CreateProductRequest is the creation payload.
type CreateProductRequest struct {
	Name        string  `json:"name"        validate:"required,min=2,max=120"`
	Description string  `json:"description" validate:"omitempty,max=500"`
	Price       float64 `json:"price"       validate:"required,gt=0"`
	Stock       int     `json:"stock"       validate:"gte=0"`
}

// UpdateProductRequest is the update payload.
type UpdateProductRequest struct {
	Name        string  `json:"name"        validate:"omitempty,min=2,max=120"`
	Description string  `json:"description" validate:"omitempty,max=500"`
	Price       float64 `json:"price"       validate:"omitempty,gt=0"`
	Stock       *int    `json:"stock"       validate:"omitempty,gte=0"`
}

func main() {
	port := config.GetEnvIntOrDefault("PORT", 7331)
	env := config.GetEnvOrDefault("APP_ENV", "development")
	serviceName := config.GetEnvOrDefault("SERVICE_NAME", "gorm-postgres")
	dbPort := config.GetEnvIntOrDefault("DB_PORT", 5432)
	dbHost := config.GetEnvOrDefault("DB_HOST", "localhost")
	dbUser := config.GetEnvOrDefault("DB_USER", "postgres")
	dbPassword := config.GetEnvOrDefault("DB_PASSWORD", "postgres")
	dbName := config.GetEnvOrDefault("DB_NAME", "keelexamples")
	dbSSLMode := config.GetEnvOrDefault("DB_SSLMODE", "disable")

	log := logger.NewLogger(env == "production")

	// Connect to PostgreSQL using ss-keel-gorm.
	dbInstance, err := database.New(database.Config{
		Engine:   database.EnginePostgres,
		Host:     dbHost,
		Port:     dbPort,
		User:     dbUser,
		Password: dbPassword,
		Database: dbName,
		SSLMode:  dbSSLMode,
	})
	if err != nil {
		log.Error("failed to connect to database: %v", err)
		os.Exit(1)
	}

	db := dbInstance.DB

	// NOTE: Keel does not run automatic migrations.
	// You are responsible for managing schema changes manually.
	//
	// Options:
	//   Option 1 (recommended): raw SQL files — up.sql / down.sql
	//   Option 2: external tools — goose, atlas, dbmate
	//   Option 3: CI-driven — apply SQL scripts in your pipeline
	//
	// AutoMigrate is shown here only for development convenience — do NOT use it in production.
	if err := db.AutoMigrate(&Product{}); err != nil {
		log.Error("migration failed: %v", err)
		os.Exit(1)
	}

	app := core.New(core.KConfig{
		ServiceName: serviceName,
		Port:        port,
		Env:         env,
		Docs: core.DocsConfig{
			Title:       "GORM + PostgreSQL API",
			Version:     "1.0.0",
			Description: "Database-backed CRUD using ss-keel-gorm and PostgreSQL.",
			Tags: []core.DocsTag{
				{Name: "products", Description: "Product catalog"},
			},
		},
	})

	// Register the built-in DB health checker from ss-keel-gorm.
	app.RegisterHealthChecker(database.NewHealthChecker(dbInstance))

	v1 := app.Group("/api/v1")
	v1.RegisterController(contracts.ControllerFunc[httpx.Route](func() []httpx.Route {
		return []httpx.Route{
			// List products
			httpx.GET("/products", func(c *httpx.Ctx) error {
				var products []Product
				if err := db.Find(&products).Error; err != nil {
					return core.Internal("could not fetch products", err)
				}
				return c.OK(map[string]any{
					"data":  products,
					"total": len(products),
				})
			}).
				Tag("products").
				Describe("List products").
				WithResponse(httpx.WithResponse[map[string]any](200)),

			// Get product by ID
			httpx.GET("/products/:id", func(c *httpx.Ctx) error {
				var product Product
				if err := db.First(&product, c.Params("id")).Error; err != nil {
					return core.NotFound("product not found")
				}
				return c.OK(product)
			}).
				Tag("products").
				Describe("Get product by ID").
				WithResponse(httpx.WithResponse[Product](200)),

			// Create product
			httpx.POST("/products", func(c *httpx.Ctx) error {
				var req CreateProductRequest
				if err := c.ParseBody(&req); err != nil {
					return err
				}
				product := Product{
					Name:        req.Name,
					Description: req.Description,
					Price:       req.Price,
					Stock:       req.Stock,
				}
				if err := db.Create(&product).Error; err != nil {
					return core.Internal("could not create product", err)
				}
				return c.Created(product)
			}).
				Tag("products").
				Describe("Create product").
				WithBody(httpx.WithBody[CreateProductRequest]()).
				WithResponse(httpx.WithResponse[Product](201)),

			// Update product
			httpx.PATCH("/products/:id", func(c *httpx.Ctx) error {
				var product Product
				if err := db.First(&product, c.Params("id")).Error; err != nil {
					return core.NotFound("product not found")
				}
				var req UpdateProductRequest
				if err := c.ParseBody(&req); err != nil {
					return err
				}
				updates := map[string]any{}
				if req.Name != "" {
					updates["name"] = req.Name
				}
				if req.Description != "" {
					updates["description"] = req.Description
				}
				if req.Price > 0 {
					updates["price"] = req.Price
				}
				if req.Stock != nil {
					updates["stock"] = *req.Stock
				}
				if err := db.Model(&product).Updates(updates).Error; err != nil {
					return core.Internal("could not update product", err)
				}
				return c.OK(product)
			}).
				Tag("products").
				Describe("Update product").
				WithBody(httpx.WithBody[UpdateProductRequest]()).
				WithResponse(httpx.WithResponse[Product](200)),

			// Delete product (soft delete via GORM's DeletedAt)
			httpx.DELETE("/products/:id", func(c *httpx.Ctx) error {
				var product Product
				if err := db.First(&product, c.Params("id")).Error; err != nil {
					return core.NotFound("product not found")
				}
				if err := db.Delete(&product).Error; err != nil {
					return core.Internal("could not delete product", err)
				}
				return c.NoContent()
			}).
				Tag("products").
				Describe("Delete product", "Soft deletes the product via GORM's DeletedAt field."),
		}
	}))

	log.Info("starting %s on port %d (env=%s)", serviceName, port, env)

	if err := app.Listen(); err != nil {
		app.Logger().Error("server error: %v", err)
	}
}
