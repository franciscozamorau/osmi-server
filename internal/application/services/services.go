// internal/application/services/services.go
package services

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	osmi "github.com/franciscozamorau/osmi-protobuf/gen/pb"
	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
	"github.com/franciscozamorau/osmi-server/internal/domain/enums"
	"github.com/franciscozamorau/osmi-server/internal/domain/repository"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// isDateRangeValid valida que endDate sea después de startDate
func isDateRangeValid(start, end time.Time) bool {
	return !end.Before(start)
}

// =============================================================================
// FUNCIONES HELPER
// =============================================================================

// isValidUUID valida si un string es un UUID válido
func isValidUUID(u string) bool {
	if u == "" {
		return false
	}
	// UUID v4 pattern
	pattern := `^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`
	match, _ := regexp.MatchString(pattern, strings.ToLower(u))
	return match
}

// isValidEmail valida si un string es un email válido
func isValidEmail(email string) bool {
	if email == "" {
		return false
	}
	// Email pattern básico
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	match, _ := regexp.MatchString(pattern, email)
	return match
}

// isValidPhone valida si un string es un teléfono válido (E.164 o formato básico)
func isValidPhone(phone string) bool {
	if phone == "" {
		return true // Teléfono opcional
	}
	// E.164: +1234567890 o formato básico: 1234567890
	pattern := `^\+?[0-9]{8,15}$`
	match, _ := regexp.MatchString(pattern, phone)
	return match
}

// truncateString trunca un string para logging seguro (evita strings enormes)
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// Server implementa el servicio gRPC
type Server struct {
	osmi.UnimplementedOsmiServiceServer
	CustomerRepo repository.CustomerRepository
	TicketRepo   repository.TicketRepository
	EventRepo    repository.EventRepository
	UserRepo     repository.UserRepository
	CategoryRepo repository.CategoryRepository
}

// NewServer crea una nueva instancia del servidor
func NewServer(
	customerRepo repository.CustomerRepository,
	ticketRepo repository.TicketRepository,
	eventRepo repository.EventRepository,
	userRepo repository.UserRepository,
	categoryRepo repository.CategoryRepository,
) *Server {
	return &Server{
		CustomerRepo: customerRepo,
		TicketRepo:   ticketRepo,
		EventRepo:    eventRepo,
		UserRepo:     userRepo,
		CategoryRepo: categoryRepo,
	}
}

// ============================================================================
// MÉTODOS DE CATEGORÍAS
// ============================================================================

// CreateCategory implementa el método gRPC para crear categorías
func (s *Server) CreateCategory(ctx context.Context, req *osmi.CreateCategoryRequest) (*osmi.CategoryResponse, error) {
	log.Printf("Creating category for event: %s, name: %s", req.EventId, req.Name)

	// Validaciones básicas
	if strings.TrimSpace(req.EventId) == "" {
		return nil, fmt.Errorf("event_id is required")
	}
	if strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("name is required")
	}

	// Obtener el evento
	event, err := s.EventRepo.GetByPublicID(ctx, req.EventId)
	if err != nil {
		return nil, fmt.Errorf("event not found: %w", err)
	}

	// Generar slug a partir del nombre
	slug := strings.ToLower(strings.ReplaceAll(req.Name, " ", "-"))

	// Crear categoría
	category := &entities.Category{
		PublicID:         uuid.New().String(),
		Name:             strings.TrimSpace(req.Name),
		Slug:             slug,
		Description:      &req.Description,
		Icon:             nil,
		ColorHex:         "#3498db",
		ParentID:         nil,
		Level:            1,
		Path:             "",
		TotalEvents:      0,
		TotalTicketsSold: 0,
		TotalRevenue:     0,
		IsActive:         req.IsActive,
		IsFeatured:       false,
		SortOrder:        0,
		MetaTitle:        nil,
		MetaDescription:  nil,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Guardar categoría
	err = s.CategoryRepo.Create(ctx, category)
	if err != nil {
		log.Printf("Error creating category: %v", err)
		return nil, fmt.Errorf("error creating category: %w", err)
	}

	// Asociar con el evento
	err = s.CategoryRepo.AddEventToCategory(ctx, event.ID, category.ID, true)
	if err != nil {
		log.Printf("Warning: could not associate category with event: %v", err)
	}

	log.Printf("Category created successfully: %s", category.PublicID)

	// Construir respuesta
	response := &osmi.CategoryResponse{
		PublicId:           category.PublicID,
		EventId:            req.EventId,
		Name:               category.Name,
		Description:        safeStringPtr(category.Description),
		Price:              req.Price,
		QuantityAvailable:  req.QuantityAvailable,
		QuantitySold:       0,
		MaxTicketsPerOrder: req.MaxTicketsPerOrder,
		SalesStart:         req.SalesStart,
		SalesEnd:           req.SalesEnd,
		Benefits:           req.Benefits,
		IsActive:           category.IsActive,
		CreatedAt:          timestamppb.New(category.CreatedAt),
		UpdatedAt:          timestamppb.New(category.UpdatedAt),
	}

	return response, nil
}

