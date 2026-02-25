// internal/application/handlers/grpc/ticket_handler.go
package grpc

import (
	"context"
	"strconv"

	osmi "github.com/franciscozamorau/osmi-protobuf/gen/pb"
	"github.com/franciscozamorau/osmi-server/internal/api/dto"
	"github.com/franciscozamorau/osmi-server/internal/api/helpers"
	"github.com/franciscozamorau/osmi-server/internal/application/services"
	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type TicketHandler struct {
	osmi.UnimplementedOsmiServiceServer
	ticketService *services.TicketService
}

func NewTicketHandler(ticketService *services.TicketService) *TicketHandler {
	return &TicketHandler{
		ticketService: ticketService,
	}
}

// ============================================================================
// MÉTODOS PRINCIPALES
// ============================================================================

// CreateTicket maneja la creación de tickets
func (h *TicketHandler) CreateTicket(ctx context.Context, req *osmi.CreateTicketRequest) (*osmi.TicketResponse, error) {
	// Convertir protobuf a DTO de solicitud
	createReq := &dto.CreateTicketRequest{
		EventID:    req.EventId,
		UserID:     req.UserId,
		CategoryID: req.CategoryId,
		Quantity:   req.Quantity,
		// CustomerID se obtendrá del contexto o se asignará después
	}

	// Llamar al servicio
	ticket, err := h.ticketService.CreateTicket(ctx, createReq)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return h.ticketToProto(ticket), nil
}

// ReserveTicket maneja la reserva de tickets
func (h *TicketHandler) ReserveTicket(ctx context.Context, req *osmi.ReserveTicketRequest) (*osmi.TicketResponse, error) {
	// Convertir protobuf a DTO - usando los campos que existen en el proto
	reserveReq := &dto.ReserveTicketRequest{
		TicketID:  req.TicketTypeId, // Mapeo: TicketTypeId del proto → TicketID del DTO
		UserID:    req.UserId,
		ExpiresAt: req.ExpiresAt.AsTime(),
	}

	// Llamar al servicio
	ticket, err := h.ticketService.ReserveTicket(ctx, reserveReq)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return h.ticketToProto(ticket), nil
}

// CheckInTicket maneja el check-in de tickets
func (h *TicketHandler) CheckInTicket(ctx context.Context, req *osmi.CheckInTicketRequest) (*osmi.TicketResponse, error) {
	// Convertir protobuf a DTO
	checkinReq := &dto.CheckInTicketRequest{
		TicketID:  req.TicketId,
		CheckedBy: req.CheckedBy,
		Method:    req.Method,
		Location:  req.Location,
	}

	// Llamar al servicio
	ticket, err := h.ticketService.CheckInTicket(ctx, checkinReq)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return h.ticketToProto(ticket), nil
}

// TransferTicket maneja la transferencia de tickets
func (h *TicketHandler) TransferTicket(ctx context.Context, req *osmi.TransferTicketRequest) (*osmi.TicketResponse, error) {
	// Convertir protobuf a DTO
	transferReq := &dto.TransferTicketRequest{
		TicketID:       req.TicketId,
		FromCustomerID: req.FromCustomerId,
		ToCustomerID:   req.ToCustomerId,
		Token:          req.Token,
	}

	// Llamar al servicio
	ticket, err := h.ticketService.TransferTicket(ctx, transferReq)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return h.ticketToProto(ticket), nil
}

// UpdateTicket actualiza información de un ticket
func (h *TicketHandler) UpdateTicket(ctx context.Context, req *osmi.UpdateTicketRequest) (*osmi.TicketResponse, error) {
	// Validar que se proporcione el ID
	if req.TicketId == "" {
		return nil, status.Error(codes.InvalidArgument, "ticket_id is required")
	}

	// Convertir protobuf a DTO
	updateReq := &dto.UpdateTicketRequest{
		AttendeeName:  req.AttendeeName,
		AttendeeEmail: req.AttendeeEmail,
		AttendeePhone: req.AttendeePhone,
		Status:        req.Status,
	}

	// Llamar al servicio
	ticket, err := h.ticketService.UpdateTicket(ctx, req.TicketId, updateReq)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return h.ticketToProto(ticket), nil
}

// GetTicket obtiene un ticket por ID
func (h *TicketHandler) GetTicket(ctx context.Context, req *osmi.GetTicketRequest) (*osmi.TicketResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "ticket id is required")
	}

	// Llamar al servicio
	ticket, err := h.ticketService.GetTicket(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return h.ticketToProto(ticket), nil
}

