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
	"github.com/jackc/pgx/v5/pgtype"
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

// CreateTicket crea un nuevo ticket - CORREGIDO
func (r *TicketRepository) CreateTicket(ctx context.Context, req *pb.TicketRequest) (string, error) {
	// Validar y limpiar datos
	eventID := strings.TrimSpace(req.EventId)
	userID := strings.TrimSpace(req.UserId)
	categoryID := strings.TrimSpace(req.CategoryId)

	if eventID == "" {
		return "", fmt.Errorf("event_id is required")
	}
	if userID == "" {
		return "", fmt.Errorf("user_id is required")
	}
	if categoryID == "" {
		return "", fmt.Errorf("category_id is required")
	}

	// Validar UUIDs
	if _, err := uuid.Parse(eventID); err != nil {
		return "", fmt.Errorf("invalid event ID format: must be a valid UUID")
	}
	if _, err := uuid.Parse(userID); err != nil {
		return "", fmt.Errorf("invalid user ID format: must be a valid UUID")
	}
	if _, err := uuid.Parse(categoryID); err != nil {
		return "", fmt.Errorf("invalid category ID format: must be a valid UUID")
	}

	// Validar existencia del event_id y category_id
	if err := r.validateEventAndCategory(ctx, eventID, categoryID); err != nil {
		return "", err
	}

	// Obtener el customer_id basado en el user_id (public_id del customer)
	customerID, err := r.getCustomerIDByPublicID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("error finding customer: %v", err)
	}

	// Obtener category_id interno basado en public_id
	categoryInternalID, err := r.getCategoryIDByPublicID(ctx, categoryID)
	if err != nil {
		return "", fmt.Errorf("error finding category: %v", err)
	}

	// Obtener event_id interno basado en public_id
	eventInternalID, err := r.getEventIDByPublicID(ctx, eventID)
	if err != nil {
		return "", fmt.Errorf("error finding event: %v", err)
	}

	// Verificar disponibilidad en la categoría
	if err := r.checkCategoryAvailability(ctx, categoryInternalID); err != nil {
		return "", err
	}

	// Obtener precio de la categoría
	categoryPrice, err := r.getCategoryPrice(ctx, categoryInternalID)
	if err != nil {
		return "", fmt.Errorf("error getting category price: %v", err)
	}

	// Crear múltiples tickets según quantity
	quantity := int(req.Quantity)
	if quantity <= 0 {
		quantity = 1
	}

	var createdTicketPublicID string

	// Usar transacción para crear múltiples tickets atómicamente
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback(ctx)

	for i := 0; i < quantity; i++ {
		publicID := uuid.New().String()
		code := generateTicketCode(eventID, userID, i)

		query := `
			INSERT INTO tickets (
				public_id, category_id, event_id, code, status, price, 
				created_at, updated_at
			) 
			VALUES ($1, $2, $3, $4, 'available', $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
			RETURNING public_id
		`

		var ticketPublicID string
		err := tx.QueryRow(ctx, query,
			publicID,
			categoryInternalID,
			eventInternalID,
			code,
			categoryPrice,
		).Scan(&ticketPublicID)

		if err != nil {
			return "", fmt.Errorf("error creating ticket %d/%d: %v", i+1, quantity, err)
		}

		// Guardar el public_id del primer ticket creado para retornar
		if i == 0 {
			createdTicketPublicID = ticketPublicID
		}
	}

	// Actualizar contador de tickets vendidos en la categoría
	updateQuery := `
		UPDATE categories 
		SET quantity_sold = quantity_sold + $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`
	_, err = tx.Exec(ctx, updateQuery, quantity, categoryInternalID)
	if err != nil {
		return "", fmt.Errorf("error updating category sold count: %v", err)
	}

	// Commit de la transacción
	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("error committing transaction: %v", err)
	}

	log.Printf("Tickets created: %d for event %s, customer ID: %d", quantity, eventID, customerID)
	return createdTicketPublicID, nil
}

