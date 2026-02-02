package httphandlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/franciscozamorau/osmi-server/internal/api/dto"
	"github.com/franciscozamorau/osmi-server/internal/api/dto/filter"
	"github.com/franciscozamorau/osmi-server/internal/application/services"
	"github.com/franciscozamorau/osmi-server/internal/shared/logger"
	"github.com/go-chi/chi/v5"
)

type EventHTTPHandler struct {
	eventService *services.EventService
}

func NewEventHTTPHandler(eventService *services.EventService) *EventHTTPHandler {
	return &EventHTTPHandler{
		eventService: eventService,
	}
}

// CreateEvent maneja la creación de eventos
func (h *EventHTTPHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req dto.CreateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Failed to decode request", logger.Field("error", err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	event, err := h.eventService.CreateEvent(ctx, &req)
	if err != nil {
		logger.Error("Failed to create event", logger.Field("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := dto.EventResponse{
		ID:               event.PublicUUID,
		Name:             event.Name,
		Slug:             event.Slug,
		Description:      event.Description,
		ShortDescription: event.ShortDescription,
		EventType:        event.EventType,
		Status:           event.Status,
		IsFeatured:       event.IsFeatured,
		IsFree:           event.IsFree,
		CreatedAt:        event.CreatedAt,
		UpdatedAt:        event.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// GetEvent maneja la obtención de un evento
func (h *EventHTTPHandler) GetEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	eventID := chi.URLParam(r, "id")

	if eventID == "" {
		http.Error(w, "Event ID is required", http.StatusBadRequest)
		return
	}

	event, err := h.eventService.GetEventByPublicID(ctx, eventID)
	if err != nil {
		logger.Error("Failed to get event", logger.Field("error", err))
		http.Error(w, "Event not found", http.StatusNotFound)
		return
	}

	resp := dto.EventResponse{
		ID:               event.PublicUUID,
		Name:             event.Name,
		Slug:             event.Slug,
		Description:      event.Description,
		ShortDescription: event.ShortDescription,
		EventType:        event.EventType,
		Status:           event.Status,
		IsFeatured:       event.IsFeatured,
		IsFree:           event.IsFree,
		CreatedAt:        event.CreatedAt,
		UpdatedAt:        event.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ListEvents maneja la lista de eventos
func (h *EventHTTPHandler) ListEvents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parsear filtros
	var eventFilter dto.EventFilter
	if err := json.NewDecoder(r.Body).Decode(&eventFilter); err != nil && r.ContentLength > 0 {
		logger.Error("Failed to decode filters", logger.Field("error", err))
		http.Error(w, "Invalid filter data", http.StatusBadRequest)
		return
	}

	// Parsear paginación
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	pagination := filter.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	events, total, err := h.eventService.ListEvents(ctx, eventFilter, pagination)
	if err != nil {
		logger.Error("Failed to list events", logger.Field("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Convertir a DTOs de respuesta
	eventResponses := make([]dto.EventResponse, len(events))
	for i, event := range events {
		eventResponses[i] = dto.EventResponse{
			ID:               event.PublicUUID,
			Name:             event.Name,
			Slug:             event.Slug,
			ShortDescription: event.ShortDescription,
			EventType:        event.EventType,
			Status:           event.Status,
			IsFeatured:       event.IsFeatured,
			IsFree:           event.IsFree,
			CreatedAt:        event.CreatedAt,
			UpdatedAt:        event.UpdatedAt,
		}
	}

	resp := dto.EventListResponse{
		Events:     eventResponses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int((total + int64(pageSize) - 1) / int64(pageSize)),
		HasNext:    page*pageSize < int(total),
		HasPrev:    page > 1,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// UpdateEvent maneja la actualización de eventos
func (h *EventHTTPHandler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	eventID := chi.URLParam(r, "id")

	if eventID == "" {
		http.Error(w, "Event ID is required", http.StatusBadRequest)
		return
	}

	var req dto.UpdateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Failed to decode request", logger.Field("error", err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	event, err := h.eventService.UpdateEvent(ctx, eventID, &req)
	if err != nil {
		logger.Error("Failed to update event", logger.Field("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := dto.EventResponse{
		ID:               event.PublicUUID,
		Name:             event.Name,
		Slug:             event.Slug,
		Description:      event.Description,
		ShortDescription: event.ShortDescription,
		EventType:        event.EventType,
		Status:           event.Status,
		IsFeatured:       event.IsFeatured,
		IsFree:           event.IsFree,
		CreatedAt:        event.CreatedAt,
		UpdatedAt:        event.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// PublishEvent maneja la publicación de eventos
func (h *EventHTTPHandler) PublishEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	eventID := chi.URLParam(r, "id")

	if eventID == "" {
		http.Error(w, "Event ID is required", http.StatusBadRequest)
		return
	}

	event, err := h.eventService.PublishEvent(ctx, eventID)
	if err != nil {
		logger.Error("Failed to publish event", logger.Field("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := dto.EventResponse{
		ID:               event.PublicUUID,
		Name:             event.Name,
		Slug:             event.Slug,
		Description:      event.Description,
		ShortDescription: event.ShortDescription,
		EventType:        event.EventType,
		Status:           event.Status,
		IsFeatured:       event.IsFeatured,
		IsFree:           event.IsFree,
		CreatedAt:        event.CreatedAt,
		UpdatedAt:        event.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// CancelEvent maneja la cancelación de eventos
func (h *EventHTTPHandler) CancelEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	eventID := chi.URLParam(r, "id")

	if eventID == "" {
		http.Error(w, "Event ID is required", http.StatusBadRequest)
		return
	}

	event, err := h.eventService.CancelEvent(ctx, eventID)
	if err != nil {
		logger.Error("Failed to cancel event", logger.Field("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := dto.EventResponse{
		ID:               event.PublicUUID,
		Name:             event.Name,
		Slug:             event.Slug,
		Description:      event.Description,
		ShortDescription: event.ShortDescription,
		EventType:        event.EventType,
		Status:           event.Status,
		IsFeatured:       event.IsFeatured,
		IsFree:           event.IsFree,
		CreatedAt:        event.CreatedAt,
		UpdatedAt:        event.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetEventStats maneja las estadísticas de eventos
func (h *EventHTTPHandler) GetEventStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	eventID := chi.URLParam(r, "id")

	if eventID == "" {
		http.Error(w, "Event ID is required", http.StatusBadRequest)
		return
	}

	stats, err := h.eventService.GetEventStats(ctx, eventID)
	if err != nil {
		logger.Error("Failed to get event stats", logger.Field("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
