package main

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/slice-soft/ss-keel-core/config"
	"github.com/slice-soft/ss-keel-core/contracts"
	"github.com/slice-soft/ss-keel-core/core"
	"github.com/slice-soft/ss-keel-core/core/httpx"
	"github.com/slice-soft/ss-keel-core/logger"
	ssredis "github.com/slice-soft/ss-keel-redis/redis"
)

// Note is the source-of-truth record stored in the in-memory repository.
type Note struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	UpdatedAt int64  `json:"updated_at"`
}

// GetNoteResponse returns the resolved note and whether it came from Redis or the source store.
type GetNoteResponse struct {
	Note   *Note  `json:"note"`
	Source string `json:"source"`
}

// WriteNoteRequest stores a note in the source store and invalidates the Redis entry.
type WriteNoteRequest struct {
	ID    string `json:"id"    validate:"required,min=2,max=64"`
	Title string `json:"title" validate:"required,min=3,max=120"`
	Body  string `json:"body"  validate:"required,max=2000"`
}

// WriteNoteResponse echoes the stored note and the cache action taken by the service.
type WriteNoteResponse struct {
	Note   *Note  `json:"note"`
	Cache  string `json:"cache"`
	Source string `json:"source"`
}

type NotesStore struct {
	mu    sync.RWMutex
	items map[string]Note
}

func NewNotesStore() *NotesStore {
	now := time.Now().UnixMilli()
	return &NotesStore{
		items: map[string]Note{
			"note-1": {
				ID:        "note-1",
				Title:     "Cache aside",
				Body:      "The first read comes from the source store and the second one comes from Redis.",
				UpdatedAt: now,
			},
			"note-2": {
				ID:        "note-2",
				Title:     "Invalidate on write",
				Body:      "POST and DELETE remove the cached entry so the next GET repopulates it.",
				UpdatedAt: now,
			},
		},
	}
}

func (s *NotesStore) FindByID(id string) (*Note, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	note, ok := s.items[id]
	if !ok {
		return nil, false
	}
	return &note, true
}

func (s *NotesStore) Save(note Note) Note {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.items[note.ID] = note
	return note
}

func (s *NotesStore) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.items[id]; !ok {
		return false
	}
	delete(s.items, id)
	return true
}

type NotesService struct {
	store *NotesStore
	cache contracts.Cache
	ttl   time.Duration
	log   *logger.Logger
}

func NewNotesService(store *NotesStore, cache contracts.Cache, ttl time.Duration, log *logger.Logger) *NotesService {
	return &NotesService{
		store: store,
		cache: cache,
		ttl:   ttl,
		log:   log,
	}
}

func (s *NotesService) GetByID(ctx context.Context, id string) (*Note, string, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, "", nil
	}

	cacheKey := noteCacheKey(id)
	cached, err := s.cache.Get(ctx, cacheKey)
	if err != nil {
		s.log.Warn("redis get failed [key=%s]: %v", cacheKey, err)
	} else if cached != nil {
		var note Note
		if err := json.Unmarshal(cached, &note); err != nil {
			s.log.Warn("redis payload decode failed [key=%s]: %v", cacheKey, err)
			if deleteErr := s.cache.Delete(ctx, cacheKey); deleteErr != nil {
				s.log.Warn("redis delete failed [key=%s]: %v", cacheKey, deleteErr)
			}
		} else {
			return &note, "cache", nil
		}
	}

	note, ok := s.store.FindByID(id)
	if !ok {
		return nil, "", nil
	}

	payload, err := json.Marshal(note)
	if err != nil {
		return nil, "", err
	}
	if err := s.cache.Set(ctx, cacheKey, payload, s.ttl); err != nil {
		s.log.Warn("redis set failed [key=%s]: %v", cacheKey, err)
	}

	return note, "store", nil
}

func (s *NotesService) Write(ctx context.Context, req WriteNoteRequest) (*Note, error) {
	note := Note{
		ID:        strings.TrimSpace(req.ID),
		Title:     strings.TrimSpace(req.Title),
		Body:      strings.TrimSpace(req.Body),
		UpdatedAt: time.Now().UnixMilli(),
	}

	saved := s.store.Save(note)
	if err := s.cache.Delete(ctx, noteCacheKey(saved.ID)); err != nil {
		s.log.Warn("redis delete failed [key=%s]: %v", noteCacheKey(saved.ID), err)
	}

	return &saved, nil
}

