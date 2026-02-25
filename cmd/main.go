package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // Driver para sqlx
	"github.com/jmoiron/sqlx"

	pb "github.com/franciscozamorau/osmi-protobuf/gen/pb"
	handlersgrpc "github.com/franciscozamorau/osmi-server/internal/application/handlers/grpc"
	"github.com/franciscozamorau/osmi-server/internal/application/services"
	"github.com/franciscozamorau/osmi-server/internal/infrastructure/repositories/postgres"
	"github.com/joho/godotenv"
	"google.golang.org/grpc" // Paquete estándar de gRPC, sin alias
	"google.golang.org/grpc/reflection"
)

func main() {
	log.Println("🚀 OSMI Server - ARQUITECTURA REAL")
	log.Println("===================================")

	_ = godotenv.Load()

	// 1. Conectar PostgreSQL con sqlx
	db, err := connectPostgreSQL()
	if err != nil {
		log.Fatalf("❌ PostgreSQL: %v", err)
	}
	defer db.Close()
	log.Println("✅ PostgreSQL conectado")

	// 2. Crear repositorios
	customerRepo := postgres.NewCustomerRepository(db)
	log.Println("✅ Customer Repository creado")

	// 3. Crear servicios
	customerService := services.NewCustomerService(customerRepo)
	log.Println("✅ Customer Service creado")

	// 4. Crear handlers (usando el alias handlersgrpc)
	customerHandler := handlersgrpc.NewCustomerHandler(customerService)
	log.Println("✅ Customer Handler creado")

	// 5. Iniciar servidor
	startServer(customerHandler, db, ":50051")
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

func startServer(customerHandler *handlersgrpc.CustomerHandler, db *sqlx.DB, address string) {
	// Crear servidor gRPC (usando el paquete estándar)
	server := grpc.NewServer()

	// Registrar handlers
	pb.RegisterOsmiServiceServer(server, customerHandler)
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

	// Iniciar gRPC
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Error escuchando: %v", err)
	}

	log.Printf("gRPC server en %s", address)
	log.Println("")
	log.Println("✅ ARQUITECTURA FUNCIONAL:")
	log.Println("   PostgreSQL → Repository → Service → Handler → gRPC")
	log.Println("")
	log.Println("📡 Endpoints:")
	log.Println("   Health:    curl http://localhost:8081/health")
	log.Println("   gRPC:      localhost:50051")
	log.Println("")

	if err := server.Serve(lis); err != nil {
		log.Fatalf("Error sirviendo: %v", err)
	}
}
