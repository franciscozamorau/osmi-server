// internal/infrastructure/repositories/postgres/ticket_repository.go
package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
	"github.com/franciscozamorau/osmi-server/internal/domain/enums"
	"github.com/franciscozamorau/osmi-server/internal/domain/repository"
)

// TicketRepository implementa la interfaz repository.TicketRepository usando PostgreSQL
type TicketRepository struct {
	db *pgxpool.Pool
}

// NewTicketRepository crea una nueva instancia del repositorio
func NewTicketRepository(db *pgxpool.Pool) *TicketRepository {
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
	if errors.Is(err, pgx.ErrNoRows) {
		return repository.ErrTicketNotFound
	}

	// Verificar si es un error de PostgreSQL con código
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // Unique violation
			if strings.Contains(pgErr.ConstraintName, "tickets_code_key") {
				return repository.ErrTicketDuplicateCode
			}
			if strings.Contains(pgErr.ConstraintName, "tickets_public_uuid_key") {
				return repository.ErrTicketAlreadyExists
			}
		case "23503": // Foreign key violation
			return fmt.Errorf("referenced record not found: %w", err)
		}
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
	args := pgx.NamedArgs{}
	argPos := 1

	// Aplicar filtros
	if filter != nil {
		// Filtro por IDs
		if len(filter.IDs) > 0 {
			conditions = append(conditions, fmt.Sprintf("id = ANY(@id_%d)", argPos))
			args[fmt.Sprintf("id_%d", argPos)] = filter.IDs
			argPos++
		}

		// Filtro por PublicIDs
		if len(filter.PublicIDs) > 0 {
			conditions = append(conditions, fmt.Sprintf("public_uuid = ANY(@public_%d)", argPos))
			args[fmt.Sprintf("public_%d", argPos)] = filter.PublicIDs
			argPos++
		}

		// Filtro por EventID
		if filter.EventID != nil {
			conditions = append(conditions, fmt.Sprintf("event_id = @event_%d", argPos))
			args[fmt.Sprintf("event_%d", argPos)] = *filter.EventID
			argPos++
		}

		// Filtro por TicketTypeID
		if filter.TicketTypeID != nil {
			conditions = append(conditions, fmt.Sprintf("ticket_type_id = @type_%d", argPos))
			args[fmt.Sprintf("type_%d", argPos)] = *filter.TicketTypeID
			argPos++
		}

		// Filtro por CustomerID
		if filter.CustomerID != nil {
			conditions = append(conditions, fmt.Sprintf("customer_id = @customer_%d", argPos))
			args[fmt.Sprintf("customer_%d", argPos)] = *filter.CustomerID
			argPos++
		}

		// Filtro por OrderID
		if filter.OrderID != nil {
			conditions = append(conditions, fmt.Sprintf("order_id = @order_%d", argPos))
			args[fmt.Sprintf("order_%d", argPos)] = *filter.OrderID
			argPos++
		}

		// Filtro por Code
		if filter.Code != nil {
			conditions = append(conditions, fmt.Sprintf("code = @code_%d", argPos))
			args[fmt.Sprintf("code_%d", argPos)] = *filter.Code
			argPos++
		}

		// Filtro por Status (múltiples estados)
		if len(filter.Status) > 0 {
			statusStrings := make([]string, len(filter.Status))
			for i, s := range filter.Status {
				statusStrings[i] = string(s)
			}
			conditions = append(conditions, fmt.Sprintf("status = ANY(@status_%d)", argPos))
			args[fmt.Sprintf("status_%d", argPos)] = statusStrings
			argPos++
		}

		// Filtro por TransferToken
		if filter.TransferToken != nil {
			conditions = append(conditions, fmt.Sprintf("transfer_token = @token_%d", argPos))
			args[fmt.Sprintf("token_%d", argPos)] = *filter.TransferToken
			argPos++
		}

		// Filtros por fechas
		if filter.CreatedFrom != nil {
			conditions = append(conditions, fmt.Sprintf("created_at >= @created_from_%d", argPos))
			args[fmt.Sprintf("created_from_%d", argPos)] = *filter.CreatedFrom
			argPos++
		}
		if filter.CreatedTo != nil {
			conditions = append(conditions, fmt.Sprintf("created_at <= @created_to_%d", argPos))
			args[fmt.Sprintf("created_to_%d", argPos)] = *filter.CreatedTo
			argPos++
		}
		if filter.SoldFrom != nil {
			conditions = append(conditions, fmt.Sprintf("sold_at >= @sold_from_%d", argPos))
			args[fmt.Sprintf("sold_from_%d", argPos)] = *filter.SoldFrom
			argPos++
		}
		if filter.SoldTo != nil {
			conditions = append(conditions, fmt.Sprintf("sold_at <= @sold_to_%d", argPos))
			args[fmt.Sprintf("sold_to_%d", argPos)] = *filter.SoldTo
			argPos++
		}
		if filter.CheckedInFrom != nil {
			conditions = append(conditions, fmt.Sprintf("checked_in_at >= @checked_from_%d", argPos))
			args[fmt.Sprintf("checked_from_%d", argPos)] = *filter.CheckedInFrom
			argPos++
		}
		if filter.CheckedInTo != nil {
			conditions = append(conditions, fmt.Sprintf("checked_in_at <= @checked_to_%d", argPos))
			args[fmt.Sprintf("checked_to_%d", argPos)] = *filter.CheckedInTo
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
	err := r.db.QueryRow(ctx, countQuery, args).Scan(&total)
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
			baseQuery += fmt.Sprintf(" LIMIT @limit")
			args["limit"] = filter.Limit
		}
		if filter.Offset > 0 {
			baseQuery += fmt.Sprintf(" OFFSET @offset")
			args["offset"] = filter.Offset
		}
	} else {
		baseQuery += " ORDER BY created_at DESC LIMIT 20"
	}

	// Ejecutar query
	rows, err := r.db.Query(ctx, baseQuery, args)
	if err != nil {
		return nil, 0, r.handleError(err, "failed to find tickets")
	}
	defer rows.Close()

	var tickets []*entities.Ticket
	for rows.Next() {
		var ticket entities.Ticket
		var attendeeName, attendeeEmail, attendeePhone, qrCodeData *string
		var checkedInBy, reservedBy *int64
		var checkinMethod, checkinLocation *string
		var checkedInAt, reservedAt, reservationExpiresAt, soldAt, cancelledAt, refundedAt, lastValidatedAt *time.Time
		var transferredFrom *int64
		var transferToken *string

		err = rows.Scan(
			&ticket.ID, &ticket.PublicID, &ticket.TicketTypeID, &ticket.EventID, &ticket.CustomerID, &ticket.OrderID,
			&ticket.Code, &ticket.SecretHash, &qrCodeData, &ticket.Status, &ticket.FinalPrice, &ticket.Currency, &ticket.TaxAmount,
			&attendeeName, &attendeeEmail, &attendeePhone,
			&checkedInAt, &checkedInBy, &checkinMethod, &checkinLocation,
			&reservedAt, &reservedBy, &reservationExpiresAt,
			&transferToken, &transferredFrom, &ticket.TransferredAt,
			&ticket.ValidationCount, &lastValidatedAt,
			&soldAt, &cancelledAt, &refundedAt,
			&ticket.CreatedAt, &ticket.UpdatedAt,
		)
		if err != nil {
			return nil, 0, r.handleError(err, "failed to scan ticket row")
		}

		// Asignar campos NULL
		ticket.AttendeeName = attendeeName
		ticket.AttendeeEmail = attendeeEmail
		ticket.AttendeePhone = attendeePhone
		ticket.QRCodeData = qrCodeData
		ticket.CheckedInAt = checkedInAt
		ticket.CheckedInBy = checkedInBy
		ticket.CheckinMethod = checkinMethod
		ticket.CheckinLocation = checkinLocation
		ticket.ReservedAt = reservedAt
		ticket.ReservedBy = reservedBy
		ticket.ReservationExpiresAt = reservationExpiresAt
		ticket.TransferToken = transferToken
		ticket.TransferredFrom = transferredFrom
		ticket.LastValidatedAt = lastValidatedAt
		ticket.SoldAt = soldAt
		ticket.CancelledAt = cancelledAt
		ticket.RefundedAt = refundedAt

		tickets = append(tickets, &ticket)
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

	err := r.db.QueryRow(ctx, query,
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
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return r.handleError(err, "failed to begin transaction")
	}
	defer tx.Rollback(ctx)

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

		_, err = tx.Exec(ctx, query,
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

	return tx.Commit(ctx)
}

// Update actualiza un ticket existente
func (r *TicketRepository) Update(ctx context.Context, ticket *entities.Ticket) error {
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

	err := r.db.QueryRow(ctx, query,
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
	cmdTag, err := r.db.Exec(ctx, `DELETE FROM ticketing.tickets WHERE id = $1`, id)
	if err != nil {
		return r.handleError(err, "failed to delete ticket")
	}

	if cmdTag.RowsAffected() == 0 {
		return repository.ErrTicketNotFound
	}

	return nil
}

// Exists verifica si existe un ticket con el ID dado
func (r *TicketRepository) Exists(ctx context.Context, id int64) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM ticketing.tickets WHERE id = $1)`
	err := r.db.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, r.handleError(err, "failed to check ticket existence")
	}
	return exists, nil
}

// ExistsByCode verifica si existe un ticket con el código dado
func (r *TicketRepository) ExistsByCode(ctx context.Context, code string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM ticketing.tickets WHERE code = $1)`
	err := r.db.QueryRow(ctx, query, code).Scan(&exists)
	if err != nil {
		return false, r.handleError(err, "failed to check ticket code existence")
	}
	return exists, nil
}

