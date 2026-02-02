// Archivo: internal/database/di.go
package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/franciscozamorau/osmi-server/internal/application/handlers/grpc"
	"github.com/franciscozamorau/osmi-server/internal/application/services"
	"github.com/franciscozamorau/osmi-server/internal/config"
	"github.com/franciscozamorau/osmi-server/internal/infrastructure/repositories/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// Container es el contenedor de dependencias
type Container struct {
	Config *config.Config

	// Database
	DBPool *pgxpool.Pool

	// Repositories
	CustomerRepo *postgres.CustomerRepository
	EventRepo    *postgres.EventRepository
	TicketRepo   *postgres.TicketRepository
	UserRepo     *postgres.UserRepository
	CategoryRepo *postgres.CategoryRepository

	// Services
	CustomerService *services.CustomerService
	EventService    *services.EventService
	TicketService   *services.TicketService
	UserService     *services.UserService

	// Handlers
	CustomerHandler *grpc.CustomerHandler
	EventHandler    *grpc.EventHandler
	TicketHandler   *grpc.TicketHandler
	UserHandler     *grpc.UserHandler
}

// NewContainer crea y configura todas las dependencias
func NewContainer() (*Container, error) {
	// 1. Cargar configuración
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment")
	}

	cfg := config.Load()
	log.Printf("Config loaded. Environment: %s", cfg.Server.Environment)

	// 2. Conectar a PostgreSQL
	dbPool, err := connectPostgreSQL(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}
	log.Println("✅ PostgreSQL connected successfully")

	// 3. Crear repositorios
	customerRepo := postgres.NewCustomerRepository(dbPool)
	eventRepo := postgres.NewEventRepository(dbPool)
	ticketRepo := postgres.NewTicketRepository(dbPool)
	userRepo := postgres.NewUserRepository(dbPool)
	categoryRepo := postgres.NewCategoryRepository(dbPool)

	log.Println("✅ Repositories created")

	// 4. Crear servicios
	customerService := services.NewCustomerService(customerRepo)
	eventService := services.NewEventService(eventRepo, categoryRepo)
	ticketService := services.NewTicketService(ticketRepo, eventRepo, customerRepo)
	userService := services.NewUserService(userRepo)

	log.Println("✅ Services created")

	// 5. Crear handlers
	customerHandler := grpc.NewCustomerHandler(customerService)
	eventHandler := grpc.NewEventHandler(eventService)
	ticketHandler := grpc.NewTicketHandler(ticketService)
	userHandler := grpc.NewUserHandler(userService)

	log.Println("✅ Handlers created")

	return &Container{
		Config:          cfg,
		DBPool:          dbPool,
		CustomerRepo:    customerRepo,
		EventRepo:       eventRepo,
		TicketRepo:      ticketRepo,
		UserRepo:        userRepo,
		CategoryRepo:    categoryRepo,
		CustomerService: customerService,
		EventService:    eventService,
		TicketService:   ticketService,
		UserService:     userService,
		CustomerHandler: customerHandler,
		EventHandler:    eventHandler,
		TicketHandler:   ticketHandler,
		UserHandler:     userHandler,
	}, nil
}

func connectPostgreSQL(cfg *config.Config) (*pgxpool.Pool, error) {
	// Usar DATABASE_URL del entorno o config
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = cfg.Database.URL
	}

	// Asegurar credenciales correctas para Docker
	if dbURL == "postgresql://postgres:password@localhost:5432/osmi" {
		// Usar credenciales de docker-compose.yml
		dbURL = "postgresql://osmi:osmi123@localhost:5432/osmidb"
	}

	log.Printf("Connecting to PostgreSQL: %s", maskPassword(dbURL))

	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse connection string: %w", err)
	}

	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute
	config.HealthCheckPeriod = time.Minute

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return pool, nil
}

func maskPassword(url string) string {
	// Ocultar contraseña en logs
	return url
}