// GetEventCategories obtiene categorías de un evento
func (s *Server) GetEventCategories(ctx context.Context, req *osmi.EventLookup) (*osmi.CategoryListResponse, error) {
	log.Printf("Getting categories for event: %s", req.PublicId)

	// Obtener el evento
	event, err := s.EventRepo.GetByPublicID(ctx, req.PublicId)
	if err != nil {
		return nil, fmt.Errorf("event not found: %w", err)
	}

	// Obtener categorías del evento
	categories, err := s.CategoryRepo.GetEventCategories(ctx, event.ID)
	if err != nil {
		log.Printf("Error getting event categories: %v", err)
		return nil, fmt.Errorf("error retrieving categories: %w", err)
	}

	// Convertir categorías a protobuf
	pbCategories := make([]*osmi.CategoryResponse, 0, len(categories))
	for _, category := range categories {
		catResponse := &osmi.CategoryResponse{
			PublicId:           category.PublicID,
			EventId:            req.PublicId,
			Name:               category.Name,
			Description:        safeStringPtr(category.Description),
			Price:              0,
			QuantityAvailable:  0,
			QuantitySold:       0,
			MaxTicketsPerOrder: 0,
			SalesStart:         timestamppb.New(time.Now()),
			Benefits:           []string{},
			IsActive:           category.IsActive,
			CreatedAt:          timestamppb.New(category.CreatedAt),
			UpdatedAt:          timestamppb.New(category.UpdatedAt),
		}
		pbCategories = append(pbCategories, catResponse)
	}

	return &osmi.CategoryListResponse{
		Categories:    pbCategories,
		EventName:     event.Name,
		EventPublicId: event.PublicID,
	}, nil
}

// ============================================================================
// MÉTODOS DE EVENTOS
// ============================================================================

