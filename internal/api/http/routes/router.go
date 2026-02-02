package routes

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"

	"github.com/franciscozamorau/osmi-server/internal/api/http/middleware"
	"github.com/franciscozamorau/osmi-server/internal/config"
)

type Router struct {
	router *chi.Mux
	logger *zap.Logger
	config *config.Config
}

func NewRouter(
	logger *zap.Logger,
	cfg *config.Config,
) *Router {
	r := chi.NewRouter()

	// Middleware base
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(httpmiddleware.Logger(logger))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORS.AllowedOrigins,
		AllowedMethods:   cfg.CORS.AllowedMethods,
		AllowedHeaders:   cfg.CORS.AllowedHeaders,
		ExposedHeaders:   cfg.CORS.ExposedHeaders,
		AllowCredentials: cfg.CORS.AllowCredentials,
		MaxAge:           cfg.CORS.MaxAge,
	}))

	// Rate limiting
	if cfg.RateLimit.Enabled {
		r.Use(httpmiddleware.RateLimiter(cfg.RateLimit))
	}

	// Timeouts
	r.Use(middleware.Timeout(time.Duration(cfg.HTTPTimeout) * time.Second))

	return &Router{
		router: r,
		logger: logger,
		config: cfg,
	}
}

func (r *Router) RegisterGateway(gwMux *runtime.ServeMux) *chi.Mux {
	// API v1
	r.router.Route("/api/v1", func(v1 chi.Router) {
		// Rutas protegidas
		v1.Group(func(auth chi.Router) {
			auth.Use(httpmiddleware.Auth(r.config.JWT.Secret))

			// Tickets
			auth.Route("/tickets", func(tickets chi.Router) {
				tickets.Get("/", r.handleGateway(gwMux))  // List tickets
				tickets.Post("/", r.handleGateway(gwMux)) // Create ticket
				tickets.Route("/{id}", func(ticket chi.Router) {
					ticket.Get("/", r.handleGateway(gwMux))         // Get ticket
					ticket.Patch("/status", r.handleGateway(gwMux)) // Update status
				})
			})

			// Events
			auth.Route("/events", func(events chi.Router) {
				events.Get("/", r.handleGateway(gwMux))     // List events
				events.Post("/", r.handleGateway(gwMux))    // Create event
				events.Get("/{id}", r.handleGateway(gwMux)) // Get event
			})
		})

		// Rutas públicas
		v1.Group(func(public chi.Router) {
			public.Get("/health", r.handleGateway(gwMux))
			public.Post("/auth/login", r.handleGateway(gwMux))
			public.Post("/users", r.handleGateway(gwMux)) // Register
		})
	})

	// Swagger UI
	if r.config.Environment == "development" {
		r.serveSwagger()
	}

	// Ruta por defecto
	r.router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "not found", "code": "NOT_FOUND"}`))
	})

	return r.router
}

func (r *Router) handleGateway(gwMux *runtime.ServeMux) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		gwMux.ServeHTTP(w, req)
	}
}

func (r *Router) serveSwagger() {
	swaggerHandler := http.FileServer(http.Dir("./gen/swagger"))

	r.router.Get("/swagger/*", func(w http.ResponseWriter, req *http.Request) {
		http.StripPrefix("/swagger/", swaggerHandler).ServeHTTP(w, req)
	})

	r.router.Get("/docs", func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, "/swagger/", http.StatusMovedPermanently)
	})
}

func (r *Router) GetHandler() http.Handler {
	return r.router
}
