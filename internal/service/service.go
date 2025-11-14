package service

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	pb "github.com/franciscozamorau/osmi-server/gen"
	"github.com/franciscozamorau/osmi-server/internal/models"
	"github.com/franciscozamorau/osmi-server/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Server implementa el servicio gRPC
type Server struct {
	pb.UnimplementedOsmiServiceServer
	CustomerRepo *repository.CustomerRepository
	TicketRepo   *repository.TicketRepository
	EventRepo    *repository.EventRepository
	UserRepo     *repository.UserRepository // NUEVO: Repositorio de usuarios
}

func NewServer(customerRepo *repository.CustomerRepository, ticketRepo *repository.TicketRepository, eventRepo *repository.EventRepository, userRepo *repository.UserRepository) *Server {
	return &Server{
		CustomerRepo: customerRepo,
		TicketRepo:   ticketRepo,
		EventRepo:    eventRepo,
		UserRepo:     userRepo,
	}
}

// CreateEvent - COMPLETAMENTE CORREGIDA para el nuevo proto
func (s *Server) CreateEvent(ctx context.Context, req *pb.EventRequest) (*pb.EventResponse, error) {
	log.Printf("Creating event: %s", req.Name)

	// Validaciones básicas
	if strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("event name is required")
	}
	if strings.TrimSpace(req.Location) == "" {
		return nil, fmt.Errorf("location is required")
	}
	if strings.TrimSpace(req.StartDate) == "" {
		return nil, fmt.Errorf("start_date is required")
	}
	if strings.TrimSpace(req.EndDate) == "" {
		return nil, fmt.Errorf("end_date is required")
	}

	// Parsear fechas desde string (RFC3339)
	startTime, err := time.Parse(time.RFC3339, req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start_date format: %v", err)
	}

	endTime, err := time.Parse(time.RFC3339, req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end_date format: %v", err)
	}

	// Validar lógica de fechas
	if endTime.Before(startTime) {
		return nil, fmt.Errorf("end_date cannot be before start_date")
	}
	if startTime.Before(time.Now()) {
		return nil, fmt.Errorf("start_date cannot be in the past")
	}

	// Generar UUID para el evento
	publicID := uuid.New().String()

	// Mapear request a modelo Event
	event := &models.Event{
		PublicID:         publicID,
		Name:             strings.TrimSpace(req.Name),
		Description:      toPgText(req.Description),
		ShortDescription: toPgText(req.ShortDescription),
		StartDate:        startTime,
		EndDate:          endTime,
		Location:         strings.TrimSpace(req.Location),
		VenueDetails:     toPgText(req.VenueDetails),
		Category:         toPgText(req.Category),
		Tags:             req.Tags, // CORREGIDO: Usar array directamente
		IsActive:         req.IsActive,
		IsPublished:      req.IsPublished,
		ImageURL:         toPgText(req.ImageUrl),
		BannerURL:        toPgText(req.BannerUrl),
		MaxAttendees:     toPgInt4(req.MaxAttendees),
	}

	// Crear evento en la base de datos
	eventID, err := s.EventRepo.CreateEvent(ctx, event)
	if err != nil {
		log.Printf("Error creating event: %v", err)
		return nil, fmt.Errorf("error creating event: %v", err)
	}

	log.Printf("Event created successfully: %s (ID: %d, PublicID: %s)", req.Name, eventID, publicID)

	// Construir respuesta
	return &pb.EventResponse{
		PublicId:         publicID,
		Name:             event.Name,
		Description:      req.Description,
		ShortDescription: req.ShortDescription,
		StartDate:        req.StartDate, // Mantener como string
		EndDate:          req.EndDate,   // Mantener como string
		Location:         event.Location,
		VenueDetails:     req.VenueDetails,
		Category:         req.Category,
		Tags:             req.Tags,
		IsActive:         event.IsActive,
		IsPublished:      event.IsPublished,
		ImageUrl:         req.ImageUrl,
		BannerUrl:        req.BannerUrl,
		MaxAttendees:     req.MaxAttendees,
		CreatedAt:        timestamppb.Now(),
		UpdatedAt:        timestamppb.Now(),
	}, nil
}

