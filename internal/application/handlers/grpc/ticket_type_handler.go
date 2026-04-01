// internal/application/handlers/grpc/ticket_type_handler.go
package grpc

import (
	"context"
	"time"

	osmi "github.com/franciscozamorau/osmi-protobuf/gen/pb"
	tickettypedto "github.com/franciscozamorau/osmi-server/internal/api/dto/ticket_type"
	"github.com/franciscozamorau/osmi-server/internal/application/services"
	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type TicketTypeHandler struct {
	osmi.UnimplementedOsmiServiceServer
	ticketTypeService *services.TicketTypeService
}

func NewTicketTypeHandler(ticketTypeService *services.TicketTypeService) *TicketTypeHandler {
	return &TicketTypeHandler{
		ticketTypeService: ticketTypeService,
	}
}

// CreateTicketType maneja la creación de un tipo de ticket
func (h *TicketTypeHandler) CreateTicketType(ctx context.Context, req *osmi.CreateTicketTypeRequest) (*osmi.TicketTypeResponse, error) {
	if req.EventId == "" {
		return nil, status.Error(codes.InvalidArgument, "event_id is required")
	}
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if req.BasePrice <= 0 {
		return nil, status.Error(codes.InvalidArgument, "base_price must be greater than 0")
	}
	if req.TotalQuantity <= 0 {
		return nil, status.Error(codes.InvalidArgument, "total_quantity must be greater than 0")
	}

	saleStartsAt := req.SaleStartsAt.AsTime()
	var saleEndsAt *time.Time
	if req.SaleEndsAt != nil {
		t := req.SaleEndsAt.AsTime()
		saleEndsAt = &t
	}

	createReq := &tickettypedto.CreateTicketTypeRequest{
		EventID:          req.EventId,
		Name:             req.Name,
		Description:      req.Description,
		TicketClass:      req.TicketClass,
		BasePrice:        req.BasePrice,
		Currency:         req.Currency,
		TaxRate:          req.TaxRate,
		ServiceFeeType:   req.ServiceFeeType,
		ServiceFeeValue:  req.ServiceFeeValue,
		TotalQuantity:    int(req.TotalQuantity),
		MaxPerOrder:      int(req.MaxPerOrder),
		MinPerOrder:      int(req.MinPerOrder),
		SaleStartsAt:     saleStartsAt.Format(time.RFC3339),
		IsActive:         req.IsActive,
		RequiresApproval: req.RequiresApproval,
		IsHidden:         req.IsHidden,
		SalesChannel:     req.SalesChannel,
		AccessType:       req.AccessType,
		Benefits:         "",
		ValidationRules:  req.ValidationRules,
	}

	if saleEndsAt != nil {
		createReq.SaleEndsAt = saleEndsAt.Format(time.RFC3339)
	}

	ticketType, err := h.ticketTypeService.CreateTicketType(ctx, createReq)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return h.ticketTypeToProto(ticketType), nil
}

// GetTicketType obtiene un tipo de ticket por ID
func (h *TicketTypeHandler) GetTicketType(ctx context.Context, req *osmi.GetTicketTypeRequest) (*osmi.TicketTypeResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "ticket type id is required")
	}

	if _, err := uuid.Parse(req.Id); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid ticket type id format")
	}

	ticketType, err := h.ticketTypeService.GetTicketType(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, "ticket type not found")
	}

	return h.ticketTypeToProto(ticketType), nil
}

// ListTicketTypes lista tipos de ticket con filtros
func (h *TicketTypeHandler) ListTicketTypes(ctx context.Context, req *osmi.ListTicketTypesRequest) (*osmi.TicketTypeListResponse, error) {
	filter := &tickettypedto.TicketTypeFilter{}

	if req.EventId != "" {
		// Necesitarías convertir el event_id a int64
	}

	if req.IsActive {
		active := true
		filter.IsActive = &active
	}

	page := int(req.Page)
	if page <= 0 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize <= 0 {
		pageSize = 20
	}

	ticketTypes, total, err := h.ticketTypeService.ListTicketTypes(ctx, filter, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbTicketTypes := make([]*osmi.TicketTypeResponse, len(ticketTypes))
	for i, tt := range ticketTypes {
		pbTicketTypes[i] = h.ticketTypeToProto(tt)
	}

	totalPages := int32(0)
	if pageSize > 0 {
		totalPages = int32((int(total) + pageSize - 1) / pageSize)
	}

	return &osmi.TicketTypeListResponse{
		TicketTypes: pbTicketTypes,
		TotalCount:  int32(total),
		Page:        int32(page),
		PageSize:    int32(pageSize),
		TotalPages:  totalPages,
	}, nil
}

// ticketTypeToProto convierte entidad a proto
func (h *TicketTypeHandler) ticketTypeToProto(tt *entities.TicketType) *osmi.TicketTypeResponse {
	if tt == nil {
		return nil
	}

	benefits := make([]string, len(tt.Benefits))
	copy(benefits, tt.Benefits)

	resp := &osmi.TicketTypeResponse{
		Id:                tt.PublicID,
		EventId:           "",
		Name:              tt.Name,
		Description:       safeString(tt.Description),
		TicketClass:       tt.TicketClass,
		BasePrice:         tt.BasePrice,
		Currency:          tt.Currency,
		TaxRate:           tt.TaxRate,
		TotalQuantity:     int32(tt.TotalQuantity),
		AvailableQuantity: int32(tt.AvailableQuantity),
		SoldQuantity:      int32(tt.SoldQuantity),
		ReservedQuantity:  int32(tt.ReservedQuantity),
		MaxPerOrder:       int32(tt.MaxPerOrder),
		MinPerOrder:       int32(tt.MinPerOrder),
		SaleStartsAt:      timestamppb.New(tt.SaleStartsAt),
		IsActive:          tt.IsActive,
		IsSoldOut:         tt.IsSoldOut,
		Benefits:          benefits,
		CreatedAt:         timestamppb.New(tt.CreatedAt),
		UpdatedAt:         timestamppb.New(tt.UpdatedAt),
	}

	if tt.SaleEndsAt != nil {
		resp.SaleEndsAt = timestamppb.New(*tt.SaleEndsAt)
	}

	return resp
}

func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