// CreateEvent crea un nuevo evento
func (s *Server) CreateEvent(ctx context.Context, req *osmi.CreateEventRequest) (*osmi.EventResponse, error) {
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
	startsAt, err := time.Parse(time.RFC3339, req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start_date format: %w", err)
	}

	endsAt, err := time.Parse(time.RFC3339, req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end_date format: %w", err)
	}

	// Validar lógica de fechas
	if endsAt.Before(startsAt) {
		return nil, fmt.Errorf("end_date cannot be before start_date")
	}

	// Generar UUID para el evento
	publicID := uuid.New().String()

	// Si se necesitan tags, debere agregarlos al proto primero
	// Por ahora, lo dejo como nil
	var tags *[]string = nil

	// Crear evento
	event := &entities.Event{
		PublicID:         publicID,
		Name:             strings.TrimSpace(req.Name),
		Slug:             strings.ToLower(strings.ReplaceAll(req.Name, " ", "-")),
		ShortDescription: &req.ShortDescription,
		Description:      &req.Description,
		EventType:        "in_person",
		Timezone:         "UTC",
		StartsAt:         startsAt,
		EndsAt:           endsAt,
		VenueName:        &req.Location,
		AddressFull:      &req.VenueDetails,
		City:             nil,
		State:            nil,
		Country:          nil,
		Status:           "draft",
		Visibility:       "public",
		IsFeatured:       false,
		IsFree:           false,
		Tags:             tags,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Manejar campos opcionales
	if req.CoverImageUrl != "" {
		event.CoverImageURL = &req.CoverImageUrl
	}
	if req.BannerImageUrl != "" {
		event.BannerImageURL = &req.BannerImageUrl
	}
	if req.MaxAttendees > 0 {
		maxAttendees := int(req.MaxAttendees)
		event.MaxAttendees = &maxAttendees
	}

	// Crear evento en la base de datos
	err = s.EventRepo.Create(ctx, event)
	if err != nil {
		log.Printf("Error creating event: %v", err)
		return nil, fmt.Errorf("error creating event: %w", err)
	}

	log.Printf("Event created successfully: %s (PublicID: %s)", req.Name, publicID)

	// Obtener el evento creado para la respuesta
	createdEvent, err := s.EventRepo.GetByPublicID(ctx, publicID)
	if err != nil {
		log.Printf("Error retrieving created event: %v", err)
		return nil, fmt.Errorf("event created but retrieval failed: %w", err)
	}

	return s.mapEventToResponse(createdEvent), nil
}

// GetEvent implementa el método gRPC para obtener eventos
func (s *Server) GetEvent(ctx context.Context, req *osmi.EventLookup) (*osmi.EventResponse, error) {
	log.Printf("Getting event: %s", req.PublicId)

	if !isValidUUID(req.PublicId) {
		return nil, fmt.Errorf("invalid event ID format: must be a valid UUID")
	}

	event, err := s.EventRepo.GetByPublicID(ctx, req.PublicId)
	if err != nil {
		log.Printf("Error getting event: %v", err)
		return nil, fmt.Errorf("event not found with id: %s", req.PublicId)
	}

	return s.mapEventToResponse(event), nil
}

// ListEvents implementa el método gRPC para listar eventos
func (s *Server) ListEvents(ctx context.Context, req *osmi.ListEventsRequest) (*osmi.EventListResponse, error) {
	log.Println("Listing events with filters")

	// Crear filtro para eventos
	filter := make(map[string]interface{})

	if req.Name != "" {
		filter["name"] = req.Name
	}
	if req.OrganizerId != "" {
		filter["organizer_id"] = req.OrganizerId
	}
	if req.CategoryId != "" {
		filter["category_id"] = req.CategoryId
	}
	if req.Status != "" {
		filter["status"] = req.Status
	}
	if req.DateFrom != "" {
		filter["date_from"] = req.DateFrom
	}
	if req.DateTo != "" {
		filter["date_to"] = req.DateTo
	}
	if req.City != "" {
		filter["city"] = req.City
	}
	if req.Country != "" {
		filter["country"] = req.Country
	}
	if req.Search != "" {
		filter["search"] = req.Search
	}
	if req.IsFeatured {
		filter["is_featured"] = true
	}
	if req.IsFree {
		filter["is_free"] = true
	}

	// Paginación
	limit := int(req.PageSize)
	if limit <= 0 {
		limit = 20
	}
	offset := (int(req.Page) - 1) * limit
	if offset < 0 {
		offset = 0
	}

	events, total, err := s.EventRepo.List(ctx, filter, limit, offset)
	if err != nil {
		log.Printf("Error listing events: %v", err)
		return nil, fmt.Errorf("error retrieving events: %w", err)
	}

	pbEvents := make([]*osmi.EventResponse, 0, len(events))
	for _, event := range events {
		pbEvents = append(pbEvents, s.mapEventToResponse(event))
	}

	totalPages := (int(total) + limit - 1) / limit

	return &osmi.EventListResponse{
		Events:     pbEvents,
		TotalCount: int32(total),
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: int32(totalPages),
	}, nil
}

// ============================================================================
// MÉTODOS DE CLIENTES
// ============================================================================

// CreateCustomer implementa el método gRPC para crear clientes
func (s *Server) CreateCustomer(ctx context.Context, req *osmi.CreateCustomerRequest) (*osmi.CustomerResponse, error) {
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

	// Usar entities.Customer directamente
	customer := &entities.Customer{
		PublicID:        uuid.New().String(),
		FullName:        strings.TrimSpace(req.Name),
		Email:           strings.TrimSpace(req.Email),
		Phone:           &phone,
		IsActive:        true,
		CustomerSegment: "new",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Usar Create
	err := s.CustomerRepo.Create(ctx, customer)
	if err != nil {
		log.Printf("Error creating customer: %v", err)

		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "23505") {
			return nil, fmt.Errorf("customer with email %s already exists", req.Email)
		}
		return nil, fmt.Errorf("error creating customer: %w", err)
	}

	log.Printf("Customer created successfully: %s (ID: %d, PublicID: %s)",
		req.Email, customer.ID, customer.PublicID)

	// Incluir campo Id según el proto
	return &osmi.CustomerResponse{
		Id:        int32(customer.ID),
		PublicId:  customer.PublicID,
		Name:      customer.FullName,
		Email:     customer.Email,
		Phone:     safeStringPtr(customer.Phone),
		CreatedAt: timestamppb.New(customer.CreatedAt),
		UpdatedAt: timestamppb.New(customer.UpdatedAt),
	}, nil
}

// GetCustomer obtiene un cliente
func (s *Server) GetCustomer(ctx context.Context, req *osmi.CustomerLookup) (*osmi.CustomerResponse, error) {
	var customer *entities.Customer
	var err error

	switch lookup := req.Lookup.(type) {
	case *osmi.CustomerLookup_Id:
		log.Printf("Getting customer by ID: %d", lookup.Id)
		if lookup.Id <= 0 {
			return nil, fmt.Errorf("customer ID must be positive")
		}
		customer, err = s.CustomerRepo.GetByID(ctx, int64(lookup.Id))

	case *osmi.CustomerLookup_PublicId:
		log.Printf("Getting customer by PublicId: %s", lookup.PublicId)
		if !isValidUUID(lookup.PublicId) {
			return nil, fmt.Errorf("invalid public_id format: must be a valid UUID")
		}
		customer, err = s.CustomerRepo.GetByPublicID(ctx, lookup.PublicId)

	case *osmi.CustomerLookup_Email:
		log.Printf("Getting customer by Email: %s", lookup.Email)
		if strings.TrimSpace(lookup.Email) == "" {
			return nil, fmt.Errorf("email cannot be empty")
		}
		customer, err = s.CustomerRepo.GetByEmail(ctx, lookup.Email)

	default:
		return nil, fmt.Errorf("no valid lookup parameter provided")
	}

	if err != nil {
		log.Printf("Error getting customer: %v", err)
		return nil, fmt.Errorf("customer not found")
	}

	return &osmi.CustomerResponse{
		Id:        int32(customer.ID),
		PublicId:  customer.PublicID,
		Name:      customer.FullName,
		Email:     customer.Email,
		Phone:     safeStringPtr(customer.Phone),
		CreatedAt: timestamppb.New(customer.CreatedAt),
		UpdatedAt: timestamppb.New(customer.UpdatedAt),
	}, nil
}

// ============================================================================
// MÉTODOS DE USUARIOS
// ============================================================================

// CreateUser crea un nuevo usuario
func (s *Server) CreateUser(ctx context.Context, req *osmi.CreateUserRequest) (*osmi.UserResponse, error) {
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

	// Validar role usando el enum
	role := strings.TrimSpace(req.Role)
	if role == "" {
		role = "customer"
	}
	if !enums.UserRole(role).IsValid() {
		return nil, fmt.Errorf("invalid role. Must be one of: admin, organizer, customer, staff, guest")
	}

	// Crear entidad User directamente
	user := &entities.User{
		PublicID:            uuid.New().String(),
		Email:               strings.TrimSpace(req.Email),
		Username:            &req.Name,
		PasswordHash:        hashPassword(req.Password),
		Role:                role,
		IsActive:            true,
		EmailVerified:       false,
		PhoneVerified:       false,
		PreferredLanguage:   "es",
		PreferredCurrency:   "MXN",
		Timezone:            "UTC",
		FailedLoginAttempts: 0,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	err := s.UserRepo.Create(ctx, user)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "23505") {
			return nil, fmt.Errorf("user with email %s already exists", req.Email)
		}
		return nil, fmt.Errorf("error creating user: %w", err)
	}

	log.Printf("User created successfully: %s (ID: %d, PublicID: %s)",
		req.Email, user.ID, user.PublicID)

	return &osmi.UserResponse{
		UserId:    user.PublicID,
		Status:    "active",
		Name:      *user.Username,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: timestamppb.New(user.CreatedAt),
	}, nil
}

