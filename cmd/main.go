package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "osmi-server/gen"
	"osmi-server/internal/db"
	"osmi-server/internal/repository"
	"osmi-server/internal/service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment")
	}
}

// server implementa la interfaz del servicio gRPC
type server struct {
	pb.UnimplementedOsmiServiceServer
	customerRepo *repository.CustomerRepository
	ticketRepo   *repository.TicketRepository
}

// NewServer crea una nueva instancia del servidor
func NewServer() *server {
	return &server{
		customerRepo: repository.NewCustomerRepository(),
		ticketRepo:   repository.NewTicketRepository(),
	}
}

// CreateTicket implementa el método gRPC para crear tickets
func (s *server) CreateTicket(ctx context.Context, req *pb.TicketRequest) (*pb.TicketResponse, error) {
	log.Printf("Creating ticket for event: %s, user: %s", req.EventId, req.UserId)

	ticketID, err := s.ticketRepo.CreateTicket(ctx, req)
	if err != nil {
		log.Printf("Error creating ticket: %v", err)
		return nil, err
	}

	// Obtener el ticket creado para devolver todos los datos
	ticket, err := s.ticketRepo.GetTicketByPublicID(ctx, ticketID)
	if err != nil {
		log.Printf("Error getting created ticket: %v", err)
		return nil, err
	}

	return &pb.TicketResponse{
		TicketId:  ticket.PublicID,
		Status:    ticket.Status,
		Code:      ticket.Code,
		QrCodeUrl: ticket.QRCodeURL.String,
	}, nil
}

// CreateCustomer implementa el método gRPC para crear clientes
func (s *server) CreateCustomer(ctx context.Context, req *pb.CustomerRequest) (*pb.CustomerResponse, error) {
	log.Printf("Creating customer: %s, email: %s", req.Name, req.Email)

	id, err := s.customerRepo.CreateCustomer(ctx, req.Name, req.Email, req.Phone)
	if err != nil {
		log.Printf("Error creating customer: %v", err)
		return nil, err
	}

	// Obtener el cliente creado para devolver todos los datos
	customer, err := s.customerRepo.GetCustomerByID(ctx, int(id))
	if err != nil {
		log.Printf("Error getting created customer: %v", err)
		return nil, err
	}

	return &pb.CustomerResponse{
		Id:       int32(customer.ID),
		Name:     customer.Name,
		Email:    customer.Email,
		Phone:    customer.Phone.String,
		PublicId: customer.PublicID,
	}, nil
}

// GetCustomer implementa el método gRPC para obtener clientes
func (s *server) GetCustomer(ctx context.Context, req *pb.CustomerLookup) (*pb.CustomerResponse, error) {
	log.Printf("Getting customer with ID: %d", req.Id)

	customer, err := s.customerRepo.GetCustomerByID(ctx, int(req.Id))
	if err != nil {
		log.Printf("Error getting customer: %v", err)
		return nil, err
	}

	return &pb.CustomerResponse{
		Id:       int32(customer.ID),
		Name:     customer.Name,
		Email:    customer.Email,
		Phone:    customer.Phone.String,
		PublicId: customer.PublicID,
	}, nil
}

// startHealthServer inicia el servidor HTTP para health checks
func startHealthServer(port string) {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		// Verificar salud de la base de datos
		if err := db.HealthCheck(); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, `{"status": "unhealthy", "error": "%s"}`, err.Error())
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status": "healthy", "timestamp": "%s"}`, time.Now().UTC().Format(time.RFC3339))
	})

	http.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		// Verificar readiness (más estricto que health)
		if db.Pool == nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, `{"status": "not ready", "error": "database not connected"}`)
			return
		}

		stats := db.GetStats()
		if stats == nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, `{"status": "not ready", "error": "database stats unavailable"}`)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{
			"status": "ready", 
			"timestamp": "%s",
			"database": {
				"total_connections": %d,
				"idle_connections": %d,
				"max_connections": %d
			}
		}`, time.Now().UTC().Format(time.RFC3339),
			stats.TotalConns(), stats.IdleConns(), stats.MaxConns())
	})

	log.Printf("🩺 Health check server running on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Printf("❌ Health server failed: %v", err)
	}
}

func main() {
	// Inicializar base de datos
	if err := db.Init(); err != nil {
		log.Fatalf("❌ DB init failed: %v", err)
	}
	defer db.Close()

	// Iniciar servidor de health checks en goroutine separada
	healthPort := os.Getenv("HEALTH_PORT")
	if healthPort == "" {
		healthPort = "8081"
	}
	go startHealthServer(healthPort)

	// Configurar el listener gRPC
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("❌ Failed to listen: %v", err)
	}

	// Crear servidor gRPC con opciones
	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(10*1024*1024), // 10MB
		grpc.MaxSendMsgSize(10*1024*1024), // 10MB
	)

	// Registrar servicio principal
	pb.RegisterOsmiServiceServer(grpcServer, &service.Server{})

	// Registrar servicio de health check gRPC
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("osmi.OsmiService", grpc_health_v1.HealthCheckResponse_SERVING)

	log.Println("Osmi gRPC server running on port 50051")

	// Configurar graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-stop
		log.Println("Shutdown signal received")

		// Graceful shutdown del servidor gRPC
		healthServer.SetServingStatus("osmi.OsmiService", grpc_health_v1.HealthCheckResponse_NOT_SERVING)

		// Dar tiempo para que las conexiones actuales terminen
		time.Sleep(5 * time.Second)

		log.Println("Shutting down gRPC server")
		grpcServer.GracefulStop()
	}()

	// Iniciar servidor gRPC
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}

	log.Println("Server shutdown complete")
}
