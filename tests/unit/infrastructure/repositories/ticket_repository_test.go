// ticket_repository.go - COMPLETAMENTE CORREGIDO Y FUNCIONAL
package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	pb "github.com/franciscozamorau/osmi-server/gen"
	"github.com/franciscozamorau/osmi-server/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TicketRepository struct {
	db *pgxpool.Pool
}

func NewTicketRepository(db *pgxpool.Pool) *TicketRepository {
	return &TicketRepository{db: db}
}

// Valid ticket statuses (ACTUALIZADO con la nueva base de datos)
var validTicketStatuses = map[string]bool{
	"available":   true,
	"reserved":    true,
	"sold":        true,
	"used":        true,
	"cancelled":   true,
	"transferred": true,
	"refunded":    true, // ✅ NUEVO ESTADO de la base de datos
}

// Valid status transitions (ACTUALIZADO)
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
		"refunded":    true, // ✅ NUEVA TRANSICIÓN
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
	"refunded": {
		// No transitions from refunded
	},
}

// CreateTicket crea un nuevo ticket (COMPLETAMENTE CORREGIDO Y OPTIMIZADO)
func (r *TicketRepository) CreateTicket(ctx context.Context, req *pb.TicketRequest) (string, error) {
	// Validar y limpiar datos
	eventID := strings.TrimSpace(req.EventId)
	categoryID := strings.TrimSpace(req.CategoryId)
	customerID := strings.TrimSpace(req.CustomerId) // ✅ OBLIGATORIO
	userID := strings.TrimSpace(req.UserId)         // ✅ OPCIONAL

	// Validaciones básicas
	if eventID == "" {
		return "", fmt.Errorf("event_id is required")
	}
	if categoryID == "" {
		return "", fmt.Errorf("category_id is required")
	}
	if customerID == "" {
		return "", fmt.Errorf("customer_id is required") // ✅ CUSTOMER_ID OBLIGATORIO
	}

	// Validar UUIDs usando helpers
	if !IsValidUUID(eventID) { // ✅ USANDO HELPER
		return "", fmt.Errorf("invalid event ID format: must be a valid UUID")
	}
	if !IsValidUUID(categoryID) { // ✅ USANDO HELPER
		return "", fmt.Errorf("invalid category ID format: must be a valid UUID")
	}
	if !IsValidUUID(customerID) { // ✅ USANDO HELPER
		return "", fmt.Errorf("invalid customer ID format: must be a valid UUID")
	}
	if userID != "" && !IsValidUUID(userID) { // ✅ USANDO HELPER
		return "", fmt.Errorf("invalid user ID format: must be a valid UUID")
	}

	// Validar existencia del event_id y category_id
	if err := r.validateEventAndCategory(ctx, eventID, categoryID); err != nil {
		return "", err
	}

	// ✅ CORREGIDO: Buscar customer_id interno (OBLIGATORIO)
	customerInternalID, err := r.getCustomerIDByPublicID(ctx, customerID)
	if err != nil {
		return "", fmt.Errorf("error finding customer: %w", err)
	}

	// ✅ CORREGIDO: pgtype.Int4 porque user_id es integer en BD
	var userInternalID pgtype.Int4
	if userID != "" {
		uid, err := r.getUserIDByPublicID(ctx, userID)
		if err != nil {
			return "", fmt.Errorf("error finding user: %w", err)
		}
		userInternalID = ToPgInt4FromInt64(uid) // ✅ Usar ToPgInt4FromInt64
	} else {
		userInternalID = pgtype.Int4{Valid: false} // ✅ Explícitamente NULL
	}

	// Obtener IDs internos
	categoryInternalID, err := r.getCategoryIDByPublicID(ctx, categoryID)
	if err != nil {
		return "", fmt.Errorf("error finding category: %w", err)
	}

	eventInternalID, err := r.getEventIDByPublicID(ctx, eventID)
	if err != nil {
		return "", fmt.Errorf("error finding event: %w", err)
	}

	// Verificar disponibilidad en la categoría
	if err := r.checkCategoryAvailability(ctx, categoryInternalID); err != nil {
		return "", err
	}

	// Obtener precio de la categoría
	categoryPrice, err := r.getCategoryPrice(ctx, categoryInternalID)
	if err != nil {
		return "", fmt.Errorf("error getting category price: %w", err)
	}

	// Crear múltiples tickets según quantity
	quantity := int(req.Quantity)
	if quantity <= 0 {
		quantity = 1
	}
	if quantity > 10 { // Límite razonable
		return "", fmt.Errorf("cannot create more than 10 tickets at once")
	}

	var createdTicketPublicID string

	// Usar transacción para crear múltiples tickets atómicamente
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	for i := 0; i < quantity; i++ {
		publicID := uuid.New().String()
		code, err := r.generateUniqueTicketCode(ctx, tx, eventID, customerID, i)
		if err != nil {
			return "", fmt.Errorf("error generating ticket code: %w", err)
		}

		// ✅ CORREGIDO: Insertar con customer_id (obligatorio) y user_id (opcional)
		query := `INSERT INTO tickets (
			public_id, category_id, event_id, customer_id, user_id,
			code, status, price, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, 'available', $7, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING public_id`

		var ticketPublicID string
		err = tx.QueryRow(ctx, query,
			publicID,
			categoryInternalID,
			eventInternalID,
			customerInternalID, // $4 - OBLIGATORIO
			userInternalID,     // $5 - OPCIONAL (puede ser NULL)
			code,               // $6
			categoryPrice,      // $7
		).Scan(&ticketPublicID)

		if err != nil {
			if IsDuplicateKeyError(err) { // ✅ USANDO HELPER
				return "", fmt.Errorf("ticket with code %s already exists", SafeStringForLog(code))
			}
			return "", fmt.Errorf("error creating ticket %d/%d: %w", i+1, quantity, err)
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
		return "", fmt.Errorf("error updating category sold count: %w", err)
	}

	// Commit de la transacción
	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("error committing transaction: %w", err)
	}

	log.Printf("Tickets created: %d for event %s, customer ID: %d",
		quantity, SafeStringForLog(eventID), customerInternalID)
	return createdTicketPublicID, nil
}