// UpdateStatus actualiza el estado de un ticket
func (r *TicketRepository) UpdateStatus(ctx context.Context, ticketID int64, status enums.TicketStatus) error {
	// Verificar transición válida
	var currentStatus string
	err := r.db.QueryRow(ctx, `SELECT status FROM ticketing.tickets WHERE id = $1`, ticketID).Scan(&currentStatus)
	if err != nil {
		return r.handleError(err, "failed to get current status")
	}

	if !enums.CanTransitionTicket(enums.TicketStatus(currentStatus), status) {
		return repository.ErrInvalidTicketStatus
	}

	query := `
		UPDATE ticketing.tickets 
		SET status = $1, updated_at = NOW() 
		WHERE id = $2
	`
	cmdTag, err := r.db.Exec(ctx, query, string(status), ticketID)
	if err != nil {
		return r.handleError(err, "failed to update ticket status")
	}

	if cmdTag.RowsAffected() == 0 {
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
	cmdTag, err := r.db.Exec(ctx, query, now, checkedBy, method, location, ticketID)
	if err != nil {
		return r.handleError(err, "failed to check in ticket")
	}

	if cmdTag.RowsAffected() == 0 {
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
	cmdTag, err := r.db.Exec(ctx, query, now, reservedBy, expiresAt, ticketID)
	if err != nil {
		return r.handleError(err, "failed to reserve ticket")
	}

	if cmdTag.RowsAffected() == 0 {
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
	cmdTag, err := r.db.Exec(ctx, query, ticketID)
	if err != nil {
		return r.handleError(err, "failed to release reservation")
	}

	if cmdTag.RowsAffected() == 0 {
		return repository.ErrTicketNotAvailable
	}

	return nil
}

// Transfer transfiere un ticket a otro cliente
func (r *TicketRepository) Transfer(ctx context.Context, ticketID int64, toCustomerID int64, transferToken string) error {
	// Obtener el customer_id actual
	var fromCustomerID int64
	err := r.db.QueryRow(ctx, `SELECT customer_id FROM ticketing.tickets WHERE id = $1`, ticketID).Scan(&fromCustomerID)
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
	cmdTag, err := r.db.Exec(ctx, query, toCustomerID, fromCustomerID, transferToken, ticketID)
	if err != nil {
		return r.handleError(err, "failed to transfer ticket")
	}

	if cmdTag.RowsAffected() == 0 {
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
	cmdTag, err := r.db.Exec(ctx, query, now, ticketID)
	if err != nil {
		return r.handleError(err, "failed to cancel ticket")
	}

	if cmdTag.RowsAffected() == 0 {
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
	cmdTag, err := r.db.Exec(ctx, query, now, ticketID)
	if err != nil {
		return r.handleError(err, "failed to refund ticket")
	}

	if cmdTag.RowsAffected() == 0 {
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
	err := r.db.QueryRow(ctx, query, code, secretHash).Scan(
		&ticket.ID, &ticket.PublicID, &ticket.TicketTypeID, &ticket.EventID, &ticket.CustomerID, &ticket.OrderID,
		&ticket.Code, &ticket.SecretHash, &ticket.QRCodeData, &ticket.Status, &ticket.FinalPrice, &ticket.Currency, &ticket.TaxAmount,
		&ticket.AttendeeName, &ticket.AttendeeEmail, &ticket.AttendeePhone,
		&ticket.CheckedInAt, &ticket.CheckedInBy, &ticket.CheckinMethod, &ticket.CheckinLocation,
		&ticket.ReservedAt, &ticket.ReservedBy, &ticket.ReservationExpiresAt,
		&ticket.TransferToken, &ticket.TransferredFrom, &ticket.TransferredAt,
		&ticket.ValidationCount, &ticket.LastValidatedAt,
		&ticket.SoldAt, &ticket.CancelledAt, &ticket.RefundedAt,
		&ticket.CreatedAt, &ticket.UpdatedAt,
	)

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
	err := r.db.QueryRow(ctx, query, eventID).Scan(
		&stats.TotalTickets,
		&stats.AvailableTickets,
		&stats.ReservedTickets,
		&stats.SoldTickets,
		&stats.CheckedInTickets,
		&stats.CancelledTickets,
		&stats.RefundedTickets,
		&stats.TotalRevenue,
		&stats.AvgTicketPrice,
	)
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

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, r.handleError(err, "failed to get expired reservations")
	}
	defer rows.Close()

	var tickets []*entities.Ticket
	for rows.Next() {
		var ticket entities.Ticket
		err = rows.Scan(
			&ticket.ID, &ticket.PublicID, &ticket.TicketTypeID, &ticket.EventID, &ticket.CustomerID, &ticket.OrderID,
			&ticket.Code, &ticket.SecretHash, &ticket.QRCodeData, &ticket.Status, &ticket.FinalPrice, &ticket.Currency, &ticket.TaxAmount,
			&ticket.AttendeeName, &ticket.AttendeeEmail, &ticket.AttendeePhone,
			&ticket.CheckedInAt, &ticket.CheckedInBy, &ticket.CheckinMethod, &ticket.CheckinLocation,
			&ticket.ReservedAt, &ticket.ReservedBy, &ticket.ReservationExpiresAt,
			&ticket.TransferToken, &ticket.TransferredFrom, &ticket.TransferredAt,
			&ticket.ValidationCount, &ticket.LastValidatedAt,
			&ticket.SoldAt, &ticket.CancelledAt, &ticket.RefundedAt,
			&ticket.CreatedAt, &ticket.UpdatedAt,
		)
		if err != nil {
			return nil, r.handleError(err, "failed to scan expired reservation")
		}
		tickets = append(tickets, &ticket)
	}

	return tickets, nil
}
