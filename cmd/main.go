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
	"github.com/franciscozamorau/osmi-server/internal/repository"
	"github.com/franciscozamorau/osmi-server/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/joho/godotenv"
)

func init() {
	// Cargar variables de entorno
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment variables")
	}
}

// server implementa la interfaz del servicio gRPC
type server struct {
	pb.UnimplementedOsmiServiceServer
	serviceServer *service.Server // ✅ Implementación centralizada del service layer
}

// NewServer crea una nueva instancia del servidor
func NewServer() *server {
	// ✅ VERIFICAR QUE LA CONEXIÓN A BD ESTÉ INICIALIZADA
	if db.Pool == nil {
		log.Fatal("Database connection not initialized. Call db.Init() first")
	}

	// Inicializar repositorios con la conexión a la base de datos
	customerRepo := repository.NewCustomerRepository(db.Pool)
	ticketRepo := repository.NewTicketRepository(db.Pool)
	eventRepo := repository.NewEventRepository(db.Pool)
	userRepo := repository.NewUserRepository(db.Pool)
	categoryRepo := repository.NewCategoryRepository(db.Pool)

	// ✅ VERIFICAR QUE LOS REPOSITORIOS SE CREARON CORRECTAMENTE
	if customerRepo == nil || ticketRepo == nil || eventRepo == nil || userRepo == nil || categoryRepo == nil {
		log.Fatal("Failed to initialize one or more repositories")
	}

	// Crear el service.Server completo para TODOS los métodos
	serviceServer := service.NewServer(
		customerRepo,
		ticketRepo,
		eventRepo,
		userRepo,
		categoryRepo,
	)

	// ✅ VERIFICAR QUE EL SERVICE SERVER SE CREÓ CORRECTAMENTE
	if serviceServer == nil {
		log.Fatal("Failed to initialize service server")
	}

	log.Println("✅ All repositories and service layer initialized successfully")

	return &server{
		serviceServer: serviceServer, // ✅ Única fuente de verdad para la lógica de negocio
	}
}

// =============================================================================
// DELEGACIÓN DE MÉTODOS AL SERVICE LAYER
// =============================================================================

// CreateCategory implementa el método gRPC para crear categorías
func (s *server) CreateCategory(ctx context.Context, req *pb.CategoryRequest) (*pb.CategoryResponse, error) {
	return s.serviceServer.CreateCategory(ctx, req)
}

// GetEventCategories obtiene categorías de un evento
func (s *server) GetEventCategories(ctx context.Context, req *pb.EventLookup) (*pb.CategoryListResponse, error) {
	return s.serviceServer.GetEventCategories(ctx, req)
}

// CreateTicket implementa el método gRPC para crear tickets
func (s *server) CreateTicket(ctx context.Context, req *pb.TicketRequest) (*pb.TicketResponse, error) {
	// ✅ USA LA IMPLEMENTACIÓN CORREGIDA con customer_id obligatorio y user_id opcional
	return s.serviceServer.CreateTicket(ctx, req)
}

// CreateCustomer implementa el método gRPC para crear clientes
func (s *server) CreateCustomer(ctx context.Context, req *pb.CustomerRequest) (*pb.CustomerResponse, error) {
	return s.serviceServer.CreateCustomer(ctx, req)
}

// GetCustomer obtiene un cliente
func (s *server) GetCustomer(ctx context.Context, req *pb.CustomerLookup) (*pb.CustomerResponse, error) {
	return s.serviceServer.GetCustomer(ctx, req)
}

// CreateUser crea un nuevo usuario
func (s *server) CreateUser(ctx context.Context, req *pb.UserRequest) (*pb.UserResponse, error) {
	return s.serviceServer.CreateUser(ctx, req)
}

// CreateEvent crea un nuevo evento
func (s *server) CreateEvent(ctx context.Context, req *pb.EventRequest) (*pb.EventResponse, error) {
	return s.serviceServer.CreateEvent(ctx, req)
}

// GetEvent obtiene un evento
func (s *server) GetEvent(ctx context.Context, req *pb.EventLookup) (*pb.EventResponse, error) {
	return s.serviceServer.GetEvent(ctx, req)
}

// ListEvents lista todos los eventos
func (s *server) ListEvents(ctx context.Context, req *pb.Empty) (*pb.EventListResponse, error) {
	return s.serviceServer.ListEvents(ctx, req)
}

// ListTickets lista tickets por usuario o cliente
func (s *server) ListTickets(ctx context.Context, req *pb.TicketLookup) (*pb.TicketListResponse, error) {
	// ✅ CORREGIDO: Usa la implementación que diferencia entre user_id y customer_id
	return s.serviceServer.ListTickets(ctx, req)
}

