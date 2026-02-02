package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	pb "github.com/franciscozamorau/osmi-protobuf/gen/pb"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string
	DBSSLMode  string
	GRPCPort   string
	HTTPPort   string
}

func loadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: No .env file found, using environment variables")
	}

	return &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBName:     getEnv("DB_NAME", "osmidb"),
		DBUser:     getEnv("DB_USER", "osmi"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
		GRPCPort:   getEnv("GRPC_PORT", "50051"),
		HTTPPort:   getEnv("HTTP_HEALTH_PORT", "8081"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	log.Println("OSMI Server - Iniciando sistema con configuración segura")
	log.Println("==========================================================")

	config := loadConfig()

	if config.DBPassword == "" {
		log.Fatal("Error: DB_PASSWORD no está configurado en variables de entorno")
	}

	dbPool, err := connectPostgreSQL(config)
	if err != nil {
		log.Fatalf("Error conectando a PostgreSQL: %v", err)
	}
	defer dbPool.Close()

	log.Println("PostgreSQL conectado exitosamente")
	startGRPCServer(dbPool, config)
}

func connectPostgreSQL(config *Config) (*pgxpool.Pool, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.DBHost,
		config.DBPort,
		config.DBUser,
		config.DBPassword,
		config.DBName,
		config.DBSSLMode,
	)

	configPool, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("error parse config: %v", err)
	}

	configPool.MaxConns = 25
	configPool.MinConns = 5
	configPool.MaxConnLifetime = time.Hour
	configPool.MaxConnIdleTime = 30 * time.Minute

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, configPool)
	if err != nil {
		return nil, fmt.Errorf("error create pool: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("error ping database: %v", err)
	}

	return pool, nil
}

func startGRPCServer(dbPool *pgxpool.Pool, config *Config) {
	server := grpc.NewServer()
	pb.RegisterOsmiServiceServer(server, &osmiServer{dbPool: dbPool})
	reflection.Register(server)

	go func() {
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			if err := dbPool.Ping(ctx); err != nil {
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Write([]byte(`{"status":"unhealthy","error":"database"}`))
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"status":"healthy","service":"osmi-server"}`))
		})

		log.Printf("Health check en :%s/health", config.HTTPPort)
		if err := http.ListenAndServe(":"+config.HTTPPort, nil); err != nil {
			log.Printf("Error en health server: %v", err)
		}
	}()

	address := ":" + config.GRPCPort
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Error escuchando: %v", err)
	}

	log.Printf("gRPC server en %s", address)
	log.Println("Para probar:")
	log.Printf("  curl http://localhost:%s/health", config.HTTPPort)
	log.Println("  curl http://localhost:8083/health")
	log.Println("  curl -X POST http://localhost:8083/customers \\")
	log.Println("    -H 'Content-Type: application/json' \\")
	log.Println("    -d '{\"name\":\"Test\",\"email\":\"test@test.com\"}'")

	if err := server.Serve(lis); err != nil {
		log.Fatalf("Error en servidor: %v", err)
	}
}

type osmiServer struct {
	pb.UnimplementedOsmiServiceServer
	dbPool *pgxpool.Pool
}

func (s *osmiServer) HealthCheck(ctx context.Context, req *pb.Empty) (*pb.HealthResponse, error) {
	return &pb.HealthResponse{
		Status:  "healthy",
		Service: "osmi-server",
		Version: "1.0",
	}, nil
}

func (s *osmiServer) CreateCustomer(ctx context.Context, req *pb.CustomerRequest) (*pb.CustomerResponse, error) {
	log.Printf("Creando cliente: %s", req.Email)

	query := `INSERT INTO crm.customers (public_uuid, full_name, email, phone, is_active, created_at, updated_at) 
	          VALUES (gen_random_uuid(), $1, $2, $3, true, NOW(), NOW()) RETURNING id, public_uuid`

	var id int32
	var publicID string
	err := s.dbPool.QueryRow(ctx, query, req.Name, req.Email, req.Phone).Scan(&id, &publicID)

	if err != nil {
		log.Printf("Error en query: %v", err)
		return nil, fmt.Errorf("no se pudo crear cliente: %v", err)
	}

	log.Printf("Cliente creado: ID=%d, PublicID=%s", id, publicID)

	return &pb.CustomerResponse{
		Id:       id,
		PublicId: publicID,
		Name:     req.Name,
		Email:    req.Email,
		Phone:    req.Phone,
	}, nil
}

func (s *osmiServer) GetCustomer(ctx context.Context, req *pb.CustomerLookup) (*pb.CustomerResponse, error) {
	return nil, fmt.Errorf("not implemented yet")
}

func (s *osmiServer) CreateUser(ctx context.Context, req *pb.UserRequest) (*pb.UserResponse, error) {
	return nil, fmt.Errorf("not implemented yet")
}

func (s *osmiServer) CreateEvent(ctx context.Context, req *pb.EventRequest) (*pb.EventResponse, error) {
	return nil, fmt.Errorf("not implemented yet")
}

func (s *osmiServer) GetEvent(ctx context.Context, req *pb.EventLookup) (*pb.EventResponse, error) {
	return nil, fmt.Errorf("not implemented yet")
}

func (s *osmiServer) ListEvents(ctx context.Context, req *pb.Empty) (*pb.EventListResponse, error) {
	return nil, fmt.Errorf("not implemented yet")
}

func (s *osmiServer) CreateTicket(ctx context.Context, req *pb.TicketRequest) (*pb.TicketResponse, error) {
	return nil, fmt.Errorf("not implemented yet")
}

func (s *osmiServer) ListTickets(ctx context.Context, req *pb.UserLookup) (*pb.TicketListResponse, error) {
	return nil, fmt.Errorf("not implemented yet")
}

func (s *osmiServer) CreateCategory(ctx context.Context, req *pb.CategoryRequest) (*pb.CategoryResponse, error) {
	return nil, fmt.Errorf("not implemented yet")
}

func (s *osmiServer) GetEventCategories(ctx context.Context, req *pb.EventLookup) (*pb.CategoryListResponse, error) {
	return nil, fmt.Errorf("not implemented yet")
}