// GetTicketsByCustomerID obtiene todos los tickets de un cliente por customer_id - CORREGIDO
func (r *TicketRepository) GetTicketsByCustomerID(ctx context.Context, customerPublicID string) ([]*models.Ticket, error) {
	// Consulta mejorada que une tickets con transactions y customers
	query := `
		SELECT t.id, t.public_id, t.category_id, t.transaction_id, t.event_id, 
		       t.code, t.status, t.seat_number, t.qr_code_url, t.price, 
		       t.used_at, t.transferred_from_ticket_id, t.created_at, t.updated_at
		FROM tickets t
		LEFT JOIN transactions tr ON t.transaction_id = tr.id
		LEFT JOIN customers c ON tr.customer_id = c.id
		WHERE c.public_id = $1 OR (
			-- También incluir tickets que estén reservados para el cliente
			t.transaction_id IS NULL 
			AND EXISTS (
				SELECT 1 FROM customers c2 
				WHERE c2.public_id = $1 
				AND t.status IN ('reserved', 'sold')
			)
		)
		ORDER BY t.created_at DESC
	`

	rows, err := db.Pool.Query(ctx, query, customerPublicID)
	if err != nil {
		log.Printf("Error querying tickets by customer: %v", err)
		// Fallback a consulta simplificada
		return r.getTicketsDirectFallback(ctx, customerPublicID)
	}
	defer rows.Close()

	var tickets []*models.Ticket
	for rows.Next() {
		var ticket models.Ticket
		var transactionID pgtype.Int4
		var usedAt pgtype.Timestamp
		var transferredFromTicketID pgtype.Int4

		err := rows.Scan(
			&ticket.ID,
			&ticket.PublicID,
			&ticket.CategoryID,
			&transactionID,
			&ticket.EventID,
			&ticket.Code,
			&ticket.Status,
			&ticket.SeatNumber,
			&ticket.QRCodeURL,
			&ticket.Price,
			&usedAt,
			&transferredFromTicketID,
			&ticket.CreatedAt,
			&ticket.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning ticket: %w", err)
		}

		// Asignar campos pgtype si son válidos
		if transactionID.Valid {
			ticket.TransactionID = transactionID
		}
		if usedAt.Valid {
			ticket.UsedAt = usedAt
		}
		if transferredFromTicketID.Valid {
			ticket.TransferredFromTicketID = transferredFromTicketID
		}

		tickets = append(tickets, &ticket)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tickets: %w", err)
	}

	log.Printf("Found %d tickets for customer: %s", len(tickets), customerPublicID)
	return tickets, nil
}

// getTicketsDirectFallback es un fallback si la consulta principal falla - MEJORADO
func (r *TicketRepository) getTicketsDirectFallback(ctx context.Context, customerPublicID string) ([]*models.Ticket, error) {
	// Consulta simplificada que devuelve tickets disponibles como fallback
	query := `
		SELECT id, public_id, category_id, transaction_id, event_id, 
		       code, status, seat_number, qr_code_url, price, 
		       used_at, transferred_from_ticket_id, created_at, updated_at
		FROM tickets 
		WHERE status IN ('available', 'reserved')
		ORDER BY created_at DESC
		LIMIT 20
	`

	rows, err := db.Pool.Query(ctx, query)
	if err != nil {
		log.Printf("Error in fallback ticket query: %v", err)
		return []*models.Ticket{}, nil
	}
	defer rows.Close()

	var tickets []*models.Ticket
	for rows.Next() {
		var ticket models.Ticket
		var transactionID pgtype.Int4
		var usedAt pgtype.Timestamp
		var transferredFromTicketID pgtype.Int4

		err := rows.Scan(
			&ticket.ID,
			&ticket.PublicID,
			&ticket.CategoryID,
			&transactionID,
			&ticket.EventID,
			&ticket.Code,
			&ticket.Status,
			&ticket.SeatNumber,
			&ticket.QRCodeURL,
			&ticket.Price,
			&usedAt,
			&transferredFromTicketID,
			&ticket.CreatedAt,
			&ticket.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning ticket: %w", err)
		}

		// Asignar campos pgtype si son válidos
		if transactionID.Valid {
			ticket.TransactionID = transactionID
		}
		if usedAt.Valid {
			ticket.UsedAt = usedAt
		}
		if transferredFromTicketID.Valid {
			ticket.TransferredFromTicketID = transferredFromTicketID
		}

		tickets = append(tickets, &ticket)
	}

	log.Printf("Fallback: Found %d tickets", len(tickets))
	return tickets, nil
}

// validateEventAndCategory valida que el evento y categoría existan - CORREGIDO
func (r *TicketRepository) validateEventAndCategory(ctx context.Context, eventPublicID, categoryPublicID string) error {
	// Validar que el evento existe y está activo
	var eventExists bool
	err := db.Pool.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM events WHERE public_id = $1 AND is_active = true AND is_published = true)",
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

// checkCategoryAvailability verifica si hay tickets disponibles en la categoría - CORREGIDO
func (r *TicketRepository) checkCategoryAvailability(ctx context.Context, categoryID int64) error {
	var available, sold int32
	err := db.Pool.QueryRow(ctx,
		"SELECT quantity_available, quantity_sold FROM categories WHERE id = $1",
		categoryID).Scan(&available, &sold)

	if err != nil {
		return fmt.Errorf("error checking category availability: %v", err)
	}

	if available <= sold {
		return fmt.Errorf("no tickets available in this category")
	}

	return nil
}

