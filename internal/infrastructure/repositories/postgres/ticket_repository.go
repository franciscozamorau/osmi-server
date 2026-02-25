// internal/infrastructure/repositories/postgres/ticket_repository.go
package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
	"github.com/franciscozamorau/osmi-server/internal/domain/enums"
	"github.com/franciscozamorau/osmi-server/internal/domain/repository"
)

// TicketRepository implementa la interfaz repository.TicketRepository usando PostgreSQL
type TicketRepository struct {
	db *sqlx.DB
}

// NewTicketRepository crea una nueva instancia del repositorio
func NewTicketRepository(db *sqlx.DB) *TicketRepository {
	return &TicketRepository{
		db: db,
	}
}

// handleError mapea errores de PostgreSQL a nuestros errores de dominio
func (r *TicketRepository) handleError(err error, context string) error {
	if err == nil {
		return nil
	}

	// Errores específicos de PostgreSQL
	if pqErr, ok := err.(*pq.Error); ok {
		switch pqErr.Code {
		case "23505": // Unique violation
			if strings.Contains(pqErr.Constraint, "tickets_code_key") {
				return repository.ErrTicketDuplicateCode
			}
			if strings.Contains(pqErr.Constraint, "tickets_public_uuid_key") {
				return repository.ErrTicketAlreadyExists
			}
		case "23503": // Foreign key violation
			return fmt.Errorf("referenced record not found: %w", err)
		}
	}

	// Wrap el error para dar contexto
	if errors.Is(err, sql.ErrNoRows) {
		return repository.ErrTicketNotFound
	}

	return fmt.Errorf("%s: %w", context, err)
}

