package service

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/franciscozamorau/osmi-server/gen"
	"github.com/franciscozamorau/osmi-server/internal/models"
	"github.com/franciscozamorau/osmi-server/internal/repository"
	"github.com/jackc/pgx/v5/pgtype"
)

// Server implementa el servicio gRPC
type Server struct {
	pb.UnimplementedOsmiServiceServer
	CustomerRepo *repository.CustomerRepository // ✅ PÚBLICO (mayúscula)
	TicketRepo   *repository.TicketRepository   // ✅ PÚBLICO (mayúscula)
	EventRepo    *repository.EventRepository    // ✅ PÚBLICO (mayúscula)
}

func NewServer(customerRepo *repository.CustomerRepository, ticketRepo *repository.TicketRepository, eventRepo *repository.EventRepository) *Server {
	return &Server{
		CustomerRepo: customerRepo, // ✅ PÚBLICO
		TicketRepo:   ticketRepo,   // ✅ PÚBLICO
		EventRepo:    eventRepo,    // ✅ PÚBLICO
	}
}

// Y LUEGO ACTUALIZA TODOS LOS MÉTODOS para usar los campos públicos:

// CreateEvent implementa el método gRPC para crear eventos
func (s *Server) CreateEvent(ctx context.Context, req *pb.EventRequest) (*pb.EventResponse, error) {
	log.Printf("Creating event: %s", req.Name)

	// Generar public_id temporal
	publicID := fmt.Sprintf("event-%d", time.Now().Unix())

	// Convertir pb.EventRequest a models.Event (RESPETANDO los tipos exactos de tus modelos)
	event := &models.Event{
		// Campos que son STRING en tu modelo:
		PublicID: publicID,
		Name:     req.Name,
		Location: req.Location,

		// Campos que son PGTYPE.TEXT en tu modelo:
		Description:      toPgText(req.Description),
		ShortDescription: toPgText(req.ShortDescription),
		VenueDetails:     toPgText(req.VenueDetails),
		Category:         toPgText(req.Category),
		Tags:             toPgText(req.Tags),
		ImageURL:         toPgText(req.ImageUrl),
		BannerURL:        toPgText(req.BannerUrl),

		// Otros campos:
		IsActive:     req.IsActive,
		IsPublished:  req.IsPublished,
		MaxAttendees: toPgInt4(req.MaxAttendees),
	}

	// Parsear fechas si están presentes
	if req.StartDate != "" {
		if startTime, err := time.Parse(time.RFC3339, req.StartDate); err == nil {
			event.StartDate = startTime
		}
	}
	if req.EndDate != "" {
		if endTime, err := time.Parse(time.RFC3339, req.EndDate); err == nil {
			event.EndDate = endTime
		}
	}

	// Crear evento
	_, err := s.EventRepo.CreateEvent(ctx, event) // ✅ Usar EventRepo (público)
	if err != nil {
		log.Printf("Error creating event: %v", err)
		return nil, err
	}

	// Obtener el evento completo
	createdEvent, err := s.EventRepo.GetEventByPublicID(ctx, publicID) // ✅ Usar EventRepo (público)
	if err != nil {
		log.Printf("Error getting created event: %v", err)
		return nil, err
	}

	// Convertir models.Event a pb.EventResponse (RESPETANDO los tipos exactos)
	return &pb.EventResponse{
		Id:               int32(createdEvent.ID),
		PublicId:         createdEvent.PublicID, // STRING
		Name:             createdEvent.Name,     // STRING
		Description:      pgTextToStr(createdEvent.Description),
		ShortDescription: pgTextToStr(createdEvent.ShortDescription),
		StartDate:        createdEvent.StartDate.Format(time.RFC3339),
		EndDate:          createdEvent.EndDate.Format(time.RFC3339),
		Location:         createdEvent.Location, // STRING
		VenueDetails:     pgTextToStr(createdEvent.VenueDetails),
		Category:         pgTextToStr(createdEvent.Category),
		Tags:             pgTextToStr(createdEvent.Tags),
		IsActive:         createdEvent.IsActive,
		IsPublished:      createdEvent.IsPublished,
		ImageUrl:         pgTextToStr(createdEvent.ImageURL),
		BannerUrl:        pgTextToStr(createdEvent.BannerURL),
		MaxAttendees:     pgInt4ToInt32(createdEvent.MaxAttendees),
	}, nil
}

