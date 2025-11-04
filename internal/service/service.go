package service

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	pb "github.com/franciscozamorau/osmi-server/gen"
	"github.com/franciscozamorau/osmi-server/internal/models"
	"github.com/franciscozamorau/osmi-server/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// Server implementa el servicio gRPC
type Server struct {
	pb.UnimplementedOsmiServiceServer
	CustomerRepo *repository.CustomerRepository
	TicketRepo   *repository.TicketRepository
	EventRepo    *repository.EventRepository
}

func NewServer(customerRepo *repository.CustomerRepository, ticketRepo *repository.TicketRepository, eventRepo *repository.EventRepository) *Server {
	return &Server{
		CustomerRepo: customerRepo,
		TicketRepo:   ticketRepo,
		EventRepo:    eventRepo,
	}
}

// CreateEvent implementa el método gRPC para crear eventos
func (s *Server) CreateEvent(ctx context.Context, req *pb.EventRequest) (*pb.EventResponse, error) {
	log.Printf("Creating event: %s", req.Name)

	// ✅ PROFESIONAL: Validación exhaustiva de campos de negocio
	if req.Name == "" {
		return nil, fmt.Errorf("event name is required")
	}
	if req.Location == "" {
		return nil, fmt.Errorf("location is required")
	}
	if req.MaxAttendees < 0 {
		return nil, fmt.Errorf("max_attendees must be non-negative")
	}

	// ✅ PROFESIONAL: Generación controlada de UUID
	publicID := uuid.New()
	publicIDStr := publicID.String()

	// ✅ PROFESIONAL: Mapeo estructurado con valores por defecto
	event := &models.Event{
		PublicID:         publicIDStr,
		Name:             req.Name,
		Location:         req.Location,
		Description:      toPgText(req.Description),
		ShortDescription: toPgText(req.ShortDescription),
		VenueDetails:     toPgText(req.VenueDetails),
		Category:         toPgText(req.Category),
		Tags:             toPgText(req.Tags),
		ImageURL:         toPgText(req.ImageUrl),
		BannerURL:        toPgText(req.BannerUrl),
		IsActive:         req.IsActive,
		IsPublished:      req.IsPublished,
		MaxAttendees:     toPgInt4(req.MaxAttendees),
	}

	// ✅ PROFESIONAL: Parseo robusto de fechas con timezone awareness
	event.StartDate, event.EndDate = s.parseEventDates(req.StartDate, req.EndDate)

	// ✅ PROFESIONAL: Auditoría de creación
	log.Printf("Creating event with public_id: %s", publicIDStr)

	// Crear evento - el ID interno se ignora intencionalmente (solo usamos public_id)
	_, err := s.EventRepo.CreateEvent(ctx, event)
	if err != nil {
		log.Printf("Error creating event: %v", err)
		return nil, fmt.Errorf("error inserting event: %v", err)
	}

	// ✅ PROFESIONAL: Obtener evento completo para respuesta (evita data inconsistency)
	createdEvent, err := s.EventRepo.GetEventByPublicID(ctx, publicIDStr)
	if err != nil {
		log.Printf("Error retrieving created event: %v", err)
		return nil, fmt.Errorf("event created but retrieval failed: %v", err)
	}

	// ✅ PROFESIONAL: Auditoría de éxito
	log.Printf("Event created successfully: %s (ID: %d)", createdEvent.PublicID, createdEvent.ID)

	return s.mapEventToResponse(createdEvent), nil
}

// GetEvent implementa el método gRPC para obtener eventos
func (s *Server) GetEvent(ctx context.Context, req *pb.EventLookup) (*pb.EventResponse, error) {
	log.Printf("Getting event: %s", req.PublicId)

	// ✅ PROFESIONAL: Validación estricta de UUID
	if _, err := uuid.Parse(req.PublicId); err != nil {
		return nil, fmt.Errorf("invalid event ID format: must be a valid UUID")
	}

	event, err := s.EventRepo.GetEventByPublicID(ctx, req.PublicId)
	if err != nil {
		log.Printf("Error getting event: %v", err)
		return nil, fmt.Errorf("event not found with id: %s", req.PublicId)
	}

	return s.mapEventToResponse(event), nil
}