// GetEvent implementa el método gRPC para obtener eventos
func (s *Server) GetEvent(ctx context.Context, req *pb.EventLookup) (*pb.EventResponse, error) {
	log.Printf("Getting event: %s", req.PublicId)

	// Validar que el public_id sea UUID válido
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

	log.Printf("Retrieved %d active events", len(pbEvents))

	return &pb.EventListResponse{
		Events:     pbEvents,
		TotalCount: int32(len(pbEvents)),
	}, nil
}

// CreateCustomer implementa el método gRPC para crear clientes
func (s *Server) CreateCustomer(ctx context.Context, req *pb.CustomerRequest) (*pb.CustomerResponse, error) {
	log.Printf("Creating customer: %s, email: %s", req.Name, req.Email)

	// Validaciones
	if strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("name is required")
	}
	if strings.TrimSpace(req.Email) == "" {
		return nil, fmt.Errorf("email is required")
	}
	if !isValidEmail(strings.TrimSpace(req.Email)) {
		return nil, fmt.Errorf("invalid email format")
	}

	// Validar formato de teléfono
	phone := strings.TrimSpace(req.Phone)
	if phone != "" && !isValidPhone(phone) {
		return nil, fmt.Errorf("invalid phone format. Use E.164 format: +1234567890 or standard format")
	}

	// Crear cliente
	customerID, err := s.CustomerRepo.CreateCustomer(ctx, strings.TrimSpace(req.Name), strings.TrimSpace(req.Email), phone)
	if err != nil {
		log.Printf("Error creating customer: %v", err)
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "already exists") {
			return nil, fmt.Errorf("customer with email %s already exists", strings.TrimSpace(req.Email))
		}
		return nil, fmt.Errorf("error creating customer: %v", err)
	}

	// Obtener el cliente creado
	customer, err := s.CustomerRepo.GetCustomerByID(ctx, customerID)
	if err != nil {
		log.Printf("Error retrieving created customer: %v", err)
		return nil, fmt.Errorf("customer created but retrieval failed: %v", err)
	}

	log.Printf("Customer created successfully: %s (ID: %d, PublicID: %s)", customer.Email, customer.ID, customer.PublicID)

	return &pb.CustomerResponse{
		Id:        int32(customer.ID),
		PublicId:  customer.PublicID,
		Name:      customer.Name,
		Email:     customer.Email,
		Phone:     customer.Phone.String,
		CreatedAt: timestamppb.New(customer.CreatedAt),
		UpdatedAt: timestamppb.New(customer.UpdatedAt),
	}, nil
}

// GetCustomer - CORREGIDA para manejar oneof lookup
func (s *Server) GetCustomer(ctx context.Context, req *pb.CustomerLookup) (*pb.CustomerResponse, error) {
	var customer *models.Customer
	var err error

	// Manejar el oneof lookup CORRECTAMENTE
	switch lookup := req.Lookup.(type) {
	case *pb.CustomerLookup_Id:
		log.Printf("Getting customer by ID: %d", lookup.Id)
		if lookup.Id <= 0 {
			return nil, fmt.Errorf("customer ID must be positive")
		}
		customer, err = s.CustomerRepo.GetCustomerByID(ctx, int64(lookup.Id))

	case *pb.CustomerLookup_PublicId:
		log.Printf("Getting customer by PublicId: %s", lookup.PublicId)
		if _, err := uuid.Parse(lookup.PublicId); err != nil {
			return nil, fmt.Errorf("invalid public_id format: must be a valid UUID")
		}
		customer, err = s.CustomerRepo.GetCustomerByPublicID(ctx, lookup.PublicId)

	case *pb.CustomerLookup_Email:
		log.Printf("Getting customer by Email: %s", lookup.Email)
		if strings.TrimSpace(lookup.Email) == "" {
			return nil, fmt.Errorf("email cannot be empty")
		}
		customer, err = s.CustomerRepo.GetCustomerByEmail(ctx, lookup.Email)

	default:
		return nil, fmt.Errorf("no valid lookup parameter provided")
	}

	if err != nil {
		log.Printf("Error getting customer: %v", err)
		return nil, fmt.Errorf("customer not found")
	}

	return &pb.CustomerResponse{
		Id:        int32(customer.ID),
		PublicId:  customer.PublicID,
		Name:      customer.Name,
		Email:     customer.Email,
		Phone:     customer.Phone.String,
		CreatedAt: timestamppb.New(customer.CreatedAt),
		UpdatedAt: timestamppb.New(customer.UpdatedAt),
	}, nil
}

