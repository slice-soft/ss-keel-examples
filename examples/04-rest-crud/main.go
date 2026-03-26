package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/slice-soft/ss-keel-core/config"
	"github.com/slice-soft/ss-keel-core/contracts"
	"github.com/slice-soft/ss-keel-core/core"
	"github.com/slice-soft/ss-keel-core/core/httpx"
	"github.com/slice-soft/ss-keel-core/logger"
)

type AppConfig struct {
	Name string `keel:"app.name"`
	Env  string `keel:"app.env"`
	Port int    `keel:"server.port"`
}

// Task is the domain model.
type Task struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Done        bool      `json:"done"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateTaskRequest is the payload for creating a task.
type CreateTaskRequest struct {
	Title       string `json:"title"       validate:"required,min=2,max=120"`
	Description string `json:"description" validate:"omitempty,max=500"`
}

// UpdateTaskRequest is the payload for updating a task.
type UpdateTaskRequest struct {
	Title       string `json:"title"       validate:"omitempty,min=2,max=120"`
	Description string `json:"description" validate:"omitempty,max=500"`
	Done        *bool  `json:"done"        validate:"omitempty"`
}

// store is a simple in-memory store for tasks.
type store struct {
	mu      sync.RWMutex
	items   map[string]*Task
	counter int
}

func newStore() *store {
	return &store{items: make(map[string]*Task)}
}

func (s *store) nextID() string {
	s.counter++
	return fmt.Sprintf("task_%d", s.counter)
}

func (s *store) List() []*Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*Task, 0, len(s.items))
	for _, t := range s.items {
		result = append(result, t)
	}
	return result
}

func (s *store) Get(id string) (*Task, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.items[id]
	return t, ok
}

func (s *store) Create(req CreateTaskRequest) *Task {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	t := &Task{
		ID:          s.nextID(),
		Title:       req.Title,
		Description: req.Description,
		Done:        false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	s.items[t.ID] = t
	return t
}

func (s *store) Update(id string, req UpdateTaskRequest) (*Task, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.items[id]
	if !ok {
		return nil, false
	}
	if req.Title != "" {
		t.Title = req.Title
	}
	if req.Description != "" {
		t.Description = req.Description
	}
	if req.Done != nil {
		t.Done = *req.Done
	}
	t.UpdatedAt = time.Now()
	return t, true
}

func (s *store) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.items[id]
	if ok {
		delete(s.items, id)
	}
	return ok
}

func main() {
	cfg := config.MustLoadConfig[AppConfig]()
	port := cfg.Port
	env := cfg.Env
	serviceName := cfg.Name

	log := logger.NewLogger(env == "production")
	db := newStore()

	app := core.New(core.KConfig{
		ServiceName: serviceName,
		Port:        port,
		Env:         env,
		Docs: core.DocsConfig{
			Title:       "REST CRUD API",
			Version:     "1.0.0",
			Description: "Full CRUD example for a Task resource using in-memory storage.",
			Tags: []core.DocsTag{
				{Name: "tasks", Description: "Task management"},
			},
		},
	})

	v1 := app.Group("/api/v1")
	v1.RegisterController(contracts.ControllerFunc[httpx.Route](func() []httpx.Route {
		return []httpx.Route{
			// List all tasks
			httpx.GET("/tasks", func(c *httpx.Ctx) error {
				tasks := db.List()
				return c.OK(map[string]any{
					"data":  tasks,
					"total": len(tasks),
				})
			}).
				Tag("tasks").
				Describe("List tasks").
				WithResponse(httpx.WithResponse[map[string]any](200)),

			// Get a single task
			httpx.GET("/tasks/:id", func(c *httpx.Ctx) error {
				task, ok := db.Get(c.Params("id"))
				if !ok {
					return core.NotFound("task not found")
				}
				return c.OK(task)
			}).
				Tag("tasks").
				Describe("Get task by ID").
				WithResponse(httpx.WithResponse[Task](200)),

			// Create a task
			httpx.POST("/tasks", func(c *httpx.Ctx) error {
				var req CreateTaskRequest
				if err := c.ParseBody(&req); err != nil {
					return err
				}
				task := db.Create(req)
				return c.Created(task)
			}).
				Tag("tasks").
				Describe("Create a task").
				WithBody(httpx.WithBody[CreateTaskRequest]()).
				WithResponse(httpx.WithResponse[Task](201)),

			// Update a task
			httpx.PATCH("/tasks/:id", func(c *httpx.Ctx) error {
				var req UpdateTaskRequest
				if err := c.ParseBody(&req); err != nil {
					return err
				}
				task, ok := db.Update(c.Params("id"), req)
				if !ok {
					return core.NotFound("task not found")
				}
				return c.OK(task)
			}).
				Tag("tasks").
				Describe("Update a task").
				WithBody(httpx.WithBody[UpdateTaskRequest]()).
				WithResponse(httpx.WithResponse[Task](200)),

			// Delete a task
			httpx.DELETE("/tasks/:id", func(c *httpx.Ctx) error {
				if !db.Delete(c.Params("id")) {
					return core.NotFound("task not found")
				}
				return c.NoContent()
			}).
				Tag("tasks").
				Describe("Delete a task").
				WithResponse(httpx.WithResponse[struct{}](204)),
		}
	}))

	log.Info("starting %s on port %d (env=%s)", serviceName, port, env)

	if err := app.Listen(); err != nil {
		app.Logger().Error("server error: %v", err)
	}
}
