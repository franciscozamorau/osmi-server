package repository

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	pb "github.com/franciscozamorau/osmi-server/gen"
	"github.com/franciscozamorau/osmi-server/internal/db"
	"github.com/franciscozamorau/osmi-server/internal/models"

	"github.com/jackc/pgx/v5"
)

type TicketRepository struct{}

func NewTicketRepository() *TicketRepository {
	return &TicketRepository{}
}

// Valid ticket statuses
var validTicketStatuses = map[string]bool{
	"available":   true,
	"reserved":    true,
	"sold":        true,
	"used":        true,
	"cancelled":   true,
	"transferred": true,
}

// Valid status transitions
var validStatusTransitions = map[string]map[string]bool{
	"available": {
		"reserved":  true,
		"sold":      true,
		"cancelled": true,
	},
	"reserved": {
		"sold":      true,
		"available": true,
		"cancelled": true,
	},
	"sold": {
		"used":        true,
		"cancelled":   true,
		"transferred": true,
	},
	"used": {
		// No transitions from used
	},
	"cancelled": {
		// No transitions from cancelled
	},
	"transferred": {
		"used": true,
	},
}

// CreateTicket crea un nuevo ticket
func (r *TicketRepository) CreateTicket(ctx context.Context, req *pb.TicketRequest) (string, error) {
	// Validar existencia del event_id y category_id
	if err := r.validateEventAndCategory(ctx, req.EventId, req.CategoryId); err != nil {
		return "", err
	}

	query := `
		INSERT INTO tickets (event_id, category_id, code, status, price) 
		VALUES ($1, $2, $3, 'available', 
			COALESCE((SELECT price FROM categories WHERE id = $2), 0)
		) 
		RETURNING public_id
	`

	// Generar un código único para el ticket
	code := generateTicketCode(req.EventId, req.UserId)

	var publicID string
	err := db.Pool.QueryRow(ctx, query, req.EventId, req.CategoryId, code).Scan(&publicID)
	if err != nil {
		return "", fmt.Errorf("error creating ticket: %v", err)
	}

	// Auditoría explícita
	r.auditTicketChange(ctx, "INSERT", nil, publicID)

	log.Printf("✅ Ticket created: %s for event %s", publicID, req.EventId)
	return publicID, nil
}

// validateEventAndCategory valida que el evento y categoría existan
func (r *TicketRepository) validateEventAndCategory(ctx context.Context, eventID, categoryID string) error {
	// Validar que el evento existe y está activo
	var eventExists bool
	err := db.Pool.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM events WHERE id = $1 AND is_active = true)",
		eventID).Scan(&eventExists)

	if err != nil {
		return fmt.Errorf("error validating event: %v", err)
	}
	if !eventExists {
		return fmt.Errorf("event not found or inactive: %s", eventID)
	}

	// Validar que la categoría existe y está activa
	var categoryExists bool
	err = db.Pool.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM categories WHERE id = $1 AND is_active = true)",
		categoryID).Scan(&categoryExists)

	if err != nil {
		return fmt.Errorf("error validating category: %v", err)
	}
	if !categoryExists {
		return fmt.Errorf("category not found or inactive: %s", categoryID)
	}

	// Validar que la categoría pertenece al evento
	var validCategory bool
	err = db.Pool.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM categories WHERE id = $1 AND event_id = $2)",
		categoryID, eventID).Scan(&validCategory)

	if err != nil {
		return fmt.Errorf("error validating category-event relationship: %v", err)
	}
	if !validCategory {
		return fmt.Errorf("category %s does not belong to event %s", categoryID, eventID)
	}

	return nil
}