// CreateUser - COMPLETAMENTE REESCRITA para usar tabla users
func (s *Server) CreateUser(ctx context.Context, req *pb.UserRequest) (*pb.UserResponse, error) {
	log.Printf("Creating user: %s, email: %s", req.Name, req.Email)

	// Validaciones
	if strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("name is required")
	}
	if strings.TrimSpace(req.Email) == "" {
		return nil, fmt.Errorf("email is required")
	}
	if !isValidEmail(strings.TrimSpace(req.Email)) {
		return nil, fmt.Errorf("invalid email format")
	}
	if strings.TrimSpace(req.Password) == "" {
		return nil, fmt.Errorf("password is required")
	}

	// Validar role
	role := strings.TrimSpace(req.Role)
	if role == "" {
		role = "customer"
	}
	if role != "customer" && role != "organizer" && role != "admin" {
		return nil, fmt.Errorf("invalid role. Must be: customer, organizer, or admin")
	}

	// Crear usuario en la base de datos
	userID, err := s.UserRepo.CreateUser(ctx, strings.TrimSpace(req.Name), strings.TrimSpace(req.Email), strings.TrimSpace(req.Password), role)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "already exists") {
			return nil, fmt.Errorf("user with email %s already exists", strings.TrimSpace(req.Email))
		}
		return nil, fmt.Errorf("error creating user: %v", err)
	}

	// Obtener el usuario creado
	user, err := s.UserRepo.GetUserByID(ctx, userID)
	if err != nil {
		log.Printf("Error retrieving created user: %v", err)
		return nil, fmt.Errorf("user created but retrieval failed: %v", err)
	}

	log.Printf("User created successfully: %s (ID: %d, PublicID: %s)", user.Email, user.ID, user.PublicID)

	return &pb.UserResponse{
		UserId:    user.PublicID,
		Status:    "active",
		Name:      user.Username,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: timestamppb.New(user.CreatedAt),
	}, nil
}