// GetUser obtiene un usuario
func (s *Server) GetUser(ctx context.Context, req *osmi.UserLookup) (*osmi.UserResponse, error) {
	log.Printf("Getting user: %s", req.UserId)

	if !isValidUUID(req.UserId) {
		return nil, fmt.Errorf("invalid user ID format: must be a valid UUID")
	}

	user, err := s.UserRepo.GetByPublicID(ctx, req.UserId)
	if err != nil {
		log.Printf("Error getting user: %v", err)
		return nil, fmt.Errorf("user not found with id: %s", req.UserId)
	}

	return &osmi.UserResponse{
		UserId:    user.PublicID,
		Status:    "active",
		Name:      *user.Username,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: timestamppb.New(user.CreatedAt),
	}, nil
}

// ============================================================================
// MÉTODOS DE TICKETS
// ============================================================================

// CreateTicket implementa el método gRPC para crear tickets
func (s *Server) CreateTicket(ctx context.Context, req *osmi.CreateTicketRequest) (*osmi.TicketResponse, error) {
	log.Printf("CreateTicket called with event_id: %s, user_id: %s, category_id: %s, quantity: %d",
		truncateString(req.EventId, 50), truncateString(req.UserId, 50),
		truncateString(req.CategoryId, 50), req.Quantity)

	// Validaciones básicas
	if strings.TrimSpace(req.EventId) == "" {
		return nil, fmt.Errorf("event_id is required")
	}
	if strings.TrimSpace(req.CategoryId) == "" {
		return nil, fmt.Errorf("category_id is required")
	}
	if req.Quantity <= 0 {
		req.Quantity = 1
	}
	if req.Quantity > 10 {
		return nil, fmt.Errorf("cannot create more than 10 tickets at once")
	}

	// TODO: Implementar lógica completa cuando el repositorio esté listo
	return nil, fmt.Errorf("CreateTicket method is under development")
}