// GetTicketsByUserID obtiene todos los tickets de un usuario por user_id (MEJORADO)
func (r *TicketRepository) GetTicketsByUserID(ctx context.Context, userPublicID string) ([]*models.Ticket, error) {
	if !IsValidUUID(userPublicID) { // ✅ USANDO HELPER
		return nil, fmt.Errorf("invalid user ID format")
	}

	query := `
		SELECT t.id, t.public_id, t.category_id, t.transaction_id, t.event_id, 
			   t.customer_id, t.user_id, t.code, t.status, t.seat_number, 
			   t.qr_code_url, t.price, t.used_at, t.transferred_from_ticket_id, 
			   t.created_at, t.updated_at
		FROM tickets t
		INNER JOIN users u ON t.user_id = u.id
		WHERE u.public_id = $1
		ORDER BY t.created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userPublicID)
	if err != nil {
		return nil, fmt.Errorf("error querying tickets by user: %w", err)
	}
	defer rows.Close()

	return r.scanTicketsFromRows(rows, "user", userPublicID)
}

// GetTicketsByCustomerID obtiene todos los tickets de un cliente por customer_id (MEJORADO)
func (r *TicketRepository) GetTicketsByCustomerID(ctx context.Context, customerPublicID string) ([]*models.Ticket, error) {
	if !IsValidUUID(customerPublicID) { // ✅ USANDO HELPER
		return nil, fmt.Errorf("invalid customer ID format")
	}

	query := `
		SELECT t.id, t.public_id, t.category_id, t.transaction_id, t.event_id, 
			   t.customer_id, t.user_id, t.code, t.status, t.seat_number, 
			   t.qr_code_url, t.price, t.used_at, t.transferred_from_ticket_id, 
			   t.created_at, t.updated_at
		FROM tickets t
		INNER JOIN customers c ON t.customer_id = c.id
		WHERE c.public_id = $1
		ORDER BY t.created_at DESC
	`

	rows, err := r.db.Query(ctx, query, customerPublicID)
	if err != nil {
		return nil, fmt.Errorf("error querying tickets by customer: %w", err)
	}
	defer rows.Close()

	return r.scanTicketsFromRows(rows, "customer", customerPublicID)
}

// GetTicketWithDetails obtiene un ticket con información completa
func (r *TicketRepository) GetTicketWithDetails(ctx context.Context, ticketPublicID string) (*models.TicketWithDetails, error) {
	if !IsValidUUID(ticketPublicID) {
		return nil, fmt.Errorf("invalid ticket ID format")
	}

	// ✅ QUERY CORREGIDO: 'c' → 'cat' y 'e' → 'ev'
	query := `
		SELECT 
			t.public_id, t.code, t.status, t.seat_number, t.price, t.created_at, t.used_at,
			cat.public_id as category_id, cat.name as category_name,
			ev.public_id as event_id, ev.name as event_name, ev.start_date, ev.location,
			cust.public_id as customer_id, cust.name as customer_name, cust.email as customer_email,
			cust.customer_type,
			u.public_id as user_id, u.username as user_name, u.role as user_role,
			trans.public_id as transaction_id, trans.status as transaction_status
		FROM tickets t
		LEFT JOIN categories cat ON t.category_id = cat.id
		LEFT JOIN events ev ON t.event_id = ev.id
		LEFT JOIN customers cust ON t.customer_id = cust.id
		LEFT JOIN users u ON t.user_id = u.id
		LEFT JOIN transactions trans ON t.transaction_id = trans.id
		WHERE t.public_id = $1
	`

	var details models.TicketWithDetails
	var usedAt pgtype.Timestamp
	var userID pgtype.Text
	var userName pgtype.Text
	var userRole pgtype.Text
	var seatNumber pgtype.Text
	var transactionID pgtype.Text
	var transactionStatus pgtype.Text

	err := r.db.QueryRow(ctx, query, ticketPublicID).Scan(
		&details.TicketID,
		&details.Code,
		&details.Status,
		&seatNumber,
		&details.Price,
		&details.CreatedAt,
		&usedAt,
		&details.CategoryID,
		&details.CategoryName,
		&details.EventID,
		&details.EventName,
		&details.StartDate,
		&details.Location,
		&details.CustomerID,
		&details.CustomerName,
		&details.CustomerEmail,
		&details.CustomerType,
		&userID,
		&userName,
		&userRole,
		&transactionID,
		&transactionStatus,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("ticket not found: %s", ticketPublicID)
		}
		return nil, fmt.Errorf("error getting ticket details: %w", err)
	}

	// Procesar valores NULL
	if usedAt.Valid {
		details.UsedAt = &usedAt.Time
	}
	if seatNumber.Valid {
		details.SeatNumber = seatNumber.String
	}
	if userID.Valid {
		details.UserID = &userID.String
	}
	if userName.Valid {
		details.UserName = &userName.String
	}
	if userRole.Valid {
		details.UserRole = &userRole.String
	}
	if transactionID.Valid {
		details.TransactionID = &transactionID.String
	}
	if transactionStatus.Valid {
		details.TransactionStatus = &transactionStatus.String
	}

	return &details, nil
}

// GetTicketByPublicID obtiene un ticket por public_id (COMPLETAMENTE CORREGIDO)
func (r *TicketRepository) GetTicketByPublicID(ctx context.Context, publicID string) (*models.Ticket, error) {
	if !IsValidUUID(publicID) { // ✅ USANDO HELPER
		return nil, fmt.Errorf("invalid ticket ID format")
	}

	query := `
		SELECT id, public_id, category_id, transaction_id, event_id, 
		       customer_id, user_id, code, status, seat_number, 
		       qr_code_url, price, used_at, transferred_from_ticket_id, 
		       created_at, updated_at
		FROM tickets 
		WHERE public_id = $1
	`

	var ticket models.Ticket
	var transactionID pgtype.Int8
	var userID pgtype.Int4 // ✅ CAMBIADO a Int4
	var seatNumber pgtype.Text
	var qrCodeURL pgtype.Text
	var usedAt pgtype.Timestamp
	var transferredFromTicketID pgtype.Int8

	err := r.db.QueryRow(ctx, query, publicID).Scan(
		&ticket.ID,
		&ticket.PublicID,
		&ticket.CategoryID,
		&transactionID,
		&ticket.EventID,
		&ticket.CustomerID, // ✅ DIRECTAMENTE al campo del modelo
		&userID,
		&ticket.Code,
		&ticket.Status,
		&seatNumber,
		&qrCodeURL,
		&ticket.Price,
		&usedAt,
		&transferredFromTicketID,
		&ticket.CreatedAt,
		&ticket.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("ticket not found with public_id: %s", publicID)
		}
		return nil, fmt.Errorf("error getting ticket by public_id: %w", err)
	}

	// Convertir pgtype a tipos nativos usando helpers
	ticket.TransactionID = ToInt64FromPgInt8(transactionID)
	if userID.Valid {
		uid := int64(userID.Int32)
		ticket.UserID = &uid
	} else {
		ticket.UserID = nil
	}
	ticket.SeatNumber = ToStringFromPgText(seatNumber)
	ticket.QRCodeURL = ToStringFromPgText(qrCodeURL)
	ticket.UsedAt = ToTimeFromPgTimestamp(usedAt)
	ticket.TransferredFromTicketID = ToInt64FromPgInt8(transferredFromTicketID)

	return &ticket, nil
}

// UpdateTicketStatus actualiza el estado de un ticket con validación
func (r *TicketRepository) UpdateTicketStatus(ctx context.Context, ticketPublicID string, status string) error {
	if !IsValidUUID(ticketPublicID) { // ✅ USANDO HELPER
		return fmt.Errorf("invalid ticket ID format")
	}

	// Validar estado
	status = strings.ToLower(strings.TrimSpace(status))
	if !r.isValidTicketStatus(status) {
		return fmt.Errorf("invalid ticket status: %s", status)
	}

	// Obtener ticket antiguo para validar transición
	oldTicket, err := r.GetTicketByPublicID(ctx, ticketPublicID)
	if err != nil {
		return err
	}

	// Validar transición de estado
	if !r.isValidStatusTransition(oldTicket.Status, status) {
		return fmt.Errorf("invalid status transition from %s to %s", oldTicket.Status, status)
	}

	query := `UPDATE tickets SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE public_id = $2`

	result, err := r.db.Exec(ctx, query, status, ticketPublicID)
	if err != nil {
		return fmt.Errorf("error updating ticket status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("ticket not found with public_id: %s", ticketPublicID)
	}

	log.Printf("Ticket status updated: %s (%s -> %s)", SafeStringForLog(ticketPublicID), oldTicket.Status, status)
	return nil
}

// UpdateTicketTransaction actualiza el transaction_id de un ticket
func (r *TicketRepository) UpdateTicketTransaction(ctx context.Context, ticketPublicID string, transactionID int64) error {
	if !IsValidUUID(ticketPublicID) { // ✅ USANDO HELPER
		return fmt.Errorf("invalid ticket ID format")
	}

	query := `UPDATE tickets SET transaction_id = $1, updated_at = CURRENT_TIMESTAMP WHERE public_id = $2`

	result, err := r.db.Exec(ctx, query, transactionID, ticketPublicID)
	if err != nil {
		return fmt.Errorf("error updating ticket transaction: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("ticket not found with public_id: %s", ticketPublicID)
	}

	log.Printf("Ticket transaction updated: %s -> transaction %d", SafeStringForLog(ticketPublicID), transactionID)
	return nil
}

// MarkTicketAsUsed marca un ticket como usado
func (r *TicketRepository) MarkTicketAsUsed(ctx context.Context, ticketPublicID string) error {
	if !IsValidUUID(ticketPublicID) { // ✅ USANDO HELPER
		return fmt.Errorf("invalid ticket ID format")
	}

	query := `
		UPDATE tickets 
		SET status = 'used', used_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP 
		WHERE public_id = $1 AND status = 'sold'
	`

	result, err := r.db.Exec(ctx, query, ticketPublicID)
	if err != nil {
		return fmt.Errorf("error marking ticket as used: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("ticket not found or not in sold status: %s", ticketPublicID)
	}

	log.Printf("Ticket marked as used: %s", SafeStringForLog(ticketPublicID))
	return nil
}

// GetTicketsByEvent obtiene todos los tickets de un evento
func (r *TicketRepository) GetTicketsByEvent(ctx context.Context, eventPublicID string) ([]*models.Ticket, error) {
	if !IsValidUUID(eventPublicID) { // ✅ USANDO HELPER
		return nil, fmt.Errorf("invalid event ID format")
	}

	query := `
		SELECT t.id, t.public_id, t.category_id, t.transaction_id, t.event_id, 
			   t.customer_id, t.user_id, t.code, t.status, t.seat_number, 
			   t.qr_code_url, t.price, t.used_at, t.transferred_from_ticket_id, 
			   t.created_at, t.updated_at
		FROM tickets t
		INNER JOIN events e ON t.event_id = e.id
		WHERE e.public_id = $1
		ORDER BY t.created_at DESC
	`

	rows, err := r.db.Query(ctx, query, eventPublicID)
	if err != nil {
		return nil, fmt.Errorf("error querying tickets by event: %w", err)
	}
	defer rows.Close()

	return r.scanTicketsFromRows(rows, "event", eventPublicID)
}

// GetTicketsByStatus obtiene tickets por estado
func (r *TicketRepository) GetTicketsByStatus(ctx context.Context, status string) ([]*models.Ticket, error) {
	if !r.isValidTicketStatus(status) {
		return nil, fmt.Errorf("invalid ticket status: %s", status)
	}

	query := `
		SELECT id, public_id, category_id, transaction_id, event_id, 
		       customer_id, user_id, code, status, seat_number, 
		       qr_code_url, price, used_at, transferred_from_ticket_id, 
		       created_at, updated_at
		FROM tickets 
		WHERE status = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, status)
	if err != nil {
		return nil, fmt.Errorf("error querying tickets by status: %w", err)
	}
	defer rows.Close()

	return r.scanTicketsFromRows(rows, "status", status)
}