// GetTicketByID obtiene un ticket por ID
func (r *TicketRepository) GetTicketByID(ctx context.Context, id int64) (*models.Ticket, error) {
	query := `
		SELECT id, public_id, category_id, event_id, customer_id, code, status,
		       seat_number, qr_code_url, price, used_at, transferred_from_ticket_id,
		       created_at, updated_at
		FROM tickets 
		WHERE id = $1
	`

	var ticket models.Ticket
	err := db.Pool.QueryRow(ctx, query, id).Scan(
		&ticket.ID,
		&ticket.PublicID,
		&ticket.CategoryID,
		&ticket.EventID,
		&ticket.CustomerID,
		&ticket.Code,
		&ticket.Status,
		&ticket.SeatNumber,
		&ticket.QRCodeURL,
		&ticket.Price,
		&ticket.UsedAt,
		&ticket.TransferredFromTicketID,
		&ticket.CreatedAt,
		&ticket.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("ticket not found with id: %d", id)
		}
		return nil, fmt.Errorf("error getting ticket: %v", err)
	}

	return &ticket, nil
}

// GetTicketByCode obtiene un ticket por código
func (r *TicketRepository) GetTicketByCode(ctx context.Context, code string) (*models.Ticket, error) {
	query := `
		SELECT id, public_id, category_id, event_id, customer_id, code, status,
		       seat_number, qr_code_url, price, used_at, transferred_from_ticket_id,
		       created_at, updated_at
		FROM tickets 
		WHERE code = $1
	`

	var ticket models.Ticket
	err := db.Pool.QueryRow(ctx, query, code).Scan(
		&ticket.ID,
		&ticket.PublicID,
		&ticket.CategoryID,
		&ticket.EventID,
		&ticket.CustomerID,
		&ticket.Code,
		&ticket.Status,
		&ticket.SeatNumber,
		&ticket.QRCodeURL,
		&ticket.Price,
		&ticket.UsedAt,
		&ticket.TransferredFromTicketID,
		&ticket.CreatedAt,
		&ticket.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("ticket not found with code: %s", code)
		}
		return nil, fmt.Errorf("error getting ticket by code: %v", err)
	}

	return &ticket, nil
}

// GetTicketByPublicID obtiene un ticket por public_id
func (r *TicketRepository) GetTicketByPublicID(ctx context.Context, publicID string) (*models.Ticket, error) {
	query := `
		SELECT id, public_id, category_id, event_id, customer_id, code, status,
		       seat_number, qr_code_url, price, used_at, transferred_from_ticket_id,
		       created_at, updated_at
		FROM tickets 
		WHERE public_id = $1
	`

	var ticket models.Ticket
	err := db.Pool.QueryRow(ctx, query, publicID).Scan(
		&ticket.ID,
		&ticket.PublicID,
		&ticket.CategoryID,
		&ticket.EventID,
		&ticket.CustomerID,
		&ticket.Code,
		&ticket.Status,
		&ticket.SeatNumber,
		&ticket.QRCodeURL,
		&ticket.Price,
		&ticket.UsedAt,
		&ticket.TransferredFromTicketID,
		&ticket.CreatedAt,
		&ticket.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("ticket not found with public_id: %s", publicID)
		}
		return nil, fmt.Errorf("error getting ticket by public_id: %v", err)
	}

	return &ticket, nil
}

