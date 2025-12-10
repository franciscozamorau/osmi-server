// service.go - COMPLETO Y CORREGIDO PARA TU PROTO
package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	osmi "github.com/franciscozamorau/osmi-server/gen"
	"github.com/franciscozamorau/osmi-server/internal/models"
	"github.com/franciscozamorau/osmi-server/internal/repository"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Server implementa el servicio gRPC
type Server struct {
	osmi.UnimplementedOsmiServiceServer
	CustomerRepo *repository.CustomerRepository
	TicketRepo   *repository.TicketRepository
	EventRepo    *repository.EventRepository
	UserRepo     *repository.UserRepository
	CategoryRepo *repository.CategoryRepository
}

// NewServer crea una nueva instancia del servidor
func NewServer(customerRepo *repository.CustomerRepository, ticketRepo *repository.TicketRepository, eventRepo *repository.EventRepository, userRepo *repository.UserRepository, categoryRepo *repository.CategoryRepository) *Server {
	return &Server{
		CustomerRepo: customerRepo,
		TicketRepo:   ticketRepo,
		EventRepo:    eventRepo,
		UserRepo:     userRepo,
		CategoryRepo: categoryRepo,
	}
}

// CreateCategory implementa el método gRPC para crear categorías
func (s *Server) CreateCategory(ctx context.Context, req *osmi.CategoryRequest) (*osmi.CategoryResponse, error) {
	log.Printf("Creating category for event: %s, name: %s", req.EventId, req.Name)

	// Validaciones
	if strings.TrimSpace(req.EventId) == "" {
		return nil, fmt.Errorf("event_id is required")
	}
	if strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("name is required")
	}
	if req.Price < 0 {
		return nil, fmt.Errorf("price must be non-negative")
	}
	if req.QuantityAvailable < 0 {
		return nil, fmt.Errorf("quantity_available must be non-negative")
	}

	// Validar que el event_id es un UUID válido
	if !repository.IsValidUUID(req.EventId) {
		return nil, fmt.Errorf("invalid event_id format: must be a valid UUID")
	}

	// Obtener el evento para conseguir su ID
	event, err := s.EventRepo.GetEventByPublicID(ctx, req.EventId)
	if err != nil {
		return nil, fmt.Errorf("event not found: %w", err)
	}

	// Mapear request al modelo Category (usando tipos nativos)
	category := &models.Category{
		EventID:            event.ID, // Usar event.ID (int64) no EventPublicID
		Name:               strings.TrimSpace(req.Name),
		Price:              req.Price,
		QuantityAvailable:  req.QuantityAvailable,
		MaxTicketsPerOrder: req.MaxTicketsPerOrder,
		IsActive:           req.IsActive,
	}

	// Manejar description como *string
	if req.Description != "" {
		desc := strings.TrimSpace(req.Description)
		category.Description = &desc
	}

	// Manejar sales_start
	if req.SalesStart != nil {
		category.SalesStart = req.SalesStart.AsTime()
	} else {
		category.SalesStart = time.Now()
	}

	// Manejar sales_end como *time.Time
	if req.SalesEnd != nil {
		salesEnd := req.SalesEnd.AsTime()
		category.SalesEnd = &salesEnd
	}

	// Llamar al repositorio para crear la categoría
	publicID, err := s.CategoryRepo.CreateCategory(ctx, category, req.Benefits)
	if err != nil {
		log.Printf("Error creating category: %v", err)
		return nil, fmt.Errorf("error creating category: %w", err)
	}

	log.Printf("Category created successfully: %s (PublicID: %s)", req.Name, publicID)

	// Obtener la categoría creada para la respuesta
	createdCategory, err := s.CategoryRepo.GetCategoryByPublicID(ctx, publicID)
	if err != nil {
		log.Printf("Error retrieving created category: %v", err)
		return nil, fmt.Errorf("category created but retrieval failed: %w", err)
	}

	// Construir respuesta con tipos correctos según tu proto
	response := &osmi.CategoryResponse{
		PublicId:           createdCategory.PublicID,
		Name:               createdCategory.Name,
		Price:              createdCategory.Price,
		QuantityAvailable:  createdCategory.QuantityAvailable,
		QuantitySold:       createdCategory.QuantitySold,
		MaxTicketsPerOrder: createdCategory.MaxTicketsPerOrder,
		SalesStart:         timestamppb.New(createdCategory.SalesStart),
		Benefits:           req.Benefits,
		IsActive:           createdCategory.IsActive,
		CreatedAt:          timestamppb.New(createdCategory.CreatedAt),
		UpdatedAt:          timestamppb.New(createdCategory.UpdatedAt),
	}

	// Manejar description en respuesta
	if createdCategory.Description != nil {
		response.Description = *createdCategory.Description
	}

	// Manejar sales_end en respuesta
	if createdCategory.SalesEnd != nil {
		response.SalesEnd = timestamppb.New(*createdCategory.SalesEnd)
	}

	return response, nil
}