// =============================================================================
// MÉTODOS PRIVADOS
// =============================================================================

// scanTicketsFromRows escanea filas de tickets (método helper reutilizable)
func (r *TicketRepository) scanTicketsFromRows(rows pgx.Rows, entityType, entityID string) ([]*models.Ticket, error) {
	var tickets []*models.Ticket

	for rows.Next() {
		var ticket models.Ticket
		var transactionID pgtype.Int8
		var userID pgtype.Int4 // ✅ CAMBIADO a Int4
		var seatNumber pgtype.Text
		var qrCodeURL pgtype.Text
		var usedAt pgtype.Timestamp
		var transferredFromTicketID pgtype.Int8

		err := rows.Scan(
			&ticket.ID,
			&ticket.PublicID,
			&ticket.CategoryID,
			&transactionID,
			&ticket.EventID,
			&ticket.CustomerID, // ✅ DIRECTAMENTE al campo del modelo
			&userID,
			&ticket.Code,
			&ticket.Status,
			&seatNumber,
			&qrCodeURL,
			&ticket.Price,
			&usedAt,
			&transferredFromTicketID,
			&ticket.CreatedAt,
			&ticket.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning ticket: %w", err)
		}

		// Convertir pgtype a tipos nativos usando helpers
		ticket.TransactionID = ToInt64FromPgInt8(transactionID)
		if userID.Valid {
			uid := int64(userID.Int32)
			ticket.UserID = &uid
		} else {
			ticket.UserID = nil
		}
		ticket.SeatNumber = ToStringFromPgText(seatNumber)
		ticket.QRCodeURL = ToStringFromPgText(qrCodeURL)
		ticket.UsedAt = ToTimeFromPgTimestamp(usedAt)
		ticket.TransferredFromTicketID = ToInt64FromPgInt8(transferredFromTicketID)

		tickets = append(tickets, &ticket)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tickets: %w", err)
	}

	log.Printf("Found %d tickets for %s: %s", len(tickets), entityType, SafeStringForLog(entityID))
	return tickets, nil
}