// Find busca tickets según los criterios del filtro
func (r *TicketRepository) Find(ctx context.Context, filter *repository.TicketFilter) ([]*entities.Ticket, int64, error) {
	// Query base
	baseQuery := `
		SELECT 
			id, public_uuid, ticket_type_id, event_id, customer_id, order_id,
			code, secret_hash, qr_code_data, status, final_price, currency, tax_amount,
			attendee_name, attendee_email, attendee_phone,
			checked_in_at, checked_in_by, checkin_method, checkin_location,
			reserved_at, reserved_by, reservation_expires_at,
			transfer_token, transferred_from, transferred_at,
			validation_count, last_validated_at,
			sold_at, cancelled_at, refunded_at,
			created_at, updated_at
		FROM ticketing.tickets
		WHERE 1=1
	`

	countQuery := `SELECT COUNT(*) FROM ticketing.tickets WHERE 1=1`

	var conditions []string
	var args []interface{}
	argPos := 1

	// Aplicar filtros
	if filter != nil {
		// Filtro por IDs
		if len(filter.IDs) > 0 {
			conditions = append(conditions, fmt.Sprintf("id = ANY($%d)", argPos))
			args = append(args, pq.Array(filter.IDs))
			argPos++
		}

		// Filtro por PublicIDs
		if len(filter.PublicIDs) > 0 {
			conditions = append(conditions, fmt.Sprintf("public_uuid = ANY($%d)", argPos))
			args = append(args, pq.Array(filter.PublicIDs))
			argPos++
		}

		// Filtro por EventID
		if filter.EventID != nil {
			conditions = append(conditions, fmt.Sprintf("event_id = $%d", argPos))
			args = append(args, *filter.EventID)
			argPos++
		}

		// Filtro por TicketTypeID
		if filter.TicketTypeID != nil {
			conditions = append(conditions, fmt.Sprintf("ticket_type_id = $%d", argPos))
			args = append(args, *filter.TicketTypeID)
			argPos++
		}

		// Filtro por CustomerID
		if filter.CustomerID != nil {
			conditions = append(conditions, fmt.Sprintf("customer_id = $%d", argPos))
			args = append(args, *filter.CustomerID)
			argPos++
		}

		// Filtro por OrderID
		if filter.OrderID != nil {
			conditions = append(conditions, fmt.Sprintf("order_id = $%d", argPos))
			args = append(args, *filter.OrderID)
			argPos++
		}

		// Filtro por Code
		if filter.Code != nil {
			conditions = append(conditions, fmt.Sprintf("code = $%d", argPos))
			args = append(args, *filter.Code)
			argPos++
		}

		// Filtro por Status (múltiples estados)
		if len(filter.Status) > 0 {
			statusStrings := make([]string, len(filter.Status))
			for i, s := range filter.Status {
				statusStrings[i] = string(s)
			}
			conditions = append(conditions, fmt.Sprintf("status = ANY($%d)", argPos))
			args = append(args, pq.Array(statusStrings))
			argPos++
		}

		// Filtro por TransferToken
		if filter.TransferToken != nil {
			conditions = append(conditions, fmt.Sprintf("transfer_token = $%d", argPos))
			args = append(args, *filter.TransferToken)
			argPos++
		}

		// Filtros por fechas
		if filter.CreatedFrom != nil {
			conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argPos))
			args = append(args, *filter.CreatedFrom)
			argPos++
		}
		if filter.CreatedTo != nil {
			conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argPos))
			args = append(args, *filter.CreatedTo)
			argPos++
		}
		if filter.SoldFrom != nil {
			conditions = append(conditions, fmt.Sprintf("sold_at >= $%d", argPos))
			args = append(args, *filter.SoldFrom)
			argPos++
		}
		if filter.SoldTo != nil {
			conditions = append(conditions, fmt.Sprintf("sold_at <= $%d", argPos))
			args = append(args, *filter.SoldTo)
			argPos++
		}
		if filter.CheckedInFrom != nil {
			conditions = append(conditions, fmt.Sprintf("checked_in_at >= $%d", argPos))
			args = append(args, *filter.CheckedInFrom)
			argPos++
		}
		if filter.CheckedInTo != nil {
			conditions = append(conditions, fmt.Sprintf("checked_in_at <= $%d", argPos))
			args = append(args, *filter.CheckedInTo)
			argPos++
		}

		// Filtros booleanos
		if filter.HasCheckedIn != nil {
			if *filter.HasCheckedIn {
				conditions = append(conditions, "checked_in_at IS NOT NULL")
			} else {
				conditions = append(conditions, "checked_in_at IS NULL")
			}
		}
		if filter.HasReservation != nil {
			if *filter.HasReservation {
				conditions = append(conditions, "reserved_at IS NOT NULL")
			} else {
				conditions = append(conditions, "reserved_at IS NULL")
			}
		}
	}

	// Unir condiciones
	if len(conditions) > 0 {
		whereClause := " AND " + strings.Join(conditions, " AND ")
		baseQuery += whereClause
		countQuery += whereClause
	}

	// Obtener total
	var total int64
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, r.handleError(err, "failed to count tickets")
	}

	// Añadir ordenamiento y paginación
	if filter != nil {
		// Ordenamiento
		sortBy := "created_at"
		sortOrder := "DESC"
		if filter.SortBy != "" {
			allowedSortColumns := map[string]bool{
				"created_at":    true,
				"sold_at":       true,
				"checked_in_at": true,
				"final_price":   true,
				"status":        true,
			}
			if allowedSortColumns[filter.SortBy] {
				sortBy = filter.SortBy
			}
		}
		if filter.SortOrder != "" {
			if strings.ToUpper(filter.SortOrder) == "ASC" {
				sortOrder = "ASC"
			}
		}
		baseQuery += fmt.Sprintf(" ORDER BY %s %s", sortBy, sortOrder)

		// Paginación
		if filter.Limit > 0 {
			baseQuery += fmt.Sprintf(" LIMIT $%d", argPos)
			args = append(args, filter.Limit)
			argPos++
		}
		if filter.Offset > 0 {
			baseQuery += fmt.Sprintf(" OFFSET $%d", argPos)
			args = append(args, filter.Offset)
			argPos++
		}
	}

	// Ejecutar query
	var tickets []*entities.Ticket
	err = r.db.SelectContext(ctx, &tickets, baseQuery, args...)
	if err != nil {
		return nil, 0, r.handleError(err, "failed to find tickets")
	}

	return tickets, total, nil
}

// GetByID obtiene un ticket por su ID numérico
func (r *TicketRepository) GetByID(ctx context.Context, id int64) (*entities.Ticket, error) {
	filter := &repository.TicketFilter{
		IDs:   []int64{id},
		Limit: 1,
	}

	tickets, _, err := r.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	if len(tickets) == 0 {
		return nil, repository.ErrTicketNotFound
	}

	return tickets[0], nil
}

