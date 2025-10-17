package service

import (
	"context"
	"log"

	pb "osmi-server/gen"
)

type Server struct {
	pb.UnimplementedOsmiServiceServer
}

// CreateUser ya está implementado
func (s *Server) CreateUser(ctx context.Context, req *pb.UserRequest) (*pb.UserResponse, error) {
	log.Printf("CreateUser called: %s (%s)", req.Name, req.Email)

	return &pb.UserResponse{
		UserId: req.UserId,
		Status: "created",
	}, nil
}

// CreateTicket implementa el método gRPC para crear tickets
func (s *Server) CreateTicket(ctx context.Context, req *pb.TicketRequest) (*pb.TicketResponse, error) {
	log.Printf("CreateTicket called: event=%s, user=%s, category=%s", req.EventId, req.UserId, req.CategoryId)

	return &pb.TicketResponse{
		TicketId:  "TICKET-123",
		Status:    "issued",
		Code:      "ABC123",
		QrCodeUrl: "https://osmi.com/qrcode/TICKET-123",
	}, nil
}

// CreateCustomer implementa el método gRPC para crear clientes
func (s *Server) CreateCustomer(ctx context.Context, req *pb.CustomerRequest) (*pb.CustomerResponse, error) {
	log.Printf("CreateCustomer called: %s (%s)", req.Name, req.Email)

	return &pb.CustomerResponse{
		Id:       1,
		Name:     req.Name,
		Email:    req.Email,
		Phone:    req.Phone,
		PublicId: "CUST-001",
	}, nil
}

// GetCustomer implementa el método gRPC para obtener clientes
func (s *Server) GetCustomer(ctx context.Context, req *pb.CustomerLookup) (*pb.CustomerResponse, error) {
	log.Printf("GetCustomer called: ID=%d", req.Id)

	return &pb.CustomerResponse{
		Id:       req.Id,
		Name:     "Francisco Zamora",
		Email:    "fran@osmi.com",
		Phone:    "+52-33-1234-5678",
		PublicId: "CUST-001",
	}, nil
}