// GetEventCategories obtiene categorías de un evento
func (s *Server) GetEventCategories(ctx context.Context, req *osmi.EventLookup) (*osmi.CategoryListResponse, error) {
	log.Printf("Getting categories for event: %s", req.PublicId)

	// Validar UUID
	if !repository.IsValidUUID(req.PublicId) {
		return nil, fmt.Errorf("invalid event ID format: must be a valid UUID")
	}

	// Obtener categorías del evento
	categories, err := s.CategoryRepo.GetCategoriesByEvent(ctx, req.PublicId)
	if err != nil {
		log.Printf("Error getting event categories: %v", err)
		return nil, fmt.Errorf("error retrieving categories: %w", err)
	}

	// Obtener información del evento para la respuesta
	event, err := s.EventRepo.GetEventByPublicID(ctx, req.PublicId)
	if err != nil {
		log.Printf("Error getting event for categories: %v", err)
		return nil, fmt.Errorf("error retrieving event: %w", err)
	}

	// Convertir categorías a protobuf con tipos correctos
	pbCategories := make([]*osmi.CategoryResponse, 0, len(categories))
	for _, category := range categories {
		catResponse := &osmi.CategoryResponse{
			PublicId:           category.PublicID,
			Name:               category.Name,
			Price:              category.Price,
			QuantityAvailable:  category.QuantityAvailable,
			QuantitySold:       category.QuantitySold,
			MaxTicketsPerOrder: category.MaxTicketsPerOrder,
			SalesStart:         timestamppb.New(category.SalesStart),
			IsActive:           category.IsActive,
			CreatedAt:          timestamppb.New(category.CreatedAt),
			UpdatedAt:          timestamppb.New(category.UpdatedAt),
		}

		// Manejar description
		if category.Description != nil {
			catResponse.Description = *category.Description
		}

		// Manejar sales_end
		if category.SalesEnd != nil {
			catResponse.SalesEnd = timestamppb.New(*category.SalesEnd)
		}

		pbCategories = append(pbCategories, catResponse)
	}

	log.Printf("Retrieved %d categories for event: %s", len(pbCategories), req.PublicId)

	return &osmi.CategoryListResponse{
		Categories:    pbCategories,
		EventName:     event.Name,
		EventPublicId: event.PublicID,
	}, nil
}