// GetEvent implementa el método gRPC para obtener eventos
func (s *Server) GetEvent(ctx context.Context, req *pb.EventLookup) (*pb.EventResponse, error) {
	log.Printf("Getting event: %s", req.PublicId)

	event, err := s.EventRepo.GetEventByPublicID(ctx, req.PublicId) // ✅ Usar EventRepo (público)
	if err != nil {
		log.Printf("Error getting event: %v", err)
		return nil, err
	}

	return &pb.EventResponse{
		Id:               int32(event.ID),
		PublicId:         event.PublicID, // STRING
		Name:             event.Name,     // STRING
		Description:      pgTextToStr(event.Description),
		ShortDescription: pgTextToStr(event.ShortDescription),
		StartDate:        event.StartDate.Format(time.RFC3339),
		EndDate:          event.EndDate.Format(time.RFC3339),
		Location:         event.Location, // STRING
		VenueDetails:     pgTextToStr(event.VenueDetails),
		Category:         pgTextToStr(event.Category),
		Tags:             pgTextToStr(event.Tags),
		IsActive:         event.IsActive,
		IsPublished:      event.IsPublished,
		ImageUrl:         pgTextToStr(event.ImageURL),
		BannerUrl:        pgTextToStr(event.BannerURL),
		MaxAttendees:     pgInt4ToInt32(event.MaxAttendees),
	}, nil
}

// ListEvents implementa el método gRPC para listar eventos
func (s *Server) ListEvents(ctx context.Context, req *pb.Empty) (*pb.EventListResponse, error) {
	log.Println("Listing all events")

	events, err := s.EventRepo.ListEvents(ctx) // ✅ Usar EventRepo (público)
	if err != nil {
		log.Printf("Error listing events: %v", err)
		return nil, err
	}

	var pbEvents []*pb.EventResponse
	for _, event := range events {
		pbEvents = append(pbEvents, &pb.EventResponse{
			Id:               int32(event.ID),
			PublicId:         event.PublicID, // STRING
			Name:             event.Name,     // STRING
			Description:      pgTextToStr(event.Description),
			ShortDescription: pgTextToStr(event.ShortDescription),
			StartDate:        event.StartDate.Format(time.RFC3339),
			EndDate:          event.EndDate.Format(time.RFC3339),
			Location:         event.Location, // STRING
			VenueDetails:     pgTextToStr(event.VenueDetails),
			Category:         pgTextToStr(event.Category),
			Tags:             pgTextToStr(event.Tags),
			IsActive:         event.IsActive,
			IsPublished:      event.IsPublished,
			ImageUrl:         pgTextToStr(event.ImageURL),
			BannerUrl:        pgTextToStr(event.BannerURL),
			MaxAttendees:     pgInt4ToInt32(event.MaxAttendees),
		})
	}

	return &pb.EventListResponse{Events: pbEvents}, nil
}

// CreateCustomer implementa el método gRPC para crear clientes
func (s *Server) CreateCustomer(ctx context.Context, req *pb.CustomerRequest) (*pb.CustomerResponse, error) {
	log.Printf("Creating customer: %s, email: %s", req.Name, req.Email)

	id, err := s.CustomerRepo.CreateCustomer(ctx, req.Name, req.Email, req.Phone) // ✅ Usar CustomerRepo (público)
	if err != nil {
		log.Printf("Error creating customer: %v", err)
		return nil, err
	}

	customer, err := s.CustomerRepo.GetCustomerByID(ctx, int(id)) // ✅ Usar CustomerRepo (público)
	if err != nil {
		log.Printf("Error getting created customer: %v", err)
		return nil, err
	}

	return &pb.CustomerResponse{
		Id:       int32(customer.ID),
		Name:     customer.Name,         // STRING
		Email:    customer.Email,        // STRING
		Phone:    customer.Phone.String, // PGTYPE.TEXT
		PublicId: customer.PublicID,     // STRING
	}, nil
}

