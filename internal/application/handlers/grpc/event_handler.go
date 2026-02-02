package grpchandlers

import (
	"context"
	"time"

	osmi "github.com/franciscozamorau/osmi-protobuf/gen/pb"
	"github.com/franciscozamorau/osmi-server/internal/api/dto"
	"github.com/franciscozamorau/osmi-server/internal/application/services"
	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type EventHandler struct {
	osmi.UnimplementedOsmiServiceServer
	eventService *services.EventService
}

func NewEventHandler(eventService *services.EventService) *EventHandler {
	return &EventHandler{
		eventService: eventService,
	}
}

func (h *EventHandler) CreateEvent(ctx context.Context, req *osmi.EventRequest) (*osmi.EventResponse, error) {
	// Convertir protobuf a DTO
	createReq := &dto.CreateEventRequest{
		OrganizerID:         req.OrganizerId,
		PrimaryCategoryID:   req.PrimaryCategoryId,
		VenueID:             req.VenueId,
		Name:                req.Name,
		Slug:                req.Slug,
		ShortDescription:    req.ShortDescription,
		Description:         req.Description,
		EventType:           req.EventType,
		CoverImageURL:       req.CoverImageUrl,
		BannerImageURL:      req.BannerImageUrl,
		Timezone:            req.Timezone,
		StartsAt:            req.StartDate,
		EndsAt:              req.EndDate,
		DoorsOpenAt:         req.DoorsOpenAt,
		DoorsCloseAt:        req.DoorsCloseAt,
		VenueName:           req.VenueName,
		AddressFull:         req.AddressFull,
		City:                req.City,
		State:               req.State,
		Country:             req.Country,
		Visibility:          req.Visibility,
		IsFeatured:          req.IsFeatured,
		IsFree:              req.IsFree,
		MaxAttendees:        int(req.MaxAttendees),
		MinAttendees:        int(req.MinAttendees),
		Tags:                req.Tags,
		AgeRestriction:      int(req.AgeRestriction),
		RequiresApproval:    req.RequiresApproval,
		AllowReservations:   req.AllowReservations,
		ReservationDuration: int(req.ReservationDuration),
		CategoryIDs:         req.CategoryIds,
	}

	// Llamar al servicio
	event, err := h.eventService.CreateEvent(ctx, createReq)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Convertir entidad a protobuf
	return h.eventToProto(event), nil
}

func (h *EventHandler) GetEvent(ctx context.Context, req *osmi.EventLookup) (*osmi.EventResponse, error) {
	// Llamar al servicio
	event, err := h.eventService.GetEvent(ctx, req.PublicId)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	// Convertir entidad a protobuf
	return h.eventToProto(event), nil
}

func (h *EventHandler) UpdateEvent(ctx context.Context, req *osmi.UpdateEventRequest) (*osmi.EventResponse, error) {
	// Convertir protobuf a DTO
	updateReq := &dto.UpdateEventRequest{
		Name:             req.Name,
		ShortDescription: req.ShortDescription,
		Description:      req.Description,
		EventType:        req.EventType,
		CoverImageURL:    req.CoverImageUrl,
		BannerImageURL:   req.BannerImageUrl,
		Timezone:         req.Timezone,
		StartsAt:         req.StartDate,
		EndsAt:           req.EndDate,
		Status:           req.Status,
		Visibility:       req.Visibility,
		IsFeatured:       req.IsFeatured,
		MaxAttendees:     req.MaxAttendees,
		AgeRestriction:   req.AgeRestriction,
		Tags:             req.Tags,
	}

	// Llamar al servicio
	event, err := h.eventService.UpdateEvent(ctx, req.Id, updateReq)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Convertir entidad a protobuf
	return h.eventToProto(event), nil
}

func (h *EventHandler) ListEvents(ctx context.Context, req *osmi.ListRequest) (*osmi.EventListResponse, error) {
	// Convertir protobuf a DTO de filtro
	filter := dto.EventFilter{
		Search:      req.Search,
		OrganizerID: req.OrganizerId,
		CategoryID:  req.CategoryId,
		VenueID:     req.VenueId,
		EventType:   req.EventType,
		Status:      req.Status,
		Country:     req.Country,
		City:        req.City,
		IsFeatured:  req.IsFeatured,
		IsFree:      req.IsFree,
		DateFrom:    req.DateFrom,
		DateTo:      req.DateTo,
		Tags:        req.Tags,
	}

	// Convertir paginación
	pagination := dto.Pagination{
		Page:     int(req.Page),
		PageSize: int(req.PageSize),
	}

	// Llamar al servicio
	events, total, err := h.eventService.ListEvents(ctx, filter, pagination)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Convertir entidades a protobuf
	pbEvents := make([]*osmi.EventResponse, 0, len(events))
	for _, event := range events {
		pbEvents = append(pbEvents, h.eventToProto(event))
	}

	return &osmi.EventListResponse{
		Events:     pbEvents,
		TotalCount: int32(total),
		Page:       int32(pagination.Page),
		PageSize:   int32(pagination.PageSize),
		TotalPages: int32((total + int64(pagination.PageSize) - 1) / int64(pagination.PageSize)),
	}, nil
}

func (h *EventHandler) PublishEvent(ctx context.Context, req *osmi.PublishEventRequest) (*osmi.EventResponse, error) {
	var publishAt *time.Time
	if req.PublishAt != "" {
		pa, err := time.Parse(time.RFC3339, req.PublishAt)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid publish date format")
		}
		publishAt = &pa
	}

	// Llamar al servicio
	event, err := h.eventService.PublishEvent(ctx, req.EventId, publishAt)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Convertir entidad a protobuf
	return h.eventToProto(event), nil
}