func (s *NotesService) Delete(ctx context.Context, id string) (bool, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return false, nil
	}

	deleted := s.store.Delete(id)
	if !deleted {
		return false, nil
	}

	if err := s.cache.Delete(ctx, noteCacheKey(id)); err != nil {
		s.log.Warn("redis delete failed [key=%s]: %v", noteCacheKey(id), err)
	}
	return true, nil
}

func noteCacheKey(id string) string {
	return "notes:" + strings.TrimSpace(id)
}

type NotesModule struct {
	log      *logger.Logger
	cache    *ssredis.Client
	store    *NotesStore
	cacheTTL time.Duration
}

func NewNotesModule(log *logger.Logger, cache *ssredis.Client, store *NotesStore, cacheTTL time.Duration) *NotesModule {
	return &NotesModule{
		log:      log,
		cache:    cache,
		store:    store,
		cacheTTL: cacheTTL,
	}
}

func (m *NotesModule) Register(app *core.App) {
	notesService := NewNotesService(m.store, m.cache, m.cacheTTL, m.log)
	api := app.Group("/api/v1")
	api.RegisterController(contracts.ControllerFunc[httpx.Route](func() []httpx.Route {
		return []httpx.Route{
			httpx.GET("/notes/:id", func(c *httpx.Ctx) error {
				note, source, err := notesService.GetByID(context.Background(), c.Params("id"))
				if err != nil {
					return core.Internal("could not fetch note", err)
				}
				if note == nil {
					return core.NotFound("note not found")
				}
				return c.OK(GetNoteResponse{Note: note, Source: source})
			}).
				Tag("notes").
				Describe("Get note by ID", "Reads from Redis first; on miss it loads the source store and caches the result with a TTL.").
				WithResponse(httpx.WithResponse[GetNoteResponse](200)),

			httpx.POST("/notes", func(c *httpx.Ctx) error {
				var req WriteNoteRequest
				if err := c.ParseBody(&req); err != nil {
					return err
				}

				note, err := notesService.Write(context.Background(), req)
				if err != nil {
					return core.Internal("could not store note", err)
				}
				return c.Created(WriteNoteResponse{
					Note:   note,
					Cache:  "invalidated",
					Source: "store",
				})
			}).
				Tag("notes").
				Describe("Create or replace note", "Writes to the source store and invalidates the Redis entry so the next GET repopulates the cache.").
				WithBody(httpx.WithBody[WriteNoteRequest]()).
				WithResponse(httpx.WithResponse[WriteNoteResponse](201)),

			httpx.DELETE("/notes/:id", func(c *httpx.Ctx) error {
				deleted, err := notesService.Delete(context.Background(), c.Params("id"))
				if err != nil {
					return core.Internal("could not delete note", err)
				}
				if !deleted {
					return core.NotFound("note not found")
				}
				return c.NoContent()
			}).
				Tag("notes").
				Describe("Delete note", "Deletes the source record and invalidates the Redis entry.").
				WithResponse(httpx.WithResponse[struct{}](204)),
		}
	}))
}

func main() {
	port := config.GetEnvIntOrDefault("PORT", 7331)
	env := config.GetEnvOrDefault("APP_ENV", "development")
	serviceName := config.GetEnvOrDefault("SERVICE_NAME", "redis-cache-example")
	redisURL := config.GetEnvOrDefault("REDIS_URL", "redis://localhost:6379")
	cacheTTL := time.Duration(config.GetEnvIntOrDefault("CACHE_TTL_SECONDS", 30)) * time.Second

	log := logger.NewLogger(env == "production")

	redisClient, err := ssredis.New(ssredis.Config{
		URL:    redisURL,
		Logger: log,
	})
	if err != nil {
		log.Error("failed to initialize redis: %v", err)
		os.Exit(1)
	}
	defer redisClient.Close()

	app := core.New(core.KConfig{
		ServiceName: serviceName,
		Port:        port,
		Env:         env,
		Docs: core.DocsConfig{
			Title:       "Redis Cache API",
			Version:     "1.0.0",
			Description: "Cache-aside notes service using ss-keel-redis and contracts.Cache in the service layer.",
			Tags: []core.DocsTag{
				{Name: "notes", Description: "Cached notes resource"},
			},
		},
	})

	app.RegisterHealthChecker(ssredis.NewHealthChecker(redisClient))
	app.Use(NewNotesModule(log, redisClient, NewNotesStore(), cacheTTL))

	log.Info("starting %s on port %d (env=%s, cache_ttl=%s)", serviceName, port, env, cacheTTL)

	if err := app.Listen(); err != nil {
		app.Logger().Error("server error: %v", err)
	}
}