// ListEvents implementa el método gRPC para listar eventos
func (s *Server) ListEvents(ctx context.Context, req *pb.Empty) (*pb.EventListResponse, error) {
	log.Println("Listing all events")

	events, err := s.EventRepo.ListEvents(ctx)
	if err != nil {
		log.Printf("Error listing events: %v", err)
		return nil, fmt.Errorf("error retrieving events: %v", err)
	}

	pbEvents := make([]*pb.EventResponse, 0, len(events))
	for _, event := range events {
		pbEvents = append(pbEvents, s.mapEventToResponse(event))
	}

	return &pb.EventListResponse{Events: pbEvents}, nil
}

// CreateCustomer implementa el método gRPC para crear clientes
func (s *Server) CreateCustomer(ctx context.Context, req *pb.CustomerRequest) (*pb.CustomerResponse, error) {
	log.Printf("Creating customer: %s, email: %s", req.Name, req.Email)

	// ✅ PROFESIONAL: Validación de dominio de negocio
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if req.Email == "" {
		return nil, fmt.Errorf("email is required")
	}
	if !isValidEmail(req.Email) {
		return nil, fmt.Errorf("invalid email format")
	}

	// ✅ PROFESIONAL: Validación de formato E.164
	if req.Phone != "" && !isValidE164(req.Phone) {
		return nil, fmt.Errorf("invalid phone format. Use E.164 format: +1234567890")
	}

	// ✅ PROFESIONAL: El repository ahora maneja int64 directamente
	customerID, err := s.CustomerRepo.CreateCustomer(ctx, req.Name, req.Email, req.Phone)
	if err != nil {
		log.Printf("Error creating customer: %v", err)
		return nil, fmt.Errorf("error creating customer: %v", err)
	}

	// ✅ PROFESIONAL: Usando int64 directamente - sin conversiones peligrosas
	customer, err := s.CustomerRepo.GetCustomerByID(ctx, customerID)
	if err != nil {
		log.Printf("Error retrieving created customer: %v", err)
		return nil, fmt.Errorf("customer created but retrieval failed: %v", err)
	}

	log.Printf("Customer created successfully: %s (ID: %d)", customer.Email, customer.ID)

	return &pb.CustomerResponse{
		Id:       int32(customer.ID),
		Name:     customer.Name,
		Email:    customer.Email,
		Phone:    customer.Phone.String,
		PublicId: customer.PublicID,
	}, nil
}

