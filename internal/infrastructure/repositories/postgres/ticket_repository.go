package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/franciscozamorau/osmi-server/internal/repositories/postgres/helpers/errors"
	"github.com/franciscozamorau/osmi-server/internal/repositories/postgres/helpers/query"
	"github.com/franciscozamorau/osmi-server/internal/repositories/postgres/helpers/scanner"
	"github.com/franciscozamorau/osmi-server/internal/repositories/postgres/helpers/types"
	"github.com/franciscozamorau/osmi-server/internal/repositories/postgres/helpers/utils"
	"github.com/franciscozamorau/osmi-server/internal/repositories/postgres/helpers/validations"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TicketRepository struct {
	db          *pgxpool.Pool
	converter   *types.Converter
	ticketConv  *types.TicketConverter
	scanner     *scanner.RowScanner
	ticketScan  *scanner.TicketScanner
	errHandler  *errors.PostgresErrorHandler
	validator   *errors.Validator
	logger      *utils.Logger
	queryBuilder *query.QueryBuilder
}

func NewTicketRepository(db *pgxpool.Pool) *TicketRepository {
	conv := types.NewConverter()
	
	return &TicketRepository{
		db:          db,
		converter:   conv,
		ticketConv:  types.NewTicketConverter(),
		scanner:     scanner.NewRowScanner(),
		ticketScan:  scanner.NewTicketScanner(),
		errHandler:  errors.NewPostgresErrorHandler(),
		validator:   errors.NewValidator(),
		logger:      utils.NewLogger("ticket-repository"),
		queryBuilder: query.NewQueryBuilder(""),
	}
}