// CreateEvent crea un nuevo evento
func (s *Server) CreateEvent(ctx context.Context, req *osmi.EventRequest) (*osmi.EventResponse, error) {
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
		return nil, fmt.Errorf("invalid start_date format: %w", err)
	}

	endTime, err := time.Parse(time.RFC3339, req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end_date format: %w", err)
	}

	// Validar lógica de fechas
	if !repository.IsDateRangeValid(startTime, endTime) {
		return nil, fmt.Errorf("end_date cannot be before start_date")
	}

	// Generar UUID para el evento
	publicID := uuid.New().String()

	// Mapear request a modelo Event con tipos nativos
	event := &models.Event{
		PublicID:    publicID,
		Name:        strings.TrimSpace(req.Name),
		StartDate:   startTime,
		EndDate:     endTime,
		Location:    strings.TrimSpace(req.Location),
		Tags:        req.Tags,
		IsActive:    req.IsActive,
		IsPublished: req.IsPublished,
		Status:      "draft", // Estado por defecto
	}

	// Manejar campos opcionales como *string
	if req.Description != "" {
		desc := strings.TrimSpace(req.Description)
		event.Description = &desc
	}
	if req.ShortDescription != "" {
		shortDesc := strings.TrimSpace(req.ShortDescription)
		event.ShortDescription = &shortDesc
	}
	if req.VenueDetails != "" {
		venue := strings.TrimSpace(req.VenueDetails)
		event.VenueDetails = &venue
	}
	if req.Category != "" {
		cat := strings.TrimSpace(req.Category)
		event.Category = &cat
	}
	if req.ImageUrl != "" {
		img := strings.TrimSpace(req.ImageUrl)
		event.ImageURL = &img
	}
	if req.BannerUrl != "" {
		banner := strings.TrimSpace(req.BannerUrl)
		event.BannerURL = &banner
	}
	if req.MaxAttendees > 0 {
		event.MaxAttendees = &req.MaxAttendees
	}

	// Crear evento en la base de datos
	createdPublicID, err := s.EventRepo.CreateEvent(ctx, event)
	if err != nil {
		log.Printf("Error creating event: %v", err)
		return nil, fmt.Errorf("error creating event: %w", err)
	}

	log.Printf("Event created successfully: %s (PublicID: %s)", req.Name, createdPublicID)

	// Obtener el evento creado para la respuesta
	createdEvent, err := s.EventRepo.GetEventByPublicID(ctx, createdPublicID)
	if err != nil {
		log.Printf("Error retrieving created event: %v", err)
		return nil, fmt.Errorf("event created but retrieval failed: %w", err)
	}

	return s.mapEventToResponse(createdEvent), nil
}

