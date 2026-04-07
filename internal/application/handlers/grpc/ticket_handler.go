// internal/application/handlers/grpc/ticket_handler.go
package grpc

import (
	"context"
	"log"
	"strconv"

	osmi "github.com/franciscozamorau/osmi-protobuf/gen/pb"
	commondto "github.com/franciscozamorau/osmi-server/internal/api/dto/common"
	ticketdto "github.com/franciscozamorau/osmi-server/internal/api/dto/ticket"
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

// CreateTicket maneja la creación de tickets (venta directa)
func (h *TicketHandler) CreateTicket(ctx context.Context, req *osmi.CreateTicketRequest) (*osmi.TicketResponse, error) {
	if req.EventId == "" {
		return nil, status.Error(codes.InvalidArgument, "event_id is required")
	}
	if req.CustomerId == "" {
		return nil, status.Error(codes.InvalidArgument, "customer_id is required")
	}
	if req.TicketTypeId == "" {
		return nil, status.Error(codes.InvalidArgument, "ticket_type_id is required")
	}
	if req.Quantity <= 0 {
		return nil, status.Error(codes.InvalidArgument, "quantity must be greater than 0")
	}

	createReq := &ticketdto.CreateTicketRequest{
		EventID:      req.EventId,
		CustomerID:   req.CustomerId,
		TicketTypeID: req.TicketTypeId,
		Quantity:     req.Quantity,
		UserID:       req.UserId,
	}

	log.Printf("📦 Creando ticket con CustomerID: %q", createReq.CustomerID)

	ticket, err := h.ticketService.CreateTicket(ctx, createReq)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return h.ticketToProto(ticket), nil
}

// ReserveTicket maneja la reserva de tickets
func (h *TicketHandler) ReserveTicket(ctx context.Context, req *osmi.ReserveTicketRequest) (*osmi.TicketResponse, error) {
	// 🔥 ELIMINADO: validación de user_id (temporalmente)
	if req.TicketTypeId == "" {
		return nil, status.Error(codes.InvalidArgument, "ticket_type_id is required")
	}
	// 🔥 COMENTADO: if req.UserId == "" { return nil, status.Error(codes.InvalidArgument, "user_id is required") }

	reserveReq := &ticketdto.ReserveTicketRequest{
		TicketID:  req.TicketTypeId,
		UserID:    req.UserId, // Puede estar vacío
		ExpiresAt: req.ExpiresAt.AsTime(),
	}

	ticket, err := h.ticketService.ReserveTicket(ctx, reserveReq)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return h.ticketToProto(ticket), nil
}

// PurchaseTicket maneja la compra de un ticket reservado
func (h *TicketHandler) PurchaseTicket(ctx context.Context, req *osmi.PurchaseTicketRequest) (*osmi.TicketResponse, error) {
	if req.TicketId == "" {
		return nil, status.Error(codes.InvalidArgument, "ticket_id is required")
	}
	if req.CustomerId == "" {
		return nil, status.Error(codes.InvalidArgument, "customer_id is required")
	}

	purchaseReq := &ticketdto.PurchaseTicketRequest{
		TicketID:   req.TicketId,
		CustomerID: req.CustomerId,
	}

	ticket, err := h.ticketService.PurchaseTicket(ctx, purchaseReq)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// 🔥 CAMBIADO: usar ticketToProto en lugar de respuesta manual
	return h.ticketToProto(ticket), nil
}

// CheckInTicket maneja el check-in de tickets
func (h *TicketHandler) CheckInTicket(ctx context.Context, req *osmi.CheckInTicketRequest) (*osmi.TicketResponse, error) {
	if req.TicketId == "" {
		return nil, status.Error(codes.InvalidArgument, "ticket_id is required")
	}
	// 🔥 COMENTADO: validación de checked_by (temporalmente)
	// if req.CheckedBy == "" {
	//     return nil, status.Error(codes.InvalidArgument, "checked_by is required")
	// }

	checkinReq := &ticketdto.CheckInTicketRequest{
		TicketID:  req.TicketId,
		CheckedBy: req.CheckedBy, // Puede estar vacío
		Method:    req.Method,
		Location:  req.Location,
	}

	ticket, err := h.ticketService.CheckInTicket(ctx, checkinReq)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return h.ticketToProto(ticket), nil
}

// TransferTicket maneja la transferencia de tickets
func (h *TicketHandler) TransferTicket(ctx context.Context, req *osmi.TransferTicketRequest) (*osmi.TicketResponse, error) {
	if req.TicketId == "" {
		return nil, status.Error(codes.InvalidArgument, "ticket_id is required")
	}
	// 🔥 COMENTADO: validación de from_customer_id (temporalmente)
	// if req.FromCustomerId == "" {
	//     return nil, status.Error(codes.InvalidArgument, "from_customer_id is required")
	// }
	if req.ToCustomerId == "" {
		return nil, status.Error(codes.InvalidArgument, "to_customer_id is required")
	}

	transferReq := &ticketdto.TransferTicketRequest{
		TicketID:       req.TicketId,
		FromCustomerID: req.FromCustomerId, // Puede estar vacío
		ToCustomerID:   req.ToCustomerId,
		Token:          req.Token,
	}

	ticket, err := h.ticketService.TransferTicket(ctx, transferReq)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return h.ticketToProto(ticket), nil
}

// UpdateTicket actualiza información de un ticket
func (h *TicketHandler) UpdateTicket(ctx context.Context, req *osmi.UpdateTicketRequest) (*osmi.TicketResponse, error) {
	if req.TicketId == "" {
		return nil, status.Error(codes.InvalidArgument, "ticket_id is required")
	}

	updateReq := &ticketdto.UpdateTicketRequest{
		AttendeeName:  req.AttendeeName,
		AttendeeEmail: req.AttendeeEmail,
		AttendeePhone: req.AttendeePhone,
		Status:        req.Status,
	}

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

	ticket, err := h.ticketService.GetTicket(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return h.ticketToProto(ticket), nil
}

// ListTickets lista tickets con filtros y paginación
func (h *TicketHandler) ListTickets(ctx context.Context, req *osmi.ListTicketsRequest) (*osmi.TicketListResponse, error) {
	filter := &ticketdto.TicketFilter{
		Status:   req.Status,
		DateFrom: req.DateFrom,
		DateTo:   req.DateTo,
	}

	pagination := commondto.Pagination{
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
	}
	if pagination.Page <= 0 {
		pagination.Page = 1
	}
	if pagination.PageSize <= 0 {
		pagination.PageSize = 20
	}

	tickets, total, err := h.ticketService.ListTickets(ctx, filter, pagination)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbTickets := make([]*osmi.TicketResponse, 0, len(tickets))
	for _, ticket := range tickets {
		pbTickets = append(pbTickets, h.ticketToProto(ticket))
	}

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

	stats, err := h.ticketService.GetTicketStats(ctx, req.EventId)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

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

// GetUserTickets obtiene tickets de un usuario
func (h *TicketHandler) GetUserTickets(ctx context.Context, req *osmi.GetUserTicketsRequest) (*osmi.TicketListResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	log.Printf("GetUserTickets llamado para user_id: %s (pendiente de implementación)", req.UserId)
	return &osmi.TicketListResponse{
		Tickets:    []*osmi.TicketResponse{},
		TotalCount: 0,
		Page:       1,
		PageSize:   20,
		TotalPages: 0,
	}, nil
}

// GetCustomerTickets obtiene tickets de un cliente
func (h *TicketHandler) GetCustomerTickets(ctx context.Context, req *osmi.GetCustomerTicketsRequest) (*osmi.TicketListResponse, error) {
	if req.PublicId == "" {
		return nil, status.Error(codes.InvalidArgument, "customer public_id is required")
	}
	log.Printf("GetCustomerTickets llamado para customer_id: %s (pendiente de implementación)", req.PublicId)
	return &osmi.TicketListResponse{
		Tickets:    []*osmi.TicketResponse{},
		TotalCount: 0,
		Page:       1,
		PageSize:   20,
		TotalPages: 0,
	}, nil
}

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
		EventName:     "",
		EventDate:     "",
		Location:      "",
		Price:         ticket.FinalPrice,
		CategoryName:  "",
		SeatNumber:    "",
		CustomerName:  "",
		CustomerEmail: "",
		UserName:      "",
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