// GetCustomer implementa el método gRPC para obtener clientes
func (s *Server) GetCustomer(ctx context.Context, req *pb.CustomerLookup) (*pb.CustomerResponse, error) {
	log.Printf("Getting customer with ID: %d", req.Id)

	customer, err := s.CustomerRepo.GetCustomerByID(ctx, int(req.Id)) // ✅ Usar CustomerRepo (público)
	if err != nil {
		log.Printf("Error getting customer: %v", err)
		return nil, err
	}

	return &pb.CustomerResponse{
		Id:       int32(customer.ID),
		Name:     customer.Name,         // STRING
		Email:    customer.Email,        // STRING
		Phone:    customer.Phone.String, // PGTYPE.TEXT
		PublicId: customer.PublicID,     // STRING
	}, nil
}

// CreateUser implementa el método gRPC para crear usuarios
func (s *Server) CreateUser(ctx context.Context, req *pb.UserRequest) (*pb.UserResponse, error) {
	log.Printf("Creating user: %s", req.Name)
	return &pb.UserResponse{
		UserId: req.UserId,
		Status: "active",
	}, nil
}

// CreateTicket implementa el método gRPC para crear tickets
func (s *Server) CreateTicket(ctx context.Context, req *pb.TicketRequest) (*pb.TicketResponse, error) {
	log.Printf("Creating ticket for event: %s, user: %s", req.EventId, req.UserId)

	ticketID, err := s.TicketRepo.CreateTicket(ctx, req) // ✅ Usar TicketRepo (público)
	if err != nil {
		log.Printf("Error creating ticket: %v", err)
		return nil, err
	}

	ticket, err := s.TicketRepo.GetTicketByPublicID(ctx, ticketID) // ✅ Usar TicketRepo (público)
	if err != nil {
		log.Printf("Error getting created ticket: %v", err)
		return nil, err
	}

	return &pb.TicketResponse{
		TicketId:  ticket.PublicID,         // STRING
		Status:    ticket.Status,           // STRING
		Code:      ticket.Code,             // STRING
		QrCodeUrl: ticket.QRCodeURL.String, // PGTYPE.TEXT
	}, nil
}

// ListTickets implementa el método gRPC para listar tickets
func (s *Server) ListTickets(ctx context.Context, req *pb.UserLookup) (*pb.TicketListResponse, error) {
	log.Printf("Listing tickets for user: %s", req.UserId)

	tickets, err := s.TicketRepo.GetTicketsByUserID(ctx, req.UserId) // ✅ Usar TicketRepo (público)
	if err != nil {
		log.Printf("Error listing tickets: %v", err)
		return nil, err
	}

	var pbTickets []*pb.TicketResponse
	for _, ticket := range tickets {
		pbTickets = append(pbTickets, &pb.TicketResponse{
			TicketId:  ticket.PublicID,         // STRING
			Status:    ticket.Status,           // STRING
			Code:      ticket.Code,             // STRING
			QrCodeUrl: ticket.QRCodeURL.String, // PGTYPE.TEXT
		})
	}

	return &pb.TicketListResponse{
		Tickets: pbTickets,
	}, nil
}

// Helper functions (se mantienen igual)
func toPgText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: s != ""}
}

func toPgInt4(i int32) pgtype.Int4 {
	return pgtype.Int4{Int32: i, Valid: true}
}

func pgTextToStr(t pgtype.Text) string {
	if t.Valid {
		return t.String
	}
	return ""
}

func pgInt4ToInt32(i pgtype.Int4) int32 {
	if i.Valid {
		return i.Int32
	}
	return 0
}