// GetEvent implementa el método gRPC para obtener eventos
func (s *Server) GetEvent(ctx context.Context, req *osmi.EventLookup) (*osmi.EventResponse, error) {
	log.Printf("Getting event: %s", req.PublicId)

	// Validar que el public_id sea UUID válido
	if !repository.IsValidUUID(req.PublicId) {
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
func (s *Server) ListEvents(ctx context.Context, req *osmi.Empty) (*osmi.EventListResponse, error) {
	log.Println("Listing all events")

	events, err := s.EventRepo.ListActiveEvents(ctx)
	if err != nil {
		log.Printf("Error listing events: %v", err)
		return nil, fmt.Errorf("error retrieving events: %w", err)
	}

	pbEvents := make([]*osmi.EventResponse, 0, len(events))
	for _, event := range events {
		pbEvents = append(pbEvents, s.mapEventToResponse(event))
	}

	log.Printf("Retrieved %d active events", len(pbEvents))

	return &osmi.EventListResponse{
		Events:     pbEvents,
		TotalCount: int32(len(pbEvents)),
	}, nil
}

// CreateCustomer implementa el método gRPC para crear clientes
func (s *Server) CreateCustomer(ctx context.Context, req *osmi.CustomerRequest) (*osmi.CustomerResponse, error) {
	log.Printf("Creating customer: %s, email: %s", req.Name, req.Email)

	// Validaciones
	if strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("name is required")
	}
	if strings.TrimSpace(req.Email) == "" {
		return nil, fmt.Errorf("email is required")
	}
	if !repository.IsValidEmail(strings.TrimSpace(req.Email)) {
		return nil, fmt.Errorf("invalid email format")
	}

	// Validar formato de teléfono
	phone := strings.TrimSpace(req.Phone)
	if phone != "" && !repository.IsValidPhone(phone) {
		return nil, fmt.Errorf("invalid phone format. Use E.164 format: +1234567890 or standard format")
	}

	// Usar el método que acepta CreateCustomerRequest
	customerReq := &models.CreateCustomerRequest{
		Name:         strings.TrimSpace(req.Name),
		Email:        strings.TrimSpace(req.Email),
		Phone:        phone,
		UserID:       strings.TrimSpace(req.UserId), // Opcional
		CustomerType: "guest",                       // Por defecto
		Source:       "web",                         // Por defecto
	}

	customer, err := s.CustomerRepo.CreateCustomer(ctx, customerReq)
	if err != nil {
		log.Printf("Error creating customer: %v", err)
		if repository.IsDuplicateKeyError(err) {
			return nil, fmt.Errorf("customer with email %s already exists", repository.SafeStringForLog(req.Email))
		}
		return nil, fmt.Errorf("error creating customer: %w", err)
	}

	log.Printf("Customer created successfully: %s (ID: %d, PublicID: %s, Type: %s)",
		repository.SafeStringForLog(customer.Email), customer.ID, customer.PublicID, customer.CustomerType)

	// ✅ CORREGIDO: Usar los campos correctos según tu proto
	return &osmi.CustomerResponse{
		PublicId:      customer.PublicID, // NO Id, ES PublicId
		Name:          customer.Name,
		Email:         customer.Email,
		Phone:         safeStringPtr(customer.Phone),
		CustomerType:  customer.CustomerType,
		LoyaltyPoints: customer.LoyaltyPoints,
		IsVerified:    customer.IsVerified,
		Source:        customer.Source,
		CreatedAt:     timestamppb.New(customer.CreatedAt),
		UpdatedAt:     timestamppb.New(customer.UpdatedAt),
	}, nil
}

// GetCustomer obtiene un cliente
func (s *Server) GetCustomer(ctx context.Context, req *osmi.CustomerLookup) (*osmi.CustomerResponse, error) {
	var customer *models.Customer
	var err error

	// Manejar el oneof lookup CORRECTAMENTE
	switch lookup := req.Lookup.(type) {
	case *osmi.CustomerLookup_Id:
		log.Printf("Getting customer by ID: %d", lookup.Id)
		if lookup.Id <= 0 {
			return nil, fmt.Errorf("customer ID must be positive")
		}
		customer, err = s.CustomerRepo.GetCustomerByID(ctx, int64(lookup.Id))

	case *osmi.CustomerLookup_PublicId:
		log.Printf("Getting customer by PublicId: %s", lookup.PublicId)
		if !repository.IsValidUUID(lookup.PublicId) {
			return nil, fmt.Errorf("invalid public_id format: must be a valid UUID")
		}
		customer, err = s.CustomerRepo.GetCustomerByPublicID(ctx, lookup.PublicId)

	case *osmi.CustomerLookup_Email:
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

	// ✅ CORREGIDO: Usar los campos correctos según tu proto
	return &osmi.CustomerResponse{
		PublicId:      customer.PublicID, // NO Id, ES PublicId
		Name:          customer.Name,
		Email:         customer.Email,
		Phone:         safeStringPtr(customer.Phone),
		CustomerType:  customer.CustomerType,
		LoyaltyPoints: customer.LoyaltyPoints,
		IsVerified:    customer.IsVerified,
		Source:        customer.Source,
		CreatedAt:     timestamppb.New(customer.CreatedAt),
		UpdatedAt:     timestamppb.New(customer.UpdatedAt),
	}, nil
}

// CreateUser crea un nuevo usuario
func (s *Server) CreateUser(ctx context.Context, req *osmi.UserRequest) (*osmi.UserResponse, error) {
	log.Printf("Creating user: %s, email: %s", req.Name, req.Email)

	// Validaciones
	if strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("name is required")
	}
	if strings.TrimSpace(req.Email) == "" {
		return nil, fmt.Errorf("email is required")
	}
	if !repository.IsValidEmail(strings.TrimSpace(req.Email)) {
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
	if !repository.IsValidUserRole(role) {
		return nil, fmt.Errorf("invalid role. Must be: customer, organizer, admin, or guest")
	}

	// Usar el método que acepta CreateUserRequest
	userReq := &models.CreateUserRequest{
		Username: strings.TrimSpace(req.Name),
		Email:    strings.TrimSpace(req.Email),
		Password: strings.TrimSpace(req.Password),
		Role:     role,
	}

	user, err := s.UserRepo.CreateUser(ctx, userReq)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		if repository.IsDuplicateKeyError(err) {
			return nil, fmt.Errorf("user with email %s already exists", repository.SafeStringForLog(req.Email))
		}
		return nil, fmt.Errorf("error creating user: %w", err)
	}

	log.Printf("User created successfully: %s (ID: %d, PublicID: %s)",
		repository.SafeStringForLog(user.Email), user.ID, user.PublicID)

	return &osmi.UserResponse{
		UserId:    user.PublicID,
		Status:    "active",
		Name:      user.Username,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}, nil
}

// CreateTicket implementa el método gRPC para crear tickets
func (s *Server) CreateTicket(ctx context.Context, req *osmi.TicketRequest) (*osmi.TicketResponse, error) {
	log.Printf("Creating ticket for event: %s, customer: %s, category: %s, quantity: %d",
		repository.SafeStringForLog(req.EventId), repository.SafeStringForLog(req.CustomerId),
		repository.SafeStringForLog(req.CategoryId), req.Quantity)

	// Validar customer_id obligatorio
	if strings.TrimSpace(req.CustomerId) == "" {
		return nil, fmt.Errorf("customer_id is required")
	}
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

	// Validar UUIDs
	if !repository.IsValidUUID(req.CustomerId) {
		return nil, fmt.Errorf("invalid customer ID format: must be a valid UUID")
	}
	if !repository.IsValidUUID(req.EventId) {
		return nil, fmt.Errorf("invalid event ID format: must be a valid UUID")
	}
	if !repository.IsValidUUID(req.CategoryId) {
		return nil, fmt.Errorf("invalid category ID format: must be a valid UUID")
	}
	// user_id es opcional, pero si viene debe ser UUID válido
	if req.UserId != "" && !repository.IsValidUUID(req.UserId) {
		return nil, fmt.Errorf("invalid user ID format: must be a valid UUID")
	}

	// Validar que el customer existe
	customer, err := s.CustomerRepo.GetCustomerByPublicID(ctx, strings.TrimSpace(req.CustomerId))
	if err != nil {
		log.Printf("Error validating customer: %v", err)
		return nil, fmt.Errorf("customer not found: %s", req.CustomerId)
	}

	// Si viene user_id, validar que el usuario existe y está activo
	if req.UserId != "" {
		user, err := s.UserRepo.GetUserByPublicID(ctx, strings.TrimSpace(req.UserId))
		if err != nil {
			log.Printf("Error validating user: %v", err)
			return nil, fmt.Errorf("user not found: %s", req.UserId)
		}
		if !user.IsActive {
			return nil, fmt.Errorf("user account is inactive: %s", req.UserId)
		}
		// Vincular customer con user si es necesario
		if customer.UserID == nil {
			err = s.CustomerRepo.UpdateCustomerWithUserID(ctx, customer.ID, user.ID)
			if err != nil {
				log.Printf("Warning: Could not link customer to user: %v", err)
			} else {
				log.Printf("Customer %s linked to user %s", customer.PublicID, user.PublicID)
			}
		}
	}

	// Crear ticket
	ticketPublicID, err := s.TicketRepo.CreateTicket(ctx, req)
	if err != nil {
		log.Printf("Error creating ticket: %v", err)
		return nil, fmt.Errorf("error creating ticket: %w", err)
	}

	// Obtener el ticket creado con detalles completos
	ticketDetails, err := s.TicketRepo.GetTicketWithDetails(ctx, ticketPublicID)
	if err != nil {
		log.Printf("Error retrieving created ticket: %v", err)
		return nil, fmt.Errorf("ticket created but retrieval failed: %w", err)
	}

	log.Printf("Ticket created successfully: %s for customer: %s",
		repository.SafeStringForLog(ticketPublicID), repository.SafeStringForLog(customer.Name))

	// ✅ CORREGIDO: Usar category_name (NO Category)
	return &osmi.TicketResponse{
		TicketId:     ticketDetails.TicketID,
		Status:       ticketDetails.Status,
		Code:         ticketDetails.Code,
		QrCodeUrl:    "",
		EventName:    ticketDetails.EventName,
		CategoryName: ticketDetails.CategoryName, // ✅ CORREGIDO: CategoryName
		Price:        ticketDetails.Price,
		CreatedAt:    timestamppb.New(ticketDetails.CreatedAt),
	}, nil
}

// ListTickets implementa el método gRPC para listar tickets
func (s *Server) ListTickets(ctx context.Context, req *osmi.TicketLookup) (*osmi.TicketListResponse, error) {
	var tickets []*models.Ticket
	var err error
	var lookupType string

	// Diferenciar entre tickets de usuarios y clientes
	switch lookup := req.Lookup.(type) {
	case *osmi.TicketLookup_UserId:
		lookupType = "user"
		log.Printf("Listing tickets for user: %s", repository.SafeStringForLog(lookup.UserId))
		if !repository.IsValidUUID(lookup.UserId) {
			return nil, fmt.Errorf("invalid user ID format: must be a valid UUID")
		}
		// Usar GetTicketsByUserID para usuarios registrados
		tickets, err = s.TicketRepo.GetTicketsByUserID(ctx, lookup.UserId)

	case *osmi.TicketLookup_CustomerId:
		lookupType = "customer"
		log.Printf("Listing tickets for customer: %s", repository.SafeStringForLog(lookup.CustomerId))
		if !repository.IsValidUUID(lookup.CustomerId) {
			return nil, fmt.Errorf("invalid customer ID format: must be a valid UUID")
		}
		// Usar GetTicketsByCustomerID para clientes invitados
		tickets, err = s.TicketRepo.GetTicketsByCustomerID(ctx, lookup.CustomerId)

	default:
		return nil, fmt.Errorf("no valid lookup parameter provided. Use user_id or customer_id")
	}

	if err != nil {
		log.Printf("Error listing tickets for %s: %v", lookupType, err)
		return nil, fmt.Errorf("error querying tickets: %w", err)
	}

	// Obtener detalles completos de los tickets
	pbTickets := make([]*osmi.TicketResponse, 0, len(tickets))
	for _, ticket := range tickets {
		// Obtener detalles completos para cada ticket
		ticketDetails, err := s.TicketRepo.GetTicketWithDetails(ctx, ticket.PublicID)
		if err != nil {
			log.Printf("Warning: Could not get details for ticket %s: %v", ticket.PublicID, err)
			// Usar información básica si no se pueden obtener detalles
			pbTickets = append(pbTickets, &osmi.TicketResponse{
				TicketId:  ticket.PublicID,
				Status:    ticket.Status,
				Code:      ticket.Code,
				QrCodeUrl: safeStringPtr(ticket.QRCodeURL),
				CreatedAt: timestamppb.New(ticket.CreatedAt),
			})
		} else {
			// ✅ CORREGIDO: Usar category_name (NO Category)
			pbTickets = append(pbTickets, &osmi.TicketResponse{
				TicketId:     ticketDetails.TicketID,
				Status:       ticketDetails.Status,
				Code:         ticketDetails.Code,
				QrCodeUrl:    "",
				EventName:    ticketDetails.EventName,
				CategoryName: ticketDetails.CategoryName, // ✅ CORREGIDO: CategoryName
				Price:        ticketDetails.Price,
				CreatedAt:    timestamppb.New(ticketDetails.CreatedAt),
			})
		}
	}

	log.Printf("Found %d tickets for %s", len(pbTickets), lookupType)

	return &osmi.TicketListResponse{
		Tickets:    pbTickets,
		TotalCount: int32(len(pbTickets)),
	}, nil
}

// GetTicketDetails obtiene detalles completos de un ticket
func (s *Server) GetTicketDetails(ctx context.Context, req *osmi.TicketLookup) (*osmi.TicketResponse, error) {
	var ticketPublicID string

	// Manejar diferentes formas de buscar tickets
	switch lookup := req.Lookup.(type) {
	case *osmi.TicketLookup_TicketId:
		ticketPublicID = lookup.TicketId
		if !repository.IsValidUUID(ticketPublicID) {
			return nil, fmt.Errorf("invalid ticket ID format: must be a valid UUID")
		}
	default:
		return nil, fmt.Errorf("no valid ticket identifier provided")
	}

	// Obtener detalles completos del ticket
	ticketDetails, err := s.TicketRepo.GetTicketWithDetails(ctx, ticketPublicID)
	if err != nil {
		log.Printf("Error getting ticket details: %v", err)
		return nil, fmt.Errorf("error retrieving ticket details: %w", err)
	}

	// Crear timestamp para used_at si es válido
	var usedAt *timestamppb.Timestamp
	if ticketDetails.UsedAt != nil && !ticketDetails.UsedAt.IsZero() {
		usedAt = timestamppb.New(*ticketDetails.UsedAt)
	}

	// ✅ CORREGIDO: Usar event_date como string (NO Timestamp) y category_name
	return &osmi.TicketResponse{
		TicketId:          ticketDetails.TicketID,
		Status:            ticketDetails.Status,
		Code:              ticketDetails.Code,
		QrCodeUrl:         "",
		EventName:         ticketDetails.EventName,
		EventDate:         ticketDetails.StartDate.Format(time.RFC3339), // ✅ CORREGIDO: string, NO Timestamp
		EventLocation:     ticketDetails.Location,
		Price:             ticketDetails.Price,
		CategoryName:      ticketDetails.CategoryName, // ✅ CORREGIDO: CategoryName
		SeatNumber:        ticketDetails.SeatNumber,
		CustomerName:      ticketDetails.CustomerName,
		CustomerEmail:     ticketDetails.CustomerEmail,
		CustomerType:      ticketDetails.CustomerType,
		UserName:          safeStringPtr(ticketDetails.UserName),
		UserRole:          safeStringPtr(ticketDetails.UserRole),
		TransactionStatus: safeStringPtr(ticketDetails.TransactionStatus),
		CreatedAt:         timestamppb.New(ticketDetails.CreatedAt),
		UsedAt:            usedAt,
	}, nil
}

// UpdateTicketStatus actualiza el estado de un ticket
func (s *Server) UpdateTicketStatus(ctx context.Context, req *osmi.UpdateTicketStatusRequest) (*osmi.TicketResponse, error) {
	log.Printf("Updating ticket status: %s -> %s", repository.SafeStringForLog(req.TicketId), req.Status)

	// Validaciones
	if strings.TrimSpace(req.TicketId) == "" {
		return nil, fmt.Errorf("ticket_id is required")
	}
	if strings.TrimSpace(req.Status) == "" {
		return nil, fmt.Errorf("status is required")
	}

	// Validar UUID
	if !repository.IsValidUUID(req.TicketId) {
		return nil, fmt.Errorf("invalid ticket ID format: must be a valid UUID")
	}

	// Actualizar estado del ticket
	err := s.TicketRepo.UpdateTicketStatus(ctx, req.TicketId, req.Status)
	if err != nil {
		log.Printf("Error updating ticket status: %v", err)
		return nil, fmt.Errorf("error updating ticket status: %w", err)
	}

	// Obtener ticket actualizado
	ticketDetails, err := s.TicketRepo.GetTicketWithDetails(ctx, req.TicketId)
	if err != nil {
		log.Printf("Error retrieving updated ticket: %v", err)
		return nil, fmt.Errorf("ticket updated but retrieval failed: %w", err)
	}

	log.Printf("Ticket status updated successfully: %s -> %s",
		repository.SafeStringForLog(req.TicketId), req.Status)

	// ✅ CORREGIDO: Usar category_name (NO Category)
	return &osmi.TicketResponse{
		TicketId:     ticketDetails.TicketID,
		Status:       ticketDetails.Status,
		Code:         ticketDetails.Code,
		EventName:    ticketDetails.EventName,
		CategoryName: ticketDetails.CategoryName, // ✅ CORREGIDO: CategoryName
		Price:        ticketDetails.Price,
		CreatedAt:    timestamppb.New(ticketDetails.CreatedAt),
	}, nil
}

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
// MÉTODOS HELPER - CORREGIDOS
// =============================================================================

// mapEventToResponse mapea un modelo Event a protobuf
func (s *Server) mapEventToResponse(event *models.Event) *osmi.EventResponse {
	response := &osmi.EventResponse{
		PublicId:    event.PublicID,
		Name:        event.Name,
		StartDate:   event.StartDate.Format(time.RFC3339),
		EndDate:     event.EndDate.Format(time.RFC3339),
		Location:    event.Location,
		Tags:        event.Tags,
		IsActive:    event.IsActive,
		IsPublished: event.IsPublished,
		Status:      event.Status,
		CreatedAt:   timestamppb.New(event.CreatedAt),
		UpdatedAt:   timestamppb.New(event.UpdatedAt),
	}

	// Manejar campos opcionales como *string
	if event.Description != nil {
		response.Description = *event.Description
	}
	if event.ShortDescription != nil {
		response.ShortDescription = *event.ShortDescription
	}
	if event.VenueDetails != nil {
		response.VenueDetails = *event.VenueDetails
	}
	if event.Category != nil {
		response.Category = *event.Category
	}
	if event.ImageURL != nil {
		response.ImageUrl = *event.ImageURL
	}
	if event.BannerURL != nil {
		response.BannerUrl = *event.BannerURL
	}
	if event.MaxAttendees != nil {
		response.MaxAttendees = *event.MaxAttendees
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
