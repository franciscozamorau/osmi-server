// osmi/osmi-server/cmd/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	pb "github.com/franciscozamorau/osmi-protobuf/gen/pb"
	handlersgrpc "github.com/franciscozamorau/osmi-server/internal/application/handlers/grpc"
	"github.com/franciscozamorau/osmi-server/internal/application/services"
	"github.com/franciscozamorau/osmi-server/internal/infrastructure/repositories/postgres"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	log.Println("🚀 OSMI Server - ARQUITECTURA COMPLETA")
	log.Println("=======================================")

	_ = godotenv.Load()

	// 1. Conectar PostgreSQL
	db, err := connectPostgreSQL()
	if err != nil {
		log.Fatalf("❌ PostgreSQL: %v", err)
	}
	defer db.Close()
	log.Println("✅ PostgreSQL conectado")

	// 2. Crear TODOS los repositorios
	customerRepo := postgres.NewCustomerRepository(db)
	eventRepo := postgres.NewEventRepository(db)
	userRepo := postgres.NewUserRepository(db)
	categoryRepo := postgres.NewCategoryRepository(db)

	// TicketRepository
	ticketRepo := postgres.NewTicketRepository(db)

	// TicketTypeRepository
	ticketTypeRepo := postgres.NewTicketTypeRepository(db)

	// OrganizerRepository y VenueRepository
	organizerRepo := postgres.NewOrganizerRepository(db)
	venueRepo := postgres.NewVenueRepository(db)

	log.Println("✅ Repositorios creados")

	// 3. Crear TODOS los servicios con sus dependencias correctas
	customerService := services.NewCustomerService(customerRepo)

	// TicketService
	ticketService := services.NewTicketService(
		ticketRepo,
		ticketTypeRepo,
		eventRepo,
		customerRepo,
		nil, // orderRepo (pendiente)
	)

	// NUEVO: TicketTypeService
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
		nil, // sessionRepo (pendiente)
		nil, // hasher (pendiente)
		nil, // jwtService (pendiente)
	)

	categoryService := services.NewCategoryService(categoryRepo, eventRepo)

	log.Println("✅ Servicios creados")

	// 4. Crear TODOS los handlers específicos
	customerHandler := handlersgrpc.NewCustomerHandler(customerService)
	ticketHandler := handlersgrpc.NewTicketHandler(ticketService)
	eventHandler := handlersgrpc.NewEventHandler(eventService)
	userHandler := handlersgrpc.NewUserHandler(userService, "tu-secreto-jwt-aqui")
	categoryHandler := handlersgrpc.NewCategoryHandler(categoryService)
	// NUEVO: TicketTypeHandler
	ticketTypeHandler := handlersgrpc.NewTicketTypeHandler(ticketTypeService)

	log.Println("✅ Handlers específicos creados")

	// 5. Crear handler unificado con TODOS
	handler := handlersgrpc.NewHandler(
		customerHandler,
		ticketHandler,
		userHandler,
		eventHandler,
		categoryHandler,
		ticketTypeHandler, // ← NUEVO: Añadido al final
	)

	log.Println("✅ Handler unificado creado")

	// 6. Iniciar servidor
	startServer(handler, db, ":50051")
}

func connectPostgreSQL() (*sqlx.DB, error) {
	connStr := "host=localhost port=5432 user=osmi password=osmi1405 dbname=osmidb sslmode=disable"

	db, err := sqlx.Connect("pgx", connStr)
	if err != nil {
		return nil, fmt.Errorf("connect: %v", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping: %v", err)
	}

	return db, nil
}

func startServer(handler *handlersgrpc.Handler, db *sqlx.DB, address string) {
	server := grpc.NewServer()

	// Registrar handler unificado (TODOS los métodos)
	pb.RegisterOsmiServiceServer(server, handler)
	reflection.Register(server)

	// Health check HTTP
	go func() {
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			if err := db.PingContext(ctx); err != nil {
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Write([]byte(`{"status":"unhealthy"}`))
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"status":"healthy","service":"osmi-server"}`))
		})

		log.Println("Health check en :8081/health")
		http.ListenAndServe(":8081", nil)
	}()

	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Error escuchando: %v", err)
	}

	log.Printf("gRPC server en %s", address)

	if err := server.Serve(lis); err != nil {
		log.Fatalf("Error sirviendo: %v", err)
	}
}
