package main

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/slice-soft/ss-keel-core/config"
	"github.com/slice-soft/ss-keel-core/contracts"
	"github.com/slice-soft/ss-keel-core/core"
	"github.com/slice-soft/ss-keel-core/core/httpx"
	"github.com/slice-soft/ss-keel-core/logger"
	ssotel "github.com/slice-soft/ss-keel-otel/otel"
)

// AppConfig maps application.properties keys to typed Go fields.
type AppConfig struct {
	Name string `keel:"app.name"`
	Env  string `keel:"app.env"`
	Port int    `keel:"server.port"`
}

// Order is the in-memory domain record.
type Order struct {
	ID        string `json:"id"`
	Product   string `json:"product"`
	Qty       int    `json:"qty"`
	CreatedAt int64  `json:"created_at"`
}

// CreateOrderRequest is the payload for POST /api/orders.
type CreateOrderRequest struct {
	Product string `json:"product" validate:"required,min=2,max=120"`
	Qty     int    `json:"qty"     validate:"required,min=1,max=9999"`
}

// OrderStore is a thread-safe in-memory repository.
type OrderStore struct {
	mu      sync.RWMutex
	items   map[string]Order
	counter atomic.Int64
}

func NewOrderStore() *OrderStore {
	now := time.Now().UnixMilli()
	s := &OrderStore{items: make(map[string]Order)}
	s.counter.Store(2)
	s.items["order-1"] = Order{ID: "order-1", Product: "Keel T-Shirt", Qty: 2, CreatedAt: now}
	s.items["order-2"] = Order{ID: "order-2", Product: "Gopher Plushie", Qty: 1, CreatedAt: now}
	return s
}

func (s *OrderStore) All() []Order {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Order, 0, len(s.items))
	for _, o := range s.items {
		out = append(out, o)
	}
	return out
}

func (s *OrderStore) FindByID(id string) (*Order, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	o, ok := s.items[id]
	if !ok {
		return nil, false
	}
	return &o, true
}

func (s *OrderStore) Save(o Order) Order {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[o.ID] = o
	return o
}

// OrdersModule wires up the /api/orders routes.
type OrdersModule struct {
	log   *logger.Logger
	store *OrderStore
	app   *core.App
}

func NewOrdersModule(log *logger.Logger, store *OrderStore, app *core.App) *OrdersModule {
	return &OrdersModule{log: log, store: store, app: app}
}

func (m *OrdersModule) Register(a *core.App) {
	api := a.Group("/api")
	api.RegisterController(contracts.ControllerFunc[httpx.Route](func() []httpx.Route {
		return []httpx.Route{
			// GET /api/orders — automatic HTTP span from OTel middleware; no manual span needed.
			httpx.GET("/orders", func(c *httpx.Ctx) error {
				return c.OK(m.store.All())
			}).
				Tag("orders").
				Describe("List orders", "Returns all in-memory orders. The HTTP span is created automatically by the OTel middleware.").
				WithResponse(httpx.WithResponse[[]Order](200)),

			// GET /api/orders/:id — manual child span shows nested tracing.
			httpx.GET("/orders/:id", func(c *httpx.Ctx) error {
				_, span := m.app.Tracer().Start(c.UserContext(), "OrderStore.FindByID")
				defer span.End()

				id := c.Params("id")
				span.SetAttribute("order.id", id)

				order, ok := m.store.FindByID(id)
				if !ok {
					span.RecordError(fmt.Errorf("order %s not found", id))
					return core.NotFound("order not found")
				}

				span.SetAttribute("order.product", order.Product)
				span.SetAttribute("order.qty", order.Qty)
				return c.OK(order)
			}).
				Tag("orders").
				Describe("Get order by ID", "Fetches a single order. Creates a manual child span (OrderStore.FindByID) under the automatic HTTP parent span.").
				WithResponse(httpx.WithResponse[Order](200)),

			// POST /api/orders — child span wraps the write operation.
			httpx.POST("/orders", func(c *httpx.Ctx) error {
				var req CreateOrderRequest
				if err := c.ParseBody(&req); err != nil {
					return err
				}

				_, span := m.app.Tracer().Start(c.UserContext(), "OrderStore.Save")
				defer span.End()

				id := fmt.Sprintf("order-%d", m.store.counter.Add(1))
				span.SetAttribute("order.id", id)
				span.SetAttribute("order.product", req.Product)
				span.SetAttribute("order.qty", req.Qty)

				order := m.store.Save(Order{
					ID:        id,
					Product:   req.Product,
					Qty:       req.Qty,
					CreatedAt: time.Now().UnixMilli(),
				})

				m.log.Info("order created id=%s product=%q qty=%d", order.ID, order.Product, order.Qty)
				return c.Created(order)
			}).
				Tag("orders").
				Describe("Create order", "Stores a new order in memory. Creates a manual child span (OrderStore.Save) with attributes for product and quantity.").
				WithBody(httpx.WithBody[CreateOrderRequest]()).
				WithResponse(httpx.WithResponse[Order](201)),
		}
	}))
}

// setupOtel initializes the OpenTelemetry SDK and registers Fiber HTTP middleware.
// All telemetry is skipped when OTEL_ENABLED=false (the default).
func setupOtel(app *core.App, log *logger.Logger) *ssotel.Provider {
	otelConfig := config.MustLoadConfig[ssotel.Config]()
	otelConfig.Logger = log

	provider, err := ssotel.New(otelConfig)
	if err != nil {
		log.Error("failed to initialize otel: %v", err)
		return provider
	}

	app.SetTracer(provider)
	app.Fiber().Use(provider.Middleware())
	app.OnShutdown(provider.Shutdown)

	return provider
}

func main() {
	cfg := config.MustLoadConfig[AppConfig]()
	log := logger.NewLogger(cfg.Env == "production")

	app := core.New(core.KConfig{
		ServiceName: cfg.Name,
		Port:        cfg.Port,
		Env:         cfg.Env,
		Docs: core.DocsConfig{
			Title:       "OTel Example API",
			Version:     "1.0.0",
			Description: "Demonstrates ss-keel-otel: automatic HTTP spans, manual child spans, span attributes, and error recording.",
			Tags: []core.DocsTag{
				{Name: "orders", Description: "In-memory orders — hit these to generate traces"},
			},
		},
	})

	// OTel must be set up before registering routes so the middleware is first in the chain.
	setupOtel(app, log)

	app.Use(NewOrdersModule(log, NewOrderStore(), app))

	log.Info("starting %s on :%d (env=%s, otel=%v)", cfg.Name, cfg.Port, cfg.Env, os.Getenv("OTEL_ENABLED"))

	if err := app.Listen(); err != nil {
		app.Logger().Error("server error: %v", err)
	}
}