// getCategoryPrice obtiene el precio de una categoría - NUEVO
func (r *TicketRepository) getCategoryPrice(ctx context.Context, categoryID int64) (float64, error) {
	var price float64
	err := db.Pool.QueryRow(ctx,
		"SELECT price FROM categories WHERE id = $1",
		categoryID).Scan(&price)

	if err != nil {
		return 0, fmt.Errorf("error getting category price: %v", err)
	}

	return price, nil
}

// GetTicketByID obtiene un ticket por ID interno - CORREGIDO
func (r *TicketRepository) GetTicketByID(ctx context.Context, id int64) (*models.Ticket, error) {
	query := `
		SELECT id, public_id, category_id, transaction_id, event_id, 
		       code, status, seat_number, qr_code_url, price, 
		       used_at, transferred_from_ticket_id, created_at, updated_at
		FROM tickets 
		WHERE id = $1
	`

	var ticket models.Ticket
	var transactionID pgtype.Int4
	var usedAt pgtype.Timestamp
	var transferredFromTicketID pgtype.Int4

	err := db.Pool.QueryRow(ctx, query, id).Scan(
		&ticket.ID,
		&ticket.PublicID,
		&ticket.CategoryID,
		&transactionID,
		&ticket.EventID,
		&ticket.Code,
		&ticket.Status,
		&ticket.SeatNumber,
		&ticket.QRCodeURL,
		&ticket.Price,
		&usedAt,
		&transferredFromTicketID,
		&ticket.CreatedAt,
		&ticket.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("ticket not found with id: %d", id)
		}
		return nil, fmt.Errorf("error getting ticket: %v", err)
	}

	// Asignar campos pgtype si son válidos
	if transactionID.Valid {
		ticket.TransactionID = transactionID
	}
	if usedAt.Valid {
		ticket.UsedAt = usedAt
	}
	if transferredFromTicketID.Valid {
		ticket.TransferredFromTicketID = transferredFromTicketID
	}

	return &ticket, nil
}

// GetTicketByCode obtiene un ticket por código - CORREGIDO
func (r *TicketRepository) GetTicketByCode(ctx context.Context, code string) (*models.Ticket, error) {
	query := `
		SELECT id, public_id, category_id, transaction_id, event_id, 
		       code, status, seat_number, qr_code_url, price, 
		       used_at, transferred_from_ticket_id, created_at, updated_at
		FROM tickets 
		WHERE code = $1
	`

	var ticket models.Ticket
	var transactionID pgtype.Int4
	var usedAt pgtype.Timestamp
	var transferredFromTicketID pgtype.Int4

	err := db.Pool.QueryRow(ctx, query, code).Scan(
		&ticket.ID,
		&ticket.PublicID,
		&ticket.CategoryID,
		&transactionID,
		&ticket.EventID,
		&ticket.Code,
		&ticket.Status,
		&ticket.SeatNumber,
		&ticket.QRCodeURL,
		&ticket.Price,
		&usedAt,
		&transferredFromTicketID,
		&ticket.CreatedAt,
		&ticket.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("ticket not found with code: %s", code)
		}
		return nil, fmt.Errorf("error getting ticket by code: %v", err)
	}

	// Asignar campos pgtype si son válidos
	if transactionID.Valid {
		ticket.TransactionID = transactionID
	}
	if usedAt.Valid {
		ticket.UsedAt = usedAt
	}
	if transferredFromTicketID.Valid {
		ticket.TransferredFromTicketID = transferredFromTicketID
	}

	return &ticket, nil
}

// GetTicketByPublicID obtiene un ticket por public_id - CORREGIDO
func (r *TicketRepository) GetTicketByPublicID(ctx context.Context, publicID string) (*models.Ticket, error) {
	query := `
		SELECT id, public_id, category_id, transaction_id, event_id, 
		       code, status, seat_number, qr_code_url, price, 
		       used_at, transferred_from_ticket_id, created_at, updated_at
		FROM tickets 
		WHERE public_id = $1
	`

	var ticket models.Ticket
	var transactionID pgtype.Int4
	var usedAt pgtype.Timestamp
	var transferredFromTicketID pgtype.Int4

	err := db.Pool.QueryRow(ctx, query, publicID).Scan(
		&ticket.ID,
		&ticket.PublicID,
		&ticket.CategoryID,
		&transactionID,
		&ticket.EventID,
		&ticket.Code,
		&ticket.Status,
		&ticket.SeatNumber,
		&ticket.QRCodeURL,
		&ticket.Price,
		&usedAt,
		&transferredFromTicketID,
		&ticket.CreatedAt,
		&ticket.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("ticket not found with public_id: %s", publicID)
		}
		return nil, fmt.Errorf("error getting ticket by public_id: %v", err)
	}

	// Asignar campos pgtype si son válidos
	if transactionID.Valid {
		ticket.TransactionID = transactionID
	}
	if usedAt.Valid {
		ticket.UsedAt = usedAt
	}
	if transferredFromTicketID.Valid {
		ticket.TransferredFromTicketID = transferredFromTicketID
	}

	return &ticket, nil
}