// ListTickets implementa el método gRPC para listar tickets
func (s *Server) ListTickets(ctx context.Context, req *osmi.ListTicketsRequest) (*osmi.TicketListResponse, error) {
	log.Printf("ListTickets called with filters")

	// Por ahora, retornar lista vacía
	return &osmi.TicketListResponse{
		Tickets:    []*osmi.TicketResponse{},
		TotalCount: 0,
		Page:       req.Page,
		PageSize:   req.PageSize,
	}, nil
}

// GetUserTickets obtiene tickets de un usuario específico
func (s *Server) GetUserTickets(ctx context.Context, req *osmi.UserLookup) (*osmi.TicketListResponse, error) {
	log.Printf("GetUserTickets called for user: %s", req.UserId)

	// Por ahora, retornar lista vacía
	return &osmi.TicketListResponse{
		Tickets:    []*osmi.TicketResponse{},
		TotalCount: 0,
	}, nil
}

// GetCustomerTickets obtiene tickets de un cliente específico
func (s *Server) GetCustomerTickets(ctx context.Context, req *osmi.CustomerLookup) (*osmi.TicketListResponse, error) {
	log.Printf("GetCustomerTickets called for customer")

	// Por ahora, retornar lista vacía
	return &osmi.TicketListResponse{
		Tickets:    []*osmi.TicketResponse{},
		TotalCount: 0,
	}, nil
}

