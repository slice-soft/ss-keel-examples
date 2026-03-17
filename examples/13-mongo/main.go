package main

import (
	"context"
	"os"

	"github.com/slice-soft/ss-keel-core/config"
	"github.com/slice-soft/ss-keel-core/contracts"
	"github.com/slice-soft/ss-keel-core/core"
	"github.com/slice-soft/ss-keel-core/core/httpx"
	"github.com/slice-soft/ss-keel-core/logger"
	"github.com/slice-soft/ss-keel-mongo/mongo"
)

// Note is the MongoDB document model.
// EntityBase provides ID (UUID string), CreatedAt and UpdatedAt (Unix ms).
// Call OnCreate() before Insert and OnUpdate() before Update.
type Note struct {
	mongo.EntityBase `bson:",inline"`
	Title            string `json:"title" bson:"title"`
	Body             string `json:"body"  bson:"body"`
}

// CreateNoteRequest is the creation payload.
type CreateNoteRequest struct {
	Title string `json:"title" validate:"required,min=2,max=120"`
	Body  string `json:"body"  validate:"required,max=5000"`
}

// UpdateNoteRequest is the partial update payload.
type UpdateNoteRequest struct {
	Title string `json:"title" validate:"omitempty,min=2,max=120"`
	Body  string `json:"body"  validate:"omitempty,max=5000"`
}

func main() {
	port := config.GetEnvIntOrDefault("PORT", 7331)
	env := config.GetEnvOrDefault("APP_ENV", "development")
	serviceName := config.GetEnvOrDefault("SERVICE_NAME", "mongo-example")
	mongoURI := config.GetEnvOrDefault("MONGO_URI", "mongodb://localhost:27017")
	mongoDatabase := config.GetEnvOrDefault("MONGO_DATABASE", "keelexamples")

	log := logger.NewLogger(env == "production")

	// Connect to MongoDB using the ss-keel-mongo addon.
	// Run:  keel add mongo
	client, err := mongo.New(mongo.Config{
		URI:      mongoURI,
		Database: mongoDatabase,
		Logger:   log,
	})
	if err != nil {
		log.Error("failed to connect to MongoDB: %v", err)
		os.Exit(1)
	}
	defer client.Close()

	// NewRepository creates a type-safe CRUD repository for the "notes" collection.
	// IDs are UUID strings — Keel uses UUID as the only ID strategy across all databases.
	repo := mongo.NewRepository[Note, string](client, "notes")

	app := core.New(core.KConfig{
		ServiceName: serviceName,
		Port:        port,
		Env:         env,
		Docs: core.DocsConfig{
			Title:       "MongoDB API",
			Version:     "1.0.0",
			Description: "Document CRUD using the ss-keel-mongo addon with EntityBase and a generic repository.",
			Tags: []core.DocsTag{
				{Name: "notes", Description: "Notes resource"},
			},
		},
	})

	// Register the built-in MongoDB health checker — wired into /health.
	app.RegisterHealthChecker(mongo.NewHealthChecker(client))

	v1 := app.Group("/api/v1")
	v1.RegisterController(contracts.ControllerFunc[httpx.Route](func() []httpx.Route {
		return []httpx.Route{
			// GET /api/v1/notes?page=1&limit=20 — paginated list.
			httpx.GET("/notes", func(c *httpx.Ctx) error {
				q := c.ParsePagination()
				page, err := repo.FindAll(context.Background(), q)
				if err != nil {
					return core.Internal("could not fetch notes", err)
				}
				return c.OK(page)
			}).
				Tag("notes").
				Describe("List notes", "Returns paginated notes. Use ?page=1&limit=20.").
				WithQueryParam("page", "integer", false, "Page number (default: 1)").
				WithQueryParam("limit", "integer", false, "Items per page (default: 20, max: 100)").
				WithResponse(httpx.WithResponse[httpx.Page[Note]](200)),

			// GET /api/v1/notes/:id — fetch by ObjectID hex.
			httpx.GET("/notes/:id", func(c *httpx.Ctx) error {
				note, err := repo.FindByID(context.Background(), c.Params("id"))
				if err != nil || note == nil {
					return core.NotFound("note not found")
				}
				return c.OK(note)
			}).
				Tag("notes").
				Describe("Get note by ID").
				WithResponse(httpx.WithResponse[Note](200)),

			// POST /api/v1/notes — create a new note.
			httpx.POST("/notes", func(c *httpx.Ctx) error {
				var req CreateNoteRequest
				if err := c.ParseBody(&req); err != nil {
					return err
				}
				note := &Note{Title: req.Title, Body: req.Body}
				note.OnCreate() // sets ID, CreatedAt, UpdatedAt
				if err := repo.Create(context.Background(), note); err != nil {
					return core.Internal("could not create note", err)
				}
				return c.Created(note)
			}).
				Tag("notes").
				Describe("Create a note").
				WithBody(httpx.WithBody[CreateNoteRequest]()).
				WithResponse(httpx.WithResponse[Note](201)),

			// PATCH /api/v1/notes/:id — partial update.
			httpx.PATCH("/notes/:id", func(c *httpx.Ctx) error {
				id := c.Params("id")
				existing, err := repo.FindByID(context.Background(), id)
				if err != nil || existing == nil {
					return core.NotFound("note not found")
				}
				var req UpdateNoteRequest
				if err := c.ParseBody(&req); err != nil {
					return err
				}
				if req.Title != "" {
					existing.Title = req.Title
				}
				if req.Body != "" {
					existing.Body = req.Body
				}
				existing.OnUpdate() // bumps UpdatedAt
				if err := repo.Update(context.Background(), id, existing); err != nil {
					return core.Internal("could not update note", err)
				}
				return c.OK(existing)
			}).
				Tag("notes").
				Describe("Update a note").
				WithBody(httpx.WithBody[UpdateNoteRequest]()).
				WithResponse(httpx.WithResponse[Note](200)),

			// DELETE /api/v1/notes/:id — remove a note.
			httpx.DELETE("/notes/:id", func(c *httpx.Ctx) error {
				existing, err := repo.FindByID(context.Background(), c.Params("id"))
				if err != nil || existing == nil {
					return core.NotFound("note not found")
				}
				if err := repo.Delete(context.Background(), c.Params("id")); err != nil {
					return core.Internal("could not delete note", err)
				}
				return c.NoContent()
			}).
				Tag("notes").
				Describe("Delete a note"),
		}
	}))

	log.Info("starting %s on port %d (env=%s)", serviceName, port, env)

	if err := app.Listen(); err != nil {
		app.Logger().Error("server error: %v", err)
	}
}
