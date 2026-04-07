// internal/application/handlers/grpc/handler.go
package grpc

import (
	"context"
	"log"

	osmi "github.com/franciscozamorau/osmi-protobuf/gen/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Handler unificado que implementa la interfaz completa de OsmiServiceServer
type Handler struct {
	osmi.UnimplementedOsmiServiceServer
	customerHandler   *CustomerHandler
	ticketHandler     *TicketHandler
	userHandler       *UserHandler
	eventHandler      *EventHandler
	categoryHandler   *CategoryHandler
	ticketTypeHandler *TicketTypeHandler
}

func NewHandler(
	customerHandler *CustomerHandler,
	ticketHandler *TicketHandler,
	userHandler *UserHandler,
	eventHandler *EventHandler,
	categoryHandler *CategoryHandler,
	ticketTypeHandler *TicketTypeHandler,
) *Handler {
	return &Handler{
		customerHandler:   customerHandler,
		ticketHandler:     ticketHandler,
		userHandler:       userHandler,
		eventHandler:      eventHandler,
		categoryHandler:   categoryHandler,
		ticketTypeHandler: ticketTypeHandler,
	}
}

// ============ TICKET TYPES ============
func (h *Handler) CreateTicketType(ctx context.Context, req *osmi.CreateTicketTypeRequest) (*osmi.TicketTypeResponse, error) {
	return h.ticketTypeHandler.CreateTicketType(ctx, req)
}

func (h *Handler) GetTicketType(ctx context.Context, req *osmi.GetTicketTypeRequest) (*osmi.TicketTypeResponse, error) {
	return h.ticketTypeHandler.GetTicketType(ctx, req)
}

func (h *Handler) ListTicketTypes(ctx context.Context, req *osmi.ListTicketTypesRequest) (*osmi.TicketTypeListResponse, error) {
	return h.ticketTypeHandler.ListTicketTypes(ctx, req)
}

func (h *Handler) UpdateTicketType(ctx context.Context, req *osmi.UpdateTicketTypeRequest) (*osmi.TicketTypeResponse, error) {
	return h.ticketTypeHandler.UpdateTicketType(ctx, req)
}

func (h *Handler) DeleteTicketType(ctx context.Context, req *osmi.DeleteTicketTypeRequest) (*osmi.Empty, error) {
	return h.ticketTypeHandler.DeleteTicketType(ctx, req)
}

// ============ CATEGORIES ============
func (h *Handler) CreateCategory(ctx context.Context, req *osmi.CreateCategoryRequest) (*osmi.CategoryResponse, error) {
	return h.categoryHandler.CreateCategory(ctx, req)
}

// CORREGIDO: Ahora usa GetEventCategoriesRequest
func (h *Handler) GetEventCategories(ctx context.Context, req *osmi.GetEventCategoriesRequest) (*osmi.CategoryListResponse, error) {
	return h.categoryHandler.GetEventCategories(ctx, req)
}

// ============ CUSTOMERS ============
func (h *Handler) CreateCustomer(ctx context.Context, req *osmi.CreateCustomerRequest) (*osmi.CustomerResponse, error) {
	return h.customerHandler.CreateCustomer(ctx, req)
}

// CORREGIDO: Ahora usa GetCustomerRequest
func (h *Handler) GetCustomer(ctx context.Context, req *osmi.GetCustomerRequest) (*osmi.CustomerResponse, error) {
	return h.customerHandler.GetCustomer(ctx, req)
}

func (h *Handler) UpdateCustomer(ctx context.Context, req *osmi.UpdateCustomerRequest) (*osmi.CustomerResponse, error) {
	return h.customerHandler.UpdateCustomer(ctx, req)
}

func (h *Handler) ListCustomers(ctx context.Context, req *osmi.ListCustomersRequest) (*osmi.CustomerListResponse, error) {
	return h.customerHandler.ListCustomers(ctx, req)
}

func (h *Handler) GetCustomerStats(ctx context.Context, req *osmi.Empty) (*osmi.CustomerStatsResponse, error) {
	return h.customerHandler.GetCustomerStats(ctx, req)
}

// CORREGIDO: Ahora usa GetCustomerTicketsRequest
func (h *Handler) GetCustomerTickets(ctx context.Context, req *osmi.GetCustomerTicketsRequest) (*osmi.TicketListResponse, error) {
	return h.ticketHandler.GetCustomerTickets(ctx, req)
}

// ============ TICKETS ============
func (h *Handler) CreateTicket(ctx context.Context, req *osmi.CreateTicketRequest) (*osmi.TicketResponse, error) {
	return h.ticketHandler.CreateTicket(ctx, req)
}

func (h *Handler) ReserveTicket(ctx context.Context, req *osmi.ReserveTicketRequest) (*osmi.TicketResponse, error) {
	return h.ticketHandler.ReserveTicket(ctx, req)
}

// PurchaseTicket maneja la compra de un ticket reservado
func (h *Handler) PurchaseTicket(ctx context.Context, req *osmi.PurchaseTicketRequest) (*osmi.TicketResponse, error) {
	return h.ticketHandler.PurchaseTicket(ctx, req)
}

func (h *Handler) CheckInTicket(ctx context.Context, req *osmi.CheckInTicketRequest) (*osmi.TicketResponse, error) {
	return h.ticketHandler.CheckInTicket(ctx, req)
}

func (h *Handler) TransferTicket(ctx context.Context, req *osmi.TransferTicketRequest) (*osmi.TicketResponse, error) {
	return h.ticketHandler.TransferTicket(ctx, req)
}

func (h *Handler) ListTickets(ctx context.Context, req *osmi.ListTicketsRequest) (*osmi.TicketListResponse, error) {
	return h.ticketHandler.ListTickets(ctx, req)
}

// CORREGIDO: Ahora usa GetUserTicketsRequest
func (h *Handler) GetUserTickets(ctx context.Context, req *osmi.GetUserTicketsRequest) (*osmi.TicketListResponse, error) {
	return h.ticketHandler.GetUserTickets(ctx, req)
}

func (h *Handler) UpdateTicketStatus(ctx context.Context, req *osmi.UpdateTicketStatusRequest) (*osmi.TicketResponse, error) {
	return h.ticketHandler.UpdateTicketStatus(ctx, req)
}

func (h *Handler) UpdateTicket(ctx context.Context, req *osmi.UpdateTicketRequest) (*osmi.TicketResponse, error) {
	return h.ticketHandler.UpdateTicket(ctx, req)
}

func (h *Handler) GetTicketDetails(ctx context.Context, req *osmi.GetTicketRequest) (*osmi.TicketResponse, error) {
	return h.ticketHandler.GetTicket(ctx, req)
}

func (h *Handler) GetTicketStats(ctx context.Context, req *osmi.GetTicketStatsRequest) (*osmi.TicketStatsResponse, error) {
	return h.ticketHandler.GetTicketStats(ctx, req)
}

// ============ USERS ============
func (h *Handler) CreateUser(ctx context.Context, req *osmi.CreateUserRequest) (*osmi.UserResponse, error) {
	return h.userHandler.CreateUser(ctx, req)
}

// CORREGIDO: Ahora usa GetUserRequest
func (h *Handler) GetUser(ctx context.Context, req *osmi.GetUserRequest) (*osmi.UserResponse, error) {
	return h.userHandler.GetUser(ctx, req)
}

func (h *Handler) UpdateUser(ctx context.Context, req *osmi.UpdateUserRequest) (*osmi.UserResponse, error) {
	return h.userHandler.UpdateUser(ctx, req)
}

func (h *Handler) DeleteUser(ctx context.Context, req *osmi.DeleteUserRequest) (*osmi.Empty, error) {
	return h.userHandler.DeleteUser(ctx, req)
}

func (h *Handler) Login(ctx context.Context, req *osmi.LoginRequest) (*osmi.LoginResponse, error) {
	return h.userHandler.Login(ctx, req)
}

func (h *Handler) Logout(ctx context.Context, req *osmi.LogoutRequest) (*osmi.Empty, error) {
	return h.userHandler.Logout(ctx, req)
}

func (h *Handler) RefreshToken(ctx context.Context, req *osmi.RefreshTokenRequest) (*osmi.RefreshTokenResponse, error) {
	return h.userHandler.RefreshToken(ctx, req)
}

// ============ EVENTS ============
func (h *Handler) CreateEvent(ctx context.Context, req *osmi.CreateEventRequest) (*osmi.EventResponse, error) {
	log.Println("✅ Handler.CreateEvent llamado")
	return h.eventHandler.CreateEvent(ctx, req)
}

// CORREGIDO: Ahora usa GetEventRequest
func (h *Handler) GetEvent(ctx context.Context, req *osmi.GetEventRequest) (*osmi.EventResponse, error) {
	return h.eventHandler.GetEvent(ctx, req)
}

func (h *Handler) ListEvents(ctx context.Context, req *osmi.ListEventsRequest) (*osmi.EventListResponse, error) {
	return h.eventHandler.ListEvents(ctx, req)
}

func (h *Handler) UpdateEvent(ctx context.Context, req *osmi.UpdateEventRequest) (*osmi.EventResponse, error) {
	return h.eventHandler.UpdateEvent(ctx, req)
}

// ============ HEALTH ============
func (h *Handler) HealthCheck(ctx context.Context, req *osmi.Empty) (*osmi.HealthResponse, error) {
	log.Println("✅ HealthCheck llamado")
	return &osmi.HealthResponse{
		Status:    "healthy",
		Service:   "osmi-server",
		Version:   "1.0.0",
		Timestamp: timestamppb.Now(),
	}, nil
}
