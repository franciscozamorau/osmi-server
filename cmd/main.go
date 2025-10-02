// osmi-server/cmd/main.go
package main

import (
	"context"
	"log"
	"net"

	pb "osmi-server/gen"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedOsmiServiceServer
}

func (s *server) CreateTicket(ctx context.Context, req *pb.TicketRequest) (*pb.TicketResponse, error) {
	return &pb.TicketResponse{
		TicketId: "OSMI123",
		Status:   "created",
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterOsmiServiceServer(s, &server{})
	log.Println("Osmi server running on port 50051")
	s.Serve(lis)
}
