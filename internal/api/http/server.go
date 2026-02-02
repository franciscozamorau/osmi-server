package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/franciscozamorau/osmi-server/internal/api/http/middleware"
	"github.com/franciscozamorau/osmi-server/internal/api/http/routes"
	"github.com/franciscozamorau/osmi-server/internal/shared/logger"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	router     *chi.Mux
	httpServer *http.Server
	config     *ServerConfig
	startTime  time.Time
	version    string
	commit     string
	buildTime  string
}

type ServerConfig struct {
	Host              string
	Port              string
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
	Environment       string
	EnableCORS        bool
	EnableMetrics     bool
	EnableDocs        bool
	EnableProfiling   bool
	TLSCertPath       string
	TLSKeyPath        string
	RateLimitRequests int
	RateLimitDuration time.Duration
	TrustedProxies    []string
	CompressionLevel  int
}

func NewDefaultConfig() *ServerConfig {
	return &ServerConfig{
		Host:              "0.0.0.0",
		Port:              "8080",
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		ShutdownTimeout:   30 * time.Second,
		Environment:       "development",
		EnableCORS:        true,
		EnableMetrics:     true,
		EnableDocs:        true,
		EnableProfiling:   false,
		RateLimitRequests: 100,
		RateLimitDuration: 1 * time.Minute,
		TrustedProxies:    []string{"127.0.0.1"},
		CompressionLevel:  5,
	}
}

func NewServer(config *ServerConfig, handlers *routes.Handlers, version, commit, buildTime string) *Server {
	if config == nil {
		config = NewDefaultConfig()
	}

	router := chi.NewRouter()

	// Middlewares base de Chi
	router.Use(chimiddleware.RequestID)
	router.Use(chimiddleware.RealIP)
	router.Use(chimiddleware.Logger)
	router.Use(chimiddleware.Recoverer)
	router.Use(chimiddleware.Timeout(config.WriteTimeout))
	router.Use(chimiddleware.Heartbeat("/ping"))

	// Middlewares personalizados
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	if config.EnableCORS {
		router.Use(middleware.CORS)
	}

	// Seguridad
	router.Use(middleware.SecurityHeaders)
	router.Use(middleware.RateLimit(config.RateLimitRequests, config.RateLimitDuration))

	// Compression si esta habilitada
	if config.CompressionLevel > 0 {
		router.Use(chimiddleware.Compress(config.CompressionLevel))
	}

	// Configurar rutas
	routes.SetupRoutes(router, handlers, config)

	// Rutas del sistema
	setupSystemRoutes(router, version, commit, buildTime)

	// Crear servidor HTTP
	httpServer := &http.Server{
		Addr:         config.Host + ":" + config.Port,
		Handler:      router,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		IdleTimeout:  config.IdleTimeout,
	}

	return &Server{
		router:     router,
		httpServer: httpServer,
		config:     config,
		startTime:  time.Now(),
		version:    version,
		commit:     commit,
		buildTime:  buildTime,
	}
}

func (s *Server) Start() error {
	addr := s.httpServer.Addr
	logger.Info("Iniciando servidor HTTP",
		logger.Field("address", addr),
		logger.Field("environment", s.config.Environment),
		logger.Field("version", s.version),
		logger.Field("timeouts", fmt.Sprintf("read:%v,write:%v", s.config.ReadTimeout, s.config.WriteTimeout)),
	)

	errChan := make(chan error, 1)

	// Iniciar en goroutine
	go func() {
		var err error
		if s.config.TLSCertPath != "" && s.config.TLSKeyPath != "" {
			logger.Info("Usando TLS")
			err = s.httpServer.ListenAndServeTLS(s.config.TLSCertPath, s.config.TLSKeyPath)
		} else {
			logger.Warn("TLS no configurado - usando HTTP")
			err = s.httpServer.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			logger.Error("Error en servidor HTTP", logger.Field("error", err))
			errChan <- err
		}
	}()

	// Esperar señal o error
	select {
	case err := <-errChan:
		return fmt.Errorf("error en servidor: %w", err)
	case sig := <-setupShutdownSignal():
		logger.Info("Señal recibida", logger.Field("signal", sig.String()))
		return s.Shutdown()
	}
}

func (s *Server) Shutdown() error {
	logger.Info("Apagando servidor HTTP...")

	ctx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		logger.Error("Error durante el apagado", logger.Field("error", err))
		return fmt.Errorf("shutdown error: %w", err)
	}

	logger.Info("Servidor HTTP apagado correctamente")
	return nil
}

func (s *Server) Router() *chi.Mux {
	return s.router
}

func (s *Server) HTTPServer() *http.Server {
	return s.httpServer
}

func setupShutdownSignal() <-chan os.Signal {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	return sigChan
}

type HealthResponse struct {
	Status    string                 `json:"status"`
	Timestamp string                 `json:"timestamp"`
	Service   string                 `json:"service"`
	Version   string                 `json:"version"`
	Uptime    string                 `json:"uptime"`
	GoVersion string                 `json:"go_version"`
	Checks    map[string]HealthCheck `json:"checks,omitempty"`
}

type HealthCheck struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Latency string `json:"latency,omitempty"`
}

func healthCheckHandler(version string, startTime time.Time) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		checks := make(map[string]HealthCheck)

		// Check de memoria
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)

		checks["memory"] = HealthCheck{
			Status:  "healthy",
			Message: fmt.Sprintf("Alloc: %v MB, TotalAlloc: %v MB", memStats.Alloc/1024/1024, memStats.TotalAlloc/1024/1024),
		}

		// Check de goroutines
		checks["goroutines"] = HealthCheck{
			Status:  "healthy",
			Message: fmt.Sprintf("Count: %d", runtime.NumGoroutine()),
		}

		response := HealthResponse{
			Status:    "healthy",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Service:   "osmi-server",
			Version:   version,
			Uptime:    time.Since(startTime).String(),
			GoVersion: runtime.Version(),
			Checks:    checks,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

type ReadinessResponse struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Ready     bool              `json:"ready"`
	Checks    map[string]string `json:"checks,omitempty"`
}

func readinessCheckHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		checks := make(map[string]string)
		allReady := true

		// Aqui puedes agregar checks de:
		// - Base de datos
		// - Redis
		// - Servicios externos
		// - etc.

		checks["api"] = "ready"

		status := http.StatusOK
		if !allReady {
			status = http.StatusServiceUnavailable
		}

		response := ReadinessResponse{
			Status:    "ready",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Ready:     allReady,
			Checks:    checks,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(response)
	}
}

type VersionResponse struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildTime string `json:"build_time"`
	GoVersion string `json:"go_version"`
}

func versionHandler(version, commit, buildTime string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := VersionResponse{
			Version:   version,
			Commit:    commit,
			BuildTime: buildTime,
			GoVersion: runtime.Version(),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

func setupSystemRoutes(router *chi.Mux, version, commit, buildTime string) {
	// Health check
	router.Get("/health", healthCheckHandler(version, time.Now()))

	// Readiness check
	router.Get("/ready", readinessCheckHandler())

	// Version info
	router.Get("/version", versionHandler(version, commit, buildTime))

	// Liveness probe
	router.Get("/live", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Metrics endpoint (si se habilita metrics)
	router.Get("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("# metrics would be here\n"))
	})
}