// UpdateTicketStatus actualiza el estado de un ticket
func (s *Server) UpdateTicketStatus(ctx context.Context, req *osmi.UpdateTicketStatusRequest) (*osmi.TicketResponse, error) {
	log.Printf("UpdateTicketStatus called for ticket: %s, status: %s", req.TicketId, req.Status)

	// Por ahora, retornar error
	return nil, fmt.Errorf("UpdateTicketStatus method temporarily disabled")
}

// UpdateTicket actualiza información de un ticket
func (s *Server) UpdateTicket(ctx context.Context, req *osmi.UpdateTicketRequest) (*osmi.TicketResponse, error) {
	log.Printf("UpdateTicket called for ticket: %s", req.TicketId)

	// Por ahora, retornar error
	return nil, fmt.Errorf("UpdateTicket method temporarily disabled")
}

// GetTicketDetails obtiene detalles completos de un ticket
func (s *Server) GetTicketDetails(ctx context.Context, req *osmi.GetTicketRequest) (*osmi.TicketResponse, error) {
	log.Printf("GetTicketDetails called for ticket: %s", req.Id)

	// Por ahora, retornar error
	return nil, fmt.Errorf("GetTicketDetails method temporarily disabled")
}

// GetTicketStats obtiene estadísticas de tickets para un evento
func (s *Server) GetTicketStats(ctx context.Context, req *osmi.GetTicketStatsRequest) (*osmi.TicketStatsResponse, error) {
	log.Printf("GetTicketStats called for event: %s", req.EventId)

	// Por ahora, retornar error
	return nil, fmt.Errorf("GetTicketStats method temporarily disabled")
}

// ============================================================================
// MÉTODOS DE HEALTH CHECK
// ============================================================================

// HealthCheck implementa el health check
func (s *Server) HealthCheck(ctx context.Context, req *osmi.Empty) (*osmi.HealthResponse, error) {
	return &osmi.HealthResponse{
		Status:    "healthy",
		Service:   "osmi-server",
		Version:   "1.0.0",
		Timestamp: timestamppb.Now(),
	}, nil
}

// =============================================================================
// MÉTODOS HELPER
// =============================================================================

// mapEventToResponse mapea un modelo Event a protobuf
func (s *Server) mapEventToResponse(event *entities.Event) *osmi.EventResponse {
	response := &osmi.EventResponse{
		PublicId:    event.PublicID,
		Name:        event.Name,
		StartDate:   event.StartsAt.Format(time.RFC3339),
		EndDate:     event.EndsAt.Format(time.RFC3339),
		Location:    safeStringPtr(event.VenueName),
		Tags:        []string{},
		IsActive:    event.Status != "cancelled" && event.Status != "archived",
		IsPublished: event.Status == "published" || event.Status == "live",
		CreatedAt:   timestamppb.New(event.CreatedAt),
		UpdatedAt:   timestamppb.New(event.UpdatedAt),
	}

	if event.Tags != nil {
		response.Tags = *event.Tags
	}

	if event.Description != nil {
		response.Description = *event.Description
	}
	if event.ShortDescription != nil {
		response.ShortDescription = *event.ShortDescription
	}
	if event.VenueName != nil {
		response.VenueDetails = *event.VenueName
	}
	if event.CoverImageURL != nil {
		response.ImageUrl = *event.CoverImageURL
	}
	if event.BannerImageURL != nil {
		response.BannerUrl = *event.BannerImageURL
	}
	if event.MaxAttendees != nil {
		response.MaxAttendees = int32(*event.MaxAttendees)
	}

	return response
}

// safeStringPtr convierte *string a string vacío si es nil
func safeStringPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// safeTimePtr convierte *time.Time a timestamppb.Timestamp si no es nil
func safeTimePtr(t *time.Time) *timestamppb.Timestamp {
	if t == nil || t.IsZero() {
		return nil
	}
	return timestamppb.New(*t)
}

// hashPassword hashea una contraseña (implementación temporal)
func hashPassword(password string) string {
	return "$2a$10$" + password // Temporal, debe ser implementado correctamente
}