// GetByPublicID obtiene un ticket por su UUID público
func (r *TicketRepository) GetByPublicID(ctx context.Context, publicID string) (*entities.Ticket, error) {
	filter := &repository.TicketFilter{
		PublicIDs: []string{publicID},
		Limit:     1,
	}

	tickets, _, err := r.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	if len(tickets) == 0 {
		return nil, repository.ErrTicketNotFound
	}

	return tickets[0], nil
}

// GetByCode obtiene un ticket por su código único
func (r *TicketRepository) GetByCode(ctx context.Context, code string) (*entities.Ticket, error) {
	filter := &repository.TicketFilter{
		Code:  &code,
		Limit: 1,
	}

	tickets, _, err := r.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	if len(tickets) == 0 {
		return nil, repository.ErrTicketNotFound
	}

	return tickets[0], nil
}

// Create inserta un nuevo ticket
func (r *TicketRepository) Create(ctx context.Context, ticket *entities.Ticket) error {
	// Validar el ticket
	if err := ticket.Validate(); err != nil {
		return err
	}

	query := `
		INSERT INTO ticketing.tickets (
			public_uuid, ticket_type_id, event_id, customer_id, order_id,
			code, secret_hash, qr_code_data, status, final_price, currency, tax_amount,
			attendee_name, attendee_email, attendee_phone,
			checked_in_at, checked_in_by, checkin_method, checkin_location,
			reserved_at, reserved_by, reservation_expires_at,
			transfer_token, transferred_from, transferred_at,
			validation_count, last_validated_at,
			sold_at, cancelled_at, refunded_at,
			created_at, updated_at
		) VALUES (
			gen_random_uuid(), $1, $2, $3, $4,
			$5, $6, $7, $8, $9, $10, $11,
			$12, $13, $14, $15, $16, $17, $18,
			$19, $20, $21, $22, $23, $24,
			$25, $26, $27, $28, $29,
			NOW(), NOW()
		)
		RETURNING id, public_uuid, created_at, updated_at
	`

	err := r.db.QueryRowContext(
		ctx, query,
		ticket.TicketTypeID, ticket.EventID, ticket.CustomerID, ticket.OrderID,
		ticket.Code, ticket.SecretHash, ticket.QRCodeData, ticket.Status,
		ticket.FinalPrice, ticket.Currency, ticket.TaxAmount,
		ticket.AttendeeName, ticket.AttendeeEmail, ticket.AttendeePhone,
		ticket.CheckedInAt, ticket.CheckedInBy, ticket.CheckinMethod, ticket.CheckinLocation,
		ticket.ReservedAt, ticket.ReservedBy, ticket.ReservationExpiresAt,
		ticket.TransferToken, ticket.TransferredFrom, ticket.TransferredAt,
		ticket.ValidationCount, ticket.LastValidatedAt,
		ticket.SoldAt, ticket.CancelledAt, ticket.RefundedAt,
	).Scan(&ticket.ID, &ticket.PublicID, &ticket.CreatedAt, &ticket.UpdatedAt)

	if err != nil {
		return r.handleError(err, "failed to create ticket")
	}

	return nil
}

