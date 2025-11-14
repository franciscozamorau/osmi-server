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

	pb "github.com/franciscozamorau/osmi-server/gen"
	"github.com/franciscozamorau/osmi-server/internal/db"
	"github.com/franciscozamorau/osmi-server/internal/models"
	"github.com/franciscozamorau/osmi-server/internal/repository"
	"github.com/franciscozamorau/osmi-server/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/protobuf/types/known/timestamppb"

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
	eventRepo    *repository.EventRepository
	userRepo     *repository.UserRepository
}

// NewServer crea una nueva instancia del servidor
func NewServer() *server {
	return &server{
		customerRepo: repository.NewCustomerRepository(),
		ticketRepo:   repository.NewTicketRepository(),
		eventRepo:    repository.NewEventRepository(),
		userRepo:     repository.NewUserRepository(),
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

	customerID, err := s.customerRepo.CreateCustomer(ctx, req.Name, req.Email, req.Phone)
	if err != nil {
		log.Printf("Error creating customer: %v", err)
		return nil, err
	}

	customer, err := s.customerRepo.GetCustomerByID(ctx, customerID)
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
	var customer *models.Customer
	var err error

	switch lookup := req.Lookup.(type) {
	case *pb.CustomerLookup_Id:
		log.Printf("Getting customer by ID: %d", lookup.Id)
		if lookup.Id <= 0 {
			return nil, fmt.Errorf("customer ID must be positive")
		}
		customer, err = s.customerRepo.GetCustomerByID(ctx, int64(lookup.Id))

	case *pb.CustomerLookup_PublicId:
		log.Printf("Getting customer by PublicId: %s", lookup.PublicId)
		customer, err = s.customerRepo.GetCustomerByPublicID(ctx, lookup.PublicId)

	case *pb.CustomerLookup_Email:
		log.Printf("Getting customer by Email: %s", lookup.Email)
		customer, err = s.customerRepo.GetCustomerByEmail(ctx, lookup.Email)

	default:
		return nil, fmt.Errorf("no valid lookup parameter provided")
	}

	if err != nil {
		log.Printf("Error getting customer: %v", err)
		return nil, fmt.Errorf("customer not found")
	}

	return &pb.CustomerResponse{
		Id:       int32(customer.ID),
		Name:     customer.Name,
		Email:    customer.Email,
		Phone:    customer.Phone.String,
		PublicId: customer.PublicID,
	}, nil
}

// CreateUser implementa el método gRPC para crear usuarios
func (s *server) CreateUser(ctx context.Context, req *pb.UserRequest) (*pb.UserResponse, error) {
	log.Printf("Creating user: %s, email: %s", req.Name, req.Email)

	// Usar el user repository para crear usuarios
	userID, err := s.userRepo.CreateUser(ctx, req.Name, req.Email, req.Password, req.Role)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		return nil, err
	}

	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		log.Printf("Error getting created user: %v", err)
		return nil, err
	}

	return &pb.UserResponse{
		UserId:    user.PublicID,
		Status:    "active",
		Name:      user.Username,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: timestamppb.New(user.CreatedAt),
	}, nil
}

// CreateEvent implementa el método gRPC para crear eventos
func (s *server) CreateEvent(ctx context.Context, req *pb.EventRequest) (*pb.EventResponse, error) {
	log.Printf("Creating event: %s", req.Name)

	// Usar el service.Server en lugar de implementar aquí
	serviceServer := &service.Server{
		EventRepo: s.eventRepo,
	}
	return serviceServer.CreateEvent(ctx, req)
}

// GetEvent implementa el método gRPC para obtener eventos
func (s *server) GetEvent(ctx context.Context, req *pb.EventLookup) (*pb.EventResponse, error) {
	log.Printf("Getting event: %s", req.PublicId)

	// Usar el service.Server en lugar de implementar aquí
	serviceServer := &service.Server{
		EventRepo: s.eventRepo,
	}
	return serviceServer.GetEvent(ctx, req)
}

// ListEvents implementa el método gRPC para listar eventos
func (s *server) ListEvents(ctx context.Context, req *pb.Empty) (*pb.EventListResponse, error) {
	log.Println("Listing all events")

	// Usar el service.Server en lugar de implementar aquí
	serviceServer := &service.Server{
		EventRepo: s.eventRepo,
	}
	return serviceServer.ListEvents(ctx, req)
}

// ListTickets implementa el método gRPC para listar tickets
func (s *server) ListTickets(ctx context.Context, req *pb.UserLookup) (*pb.TicketListResponse, error) {
	log.Printf("Listing tickets for user: %s", req.UserId)

	tickets, err := s.ticketRepo.GetTicketsByCustomerID(ctx, req.UserId)
	if err != nil {
		log.Printf("Error listing tickets: %v", err)
		return nil, err
	}

	pbTickets := make([]*pb.TicketResponse, 0, len(tickets))
	for _, ticket := range tickets {
		pbTickets = append(pbTickets, &pb.TicketResponse{
			TicketId:  ticket.PublicID,
			Status:    ticket.Status,
			Code:      ticket.Code,
			QrCodeUrl: ticket.QRCodeURL.String,
		})
	}

	return &pb.TicketListResponse{
		Tickets:    pbTickets,
		TotalCount: int32(len(pbTickets)),
	}, nil
}

// HealthCheck implementa el health check gRPC
func (s *server) HealthCheck(ctx context.Context, req *pb.Empty) (*pb.HealthResponse, error) {
	return &pb.HealthResponse{
		Status:  "healthy",
		Service: "osmi-server",
		Version: "1.0.0",
	}, nil
}

// GetEventCategories implementa el método para obtener categorías de evento
func (s *server) GetEventCategories(ctx context.Context, req *pb.EventLookup) (*pb.CategoryListResponse, error) {
	log.Printf("Getting categories for event: %s", req.PublicId)
	return &pb.CategoryListResponse{
		EventName:     "Event Placeholder",
		EventPublicId: req.PublicId,
		Categories:    []*pb.CategoryResponse{},
	}, nil
}

// startHealthServer inicia el servidor HTTP para health checks
func startHealthServer(port string) {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := db.HealthCheck(); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, `{"status": "unhealthy", "error": "%s"}`, err.Error())
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status": "healthy", "timestamp": "%s"}`, time.Now().UTC().Format(time.RFC3339))
	})

	http.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
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
		}`, time.Now().UTC().Format(time.RFC3339), stats.TotalConns(), stats.IdleConns(), stats.MaxConns())
	})

	log.Printf("Health check server running on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Printf("Health server failed: %v", err)
	}
}

func main() {
	// Inicializar base de datos
	if err := db.Init(); err != nil {
		log.Fatalf("DB init failed: %v", err)
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
		log.Fatalf("Failed to listen: %v", err)
	}

	// Crear servidor gRPC
	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(10*1024*1024),
		grpc.MaxSendMsgSize(10*1024*1024),
	)

	// Crear instancia del servidor
	srv := NewServer()

	// Registrar servicio principal
	pb.RegisterOsmiServiceServer(grpcServer, srv)

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

		healthServer.SetServingStatus("osmi.OsmiService", grpc_health_v1.HealthCheckResponse_NOT_SERVING)

		log.Println("Waiting for ongoing requests to complete...")
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
