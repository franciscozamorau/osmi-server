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
	"github.com/google/uuid"

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

	// Obtener el customer_id basado en el user_id (public_id del customer)
	customerID, err := r.getCustomerIDByPublicID(ctx, req.UserId)
	if err != nil {
		return "", fmt.Errorf("error finding customer: %v", err)
	}

	// Obtener category_id interno basado en public_id
	categoryInternalID, err := r.getCategoryIDByPublicID(ctx, req.CategoryId)
	if err != nil {
		return "", fmt.Errorf("error finding category: %v", err)
	}

	// Obtener event_id interno basado en public_id
	eventInternalID, err := r.getEventIDByPublicID(ctx, req.EventId)
	if err != nil {
		return "", fmt.Errorf("error finding event: %v", err)
	}

	query := `
		INSERT INTO tickets (
			public_id, category_id, event_id, code, status, price, 
			created_at, updated_at
		) 
		VALUES ($1, $2, $3, $4, 'available', 
			COALESCE((SELECT price FROM categories WHERE id = $2), 0),
			CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		) 
		RETURNING public_id
	`

	publicID := uuid.New().String()
	code := generateTicketCode(req.EventId, req.UserId)

	var resultPublicID string
	err = db.Pool.QueryRow(ctx, query, publicID, categoryInternalID, eventInternalID, code).Scan(&resultPublicID)
	if err != nil {
		return "", fmt.Errorf("error creating ticket: %v", err)
	}

	log.Printf("Ticket created: %s for event %s", resultPublicID, req.EventId)
	return resultPublicID, nil
}

// GetTicketsByUserID obtiene todos los tickets de un usuario por user_id (public_id del customer)
func (r *TicketRepository) GetTicketsByUserID(ctx context.Context, userID string) ([]*models.Ticket, error) {
	// En tu esquema, tickets se obtienen via transactions -> customers
	// Esta función necesita ser reimplementada para tu esquema complejo
	// Por ahora, devolvemos un array vacío
	log.Printf("GetTicketsByUserID called with userID: %s - Function needs adaptation for complex schema", userID)
	return []*models.Ticket{}, nil
}

// GetTicketsByCustomerID obtiene todos los tickets de un cliente por customer_id
func (r *TicketRepository) GetTicketsByCustomerID(ctx context.Context, customerPublicID string) ([]*models.Ticket, error) {
	// En tu esquema real, tickets se relacionan via transactions
	// Esta es una implementación simplificada que necesita ajustarse
	query := `
		SELECT t.id, t.public_id, t.category_id, t.event_id, t.code, t.status,
		       t.seat_number, t.qr_code_url, t.price, t.used_at, t.transferred_from_ticket_id,
		       t.created_at, t.updated_at
		FROM tickets t
		INNER JOIN transactions tr ON t.transaction_id = tr.id
		INNER JOIN customers c ON tr.customer_id = c.id
		WHERE c.public_id = $1
		ORDER BY t.created_at DESC
	`

	rows, err := db.Pool.Query(ctx, query, customerPublicID)
	if err != nil {
		// Si falla, puede ser porque la estructura no está completa
		// Por ahora, devolvemos vacío en lugar de error
		log.Printf("Error querying tickets by customer (schema may need adjustment): %v", err)
		return []*models.Ticket{}, nil
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
			return nil, fmt.Errorf("error scanning ticket: %w", err)
		}
		tickets = append(tickets, &ticket)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tickets: %w", err)
	}

	log.Printf("Found %d tickets for customer: %s", len(tickets), customerPublicID)
	return tickets, nil
}

// validateEventAndCategory valida que el evento y categoría existan
func (r *TicketRepository) validateEventAndCategory(ctx context.Context, eventPublicID, categoryPublicID string) error {
	// Validar que el evento existe y está activo
	var eventExists bool
	err := db.Pool.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM events WHERE public_id = $1 AND is_active = true)",
		eventPublicID).Scan(&eventExists)

	if err != nil {
		return fmt.Errorf("error validating event: %v", err)
	}
	if !eventExists {
		return fmt.Errorf("event not found or inactive: %s", eventPublicID)
	}

	// Validar que la categoría existe y está activa
	var categoryExists bool
	err = db.Pool.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM categories WHERE public_id = $1 AND is_active = true)",
		categoryPublicID).Scan(&categoryExists)

	if err != nil {
		return fmt.Errorf("error validating category: %v", err)
	}
	if !categoryExists {
		return fmt.Errorf("category not found or inactive: %s", categoryPublicID)
	}

	// Validar que la categoría pertenece al evento
	var validCategory bool
	err = db.Pool.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM categories c 
			INNER JOIN events e ON c.event_id = e.id 
			WHERE c.public_id = $1 AND e.public_id = $2
		)`, categoryPublicID, eventPublicID).Scan(&validCategory)

	if err != nil {
		return fmt.Errorf("error validating category-event relationship: %v", err)
	}
	if !validCategory {
		return fmt.Errorf("category %s does not belong to event %s", categoryPublicID, eventPublicID)
	}

	return nil
}

// GetTicketByID obtiene un ticket por ID interno
func (r *TicketRepository) GetTicketByID(ctx context.Context, id int64) (*models.Ticket, error) {
	query := `
		SELECT id, public_id, category_id, event_id, code, status,
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
		SELECT id, public_id, category_id, event_id, code, status,
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
		SELECT id, public_id, category_id, event_id, code, status,
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

	log.Printf("Ticket status updated: ID %d (%s -> %s)", ticketID, oldTicket.Status, status)
	return nil
}

// ListTickets lista tickets con filtros y paginación
func (r *TicketRepository) ListTickets(ctx context.Context, filters *TicketFilters, page, pageSize int) ([]*models.Ticket, *PaginationInfo, error) {
	whereClause, params := r.buildTicketWhereClause(filters)
	params = append(params, pageSize, (page-1)*pageSize)

	query := fmt.Sprintf(`
		SELECT id, public_id, category_id, event_id, code, status,
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

// Helper functions para mapear public_ids a internal_ids

func (r *TicketRepository) getCustomerIDByPublicID(ctx context.Context, customerPublicID string) (int64, error) {
	var customerID int64
	err := db.Pool.QueryRow(ctx,
		"SELECT id FROM customers WHERE public_id = $1",
		customerPublicID).Scan(&customerID)
	if err != nil {
		return 0, fmt.Errorf("customer not found with public_id: %s", customerPublicID)
	}
	return customerID, nil
}

func (r *TicketRepository) getCategoryIDByPublicID(ctx context.Context, categoryPublicID string) (int64, error) {
	var categoryID int64
	err := db.Pool.QueryRow(ctx,
		"SELECT id FROM categories WHERE public_id = $1",
		categoryPublicID).Scan(&categoryID)
	if err != nil {
		return 0, fmt.Errorf("category not found with public_id: %s", categoryPublicID)
	}
	return categoryID, nil
}

func (r *TicketRepository) getEventIDByPublicID(ctx context.Context, eventPublicID string) (int64, error) {
	var eventID int64
	err := db.Pool.QueryRow(ctx,
		"SELECT id FROM events WHERE public_id = $1",
		eventPublicID).Scan(&eventID)
	if err != nil {
		return 0, fmt.Errorf("event not found with public_id: %s", eventPublicID)
	}
	return eventID, nil
}

// Helper types and functions

type TicketFilters struct {
	EventID    *int64
	CategoryID *int64
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
	timestamp := time.Now().UnixNano()
	shortTimestamp := fmt.Sprintf("%d", timestamp)[:8]
	return fmt.Sprintf("TKT-%s-%s-%s", eventID, userID, shortTimestamp)
}