// CreateBatch crea múltiples tickets en una transacción
func (r *TicketRepository) CreateBatch(ctx context.Context, tickets []*entities.Ticket) error {
	if len(tickets) == 0 {
		return nil
	}

	// Iniciar transacción
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return r.handleError(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	query := `
		INSERT INTO ticketing.tickets (
			public_uuid, ticket_type_id, event_id, customer_id, order_id,
			code, secret_hash, qr_code_data, status, final_price, currency, tax_amount,
			attendee_name, attendee_email, attendee_phone,
			checked_in_at, checked_in_by, checkin_method, checkin_location,
			reserved_at, reserved_by, reservation_expires_at,
			transfer_token, transferred_from, transferred_at,
			validation_count, last_validated_at,
			sold_at, cancelled_at, refunded_at,
			created_at, updated_at
		) VALUES (
			gen_random_uuid(), $1, $2, $3, $4,
			$5, $6, $7, $8, $9, $10, $11,
			$12, $13, $14, $15, $16, $17, $18,
			$19, $20, $21, $22, $23, $24,
			$25, $26, $27, $28, $29,
			NOW(), NOW()
		)
	`

	for _, ticket := range tickets {
		if err := ticket.Validate(); err != nil {
			return err
		}

		_, err = tx.ExecContext(
			ctx, query,
			ticket.TicketTypeID, ticket.EventID, ticket.CustomerID, ticket.OrderID,
			ticket.Code, ticket.SecretHash, ticket.QRCodeData, ticket.Status,
			ticket.FinalPrice, ticket.Currency, ticket.TaxAmount,
			ticket.AttendeeName, ticket.AttendeeEmail, ticket.AttendeePhone,
			ticket.CheckedInAt, ticket.CheckedInBy, ticket.CheckinMethod, ticket.CheckinLocation,
			ticket.ReservedAt, ticket.ReservedBy, ticket.ReservationExpiresAt,
			ticket.TransferToken, ticket.TransferredFrom, ticket.TransferredAt,
			ticket.ValidationCount, ticket.LastValidatedAt,
			ticket.SoldAt, ticket.CancelledAt, ticket.RefundedAt,
		)
		if err != nil {
			return r.handleError(err, "failed to create ticket in batch")
		}
	}

	return tx.Commit()
}

// Update actualiza un ticket existente
func (r *TicketRepository) Update(ctx context.Context, ticket *entities.Ticket) error {
	// Verificar que existe
	exists, err := r.Exists(ctx, ticket.ID)
	if err != nil {
		return err
	}
	if !exists {
		return repository.ErrTicketNotFound
	}

	query := `
		UPDATE ticketing.tickets SET
			ticket_type_id = $1,
			event_id = $2,
			customer_id = $3,
			order_id = $4,
			qr_code_data = $5,
			status = $6,
			final_price = $7,
			currency = $8,
			tax_amount = $9,
			attendee_name = $10,
			attendee_email = $11,
			attendee_phone = $12,
			checked_in_at = $13,
			checked_in_by = $14,
			checkin_method = $15,
			checkin_location = $16,
			reserved_at = $17,
			reserved_by = $18,
			reservation_expires_at = $19,
			transfer_token = $20,
			transferred_from = $21,
			transferred_at = $22,
			validation_count = $23,
			last_validated_at = $24,
			sold_at = $25,
			cancelled_at = $26,
			refunded_at = $27,
			updated_at = NOW()
		WHERE id = $28
		RETURNING updated_at
	`

	err = r.db.QueryRowContext(
		ctx, query,
		ticket.TicketTypeID, ticket.EventID, ticket.CustomerID, ticket.OrderID,
		ticket.QRCodeData, ticket.Status, ticket.FinalPrice, ticket.Currency, ticket.TaxAmount,
		ticket.AttendeeName, ticket.AttendeeEmail, ticket.AttendeePhone,
		ticket.CheckedInAt, ticket.CheckedInBy, ticket.CheckinMethod, ticket.CheckinLocation,
		ticket.ReservedAt, ticket.ReservedBy, ticket.ReservationExpiresAt,
		ticket.TransferToken, ticket.TransferredFrom, ticket.TransferredAt,
		ticket.ValidationCount, ticket.LastValidatedAt,
		ticket.SoldAt, ticket.CancelledAt, ticket.RefundedAt,
		ticket.ID,
	).Scan(&ticket.UpdatedAt)

	if err != nil {
		return r.handleError(err, "failed to update ticket")
	}

	return nil
}

// Delete elimina un ticket (usar con precaución, mejor usar Cancel)
func (r *TicketRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM ticketing.tickets WHERE id = $1`, id)
	if err != nil {
		return r.handleError(err, "failed to delete ticket")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return repository.ErrTicketNotFound
	}

	return nil
}

// Exists verifica si existe un ticket con el ID dado
func (r *TicketRepository) Exists(ctx context.Context, id int64) (bool, error) {
	var exists bool
	err := r.db.GetContext(ctx, &exists, `SELECT EXISTS(SELECT 1 FROM ticketing.tickets WHERE id = $1)`, id)
	if err != nil {
		return false, r.handleError(err, "failed to check ticket existence")
	}
	return exists, nil
}

// ExistsByCode verifica si existe un ticket con el código dado
func (r *TicketRepository) ExistsByCode(ctx context.Context, code string) (bool, error) {
	var exists bool
	err := r.db.GetContext(ctx, &exists, `SELECT EXISTS(SELECT 1 FROM ticketing.tickets WHERE code = $1)`, code)
	if err != nil {
		return false, r.handleError(err, "failed to check ticket code existence")
	}
	return exists, nil
}

// UpdateStatus actualiza el estado de un ticket
func (r *TicketRepository) UpdateStatus(ctx context.Context, ticketID int64, status enums.TicketStatus) error {
	// Verificar transición válida
	var currentStatus string
	err := r.db.GetContext(ctx, &currentStatus, `SELECT status FROM ticketing.tickets WHERE id = $1`, ticketID)
	if err != nil {
		return r.handleError(err, "failed to get current status")
	}

	// CORREGIDO: Usar CanTransitionTicket en lugar de CanTransition
	if !enums.CanTransitionTicket(enums.TicketStatus(currentStatus), status) {
		return repository.ErrInvalidTicketStatus
	}

	query := `
		UPDATE ticketing.tickets 
		SET status = $1, updated_at = NOW() 
		WHERE id = $2
	`
	result, err := r.db.ExecContext(ctx, query, string(status), ticketID)
	if err != nil {
		return r.handleError(err, "failed to update ticket status")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return repository.ErrTicketNotFound
	}

	return nil
}

// CheckIn marca un ticket como usado (check-in)
func (r *TicketRepository) CheckIn(ctx context.Context, ticketID int64, method, location string, checkedBy *int64) error {
	now := time.Now()
	query := `
		UPDATE ticketing.tickets 
		SET status = 'checked_in', 
			checked_in_at = $1, 
			checked_in_by = $2, 
			checkin_method = $3, 
			checkin_location = $4,
			validation_count = validation_count + 1,
			last_validated_at = $1,
			updated_at = $1
		WHERE id = $5 AND status = 'sold'
	`
	result, err := r.db.ExecContext(ctx, query, now, checkedBy, method, location, ticketID)
	if err != nil {
		return r.handleError(err, "failed to check in ticket")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return repository.ErrTicketNotAvailable
	}

	return nil
}

// Reserve reserva un ticket
func (r *TicketRepository) Reserve(ctx context.Context, ticketID int64, reservedBy int64, expiresAt time.Time) error {
	now := time.Now()
	query := `
		UPDATE ticketing.tickets 
		SET status = 'reserved', 
			reserved_at = $1, 
			reserved_by = $2, 
			reservation_expires_at = $3,
			updated_at = $1
		WHERE id = $4 AND status = 'available'
	`
	result, err := r.db.ExecContext(ctx, query, now, reservedBy, expiresAt, ticketID)
	if err != nil {
		return r.handleError(err, "failed to reserve ticket")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return repository.ErrTicketNotAvailable
	}

	return nil
}

// ReleaseReservation libera una reserva
func (r *TicketRepository) ReleaseReservation(ctx context.Context, ticketID int64) error {
	query := `
		UPDATE ticketing.tickets 
		SET status = 'available', 
			reserved_at = NULL, 
			reserved_by = NULL, 
			reservation_expires_at = NULL,
			updated_at = NOW()
		WHERE id = $1 AND status = 'reserved'
	`
	result, err := r.db.ExecContext(ctx, query, ticketID)
	if err != nil {
		return r.handleError(err, "failed to release reservation")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return repository.ErrTicketNotAvailable
	}

	return nil
}

// Transfer transfiere un ticket a otro cliente
func (r *TicketRepository) Transfer(ctx context.Context, ticketID int64, toCustomerID int64, transferToken string) error {
	// Obtener el customer_id actual
	var fromCustomerID int64
	err := r.db.GetContext(ctx, &fromCustomerID, `SELECT customer_id FROM ticketing.tickets WHERE id = $1`, ticketID)
	if err != nil {
		return r.handleError(err, "failed to get current customer")
	}

	query := `
		UPDATE ticketing.tickets 
		SET customer_id = $1, 
			transferred_from = $2, 
			transferred_at = NOW(),
			transfer_token = $3,
			status = 'sold',
			updated_at = NOW()
		WHERE id = $4 AND status = 'sold'
	`
	result, err := r.db.ExecContext(ctx, query, toCustomerID, fromCustomerID, transferToken, ticketID)
	if err != nil {
		return r.handleError(err, "failed to transfer ticket")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return repository.ErrTicketNotAvailable
	}

	return nil
}

// Cancel cancela un ticket
func (r *TicketRepository) Cancel(ctx context.Context, ticketID int64) error {
	now := time.Now()
	query := `
		UPDATE ticketing.tickets 
		SET status = 'cancelled', 
			cancelled_at = $1,
			updated_at = $1
		WHERE id = $2 AND status IN ('available', 'reserved', 'sold')
	`
	result, err := r.db.ExecContext(ctx, query, now, ticketID)
	if err != nil {
		return r.handleError(err, "failed to cancel ticket")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return repository.ErrTicketNotAvailable
	}

	return nil
}

// Refund reembolsa un ticket
func (r *TicketRepository) Refund(ctx context.Context, ticketID int64) error {
	now := time.Now()
	query := `
		UPDATE ticketing.tickets 
		SET status = 'refunded', 
			refunded_at = $1,
			updated_at = $1
		WHERE id = $2 AND status = 'sold'
	`
	result, err := r.db.ExecContext(ctx, query, now, ticketID)
	if err != nil {
		return r.handleError(err, "failed to refund ticket")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return repository.ErrTicketNotAvailable
	}

	return nil
}

// ValidateTicket valida un ticket por código y hash secreto
func (r *TicketRepository) ValidateTicket(ctx context.Context, code, secretHash string) (*entities.Ticket, error) {
	query := `
		SELECT 
			id, public_uuid, ticket_type_id, event_id, customer_id, order_id,
			code, secret_hash, qr_code_data, status, final_price, currency, tax_amount,
			attendee_name, attendee_email, attendee_phone,
			checked_in_at, checked_in_by, checkin_method, checkin_location,
			reserved_at, reserved_by, reservation_expires_at,
			transfer_token, transferred_from, transferred_at,
			validation_count, last_validated_at,
			sold_at, cancelled_at, refunded_at,
			created_at, updated_at
		FROM ticketing.tickets
		WHERE code = $1 AND secret_hash = $2
	`

	var ticket entities.Ticket
	err := r.db.GetContext(ctx, &ticket, query, code, secretHash)
	if err != nil {
		return nil, r.handleError(err, "failed to validate ticket")
	}

	return &ticket, nil
}

// GetEventStats obtiene estadísticas de tickets para un evento
func (r *TicketRepository) GetEventStats(ctx context.Context, eventID int64) (*repository.TicketStats, error) {
	query := `
		SELECT 
			COUNT(*) as total_tickets,
			COUNT(CASE WHEN status = 'available' THEN 1 END) as available_tickets,
			COUNT(CASE WHEN status = 'reserved' THEN 1 END) as reserved_tickets,
			COUNT(CASE WHEN status = 'sold' THEN 1 END) as sold_tickets,
			COUNT(CASE WHEN status = 'checked_in' THEN 1 END) as checked_in_tickets,
			COUNT(CASE WHEN status = 'cancelled' THEN 1 END) as cancelled_tickets,
			COUNT(CASE WHEN status = 'refunded' THEN 1 END) as refunded_tickets,
			COALESCE(SUM(CASE WHEN status IN ('sold', 'checked_in') THEN final_price ELSE 0 END), 0) as total_revenue,
			COALESCE(AVG(CASE WHEN status IN ('sold', 'checked_in') THEN final_price END), 0) as avg_ticket_price
		FROM ticketing.tickets
		WHERE event_id = $1
	`

	var stats repository.TicketStats
	err := r.db.GetContext(ctx, &stats, query, eventID)
	if err != nil {
		return nil, r.handleError(err, "failed to get event stats")
	}

	return &stats, nil
}

// GetReservedExpired obtiene tickets con reservas expiradas
func (r *TicketRepository) GetReservedExpired(ctx context.Context) ([]*entities.Ticket, error) {
	query := `
		SELECT 
			id, public_uuid, ticket_type_id, event_id, customer_id, order_id,
			code, secret_hash, qr_code_data, status, final_price, currency, tax_amount,
			attendee_name, attendee_email, attendee_phone,
			checked_in_at, checked_in_by, checkin_method, checkin_location,
			reserved_at, reserved_by, reservation_expires_at,
			transfer_token, transferred_from, transferred_at,
			validation_count, last_validated_at,
			sold_at, cancelled_at, refunded_at,
			created_at, updated_at
		FROM ticketing.tickets
		WHERE status = 'reserved' AND reservation_expires_at < NOW()
	`

	var tickets []*entities.Ticket
	err := r.db.SelectContext(ctx, &tickets, query)
	if err != nil {
		return nil, r.handleError(err, "failed to get expired reservations")
	}

	return tickets, nil
}