func (h *EventHandler) CancelEvent(ctx context.Context, req *osmi.CancelEventRequest) (*osmi.EventResponse, error) {
	// Llamar al servicio
	event, err := h.eventService.CancelEvent(ctx, req.EventId, req.Reason)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Convertir entidad a protobuf
	return h.eventToProto(event), nil
}

func (h *EventHandler) GetEventStats(ctx context.Context, req *osmi.EventLookup) (*osmi.EventStatsResponse, error) {
	// Llamar al servicio
	stats, err := h.eventService.GetEventStats(ctx, req.PublicId)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	// Convertir estadísticas a protobuf
	return &osmi.EventStatsResponse{
		TicketsSold:      int32(stats.TicketsSold),
		TicketsAvailable: int32(stats.TicketsAvailable),
		TotalRevenue:     stats.TotalRevenue,
		AvgTicketPrice:   stats.AvgTicketPrice,
		CheckInRate:      stats.CheckInRate,
		ConversionRate:   stats.ConversionRate,
		ViewsToday:       int32(stats.ViewsToday),
		SalesVelocity:    stats.SalesVelocity,
	}, nil
}

// Helper function para convertir evento a protobuf
func (h *EventHandler) eventToProto(event *entities.Event) *osmi.EventResponse {
	return &osmi.EventResponse{
		Id:                event.PublicID,
		OrganizerId:       "", // TODO: Obtener public ID del organizador
		PrimaryCategoryId: safeStringID(event.PrimaryCategoryID),
		VenueId:           safeStringID(event.VenueID),
		Name:              event.Name,
		Slug:              event.Slug,
		ShortDescription:  safeStringPtr(event.ShortDescription),
		Description:       safeStringPtr(event.Description),
		EventType:         event.EventType,
		CoverImageUrl:     safeStringPtr(event.CoverImageURL),
		BannerImageUrl:    safeStringPtr(event.BannerImageURL),
		Timezone:          event.Timezone,
		StartDate:         event.StartsAt.Format(time.RFC3339),
		EndDate:           event.EndsAt.Format(time.RFC3339),
		DoorsOpenAt:       safeTimeString(event.DoorsOpenAt),
		DoorsCloseAt:      safeTimeString(event.DoorsCloseAt),
		VenueName:         safeStringPtr(event.VenueName),
		AddressFull:       safeStringPtr(event.AddressFull),
		City:              safeStringPtr(event.City),
		State:             safeStringPtr(event.State),
		Country:           safeStringPtr(event.Country),
		Status:            event.Status,
		Visibility:        event.Visibility,
		IsFeatured:        event.IsFeatured,
		IsFree:            event.IsFree,
		MaxAttendees:      safeInt32Ptr(event.MaxAttendees),
		MinAttendees:      int32(event.MinAttendees),
		Tags:              event.Tags,
		AgeRestriction:    safeInt32Ptr(event.AgeRestriction),
		ViewCount:         int32(event.ViewCount),
		FavoriteCount:     int32(event.FavoriteCount),
		ShareCount:        int32(event.ShareCount),
		PublishedAt:       safeTimeProto(event.PublishedAt),
		CreatedAt:         timestamppb.New(event.CreatedAt),
		UpdatedAt:         timestamppb.New(event.UpdatedAt),
	}
}

// Helper functions adicionales
func safeInt32Ptr(i *int32) int32 {
	if i == nil {
		return 0
	}
	return *i
}