// ListTickets lista tickets con filtros y paginación
func (h *TicketHandler) ListTickets(ctx context.Context, req *osmi.ListTicketsRequest) (*osmi.TicketListResponse, error) {
	// Convertir filtros del protobuf a DTO TicketFilter
	filter := &dto.TicketFilter{
		EventID:    nil, // Necesitarías convertir string a *int64
		CustomerID: nil, // Necesitarías convertir string a *int64
		Status:     req.Status,
		DateFrom:   req.DateFrom,
		DateTo:     req.DateTo,
	}

	// Paginación
	pagination := dto.Pagination{
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
	}
	if pagination.Page <= 0 {
		pagination.Page = 1
	}
	if pagination.PageSize <= 0 {
		pagination.PageSize = 20
	}

	// Llamar al servicio
	tickets, total, err := h.ticketService.ListTickets(ctx, filter, pagination)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Convertir entidades a protobuf
	pbTickets := make([]*osmi.TicketResponse, 0, len(tickets))
	for _, ticket := range tickets {
		pbTickets = append(pbTickets, h.ticketToProto(ticket))
	}

	// Calcular total de páginas
	totalPages := int32(0)
	if pagination.PageSize > 0 {
		totalPages = int32((int(total) + pagination.PageSize - 1) / pagination.PageSize)
	}

	return &osmi.TicketListResponse{
		Tickets:    pbTickets,
		TotalCount: int32(total),
		Page:       int32(pagination.Page),
		PageSize:   int32(pagination.PageSize),
		TotalPages: totalPages,
	}, nil
}

// GetTicketStats obtiene estadísticas de tickets para un evento
func (h *TicketHandler) GetTicketStats(ctx context.Context, req *osmi.GetTicketStatsRequest) (*osmi.TicketStatsResponse, error) {
	if req.EventId == "" {
		return nil, status.Error(codes.InvalidArgument, "event_id is required")
	}

	// Llamar al servicio
	stats, err := h.ticketService.GetTicketStats(ctx, req.EventId)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	// Convertir estadísticas a protobuf
	return &osmi.TicketStatsResponse{
		TotalTickets:     stats.TotalTickets,
		AvailableTickets: stats.AvailableTickets,
		SoldTickets:      stats.SoldTickets,
		ReservedTickets:  stats.ReservedTickets,
		CheckedInTickets: stats.CheckedInTickets,
		CancelledTickets: stats.CancelledTickets,
		RefundedTickets:  stats.RefundedTickets,
		TotalRevenue:     stats.TotalRevenue,
		AvgTicketPrice:   stats.AvgTicketPrice,
		CheckInRate:      stats.CheckInRate,
	}, nil
}

// ============================================================================
// MÉTODOS DE CONSULTA ESPECÍFICOS (COMENtADOS HASTA IMPLEMENTAR EN EL PROTO)
// ============================================================================

// Estos métodos están comentados porque los mensajes correspondientes
// no existen en el proto actual. Cuando se agreguen, se pueden descomentar.

/*
func (h *TicketHandler) GetUserTickets(ctx context.Context, req *osmi.UserLookup) (*osmi.TicketListResponse, error) {
	return nil, status.Error(codes.Unimplemented, "GetUserTickets not implemented in proto")
}

func (h *TicketHandler) GetCustomerTickets(ctx context.Context, req *osmi.CustomerLookup) (*osmi.TicketListResponse, error) {
	return nil, status.Error(codes.Unimplemented, "GetCustomerTickets not implemented in proto")
}
*/

// ============================================================================
// FUNCIÓN HELPER PARA CONVERSIÓN
// ============================================================================

// ticketToProto convierte una entidad Ticket a protobuf TicketResponse
// ticketToProto convierte una entidad Ticket a protobuf TicketResponse
func (h *TicketHandler) ticketToProto(ticket *entities.Ticket) *osmi.TicketResponse {
	if ticket == nil {
		return nil
	}

	return &osmi.TicketResponse{
		TicketId:      ticket.PublicID,
		Status:        ticket.Status,
		Code:          ticket.Code,
		QrCodeUrl:     helpers.SafeStringPtr(ticket.QRCodeData),
		EventName:     "", // Requiere consulta adicional
		EventDate:     "", // Requiere consulta adicional
		Location:      "", // Requiere consulta adicional
		Price:         ticket.FinalPrice,
		CategoryName:  "", // Requiere consulta adicional
		SeatNumber:    "", // No disponible en la entidad actual
		CustomerName:  "", // Requiere consulta adicional
		CustomerEmail: "", // Requiere consulta adicional
		UserName:      "", // Requiere consulta adicional
		CreatedAt:     timestamppb.New(ticket.CreatedAt),
		UsedAt:        helpers.SafeTimePtr(ticket.CheckedInAt),
	}
}

// safeStringID convierte un *int64 a string
func safeStringID(id *int64) string {
	if id == nil {
		return ""
	}
	return strconv.FormatInt(*id, 10)
}
