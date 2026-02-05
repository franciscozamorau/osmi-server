package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	pb "github.com/franciscozamorau/osmi-protobuf/gen/pb"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	log.Println("🚀 OSMI Server - FUNCIONANDO")
	log.Println("=============================")

	_ = godotenv.Load()

	dbPool, err := connectPostgreSQL()
	if err != nil {
		log.Fatalf("❌ Error conectando a PostgreSQL: %v", err)
	}
	defer dbPool.Close()

	log.Println("✅ PostgreSQL conectado")
	startGRPCServer(dbPool, ":50051")
}

func connectPostgreSQL() (*pgxpool.Pool, error) {
	connStr := "host=localhost port=5432 user=osmi password=osmi1405 dbname=osmidb sslmode=disable"

	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("error parse config: %v", err)
	}

	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("error create pool: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("error ping database: %v", err)
	}

	return pool, nil
}

func startGRPCServer(dbPool *pgxpool.Pool, address string) {
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

		log.Println("Health check en :8081/health")
		if err := http.ListenAndServe(":8081", nil); err != nil {
			log.Printf("Error en health server: %v", err)
		}
	}()

	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Error escuchando: %v", err)
	}

	log.Printf("gRPC server en %s", address)
	log.Println("")
	log.Println("✅ Sistema funcionando:")
	log.Println("   curl http://localhost:8081/health")
	log.Println("   curl -X POST http://localhost:8083/customers \\")
	log.Println("     -H 'Content-Type: application/json' \\")
	log.Println("     -d '{\"name\":\"Test\",\"email\":\"test@test.com\"}'")

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
