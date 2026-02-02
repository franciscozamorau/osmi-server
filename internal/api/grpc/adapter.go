// ADAPTADOR - Conecta proto generado con tus handlers existentes
// SIN MODIFICAR TUS HANDLERS

package grpc

import (
    "context"
    
    // TUS handlers existentes
    grpchandlers "github.com/franciscozamorau/osmi-server/internal/application/handlers/grpc"
    
    // Proto generado EN EL LUGAR CORRECTO
    pb "github.com/franciscozamorau/osmi-protobuf/gen/pb"
)

// Adaptador convierte llamadas del proto generado a tus handlers
type HandlerAdapter struct {
    pb.UnimplementedOsmiServiceServer
    userHandler     *grpchandlers.UserHandler
    customerHandler *grpchandlers.CustomerHandler
    eventHandler    *grpchandlers.EventHandler
    ticketHandler   *grpchandlers.TicketHandler
    categoryHandler *grpchandlers.CategoryHandler
}

func NewHandlerAdapter(
    userHandler *grpchandlers.UserHandler,
    customerHandler *grpchandlers.CustomerHandler,
    eventHandler *grpchandlers.EventHandler,
    ticketHandler *grpchandlers.TicketHandler,
    categoryHandler *grpchandlers.CategoryHandler,
) *HandlerAdapter {
    return &HandlerAdapter{
        userHandler:     userHandler,
        customerHandler: customerHandler,
        eventHandler:    eventHandler,
        ticketHandler:   ticketHandler,
        categoryHandler: categoryHandler,
    }
}

// Métodos que adaptan las firmas
func (a *HandlerAdapter) HealthCheck(ctx context.Context, req *pb.Empty) (*pb.HealthResponse, error) {
    // Implementación simple
    return &pb.HealthResponse{
        Status:    "healthy",
        Service:   "osmi-server",
        Version:   "1.0.0",
        Timestamp: nil, // TODO: agregar timestamp
    }, nil
}

// Los demás métodos los implementaremos uno por uno
// SIN TOCAR TUS HANDLERS EXISTENTES