// GetTicketDetails obtiene detalles completos de un ticket
func (s *server) GetTicketDetails(ctx context.Context, req *pb.TicketLookup) (*pb.TicketResponse, error) {
	return s.serviceServer.GetTicketDetails(ctx, req)
}

// UpdateTicketStatus actualiza el estado de un ticket
func (s *server) UpdateTicketStatus(ctx context.Context, req *pb.UpdateTicketStatusRequest) (*pb.TicketResponse, error) {
	return s.serviceServer.UpdateTicketStatus(ctx, req)
}

// HealthCheck implementa el health check del servicio
func (s *server) HealthCheck(ctx context.Context, req *pb.Empty) (*pb.HealthResponse, error) {
	return s.serviceServer.HealthCheck(ctx, req)
}

// =============================================================================
// SERVIDOR DE HEALTH CHECKS HTTP
// =============================================================================

// startHealthServer inicia el servidor HTTP para health checks
func startHealthServer(port string) {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		if err := db.HealthCheck(); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, `{"status": "unhealthy", "error": "%s", "timestamp": "%s"}`, 
				err.Error(), time.Now().UTC().Format(time.RFC3339))
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status": "healthy", "timestamp": "%s"}`, 
			time.Now().UTC().Format(time.RFC3339))
	})

	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
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

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, `{"error": "endpoint not found", "path": "%s"}`, r.URL.Path)
	})

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	log.Printf("🩺 Health check server running on port %s", port)
	
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("❌ Health server failed: %v", err)
	}
}

// =============================================================================
// FUNCIÓN PRINCIPAL
// =============================================================================

func main() {
	// ✅ INICIALIZACIÓN ROBUSTA DE LA BASE DE DATOS
	log.Println("🚀 Starting Osmi Ticket System Server...")
	
	if err := db.Init(); err != nil {
		log.Fatalf("❌ Database initialization failed: %v", err)
	}
	defer db.Close()

	log.Println("✅ Database connection established successfully")

	// ✅ INICIAR SERVIDOR DE HEALTH CHECKS EN GOROUTINE SEPARADA
	healthPort := os.Getenv("HEALTH_PORT")
	if healthPort == "" {
		healthPort = "8081"
	}
	go startHealthServer(healthPort)

	// ✅ CONFIGURAR EL LISTENER gRPC
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}

	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("❌ Failed to listen on port %s: %v", grpcPort, err)
	}

	// ✅ CREAR SERVIDOR gRPC CON CONFIGURACIÓN ROBUSTA
	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(16*1024*1024),  // 16MB
		grpc.MaxSendMsgSize(16*1024*1024),  // 16MB
		grpc.ConnectionTimeout(30*time.Second),
	)

	// ✅ CREAR INSTANCIA DEL SERVIDOR
	srv := NewServer()

	// ✅ REGISTRAR SERVICIOS
	pb.RegisterOsmiServiceServer(grpcServer, srv)
	
	// ✅ HABILITAR REFLECTION PARA DESARROLLO (opcional)
	if os.Getenv("ENABLE_REFLECTION") == "true" {
		reflection.Register(grpcServer)
		log.Println("🔍 gRPC reflection enabled")
	}

	// ✅ REGISTRAR SERVICIO DE HEALTH CHECK gRPC
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("osmi.OsmiService", grpc_health_v1.HealthCheckResponse_SERVING)

	log.Printf("📡 Osmi gRPC server running on port %s", grpcPort)

	// ✅ CONFIGURAR GRACEFUL SHUTDOWN
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	shutdownComplete := make(chan bool, 1)

	go func() {
		<-stop
		log.Println("🛑 Shutdown signal received")

		// ✅ CAMBIAR ESTADO DE HEALTH CHECK A NO SERVING
		healthServer.SetServingStatus("osmi.OsmiService", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
		log.Println("📊 Health status set to NOT_SERVING")

		// ✅ DAR TIEMPO PARA QUE LAS SOLICITUDES EN CURSO SE COMPLETEN
		log.Println("⏳ Waiting for ongoing requests to complete...")
		time.Sleep(5 * time.Second)

		// ✅ DETENER SERVIDOR gRPC GRACEFULLY
		log.Println("🛑 Shutting down gRPC server gracefully...")
		grpcServer.GracefulStop()
		
		log.Println("✅ gRPC server stopped successfully")
		shutdownComplete <- true
	}()

	// ✅ INICIAR SERVIDOR gRPC
	log.Println("🎯 Server is ready to accept requests")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("❌ Failed to serve gRPC: %v", err)
	}

	// ✅ ESPERAR A QUE EL SHUTDOWN SE COMPLETE
	<-shutdownComplete
	log.Println("🏁 Server shutdown complete")
}