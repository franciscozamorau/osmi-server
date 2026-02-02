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

type TicketHTTPHandler struct {
	ticketService *services.TicketService
}

func NewTicketHTTPHandler(ticketService *services.TicketService) *TicketHTTPHandler {
	return &TicketHTTPHandler{
		ticketService: ticketService,
	}
}

// CreateTicket maneja la creación de tickets
func (h *TicketHTTPHandler) CreateTicket(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req dto.CreateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Failed to decode request", logger.Field("error", err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ticket, err := h.ticketService.CreateTicket(ctx, &req)
	if err != nil {
		logger.Error("Failed to create ticket", logger.Field("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := dto.TicketResponse{
		ID:         ticket.PublicUUID,
		Code:       ticket.Code,
		Status:     ticket.Status,
		FinalPrice: ticket.FinalPrice,
		Currency:   ticket.Currency,
		EventID:    strconv.FormatInt(ticket.EventID, 10),
		CustomerID: strconv.FormatInt(*ticket.CustomerID, 10),
		CreatedAt:  ticket.CreatedAt,
		UpdatedAt:  ticket.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// GetTicket maneja la obtención de un ticket
func (h *TicketHTTPHandler) GetTicket(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ticketID := chi.URLParam(r, "id")

	if ticketID == "" {
		http.Error(w, "Ticket ID is required", http.StatusBadRequest)
		return
	}

	ticket, err := h.ticketService.GetTicket(ctx, ticketID)
	if err != nil {
		logger.Error("Failed to get ticket", logger.Field("error", err))
		http.Error(w, "Ticket not found", http.StatusNotFound)
		return
	}

	resp := dto.TicketResponse{
		ID:         ticket.PublicUUID,
		Code:       ticket.Code,
		Status:     ticket.Status,
		FinalPrice: ticket.FinalPrice,
		Currency:   ticket.Currency,
		EventID:    strconv.FormatInt(ticket.EventID, 10),
		CustomerID: strconv.FormatInt(*ticket.CustomerID, 10),
		CreatedAt:  ticket.CreatedAt,
		UpdatedAt:  ticket.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ListTickets maneja la lista de tickets
func (h *TicketHTTPHandler) ListTickets(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parsear filtros
	var ticketFilter dto.TicketFilter
	if err := json.NewDecoder(r.Body).Decode(&ticketFilter); err != nil && r.ContentLength > 0 {
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

	tickets, total, err := h.ticketService.ListTickets(ctx, ticketFilter, pagination)
	if err != nil {
		logger.Error("Failed to list tickets", logger.Field("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Convertir a DTOs de respuesta
	ticketResponses := make([]dto.TicketResponse, len(tickets))
	for i, ticket := range tickets {
		var customerID string
		if ticket.CustomerID != nil {
			customerID = strconv.FormatInt(*ticket.CustomerID, 10)
		}

		ticketResponses[i] = dto.TicketResponse{
			ID:         ticket.PublicUUID,
			Code:       ticket.Code,
			Status:     ticket.Status,
			FinalPrice: ticket.FinalPrice,
			Currency:   ticket.Currency,
			EventID:    strconv.FormatInt(ticket.EventID, 10),
			CustomerID: customerID,
			CreatedAt:  ticket.CreatedAt,
			UpdatedAt:  ticket.UpdatedAt,
		}
	}

	resp := dto.TicketListResponse{
		Tickets:    ticketResponses,
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

// UpdateTicket maneja la actualización de tickets
func (h *TicketHTTPHandler) UpdateTicket(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ticketID := chi.URLParam(r, "id")

	if ticketID == "" {
		http.Error(w, "Ticket ID is required", http.StatusBadRequest)
		return
	}

	var req dto.UpdateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Failed to decode request", logger.Field("error", err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ticket, err := h.ticketService.UpdateTicket(ctx, ticketID, &req)
	if err != nil {
		logger.Error("Failed to update ticket", logger.Field("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := dto.TicketResponse{
		ID:         ticket.PublicUUID,
		Code:       ticket.Code,
		Status:     ticket.Status,
		FinalPrice: ticket.FinalPrice,
		Currency:   ticket.Currency,
		EventID:    strconv.FormatInt(ticket.EventID, 10),
		CustomerID: strconv.FormatInt(*ticket.CustomerID, 10),
		CreatedAt:  ticket.CreatedAt,
		UpdatedAt:  ticket.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// CheckInTicket maneja el check-in de tickets
func (h *TicketHTTPHandler) CheckInTicket(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ticketID := chi.URLParam(r, "id")

	if ticketID == "" {
		http.Error(w, "Ticket ID is required", http.StatusBadRequest)
		return
	}

	var req dto.CheckInTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Failed to decode request", logger.Field("error", err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ticket, err := h.ticketService.CheckInTicket(ctx, &req)
	if err != nil {
		logger.Error("Failed to check in ticket", logger.Field("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := dto.TicketResponse{
		ID:         ticket.PublicUUID,
		Code:       ticket.Code,
		Status:     ticket.Status,
		FinalPrice: ticket.FinalPrice,
		Currency:   ticket.Currency,
		EventID:    strconv.FormatInt(ticket.EventID, 10),
		CustomerID: strconv.FormatInt(*ticket.CustomerID, 10),
		CreatedAt:  ticket.CreatedAt,
		UpdatedAt:  ticket.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ReserveTicket maneja la reserva de tickets
func (h *TicketHTTPHandler) ReserveTicket(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req dto.ReserveTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Failed to decode request", logger.Field("error", err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ticket, err := h.ticketService.ReserveTicket(ctx, &req)
	if err != nil {
		logger.Error("Failed to reserve ticket", logger.Field("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := dto.TicketResponse{
		ID:         ticket.PublicUUID,
		Code:       ticket.Code,
		Status:     ticket.Status,
		FinalPrice: ticket.FinalPrice,
		Currency:   ticket.Currency,
		EventID:    strconv.FormatInt(ticket.EventID, 10),
		CustomerID: strconv.FormatInt(*ticket.CustomerID, 10),
		CreatedAt:  ticket.CreatedAt,
		UpdatedAt:  ticket.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// TransferTicket maneja la transferencia de tickets
func (h *TicketHTTPHandler) TransferTicket(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ticketID := chi.URLParam(r, "id")

	if ticketID == "" {
		http.Error(w, "Ticket ID is required", http.StatusBadRequest)
		return
	}

	var req dto.TransferTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("Failed to decode request", logger.Field("error", err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ticket, err := h.ticketService.TransferTicket(ctx, &req)
	if err != nil {
		logger.Error("Failed to transfer ticket", logger.Field("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := dto.TicketResponse{
		ID:         ticket.PublicUUID,
		Code:       ticket.Code,
		Status:     ticket.Status,
		FinalPrice: ticket.FinalPrice,
		Currency:   ticket.Currency,
		EventID:    strconv.FormatInt(ticket.EventID, 10),
		CustomerID: strconv.FormatInt(*ticket.CustomerID, 10),
		CreatedAt:  ticket.CreatedAt,
		UpdatedAt:  ticket.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetTicketStats maneja las estadísticas de tickets
func (h *TicketHTTPHandler) GetTicketStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	eventID := r.URL.Query().Get("event_id")

	if eventID == "" {
		http.Error(w, "Event ID is required", http.StatusBadRequest)
		return
	}

	stats, err := h.ticketService.GetTicketStats(ctx, eventID)
	if err != nil {
		logger.Error("Failed to get ticket stats", logger.Field("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