// Valid ticket statuses usando validaciones del paquete validations
var validTicketStatuses = map[string]bool{
	"available":   true,
	"reserved":    true,
	"sold":        true,
	"used":        true,
	"cancelled":   true,
	"transferred": true,
	"refunded":    true,
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
		"refunded":    true,
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

// CreateTicket crea un nuevo ticket con todos los helpers
func (r *TicketRepository) CreateTicket(ctx context.Context, req *pb.TicketRequest) (string, error) {
	startTime := time.Now()
	
	// Validar y limpiar datos usando utils
	eventID := strings.TrimSpace(req.EventId)
	categoryID := strings.TrimSpace(req.CategoryId)
	customerID := strings.TrimSpace(req.CustomerId) // OBLIGATORIO
	userID := strings.TrimSpace(req.UserId)         // OPCIONAL

	// Validaciones básicas usando validator
	r.validator.Required("event_id", eventID).
		Required("category_id", categoryID).
		Required("customer_id", customerID)
	
	if validationErr := r.validator.Validate(); validationErr != nil {
		return "", validationErr
	}

	// Validar UUIDs usando validations
	if !validations.IsValidUUID(eventID) {
		return "", fmt.Errorf("invalid event ID format: must be a valid UUID")
	}
	if !validations.IsValidUUID(categoryID) {
		return "", fmt.Errorf("invalid category ID format: must be a valid UUID")
	}
	if !validations.IsValidUUID(customerID) {
		return "", fmt.Errorf("invalid customer ID format: must be a valid UUID")
	}
	if userID != "" && !validations.IsValidUUID(userID) {
		return "", fmt.Errorf("invalid user ID format: must be a valid UUID")
	}

	// Validar existencia del event_id y category_id
	if err := r.validateEventAndCategory(ctx, eventID, categoryID); err != nil {
		return "", err
	}

	// Buscar customer_id interno (OBLIGATORIO)
	customerInternalID, err := r.getCustomerIDByPublicID(ctx, customerID)
	if err != nil {
		r.logger.Error("Failed to find customer", err, map[string]interface{}{
			"customer_id": customerID,
		})
		return "", fmt.Errorf("error finding customer: %w", err)
	}

	// pgtype.Int4 porque user_id es integer en BD
	var userInternalID types.Converter.Int4Result
	if userID != "" {
		uid, err := r.getUserIDByPublicID(ctx, userID)
		if err != nil {
			r.logger.Error("Failed to find user", err, map[string]interface{}{
				"user_id": userID,
			})
			return "", fmt.Errorf("error finding user: %w", err)
		}
		userInternalID = r.converter.Int64Ptr(&uid)
	} else {
		userInternalID = r.converter.Int64Ptr(nil)
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

	// Usar transacción con transaction manager
	tm := errors.NewTransactionManager(r.errHandler)
	
	err = tm.ExecuteInTransaction(ctx, r.db, func(tx interface{}) error {
		for i := 0; i < quantity; i++ {
			publicID := uuid.New().String()
			code, err := r.generateUniqueTicketCode(ctx, tx.(pgx.Tx), eventID, customerID, i)
			if err != nil {
				return fmt.Errorf("error generating ticket code: %w", err)
			}

			// Insertar con customer_id (obligatorio) y user_id (opcional)
			query := `INSERT INTO tickets (
				public_id, category_id, event_id, customer_id, user_id,
				code, status, price, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, 'available', $7, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
			RETURNING public_id`

			var ticketPublicID string
			err = tx.(pgx.Tx).QueryRow(ctx, query,
				publicID,
				categoryInternalID,
				eventInternalID,
				customerInternalID, // $4 - OBLIGATORIO
				userInternalID,     // $5 - OPCIONAL (puede ser NULL)
				code,               // $6
				categoryPrice,      // $7
			).Scan(&ticketPublicID)

			if err != nil {
				if r.errHandler.IsDuplicateKey(err) {
					return fmt.Errorf("ticket with code %s already exists", utils.SafeStringForLog(code))
				}
				return fmt.Errorf("error creating ticket %d/%d: %w", i+1, quantity, err)
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
		_, err := tx.(pgx.Tx).Exec(ctx, updateQuery, quantity, categoryInternalID)
		if err != nil {
			return fmt.Errorf("error updating category sold count: %w", err)
		}

		return nil
	})

	if err != nil {
		r.logger.DatabaseLogger("INSERT", "tickets", time.Since(startTime), 0, err, map[string]interface{}{
			"quantity": quantity,
			"event_id": utils.SafeStringForLog(eventID),
			"category_id": utils.SafeStringForLog(categoryID),
		})
		return "", err
	}

	r.logger.DatabaseLogger("INSERT", "tickets", time.Since(startTime), int64(quantity), nil, map[string]interface{}{
		"quantity": quantity,
		"event_id": utils.SafeStringForLog(eventID),
		"category_id": utils.SafeStringForLog(categoryID),
		"customer_id": customerInternalID,
	})
	
	return createdTicketPublicID, nil
}

// GetTicketsByUserID obtiene todos los tickets de un usuario usando scanner
func (r *TicketRepository) GetTicketsByUserID(ctx context.Context, userPublicID string) ([]*models.Ticket, error) {
	startTime := time.Now()
	
	if !validations.IsValidUUID(userPublicID) {
		return nil, fmt.Errorf("invalid user ID format")
	}

	// Usar query builder
	qb := query.NewQueryBuilder(`
		SELECT t.id, t.public_id, t.category_id, t.transaction_id, t.event_id, 
			   t.customer_id, t.user_id, t.code, t.status, t.seat_number, 
			   t.qr_code_url, t.price, t.used_at, t.transferred_from_ticket_id, 
			   t.created_at, t.updated_at
		FROM tickets t
		INNER JOIN users u ON t.user_id = u.id
	`).Where("u.public_id = ?", userPublicID).
		OrderBy("t.created_at", true)

	queryStr, args := qb.Build()

	rows, err := r.db.Query(ctx, queryStr, args...)
	if err != nil {
		r.logger.DatabaseLogger("SELECT", "tickets", time.Since(startTime), 0, err, map[string]interface{}{
			"user_id": utils.SafeStringForLog(userPublicID),
		})
		return nil, r.errHandler.WrapError(err, "ticket repository", "get tickets by user")
	}
	defer rows.Close()

	tickets, err := r.ticketScan.ScanAllRows(rows, r.ticketScan.ScanTicket)
	if err != nil {
		r.logger.Error("Failed to scan tickets", err, map[string]interface{}{
			"user_id": utils.SafeStringForLog(userPublicID),
		})
		return nil, err
	}

	r.logger.DatabaseLogger("SELECT", "tickets", time.Since(startTime), int64(len(tickets)), nil, map[string]interface{}{
		"user_id": utils.SafeStringForLog(userPublicID),
		"count": len(tickets),
	})
	
	return tickets, nil
}

// GetTicketsByCustomerID obtiene todos los tickets de un cliente
func (r *TicketRepository) GetTicketsByCustomerID(ctx context.Context, customerPublicID string) ([]*models.Ticket, error) {
	startTime := time.Now()
	
	if !validations.IsValidUUID(customerPublicID) {
		return nil, fmt.Errorf("invalid customer ID format")
	}

	qb := query.NewQueryBuilder(`
		SELECT t.id, t.public_id, t.category_id, t.transaction_id, t.event_id, 
			   t.customer_id, t.user_id, t.code, t.status, t.seat_number, 
			   t.qr_code_url, t.price, t.used_at, t.transferred_from_ticket_id, 
			   t.created_at, t.updated_at
		FROM tickets t
		INNER JOIN customers c ON t.customer_id = c.id
	`).Where("c.public_id = ?", customerPublicID).
		OrderBy("t.created_at", true)

	queryStr, args := qb.Build()

	rows, err := r.db.Query(ctx, queryStr, args...)
	if err != nil {
		r.logger.DatabaseLogger("SELECT", "tickets", time.Since(startTime), 0, err, map[string]interface{}{
			"customer_id": utils.SafeStringForLog(customerPublicID),
		})
		return nil, r.errHandler.WrapError(err, "ticket repository", "get tickets by customer")
	}
	defer rows.Close()

	tickets, err := r.ticketScan.ScanAllRows(rows, r.ticketScan.ScanTicket)
	if err != nil {
		r.logger.Error("Failed to scan tickets", err, map[string]interface{}{
			"customer_id": utils.SafeStringForLog(customerPublicID),
		})
		return nil, err
	}

	r.logger.DatabaseLogger("SELECT", "tickets", time.Since(startTime), int64(len(tickets)), nil, map[string]interface{}{
		"customer_id": utils.SafeStringForLog(customerPublicID),
		"count": len(tickets),
	})
	
	return tickets, nil
}

// GetTicketWithDetails obtiene un ticket con información completa usando scanner
func (r *TicketRepository) GetTicketWithDetails(ctx context.Context, ticketPublicID string) (*models.TicketWithDetails, error) {
	startTime := time.Now()
	
	if !validations.IsValidUUID(ticketPublicID) {
		return nil, fmt.Errorf("invalid ticket ID format")
	}

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

	row := r.db.QueryRow(ctx, query, ticketPublicID)
	details, err := r.ticketScan.ScanTicketWithEvent(row)
	
	if err != nil {
		if err.Error() == "ticket not found" {
			r.logger.Debug("Ticket not found", map[string]interface{}{
				"ticket_id": ticketPublicID,
			})
			return nil, err
		}
		
		r.logger.DatabaseLogger("SELECT", "tickets", time.Since(startTime), 0, err, map[string]interface{}{
			"ticket_id": ticketPublicID,
		})
		
		return nil, r.errHandler.WrapError(err, "ticket repository", "get ticket details")
	}

	r.logger.DatabaseLogger("SELECT", "tickets", time.Since(startTime), 1, nil, map[string]interface{}{
		"ticket_id": ticketPublicID,
	})
	
	return &models.TicketWithDetails{
		TicketID: details.TicketID,
		Code: details.Code,
		Status: details.Status,
		Price: details.Price,
		CreatedAt: details.CreatedAt,
		EventID: details.EventID,
		EventName: details.EventName,
		EventDate: details.EventDate,
		VenueName: details.VenueName,
	}, nil
}

// GetTicketByPublicID obtiene un ticket por public_id usando scanner
func (r *TicketRepository) GetTicketByPublicID(ctx context.Context, publicID string) (*models.Ticket, error) {
	startTime := time.Now()
	
	if !validations.IsValidUUID(publicID) {
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

	row := r.db.QueryRow(ctx, query, publicID)
	ticket, err := r.ticketScan.ScanTicket(row)
	
	if err != nil {
		if err.Error() == "ticket not found" {
			r.logger.Debug("Ticket not found", map[string]interface{}{
				"ticket_id": publicID,
			})
			return nil, err
		}
		
		r.logger.DatabaseLogger("SELECT", "tickets", time.Since(startTime), 0, err, map[string]interface{}{
			"ticket_id": publicID,
		})
		
		return nil, r.errHandler.WrapError(err, "ticket repository", "get ticket by public ID")
	}

	r.logger.DatabaseLogger("SELECT", "tickets", time.Since(startTime), 1, nil, map[string]interface{}{
		"ticket_id": publicID,
	})
	
	return ticket, nil
}

// UpdateTicketStatus actualiza el estado de un ticket con validación
func (r *TicketRepository) UpdateTicketStatus(ctx context.Context, ticketPublicID string, status string) error {
	startTime := time.Now()
	
	if !validations.IsValidUUID(ticketPublicID) {
		return fmt.Errorf("invalid ticket ID format")
	}

	// Validar estado usando validations
	status = strings.ToLower(strings.TrimSpace(status))
	if !validations.IsValidTicketStatus(status) {
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
		r.logger.DatabaseLogger("UPDATE", "tickets", time.Since(startTime), 0, err, map[string]interface{}{
			"ticket_id": ticketPublicID,
			"old_status": oldTicket.Status,
			"new_status": status,
		})
		
		return r.errHandler.WrapError(err, "ticket repository", "update ticket status")
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		r.logger.Debug("Ticket not found for status update", map[string]interface{}{
			"ticket_id": ticketPublicID,
		})
		return fmt.Errorf("ticket not found with public_id: %s", ticketPublicID)
	}

	r.logger.DatabaseLogger("UPDATE", "tickets", time.Since(startTime), rowsAffected, nil, map[string]interface{}{
		"ticket_id": ticketPublicID,
		"old_status": oldTicket.Status,
		"new_status": status,
	})
	
	return nil
}

// UpdateTicketTransaction actualiza el transaction_id de un ticket
func (r *TicketRepository) UpdateTicketTransaction(ctx context.Context, ticketPublicID string, transactionID int64) error {
	startTime := time.Now()
	
	if !validations.IsValidUUID(ticketPublicID) {
		return fmt.Errorf("invalid ticket ID format")
	}

	query := `UPDATE tickets SET transaction_id = $1, updated_at = CURRENT_TIMESTAMP WHERE public_id = $2`

	result, err := r.db.Exec(ctx, query, transactionID, ticketPublicID)
	if err != nil {
		r.logger.DatabaseLogger("UPDATE", "tickets", time.Since(startTime), 0, err, map[string]interface{}{
			"ticket_id": ticketPublicID,
			"transaction_id": transactionID,
		})
		
		return r.errHandler.WrapError(err, "ticket repository", "update ticket transaction")
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		r.logger.Debug("Ticket not found for transaction update", map[string]interface{}{
			"ticket_id": ticketPublicID,
		})
		return fmt.Errorf("ticket not found with public_id: %s", ticketPublicID)
	}

	r.logger.DatabaseLogger("UPDATE", "tickets", time.Since(startTime), rowsAffected, nil, map[string]interface{}{
		"ticket_id": ticketPublicID,
		"transaction_id": transactionID,
	})
	
	return nil
}

// MarkTicketAsUsed marca un ticket como usado
func (r *TicketRepository) MarkTicketAsUsed(ctx context.Context, ticketPublicID string) error {
	startTime := time.Now()
	
	if !validations.IsValidUUID(ticketPublicID) {
		return fmt.Errorf("invalid ticket ID format")
	}

	query := `
		UPDATE tickets 
		SET status = 'used', used_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP 
		WHERE public_id = $1 AND status = 'sold'
	`

	result, err := r.db.Exec(ctx, query, ticketPublicID)
	if err != nil {
		r.logger.DatabaseLogger("UPDATE", "tickets", time.Since(startTime), 0, err, map[string]interface{}{
			"ticket_id": ticketPublicID,
		})
		
		return r.errHandler.WrapError(err, "ticket repository", "mark ticket as used")
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		r.logger.Debug("Ticket not found or not in sold status", map[string]interface{}{
			"ticket_id": ticketPublicID,
		})
		return fmt.Errorf("ticket not found or not in sold status: %s", ticketPublicID)
	}

	r.logger.DatabaseLogger("UPDATE", "tickets", time.Since(startTime), rowsAffected, nil, map[string]interface{}{
		"ticket_id": ticketPublicID,
	})
	
	return nil
}

// GetTicketsByEvent obtiene todos los tickets de un evento
func (r *TicketRepository) GetTicketsByEvent(ctx context.Context, eventPublicID string) ([]*models.Ticket, error) {
	startTime := time.Now()
	
	if !validations.IsValidUUID(eventPublicID) {
		return nil, fmt.Errorf("invalid event ID format")
	}

	qb := query.NewQueryBuilder(`
		SELECT t.id, t.public_id, t.category_id, t.transaction_id, t.event_id, 
			   t.customer_id, t.user_id, t.code, t.status, t.seat_number, 
			   t.qr_code_url, t.price, t.used_at, t.transferred_from_ticket_id, 
			   t.created_at, t.updated_at
		FROM tickets t
		INNER JOIN events e ON t.event_id = e.id
	`).Where("e.public_id = ?", eventPublicID).
		OrderBy("t.created_at", true)

	queryStr, args := qb.Build()

	rows, err := r.db.Query(ctx, queryStr, args...)
	if err != nil {
		r.logger.DatabaseLogger("SELECT", "tickets", time.Since(startTime), 0, err, map[string]interface{}{
			"event_id": utils.SafeStringForLog(eventPublicID),
		})
		return nil, r.errHandler.WrapError(err, "ticket repository", "get tickets by event")
	}
	defer rows.Close()

	tickets, err := r.ticketScan.ScanAllRows(rows, r.ticketScan.ScanTicket)
	if err != nil {
		r.logger.Error("Failed to scan tickets", err, map[string]interface{}{
			"event_id": utils.SafeStringForLog(eventPublicID),
		})
		return nil, err
	}

	r.logger.DatabaseLogger("SELECT", "tickets", time.Since(startTime), int64(len(tickets)), nil, map[string]interface{}{
		"event_id": utils.SafeStringForLog(eventPublicID),
		"count": len(tickets),
	})
	
	return tickets, nil
}

// GetTicketsByStatus obtiene tickets por estado
func (r *TicketRepository) GetTicketsByStatus(ctx context.Context, status string) ([]*models.Ticket, error) {
	startTime := time.Now()
	
	if !validations.IsValidTicketStatus(status) {
		return nil, fmt.Errorf("invalid ticket status: %s", status)
	}

	qb := query.NewQueryBuilder(`
		SELECT id, public_id, category_id, transaction_id, event_id, 
		       customer_id, user_id, code, status, seat_number, 
		       qr_code_url, price, used_at, transferred_from_ticket_id, 
		       created_at, updated_at
		FROM tickets 
	`).Where("status = ?", status).
		OrderBy("created_at", true)

	queryStr, args := qb.Build()

	rows, err := r.db.Query(ctx, queryStr, args...)
	if err != nil {
		r.logger.DatabaseLogger("SELECT", "tickets", time.Since(startTime), 0, err, map[string]interface{}{
			"status": status,
		})
		return nil, r.errHandler.WrapError(err, "ticket repository", "get tickets by status")
	}
	defer rows.Close()

	tickets, err := r.ticketScan.ScanAllRows(rows, r.ticketScan.ScanTicket)
	if err != nil {
		r.logger.Error("Failed to scan tickets", err, map[string]interface{}{
			"status": status,
		})
		return nil, err
	}

	r.logger.DatabaseLogger("SELECT", "tickets", time.Since(startTime), int64(len(tickets)), nil, map[string]interface{}{
		"status": status,
		"count": len(tickets),
	})
	
	return tickets, nil
}

// GetTicketByCode obtiene un ticket por su código
func (r *TicketRepository) GetTicketByCode(ctx context.Context, code string) (*models.Ticket, error) {
	startTime := time.Now()
	
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

	row := r.db.QueryRow(ctx, query, code)
	ticket, err := r.ticketScan.ScanTicket(row)
	
	if err != nil {
		if err.Error() == "ticket not found" {
			r.logger.Debug("Ticket not found by code", map[string]interface{}{
				"code": utils.SafeStringForLog(code),
			})
			return nil, err
		}
		
		r.logger.DatabaseLogger("SELECT", "tickets", time.Since(startTime), 0, err, map[string]interface{}{
			"code": utils.SafeStringForLog(code),
		})
		
		return nil, r.errHandler.WrapError(err, "ticket repository", "get ticket by code")
	}

	r.logger.DatabaseLogger("SELECT", "tickets", time.Since(startTime), 1, nil, map[string]interface{}{
		"code": utils.SafeStringForLog(code),
	})
	
	return ticket, nil
}

// GetTicketsByTransaction obtiene tickets por transaction_id
func (r *TicketRepository) GetTicketsByTransaction(ctx context.Context, transactionID int64) ([]*models.Ticket, error) {
	startTime := time.Now()
	
	qb := query.NewQueryBuilder(`
		SELECT id, public_id, category_id, transaction_id, event_id, 
		       customer_id, user_id, code, status, seat_number, 
		       qr_code_url, price, used_at, transferred_from_ticket_id, 
		       created_at, updated_at
		FROM tickets 
	`).Where("transaction_id = ?", transactionID).
		OrderBy("created_at", true)

	queryStr, args := qb.Build()

	rows, err := r.db.Query(ctx, queryStr, args...)
	if err != nil {
		r.logger.DatabaseLogger("SELECT", "tickets", time.Since(startTime), 0, err, map[string]interface{}{
			"transaction_id": transactionID,
		})
		return nil, r.errHandler.WrapError(err, "ticket repository", "get tickets by transaction")
	}
	defer rows.Close()

	tickets, err := r.ticketScan.ScanAllRows(rows, r.ticketScan.ScanTicket)
	if err != nil {
		r.logger.Error("Failed to scan tickets", err, map[string]interface{}{
			"transaction_id": transactionID,
		})
		return nil, err
	}

	r.logger.DatabaseLogger("SELECT", "tickets", time.Since(startTime), int64(len(tickets)), nil, map[string]interface{}{
		"transaction_id": transactionID,
		"count": len(tickets),
	})
	
	return tickets, nil
}

// =============================================================================
// MÉTODOS PRIVADOS CON HELPERS
// =============================================================================

// validateEventAndCategory valida que el evento y categoría existan
func (r *TicketRepository) validateEventAndCategory(ctx context.Context, eventPublicID, categoryPublicID string) error {
	// Validar que el evento existe y está activo
	var eventExists bool
	err := r.db.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM events WHERE public_id = $1 AND is_active = true AND is_published = true)",
		eventPublicID).Scan(&eventExists)

	if err != nil {
		r.logger.Error("Error validating event", err, map[string]interface{}{
			"event_id": eventPublicID,
		})
		return r.errHandler.WrapError(err, "ticket repository", "validate event")
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
		r.logger.Error("Error validating category", err, map[string]interface{}{
			"category_id": categoryPublicID,
		})
		return r.errHandler.WrapError(err, "ticket repository", "validate category")
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
		r.logger.Error("Error validating category-event relationship", err, map[string]interface{}{
			"event_id": eventPublicID,
			"category_id": categoryPublicID,
		})
		return r.errHandler.WrapError(err, "ticket repository", "validate category-event relationship")
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
		r.logger.Error("Error checking category availability", err, map[string]interface{}{
			"category_id": categoryID,
		})
		return r.errHandler.WrapError(err, "ticket repository", "check category availability")
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
		r.logger.Error("Error getting category price", err, map[string]interface{}{
			"category_id": categoryID,
		})
		return 0, r.errHandler.WrapError(err, "ticket repository", "get category price")
	}

	return price, nil
}

// generateUniqueTicketCode genera un código único para el ticket usando utils
func (r *TicketRepository) generateUniqueTicketCode(ctx context.Context, tx pgx.Tx, eventID, customerID string, index int) (string, error) {
	maxAttempts := 5
	for attempt := 0; attempt < maxAttempts; attempt++ {
		code := r.generateTicketCode(eventID, customerID, index+attempt)

		// Verificar si el código ya existe
		var exists bool
		err := tx.QueryRow(ctx,
			"SELECT EXISTS(SELECT 1 FROM tickets WHERE code = $1)",
			code).Scan(&exists)

		if err != nil {
			return "", r.errHandler.WrapError(err, "ticket repository", "check ticket code uniqueness")
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
		r.logger.Error("Customer not found", err, map[string]interface{}{
			"customer_id": customerPublicID,
		})
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
		r.logger.Error("Category not found", err, map[string]interface{}{
			"category_id": categoryPublicID,
		})
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
		r.logger.Error("Event not found", err, map[string]interface{}{
			"event_id": eventPublicID,
		})
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
		r.logger.Error("User not found", err, map[string]interface{}{
			"user_id": userPublicID,
		})
		return 0, fmt.Errorf("user not found with public_id: %s", userPublicID)
	}
	return userID, nil
}

// Helper functions de validación
func (r *TicketRepository) isValidTicketStatus(status string) bool {
	return validations.IsValidTicketStatus(status)
}

func (r *TicketRepository) isValidStatusTransition(from, to string) bool {
	if transitions, exists := validStatusTransitions[from]; exists {
		return transitions[to]
	}
	return false
}

// generateTicketCode genera un código de ticket único usando utils
func (r *TicketRepository) generateTicketCode(eventID, customerID string, index int) string {
	timestamp := time.Now().UnixNano()
	shortTimestamp := fmt.Sprintf("%d", timestamp)[:8]
	// Usar solo los primeros 8 caracteres de los UUIDs para mantener el código legible
	shortEventID := eventID[:8]
	shortCustomerID := customerID[:8]
	return fmt.Sprintf("TKT-%s-%s-%s-%d", shortEventID, shortCustomerID, shortTimestamp, index)
}