// CreateTicket implementa el método gRPC para crear tickets
func (s *Server) CreateTicket(ctx context.Context, req *pb.TicketRequest) (*pb.TicketResponse, error) {
	log.Printf("Creating ticket for event: %s, user: %s, category: %s, quantity: %d",
		req.EventId, req.UserId, req.CategoryId, req.Quantity)

	// Validaciones
	if strings.TrimSpace(req.EventId) == "" {
		return nil, fmt.Errorf("event_id is required")
	}
	if strings.TrimSpace(req.UserId) == "" {
		return nil, fmt.Errorf("user_id is required")
	}
	if strings.TrimSpace(req.CategoryId) == "" {
		return nil, fmt.Errorf("category_id is required")
	}
	if req.Quantity <= 0 {
		req.Quantity = 1 // Default value
	}

	// Validar UUIDs
	if _, err := uuid.Parse(strings.TrimSpace(req.EventId)); err != nil {
		return nil, fmt.Errorf("invalid event ID format: must be a valid UUID")
	}
	if _, err := uuid.Parse(strings.TrimSpace(req.UserId)); err != nil {
		return nil, fmt.Errorf("invalid user ID format: must be a valid UUID")
	}
	if _, err := uuid.Parse(strings.TrimSpace(req.CategoryId)); err != nil {
		return nil, fmt.Errorf("invalid category ID format: must be a valid UUID")
	}

	// Crear ticket
	ticketPublicID, err := s.TicketRepo.CreateTicket(ctx, req)
	if err != nil {
		log.Printf("Error creating ticket: %v", err)
		return nil, fmt.Errorf("error creating ticket: %v", err)
	}

	// Obtener el ticket creado
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

	// Validar UUID
	if _, err := uuid.Parse(req.UserId); err != nil {
		return nil, fmt.Errorf("invalid user ID format: must be a valid UUID")
	}

	// Obtener tickets del usuario
	tickets, err := s.TicketRepo.GetTicketsByCustomerID(ctx, req.UserId)
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
		Tickets:    pbTickets,
		TotalCount: int32(len(pbTickets)),
	}, nil
}

// HealthCheck implementa el health check
func (s *Server) HealthCheck(ctx context.Context, req *pb.Empty) (*pb.HealthResponse, error) {
	return &pb.HealthResponse{
		Status:    "healthy",
		Service:   "osmi-server",
		Version:   "1.0.0",
		Timestamp: timestamppb.Now(),
	}, nil
}

// GetEventCategories implementa el método para obtener categorías de evento
func (s *Server) GetEventCategories(ctx context.Context, req *pb.EventLookup) (*pb.CategoryListResponse, error) {
	log.Printf("Getting categories for event: %s", req.PublicId)
	return &pb.CategoryListResponse{
		EventName:     "Event Placeholder",
		EventPublicId: req.PublicId,
		Categories:    []*pb.CategoryResponse{},
	}, nil
}

// mapEventToResponse - CORREGIDA para el nuevo proto
func (s *Server) mapEventToResponse(event *models.Event) *pb.EventResponse {
	return &pb.EventResponse{
		PublicId:         event.PublicID,
		Name:             event.Name,
		Description:      pgTextToStr(event.Description),
		ShortDescription: pgTextToStr(event.ShortDescription),
		StartDate:        event.StartDate.Format(time.RFC3339), // Convertir a string RFC3339
		EndDate:          event.EndDate.Format(time.RFC3339),   // Convertir a string RFC3339
		Location:         event.Location,
		VenueDetails:     pgTextToStr(event.VenueDetails),
		Category:         pgTextToStr(event.Category),
		Tags:             event.Tags, // Usar array directamente
		IsActive:         event.IsActive,
		IsPublished:      event.IsPublished,
		ImageUrl:         pgTextToStr(event.ImageURL),
		BannerUrl:        pgTextToStr(event.BannerURL),
		MaxAttendees:     pgInt4ToInt32(event.MaxAttendees),
		CreatedAt:        timestamppb.New(event.CreatedAt),
		UpdatedAt:        timestamppb.New(event.UpdatedAt),
	}
}

// Helper functions
func toPgText(s string) pgtype.Text {
	return pgtype.Text{String: strings.TrimSpace(s), Valid: s != ""}
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

// Validaciones
func isValidPhone(phone string) bool {
	e164Regex := `^\+[1-9]\d{1,14}$`
	nationalRegex := `^[\d\s\(\)\.\-]+$`

	phone = strings.TrimSpace(phone)
	if phone == "" {
		return true
	}

	matchedE164, _ := regexp.MatchString(e164Regex, phone)
	matchedNational, _ := regexp.MatchString(nationalRegex, phone)

	digits := regexp.MustCompile(`\d`).FindAllString(phone, -1)

	return (matchedE164 || matchedNational) && len(digits) >= 6
}

func isValidEmail(email string) bool {
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(emailRegex, email)
	return matched
}