// UpdateTicketStatus actualiza el estado de un ticket con validación - CORREGIDO
func (r *TicketRepository) UpdateTicketStatus(ctx context.Context, ticketID int64, status string) error {
	// Validar estado
	status = strings.ToLower(strings.TrimSpace(status))
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

// Helper functions para mapear public_ids a internal_ids - CORREGIDOS

func (r *TicketRepository) getCustomerIDByPublicID(ctx context.Context, customerPublicID string) (int64, error) {
	var customerID int64
	err := db.Pool.QueryRow(ctx,
		"SELECT id FROM customers WHERE public_id = $1 AND is_verified = true",
		customerPublicID).Scan(&customerID)
	if err != nil {
		return 0, fmt.Errorf("customer not found with public_id: %s", customerPublicID)
	}
	return customerID, nil
}

func (r *TicketRepository) getCategoryIDByPublicID(ctx context.Context, categoryPublicID string) (int64, error) {
	var categoryID int64
	err := db.Pool.QueryRow(ctx,
		"SELECT id FROM categories WHERE public_id = $1 AND is_active = true",
		categoryPublicID).Scan(&categoryID)
	if err != nil {
		return 0, fmt.Errorf("category not found with public_id: %s", categoryPublicID)
	}
	return categoryID, nil
}

func (r *TicketRepository) getEventIDByPublicID(ctx context.Context, eventPublicID string) (int64, error) {
	var eventID int64
	err := db.Pool.QueryRow(ctx,
		"SELECT id FROM events WHERE public_id = $1 AND is_active = true AND is_published = true",
		eventPublicID).Scan(&eventID)
	if err != nil {
		return 0, fmt.Errorf("event not found with public_id: %s", eventPublicID)
	}
	return eventID, nil
}

// Helper types and functions - CORREGIDOS

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

// ListTickets lista tickets con filtros y paginación - CORREGIDO
func (r *TicketRepository) ListTickets(ctx context.Context, filters *TicketFilters, page, pageSize int) ([]*models.Ticket, *PaginationInfo, error) {
	whereClause, params := r.buildTicketWhereClause(filters)

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	params = append(params, pageSize, (page-1)*pageSize)

	query := fmt.Sprintf(`
		SELECT id, public_id, category_id, transaction_id, event_id, 
		       code, status, seat_number, qr_code_url, price, 
		       used_at, transferred_from_ticket_id, created_at, updated_at
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
		var transactionID pgtype.Int4
		var usedAt pgtype.Timestamp
		var transferredFromTicketID pgtype.Int4

		err := rows.Scan(
			&ticket.ID,
			&ticket.PublicID,
			&ticket.CategoryID,
			&transactionID,
			&ticket.EventID,
			&ticket.Code,
			&ticket.Status,
			&ticket.SeatNumber,
			&ticket.QRCodeURL,
			&ticket.Price,
			&usedAt,
			&transferredFromTicketID,
			&ticket.CreatedAt,
			&ticket.UpdatedAt,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("error scanning ticket: %v", err)
		}

		// Asignar campos pgtype si son válidos
		if transactionID.Valid {
			ticket.TransactionID = transactionID
		}
		if usedAt.Valid {
			ticket.UsedAt = usedAt
		}
		if transferredFromTicketID.Valid {
			ticket.TransferredFromTicketID = transferredFromTicketID
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
			status := strings.ToLower(strings.TrimSpace(*filters.Status))
			if r.isValidTicketStatus(status) {
				conditions = append(conditions, fmt.Sprintf("status = $%d", paramCount))
				params = append(params, status)
				paramCount++
			}
		}
		if filters.Code != nil {
			conditions = append(conditions, fmt.Sprintf("code ILIKE $%d", paramCount))
			params = append(params, "%"+strings.TrimSpace(*filters.Code)+"%")
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
	if totalPages == 0 {
		totalPages = 1
	}

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

func generateTicketCode(eventID, userID string, index int) string {
	timestamp := time.Now().UnixNano()
	shortTimestamp := fmt.Sprintf("%d", timestamp)[:8]
	// Usar solo los primeros 8 caracteres de los UUIDs para mantener el código legible
	shortEventID := eventID[:8]
	shortUserID := userID[:8]
	return fmt.Sprintf("TKT-%s-%s-%s-%d", shortEventID, shortUserID, shortTimestamp, index)
}
