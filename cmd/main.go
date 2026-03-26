// osmi/osmi-server/cmd/main.go
package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // Mantenemos esto para el driver

	pb "github.com/franciscozamorau/osmi-protobuf/gen/pb"
	handlersgrpc "github.com/franciscozamorau/osmi-server/internal/application/handlers/grpc"
	"github.com/franciscozamorau/osmi-server/internal/application/services"
	"github.com/franciscozamorau/osmi-server/internal/config"
	"github.com/franciscozamorau/osmi-server/internal/database"
	"github.com/franciscozamorau/osmi-server/internal/infrastructure/repositories/postgres"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	log.Println("🚀 OSMI Server - ARQUITECTURA COMPLETA")
	log.Println("=======================================")

	// Cargar configuración
	cfg := config.Load()
	_ = godotenv.Load()

	// Inicializar base de datos con pgxpool
	if err := database.Init(); err != nil {
		log.Fatalf("❌ Failed to initialize database pool: %v", err)
	}
	defer database.Close()

	// ================================================
	// REPOSITORIOS - TODOS DEBEN ESTAR MIGRADOS A pgxpool
	// ================================================

	// Estos constructores DEBEN aceptar *pgxpool.Pool
	customerRepo := postgres.NewCustomerRepository(database.Pool)
	eventRepo := postgres.NewEventRepository(database.Pool)
	userRepo := postgres.NewUserRepository(database.Pool)
	categoryRepo := postgres.NewCategoryRepository(database.Pool)
	ticketRepo := postgres.NewTicketRepository(database.Pool)
	ticketTypeRepo := postgres.NewTicketTypeRepository(database.Pool)
	organizerRepo := postgres.NewOrganizerRepository(database.Pool)
	venueRepo := postgres.NewVenueRepository(database.Pool)

	log.Println("✅ Repositorios creados (todos con pgxpool)")

	// Crear servicios
	customerService := services.NewCustomerService(customerRepo)
	ticketService := services.NewTicketService(
		ticketRepo,
		ticketTypeRepo,
		eventRepo,
		customerRepo,
		nil,
	)
	ticketTypeService := services.NewTicketTypeService(ticketTypeRepo, eventRepo)
	eventService := services.NewEventService(
		eventRepo,
		organizerRepo,
		venueRepo,
		categoryRepo,
		ticketTypeRepo,
	)
	userService := services.NewUserService(
		userRepo,
		customerRepo,
		nil, nil, nil,
	)
	categoryService := services.NewCategoryService(categoryRepo, eventRepo)

	log.Println("✅ Servicios creados")

	// Crear handlers
	customerHandler := handlersgrpc.NewCustomerHandler(customerService)
	ticketHandler := handlersgrpc.NewTicketHandler(ticketService)
	eventHandler := handlersgrpc.NewEventHandler(eventService)
	userHandler := handlersgrpc.NewUserHandler(userService, "tu-secreto-jwt-aqui")
	categoryHandler := handlersgrpc.NewCategoryHandler(categoryService)
	ticketTypeHandler := handlersgrpc.NewTicketTypeHandler(ticketTypeService)

	log.Println("✅ Handlers específicos creados")

	// Handler unificado
	handler := handlersgrpc.NewHandler(
		customerHandler,
		ticketHandler,
		userHandler,
		eventHandler,
		categoryHandler,
		ticketTypeHandler,
	)

	log.Println("✅ Handler unificado creado")

	// Iniciar servidor gRPC
	startServer(handler, cfg.GRPCPort)
}

func startServer(handler *handlersgrpc.Handler, port string) {
	address := ":" + port
	server := grpc.NewServer()

	pb.RegisterOsmiServiceServer(server, handler)
	reflection.Register(server)

	// Health check HTTP
	go func() {
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			if err := database.Pool.Ping(ctx); err != nil {
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Write([]byte(`{"status":"unhealthy"}`))
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"status":"healthy","service":"osmi-server"}`))
		})

		log.Printf("Health check en :%s/health", "8081")
		http.ListenAndServe(":8081", nil)
	}()

	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("❌ Error escuchando: %v", err)
	}

	log.Printf("🚀 gRPC server en %s", address)

	if err := server.Serve(lis); err != nil {
		log.Fatalf("❌ Error sirviendo: %v", err)
	}
}
