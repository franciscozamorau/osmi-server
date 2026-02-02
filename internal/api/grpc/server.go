package grpc

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	pb "github.com/franciscozamorau/osmi-protobuf/gen/pb"
	"github.com/franciscozamorau/osmi-server/internal/application/handlers/grpc"
	"github.com/franciscozamorau/osmi-server/internal/config"
)

type Server struct {
	config      *config.Config
	logger      *zap.Logger
	grpcServer  *grpc.Server
	httpServer  *http.Server
	grpcHandler *grpc.Handler
}

func NewServer(
	cfg *config.Config,
	logger *zap.Logger,
	grpcHandler *grpc.Handler,
) *Server {
	return &Server{
		config:      cfg,
		logger:      logger,
		grpcHandler: grpcHandler,
	}
}

func (s *Server) StartGRPC() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.GRPCPort))
	if err != nil {
		return fmt.Errorf("failed to listen on gRPC port: %w", err)
	}

	s.grpcServer = grpc.NewServer(
		grpc.UnaryInterceptor(s.grpcHandler.UnaryInterceptor()),
		grpc.StreamInterceptor(s.grpcHandler.StreamInterceptor()),
	)

	// Registrar servicios
	pb.RegisterHealthServiceServer(s.grpcServer, s.grpcHandler)
	pb.RegisterTicketServiceServer(s.grpcServer, s.grpcHandler)
	pb.RegisterEventServiceServer(s.grpcServer, s.grpcHandler)
	pb.RegisterUserServiceServer(s.grpcServer, s.grpcHandler)

	// Para desarrollo/testing
	reflection.Register(s.grpcServer)

	s.logger.Info("🚀 gRPC server starting",
		zap.String("address", lis.Addr().String()),
		zap.Int("port", s.config.GRPCPort),
	)

	go func() {
		if err := s.grpcServer.Serve(lis); err != nil {
			s.logger.Fatal("gRPC server failed", zap.Error(err))
		}
	}()

	return nil
}

func (s *Server) StartHTTPGateway() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Crear mux para gRPC-Gateway
	mux := runtime.NewMux(
		runtime.WithErrorHandler(s.customErrorHandler),
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{}),
	)

	// Registrar handlers HTTP
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	err := pb.RegisterHealthServiceHandlerFromEndpoint(ctx, mux,
		fmt.Sprintf("localhost:%d", s.config.GRPCPort), opts)
	if err != nil {
		return fmt.Errorf("failed to register health service: %w", err)
	}

	err = pb.RegisterTicketServiceHandlerFromEndpoint(ctx, mux,
		fmt.Sprintf("localhost:%d", s.config.GRPCPort), opts)
	if err != nil {
		return fmt.Errorf("failed to register ticket service: %w", err)
	}

	err = pb.RegisterEventServiceHandlerFromEndpoint(ctx, mux,
		fmt.Sprintf("localhost:%d", s.config.GRPCPort), opts)
	if err != nil {
		return fmt.Errorf("failed to register event service: %w", err)
	}

	err = pb.RegisterUserServiceHandlerFromEndpoint(ctx, mux,
		fmt.Sprintf("localhost:%d", s.config.GRPCPort), opts)
	if err != nil {
		return fmt.Errorf("failed to register user service: %w", err)
	}

	// Configurar router HTTP con middleware
	router := s.setupRouter(mux)

	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.config.HTTPPort),
		Handler:      router,
		ReadTimeout:  s.config.HTTPReadTimeout,
		WriteTimeout: s.config.HTTPWriteTimeout,
		IdleTimeout:  s.config.HTTPIdleTimeout,
	}

	s.logger.Info("🌐 HTTP Gateway starting",
		zap.String("address", s.httpServer.Addr),
		zap.Int("port", s.config.HTTPPort),
		zap.String("docs", fmt.Sprintf("http://localhost:%d/swagger/", s.config.HTTPPort)),
	)

	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Fatal("HTTP gateway failed", zap.Error(err))
		}
	}()

	return nil
}

func (s *Server) setupRouter(gwMux *runtime.ServeMux) http.Handler {
	// Este método será implementado en el router HTTP
	// Por ahora, retornamos el mux directamente
	return gwMux
}

func (s *Server) customErrorHandler(
	ctx context.Context,
	mux *runtime.ServeMux,
	marshaler runtime.Marshaler,
	w http.ResponseWriter,
	r *http.Request,
	err error,
) {
	// Implementar manejo de errores personalizado
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(runtime.HTTPStatusFromCode(runtime.Code(err)))

	_ = marshaler.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
		"code":  runtime.Code(err).String(),
	})
}

func (s *Server) Shutdown(ctx context.Context) error {
	var errs []error

	if s.httpServer != nil {
		if err := s.httpServer.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("HTTP shutdown error: %w", err))
		}
	}

	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}

	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}

	return nil
}