// validateEventAndCategory valida que el evento y categoría existan
func (r *TicketRepository) validateEventAndCategory(ctx context.Context, eventPublicID, categoryPublicID string) error {
	// Validar que el evento existe y está activo
	var eventExists bool
	err := r.db.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM events WHERE public_id = $1 AND is_active = true AND is_published = true)",
		eventPublicID).Scan(&eventExists)

	if err != nil {
		return fmt.Errorf("error validating event: %w", err)
	}
	if !eventExists {
		return fmt.Errorf("event not found or inactive: %s", eventPublicID)
	}

	// Validar que la categoría existe y está activa
	var categoryExists bool
	err = r.db.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM categories WHERE public_id = $1 AND is_active = true)",
		categoryPublicID).Scan(&categoryExists)

	if err != nil {
		return fmt.Errorf("error validating category: %w", err)
	}
	if !categoryExists {
		return fmt.Errorf("category not found or inactive: %s", categoryPublicID)
	}

	// Validar que la categoría pertenece al evento
	var validCategory bool
	err = r.db.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM categories c 
			INNER JOIN events e ON c.event_id = e.id 
			WHERE c.public_id = $1 AND e.public_id = $2
		)`, categoryPublicID, eventPublicID).Scan(&validCategory)

	if err != nil {
		return fmt.Errorf("error validating category-event relationship: %w", err)
	}
	if !validCategory {
		return fmt.Errorf("category %s does not belong to event %s", categoryPublicID, eventPublicID)
	}

	return nil
}

// checkCategoryAvailability verifica si hay tickets disponibles en la categoría
func (r *TicketRepository) checkCategoryAvailability(ctx context.Context, categoryID int64) error {
	var available, sold int32
	err := r.db.QueryRow(ctx,
		"SELECT quantity_available, quantity_sold FROM categories WHERE id = $1",
		categoryID).Scan(&available, &sold)

	if err != nil {
		return fmt.Errorf("error checking category availability: %w", err)
	}

	if available <= sold {
		return fmt.Errorf("no tickets available in this category")
	}

	return nil
}

// getCategoryPrice obtiene el precio de una categoría
func (r *TicketRepository) getCategoryPrice(ctx context.Context, categoryID int64) (float64, error) {
	var price float64
	err := r.db.QueryRow(ctx,
		"SELECT price FROM categories WHERE id = $1",
		categoryID).Scan(&price)

	if err != nil {
		return 0, fmt.Errorf("error getting category price: %w", err)
	}

	return price, nil
}

// generateUniqueTicketCode genera un código único para el ticket
func (r *TicketRepository) generateUniqueTicketCode(ctx context.Context, tx pgx.Tx, eventID, customerID string, index int) (string, error) {
	maxAttempts := 5
	for attempt := 0; attempt < maxAttempts; attempt++ {
		code := generateTicketCode(eventID, customerID, index+attempt)

		// Verificar si el código ya existe
		var exists bool
		err := tx.QueryRow(ctx,
			"SELECT EXISTS(SELECT 1 FROM tickets WHERE code = $1)",
			code).Scan(&exists)

		if err != nil {
			return "", fmt.Errorf("error checking ticket code uniqueness: %w", err)
		}

		if !exists {
			return code, nil
		}
	}

	return "", fmt.Errorf("failed to generate unique ticket code after %d attempts", maxAttempts)
}

// Helper functions para mapear public_ids a internal_ids
func (r *TicketRepository) getCustomerIDByPublicID(ctx context.Context, customerPublicID string) (int64, error) {
	var customerID int64
	err := r.db.QueryRow(ctx,
		"SELECT id FROM customers WHERE public_id = $1",
		customerPublicID).Scan(&customerID)
	if err != nil {
		return 0, fmt.Errorf("customer not found with public_id: %s", customerPublicID)
	}
	return customerID, nil
}

func (r *TicketRepository) getCategoryIDByPublicID(ctx context.Context, categoryPublicID string) (int64, error) {
	var categoryID int64
	err := r.db.QueryRow(ctx,
		"SELECT id FROM categories WHERE public_id = $1 AND is_active = true",
		categoryPublicID).Scan(&categoryID)
	if err != nil {
		return 0, fmt.Errorf("category not found with public_id: %s", categoryPublicID)
	}
	return categoryID, nil
}

func (r *TicketRepository) getEventIDByPublicID(ctx context.Context, eventPublicID string) (int64, error) {
	var eventID int64
	err := r.db.QueryRow(ctx,
		"SELECT id FROM events WHERE public_id = $1 AND is_active = true AND is_published = true",
		eventPublicID).Scan(&eventID)
	if err != nil {
		return 0, fmt.Errorf("event not found with public_id: %s", eventPublicID)
	}
	return eventID, nil
}

func (r *TicketRepository) getUserIDByPublicID(ctx context.Context, userPublicID string) (int64, error) {
	var userID int64
	err := r.db.QueryRow(ctx,
		"SELECT id FROM users WHERE public_id = $1 AND is_active = true",
		userPublicID).Scan(&userID)
	if err != nil {
		return 0, fmt.Errorf("user not found with public_id: %s", userPublicID)
	}
	return userID, nil
}

// Helper functions de validación
func (r *TicketRepository) isValidTicketStatus(status string) bool {
	return validTicketStatuses[status]
}

func (r *TicketRepository) isValidStatusTransition(from, to string) bool {
	if transitions, exists := validStatusTransitions[from]; exists {
		return transitions[to]
	}
	return false
}

// generateTicketCode genera un código de ticket único
func generateTicketCode(eventID, customerID string, index int) string {
	timestamp := time.Now().UnixNano()
	shortTimestamp := fmt.Sprintf("%d", timestamp)[:8]
	// Usar solo los primeros 8 caracteres de los UUIDs para mantener el código legible
	shortEventID := eventID[:8]
	shortCustomerID := customerID[:8]
	return fmt.Sprintf("TKT-%s-%s-%s-%d", shortEventID, shortCustomerID, shortTimestamp, index)
}

// GetTicketByCode obtiene un ticket por su código
func (r *TicketRepository) GetTicketByCode(ctx context.Context, code string) (*models.Ticket, error) {
	if code == "" {
		return nil, fmt.Errorf("ticket code is required")
	}

	query := `
		SELECT id, public_id, category_id, transaction_id, event_id, 
		       customer_id, user_id, code, status, seat_number, 
		       qr_code_url, price, used_at, transferred_from_ticket_id, 
		       created_at, updated_at
		FROM tickets 
		WHERE code = $1
	`

	var ticket models.Ticket
	var transactionID pgtype.Int8
	var userID pgtype.Int4 // ✅ CAMBIADO a Int4
	var seatNumber pgtype.Text
	var qrCodeURL pgtype.Text
	var usedAt pgtype.Timestamp
	var transferredFromTicketID pgtype.Int8

	err := r.db.QueryRow(ctx, query, code).Scan(
		&ticket.ID,
		&ticket.PublicID,
		&ticket.CategoryID,
		&transactionID,
		&ticket.EventID,
		&ticket.CustomerID, // ✅ DIRECTAMENTE al campo del modelo
		&userID,
		&ticket.Code,
		&ticket.Status,
		&seatNumber,
		&qrCodeURL,
		&ticket.Price,
		&usedAt,
		&transferredFromTicketID,
		&ticket.CreatedAt,
		&ticket.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("ticket not found with code: %s", code)
		}
		return nil, fmt.Errorf("error getting ticket by code: %w", err)
	}

	// Convertir pgtype a tipos nativos usando helpers
	ticket.TransactionID = ToInt64FromPgInt8(transactionID)
	if userID.Valid {
		uid := int64(userID.Int32)
		ticket.UserID = &uid
	} else {
		ticket.UserID = nil
	}
	ticket.SeatNumber = ToStringFromPgText(seatNumber)
	ticket.QRCodeURL = ToStringFromPgText(qrCodeURL)
	ticket.UsedAt = ToTimeFromPgTimestamp(usedAt)
	ticket.TransferredFromTicketID = ToInt64FromPgInt8(transferredFromTicketID)

	return &ticket, nil
}

// GetTicketsByTransaction obtiene tickets por transaction_id
func (r *TicketRepository) GetTicketsByTransaction(ctx context.Context, transactionID int64) ([]*models.Ticket, error) {
	query := `
		SELECT id, public_id, category_id, transaction_id, event_id, 
		       customer_id, user_id, code, status, seat_number, 
		       qr_code_url, price, used_at, transferred_from_ticket_id, 
		       created_at, updated_at
		FROM tickets 
		WHERE transaction_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, transactionID)
	if err != nil {
		return nil, fmt.Errorf("error querying tickets by transaction: %w", err)
	}
	defer rows.Close()

	var tickets []*models.Ticket
	for rows.Next() {
		var ticket models.Ticket
		var txID pgtype.Int8
		var userID pgtype.Int4 // ✅ CAMBIADO a Int4
		var seatNumber pgtype.Text
		var qrCodeURL pgtype.Text
		var usedAt pgtype.Timestamp
		var transferredFromTicketID pgtype.Int8

		err := rows.Scan(
			&ticket.ID,
			&ticket.PublicID,
			&ticket.CategoryID,
			&txID,
			&ticket.EventID,
			&ticket.CustomerID, // ✅ DIRECTAMENTE al campo del modelo
			&userID,
			&ticket.Code,
			&ticket.Status,
			&seatNumber,
			&qrCodeURL,
			&ticket.Price,
			&usedAt,
			&transferredFromTicketID,
			&ticket.CreatedAt,
			&ticket.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning ticket: %w", err)
		}

		// Convertir pgtype a tipos nativos usando helpers
		ticket.TransactionID = ToInt64FromPgInt8(txID)
		if userID.Valid {
			uid := int64(userID.Int32)
			ticket.UserID = &uid
		} else {
			ticket.UserID = nil
		}
		ticket.SeatNumber = ToStringFromPgText(seatNumber)
		ticket.QRCodeURL = ToStringFromPgText(qrCodeURL)
		ticket.UsedAt = ToTimeFromPgTimestamp(usedAt)
		ticket.TransferredFromTicketID = ToInt64FromPgInt8(transferredFromTicketID)

		tickets = append(tickets, &ticket)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tickets: %w", err)
	}

	log.Printf("Found %d tickets for transaction: %d", len(tickets), transactionID)
	return tickets, nil
}