// GetCustomer implementa el método gRPC para obtener clientes
func (s *Server) GetCustomer(ctx context.Context, req *pb.CustomerLookup) (*pb.CustomerResponse, error) {
	log.Printf("Getting customer with ID: %d", req.Id)

	// ✅ PROFESIONAL: Validación de ID de negocio
	if req.Id <= 0 {
		return nil, fmt.Errorf("customer ID must be positive")
	}

	// ✅ PROFESIONAL: Usando int64 directamente
	customer, err := s.CustomerRepo.GetCustomerByID(ctx, int64(req.Id))
	if err != nil {
		log.Printf("Error getting customer: %v", err)
		return nil, fmt.Errorf("customer not found with id: %d", req.Id)
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
func (s *Server) CreateUser(ctx context.Context, req *pb.UserRequest) (*pb.UserResponse, error) {
	log.Printf("Creating user: %s", req.Name)

	// ✅ PROFESIONAL: Validación de dominio
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if req.Email != "" && !isValidEmail(req.Email) {
		return nil, fmt.Errorf("invalid email format")
	}

	// ✅ PROFESIONAL: Usando int64 directamente
	userID, err := s.CustomerRepo.CreateCustomer(ctx, req.Name, req.Email, "")
	if err != nil {
		log.Printf("Error creating user: %v", err)
		return nil, fmt.Errorf("error creating user: %v", err)
	}

	// ✅ PROFESIONAL: Usando int64 directamente
	customer, err := s.CustomerRepo.GetCustomerByID(ctx, userID)
	if err != nil {
		log.Printf("Error retrieving created user: %v", err)
		return nil, fmt.Errorf("user created but retrieval failed: %v", err)
	}

	log.Printf("User created successfully: %s (Customer ID: %d)", customer.Name, customer.ID)

	return &pb.UserResponse{
		UserId: customer.PublicID,
		Status: "active",
	}, nil
}

// CreateTicket implementa el método gRPC para crear tickets
func (s *Server) CreateTicket(ctx context.Context, req *pb.TicketRequest) (*pb.TicketResponse, error) {
	log.Printf("Creating ticket for event: %s, user: %s", req.EventId, req.UserId)

	// ✅ PROFESIONAL: Validación exhaustiva
	if req.EventId == "" {
		return nil, fmt.Errorf("event_id is required")
	}
	if req.UserId == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	// ✅ PROFESIONAL: Validación estricta de UUIDs
	if _, err := uuid.Parse(req.EventId); err != nil {
		return nil, fmt.Errorf("invalid event ID format: must be a valid UUID")
	}
	if _, err := uuid.Parse(req.UserId); err != nil {
		return nil, fmt.Errorf("invalid user ID format: must be a valid UUID")
	}

	ticketPublicID, err := s.TicketRepo.CreateTicket(ctx, req)
	if err != nil {
		log.Printf("Error creating ticket: %v", err)
		return nil, fmt.Errorf("error creating ticket: %v", err)
	}

	ticket, err := s.TicketRepo.GetTicketByPublicID(ctx, ticketPublicID)
	if err != nil {
		log.Printf("Error retrieving created ticket: %v", err)
		return nil, fmt.Errorf("ticket created but retrieval failed: %v", err)
	}

	log.Printf("Ticket created successfully: %s", ticket.PublicID)

	return &pb.TicketResponse{
		TicketId:  ticket.PublicID,
		Status:    ticket.Status,
		Code:      ticket.Code,
		QrCodeUrl: ticket.QRCodeURL.String,
	}, nil
}

// ListTickets implementa el método gRPC para listar tickets
func (s *Server) ListTickets(ctx context.Context, req *pb.UserLookup) (*pb.TicketListResponse, error) {
	log.Printf("Listing tickets for user: %s", req.UserId)

	// ✅ PROFESIONAL: Validación estricta
	if _, err := uuid.Parse(req.UserId); err != nil {
		return nil, fmt.Errorf("invalid user ID format: must be a valid UUID")
	}

	tickets, err := s.TicketRepo.GetTicketsByUserID(ctx, req.UserId)
	if err != nil {
		log.Printf("Error listing tickets: %v", err)
		return nil, fmt.Errorf("error querying tickets by user: %v", err)
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

	log.Printf("Found %d tickets for user: %s", len(pbTickets), req.UserId)

	return &pb.TicketListResponse{
		Tickets: pbTickets,
	}, nil
}

// ✅ PROFESIONAL: Helper methods para mejor organización
func (s *Server) parseEventDates(startDateStr, endDateStr string) (time.Time, time.Time) {
	var startDate, endDate time.Time
	now := time.Now()

	if startDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			startDate = parsed
		} else {
			startDate = now
		}
	} else {
		startDate = now
	}

	if endDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			endDate = parsed
		} else {
			endDate = startDate.Add(2 * time.Hour)
		}
	} else {
		endDate = startDate.Add(2 * time.Hour)
	}

	// ✅ PROFESIONAL: Validación de lógica de negocio
	if endDate.Before(startDate) {
		endDate = startDate.Add(2 * time.Hour)
	}

	return startDate, endDate
}

func (s *Server) mapEventToResponse(event *models.Event) *pb.EventResponse {
	return &pb.EventResponse{
		Id:               int32(event.ID),
		PublicId:         event.PublicID,
		Name:             event.Name,
		Description:      pgTextToStr(event.Description),
		ShortDescription: pgTextToStr(event.ShortDescription),
		StartDate:        event.StartDate.Format(time.RFC3339),
		EndDate:          event.EndDate.Format(time.RFC3339),
		Location:         event.Location,
		VenueDetails:     pgTextToStr(event.VenueDetails),
		Category:         pgTextToStr(event.Category),
		Tags:             pgTextToStr(event.Tags),
		IsActive:         event.IsActive,
		IsPublished:      event.IsPublished,
		ImageUrl:         pgTextToStr(event.ImageURL),
		BannerUrl:        pgTextToStr(event.BannerURL),
		MaxAttendees:     pgInt4ToInt32(event.MaxAttendees),
	}
}

// Helper functions
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

// ✅ PROFESIONAL: Validaciones robustas
func isValidE164(phone string) bool {
	e164Regex := `^\+[1-9]\d{1,14}$`
	matched, _ := regexp.MatchString(e164Regex, phone)
	return matched
}

func isValidEmail(email string) bool {
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(emailRegex, email)
	return matched
}