// UpdateTicketStatus actualiza el estado de un ticket con validación
func (r *TicketRepository) UpdateTicketStatus(ctx context.Context, ticketID int64, status string) error {
	// Validar estado
	if !r.isValidTicketStatus(status) {
		return fmt.Errorf("invalid ticket status: %s", status)
	}

	// Obtener ticket antiguo para validar transición
	oldTicket, err := r.GetTicketByID(ctx, ticketID)
	if err != nil {
		return err
	}

	// Validar transición de estado
	if !r.isValidStatusTransition(oldTicket.Status, status) {
		return fmt.Errorf("invalid status transition from %s to %s", oldTicket.Status, status)
	}

	query := `UPDATE tickets SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`

	result, err := db.Pool.Exec(ctx, query, status, ticketID)
	if err != nil {
		return fmt.Errorf("error updating ticket status: %v", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("ticket not found with id: %d", ticketID)
	}

	// Auditoría explícita
	r.auditTicketChange(ctx, "UPDATE", oldTicket, oldTicket.PublicID)

	log.Printf("✅ Ticket status updated: ID %d (%s -> %s)", ticketID, oldTicket.Status, status)
	return nil
}

// ListTickets lista tickets con filtros y paginación
func (r *TicketRepository) ListTickets(ctx context.Context, filters *TicketFilters, page, pageSize int) ([]*models.Ticket, *PaginationInfo, error) {
	whereClause, params := r.buildTicketWhereClause(filters)
	params = append(params, pageSize, (page-1)*pageSize)

	query := fmt.Sprintf(`
		SELECT id, public_id, category_id, event_id, customer_id, code, status,
		       seat_number, qr_code_url, price, used_at, transferred_from_ticket_id,
		       created_at, updated_at
		FROM tickets 
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, len(params)-1, len(params))

	rows, err := db.Pool.Query(ctx, query, params...)
	if err != nil {
		return nil, nil, fmt.Errorf("error listing tickets: %v", err)
	}
	defer rows.Close()

	var tickets []*models.Ticket
	for rows.Next() {
		var ticket models.Ticket
		err := rows.Scan(
			&ticket.ID,
			&ticket.PublicID,
			&ticket.CategoryID,
			&ticket.EventID,
			&ticket.CustomerID,
			&ticket.Code,
			&ticket.Status,
			&ticket.SeatNumber,
			&ticket.QRCodeURL,
			&ticket.Price,
			&ticket.UsedAt,
			&ticket.TransferredFromTicketID,
			&ticket.CreatedAt,
			&ticket.UpdatedAt,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("error scanning ticket: %v", err)
		}
		tickets = append(tickets, &ticket)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("error iterating tickets: %v", err)
	}

	// Obtener información de paginación
	pagination, err := r.getTicketsPagination(ctx, filters, page, pageSize)
	if err != nil {
		return nil, nil, err
	}

	return tickets, pagination, nil
}

// AssignTicketToTransaction asigna un ticket a una transacción
func (r *TicketRepository) AssignTicketToTransaction(ctx context.Context, ticketID, transactionID int64) error {
	// Obtener ticket antiguo para auditoría
	oldTicket, err := r.GetTicketByID(ctx, ticketID)
	if err != nil {
		return err
	}

	query := `
		UPDATE tickets 
		SET transaction_id = $1, status = 'sold', updated_at = CURRENT_TIMESTAMP 
		WHERE id = $2 AND status IN ('available', 'reserved')
	`

	result, err := db.Pool.Exec(ctx, query, transactionID, ticketID)
	if err != nil {
		return fmt.Errorf("error assigning ticket to transaction: %v", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("ticket not available for transaction assignment: %d", ticketID)
	}

	// Auditoría explícita
	r.auditTicketChange(ctx, "UPDATE", oldTicket, oldTicket.PublicID)

	log.Printf("✅ Ticket assigned to transaction: Ticket %d -> Transaction %d", ticketID, transactionID)
	return nil
}

// ReserveTicket reserva un ticket para un cliente
func (r *TicketRepository) ReserveTicket(ctx context.Context, ticketID int64, customerID int64) error {
	query := `
		UPDATE tickets 
		SET status = 'reserved', customer_id = $1, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $2 AND status = 'available'
	`

	result, err := db.Pool.Exec(ctx, query, customerID, ticketID)
	if err != nil {
		return fmt.Errorf("error reserving ticket: %v", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("ticket not available for reservation: %d", ticketID)
	}

	log.Printf("✅ Ticket reserved: ID %d for customer %d", ticketID, customerID)
	return nil
}

// Helper functions

type TicketFilters struct {
	EventID    *int64
	CategoryID *int64
	CustomerID *int64
	Status     *string
	Code       *string
	DateFrom   *time.Time
	DateTo     *time.Time
}

type PaginationInfo struct {
	Total       int64 `json:"total"`
	Page        int   `json:"page"`
	PageSize    int   `json:"page_size"`
	TotalPages  int   `json:"total_pages"`
	HasNext     bool  `json:"has_next"`
	HasPrevious bool  `json:"has_previous"`
}

func (r *TicketRepository) buildTicketWhereClause(filters *TicketFilters) (string, []interface{}) {
	var conditions []string
	var params []interface{}
	paramCount := 1

	if filters != nil {
		if filters.EventID != nil {
			conditions = append(conditions, fmt.Sprintf("event_id = $%d", paramCount))
			params = append(params, *filters.EventID)
			paramCount++
		}
		if filters.CategoryID != nil {
			conditions = append(conditions, fmt.Sprintf("category_id = $%d", paramCount))
			params = append(params, *filters.CategoryID)
			paramCount++
		}
		if filters.CustomerID != nil {
			conditions = append(conditions, fmt.Sprintf("customer_id = $%d", paramCount))
			params = append(params, *filters.CustomerID)
			paramCount++
		}
		if filters.Status != nil {
			conditions = append(conditions, fmt.Sprintf("status = $%d", paramCount))
			params = append(params, *filters.Status)
			paramCount++
		}
		if filters.Code != nil {
			conditions = append(conditions, fmt.Sprintf("code ILIKE $%d", paramCount))
			params = append(params, "%"+*filters.Code+"%")
			paramCount++
		}
		if filters.DateFrom != nil {
			conditions = append(conditions, fmt.Sprintf("created_at >= $%d", paramCount))
			params = append(params, *filters.DateFrom)
			paramCount++
		}
		if filters.DateTo != nil {
			conditions = append(conditions, fmt.Sprintf("created_at <= $%d", paramCount))
			params = append(params, *filters.DateTo)
			paramCount++
		}
	}

	if len(conditions) == 0 {
		return "", params
	}

	return "WHERE " + strings.Join(conditions, " AND "), params
}

func (r *TicketRepository) getTicketsPagination(ctx context.Context, filters *TicketFilters, page, pageSize int) (*PaginationInfo, error) {
	whereClause, params := r.buildTicketWhereClause(filters)

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM tickets %s", whereClause)

	var total int64
	err := db.Pool.QueryRow(ctx, countQuery, params...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("error counting tickets: %v", err)
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	return &PaginationInfo{
		Total:       total,
		Page:        page,
		PageSize:    pageSize,
		TotalPages:  totalPages,
		HasNext:     page < totalPages,
		HasPrevious: page > 1,
	}, nil
}

func (r *TicketRepository) isValidTicketStatus(status string) bool {
	return validTicketStatuses[status]
}

func (r *TicketRepository) isValidStatusTransition(from, to string) bool {
	if transitions, exists := validStatusTransitions[from]; exists {
		return transitions[to]
	}
	return false
}

func generateTicketCode(eventID, userID string) string {
	timestamp := time.Now().Unix()
	return fmt.Sprintf("TKT-%s-%s-%d", eventID, userID, timestamp)
}

// auditTicketChange realiza auditoría explícita de cambios en tickets
func (r *TicketRepository) auditTicketChange(ctx context.Context, operation string, oldTicket *models.Ticket, ticketPublicID string) {
	// Auditoría adicional a los triggers de base de datos
	auditData := map[string]interface{}{
		"operation":  operation,
		"ticket_id":  ticketPublicID,
		"timestamp":  time.Now().UTC(),
		"old_status": nil,
		"new_status": nil,
	}

	if oldTicket != nil {
		auditData["old_status"] = oldTicket.Status
	}

	// En un entorno de producción, aquí enviarías a un servicio de auditoría
	log.Printf("📝 Ticket audit - %s: %+v", operation, auditData)
}
