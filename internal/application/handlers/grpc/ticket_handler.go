package grpchandlers

import (
	"context"

	osmi "github.com/franciscozamorau/osmi-protobuf/gen/pb"
	"github.com/franciscozamorau/osmi-server/internal/api/dto"
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

func (h *TicketHandler) CreateTicket(ctx context.Context, req *osmi.TicketRequest) (*osmi.TicketResponse, error) {
	// Convertir protobuf a DTO
	createReq := &dto.CreateTicketRequest{
		TicketTypeID:  req.TicketTypeId,
		CustomerID:    req.CustomerId,
		OrderID:       req.OrderId,
		AttendeeName:  req.AttendeeName,
		AttendeeEmail: req.AttendeeEmail,
		AttendeePhone: req.AttendeePhone,
	}

	// Llamar al servicio
	ticket, err := h.ticketService.CreateTicket(ctx, createReq)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Convertir entidad a protobuf
	return h.ticketToProto(ticket), nil
}

func (h *TicketHandler) ReserveTicket(ctx context.Context, req *osmi.ReserveTicketRequest) (*osmi.TicketResponse, error) {
	// Convertir protobuf a DTO
	reserveReq := &dto.ReserveTicketRequest{
		TicketTypeID:    req.TicketTypeId,
		CustomerID:      req.CustomerId,
		Quantity:        int(req.Quantity),
		DurationMinutes: int(req.DurationMinutes),
	}

	// Llamar al servicio
	ticket, err := h.ticketService.ReserveTicket(ctx, reserveReq)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Convertir entidad a protobuf
	return h.ticketToProto(ticket), nil
}

func (h *TicketHandler) CheckInTicket(ctx context.Context, req *osmi.CheckInTicketRequest) (*osmi.TicketResponse, error) {
	// Convertir protobuf a DTO
	checkinReq := &dto.CheckInTicketRequest{
		TicketID:        req.TicketId,
		CheckinMethod:   req.CheckinMethod,
		CheckinLocation: req.CheckinLocation,
		ValidatorID:     req.ValidatorId,
	}

	// Llamar al servicio
	ticket, err := h.ticketService.CheckInTicket(ctx, checkinReq)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Convertir entidad a protobuf
	return h.ticketToProto(ticket), nil
}

func (h *TicketHandler) TransferTicket(ctx context.Context, req *osmi.TransferTicketRequest) (*osmi.TicketResponse, error) {
	// Convertir protobuf a DTO
	transferReq := &dto.TransferTicketRequest{
		TicketID:       req.TicketId,
		FromCustomerID: req.FromCustomerId,
		ToCustomerID:   req.ToCustomerId,
		ToEmail:        req.ToEmail,
		ToName:         req.ToName,
	}

	// Llamar al servicio
	ticket, err := h.ticketService.TransferTicket(ctx, transferReq)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Convertir entidad a protobuf
	return h.ticketToProto(ticket), nil
}

func (h *TicketHandler) GetTicket(ctx context.Context, req *osmi.TicketLookup) (*osmi.TicketResponse, error) {
	var ticketID string

	// Manejar diferentes formas de búsqueda
	switch lookup := req.Lookup.(type) {
	case *osmi.TicketLookup_TicketId:
		ticketID = lookup.TicketId
	case *osmi.TicketLookup_Code:
		// TODO: Implementar búsqueda por código
		return nil, status.Error(codes.Unimplemented, "search by code not implemented")
	default:
		return nil, status.Error(codes.InvalidArgument, "no valid lookup provided")
	}

	// Llamar al servicio
	ticket, err := h.ticketService.GetTicket(ctx, ticketID)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	// Convertir entidad a protobuf
	return h.ticketToProto(ticket), nil
}

func (h *TicketHandler) UpdateTicket(ctx context.Context, req *osmi.UpdateTicketRequest) (*osmi.TicketResponse, error) {
	// Convertir protobuf a DTO
	updateReq := &dto.UpdateTicketRequest{
		Status:        req.Status,
		AttendeeName:  req.AttendeeName,
		AttendeeEmail: req.AttendeeEmail,
		AttendeePhone: req.AttendeePhone,
	}

	// Llamar al servicio
	ticket, err := h.ticketService.UpdateTicket(ctx, req.TicketId, updateReq)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Convertir entidad a protobuf
	return h.ticketToProto(ticket), nil
}

func (h *TicketHandler) ListTickets(ctx context.Context, req *osmi.ListRequest) (*osmi.TicketListResponse, error) {
	// Convertir protobuf a DTO de filtro
	filter := dto.TicketFilter{
		EventID:      req.EventId,
		CustomerID:   req.CustomerId,
		OrderID:      req.OrderId,
		Status:       req.Status,
		TicketTypeID: req.TicketTypeId,
		Code:         req.Code,
		DateFrom:     req.DateFrom,
		DateTo:       req.DateTo,
		CheckedIn:    req.CheckedIn,
	}

	// Convertir paginación
	pagination := dto.Pagination{
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
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

	return &osmi.TicketListResponse{
		Tickets:    pbTickets,
		TotalCount: int32(total),
		Page:       int32(pagination.Page),
		PageSize:   int32(pagination.PageSize),
		TotalPages: int32((total + int64(pagination.PageSize) - 1) / int64(pagination.PageSize)),
	}, nil
}

func (h *TicketHandler) GetTicketStats(ctx context.Context, req *osmi.EventLookup) (*osmi.TicketStatsResponse, error) {
	// Llamar al servicio
	stats, err := h.ticketService.GetTicketStats(ctx, req.PublicId)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	// Convertir estadísticas a protobuf
	return &osmi.TicketStatsResponse{
		TotalTickets:     int32(stats.TotalTickets),
		AvailableTickets: int32(stats.AvailableTickets),
		ReservedTickets:  int32(stats.ReservedTickets),
		SoldTickets:      int32(stats.SoldTickets),
		CheckedInTickets: int32(stats.CheckedInTickets),
		CancelledTickets: int32(stats.CancelledTickets),
		RefundedTickets:  int32(stats.RefundedTickets),
		TotalRevenue:     stats.TotalRevenue,
		CheckInRate:      stats.CheckInRate,
		AvgTicketPrice:   stats.AvgTicketPrice,
	}, nil
}

// Helper function para convertir ticket a protobuf
func (h *TicketHandler) ticketToProto(ticket *entities.Ticket) *osmi.TicketResponse {
	return &osmi.TicketResponse{
		Id:                   ticket.PublicID,
		TicketTypeId:         "", // TODO: Obtener public ID del tipo de ticket
		EventId:              "", // TODO: Obtener public ID del evento
		CustomerId:           safeStringID(ticket.CustomerID),
		OrderId:              safeStringID(ticket.OrderID),
		Code:                 ticket.Code,
		QrCodeData:           safeStringPtr(ticket.QRCodeData),
		Status:               ticket.Status,
		FinalPrice:           ticket.FinalPrice,
		Currency:             ticket.Currency,
		TaxAmount:            ticket.TaxAmount,
		AttendeeName:         safeStringPtr(ticket.AttendeeName),
		AttendeeEmail:        safeStringPtr(ticket.AttendeeEmail),
		AttendeePhone:        safeStringPtr(ticket.AttendeePhone),
		CheckedInAt:          safeTimeProto(ticket.CheckedInAt),
		CheckinMethod:        safeStringPtr(ticket.CheckinMethod),
		CheckinLocation:      safeStringPtr(ticket.CheckinLocation),
		ReservedAt:           safeTimeProto(ticket.ReservedAt),
		ReservationExpiresAt: safeTimeProto(ticket.ReservationExpiresAt),
		TransferToken:        safeStringPtr(ticket.TransferToken),
		TransferredFrom:      safeStringID(ticket.TransferredFrom),
		TransferredAt:        safeTimeProto(ticket.TransferredAt),
		ValidationCount:      int32(ticket.ValidationCount),
		LastValidatedAt:      safeTimeProto(ticket.LastValidatedAt),
		SoldAt:               safeTimeProto(ticket.SoldAt),
		CancelledAt:          safeTimeProto(ticket.CancelledAt),
		RefundedAt:           safeTimeProto(ticket.RefundedAt),
		CreatedAt:            timestamppb.New(ticket.CreatedAt),
		UpdatedAt:            timestamppb.New(ticket.UpdatedAt),
	}
}
