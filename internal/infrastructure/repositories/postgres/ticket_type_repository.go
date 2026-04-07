// internal/infrastructure/repositories/postgres/ticket_type_repository.go
package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	commondto "github.com/franciscozamorau/osmi-server/internal/api/dto/common"
	tickettypedto "github.com/franciscozamorau/osmi-server/internal/api/dto/ticket_type"
	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
	"github.com/franciscozamorau/osmi-server/internal/domain/repository"
)

// TicketTypeRepository implementa la interfaz repository.TicketTypeRepository
type TicketTypeRepository struct {
	db *pgxpool.Pool
}

// NewTicketTypeRepository crea una nueva instancia
func NewTicketTypeRepository(db *pgxpool.Pool) *TicketTypeRepository {
	return &TicketTypeRepository{
		db: db,
	}
}

// ============================================================================
// FUNCIONES HELPER
// ============================================================================

// handleError mapea errores de PostgreSQL
func (r *TicketTypeRepository) handleError(err error, context string) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return repository.ErrTicketNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			if strings.Contains(pgErr.ConstraintName, "ticket_types_public_uuid_key") {
				return repository.ErrTicketAlreadyExists
			}
		case "23503":
			return fmt.Errorf("referenced event not found: %w", err)
		}
	}

	return fmt.Errorf("%s: %w", context, err)
}

// ============================================================================
// CRUD BÁSICO
// ============================================================================

// Create inserta un nuevo tipo de ticket
func (r *TicketTypeRepository) Create(ctx context.Context, ticketType *entities.TicketType) error {
	query := `
		INSERT INTO ticketing.ticket_types (
			public_uuid, event_id, name, description, ticket_class,
			base_price, currency, tax_rate, service_fee_type, service_fee_value,
			total_quantity, reserved_quantity, sold_quantity,
			max_per_order, min_per_order,
			sale_starts_at, sale_ends_at,
			is_active, requires_approval, is_hidden, sales_channel,
			benefits, access_type, validation_rules,
			created_at, updated_at
		) VALUES (
			gen_random_uuid(), $1, $2, $3, $4,
			$5, $6, $7, $8, $9,
			$10, 0, 0,
			$11, $12,
			$13, $14,
			$15, $16, $17, $18,
			$19, $20, $21,
			NOW(), NOW()
		)
		RETURNING id, public_uuid, created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query,
		ticketType.EventID,
		ticketType.Name,
		ticketType.Description,
		ticketType.TicketClass,
		ticketType.BasePrice,
		ticketType.Currency,
		ticketType.TaxRate,
		ticketType.ServiceFeeType,
		ticketType.ServiceFeeValue,
		ticketType.TotalQuantity,
		ticketType.MaxPerOrder,
		ticketType.MinPerOrder,
		ticketType.SaleStartsAt,
		ticketType.SaleEndsAt,
		ticketType.IsActive,
		ticketType.RequiresApproval,
		ticketType.IsHidden,
		ticketType.SalesChannel,
		ticketType.Benefits,
		ticketType.AccessType,
		ticketType.ValidationRules,
	).Scan(&ticketType.ID, &ticketType.PublicID, &ticketType.CreatedAt, &ticketType.UpdatedAt)

	if err != nil {
		return r.handleError(err, "failed to create ticket type")
	}

	return nil
}

// FindByID obtiene por ID numérico
func (r *TicketTypeRepository) FindByID(ctx context.Context, id int64) (*entities.TicketType, error) {
	query := `
		SELECT 
			id, public_uuid, event_id, name, description, ticket_class,
			base_price, currency, tax_rate, service_fee_type, service_fee_value,
			total_quantity, reserved_quantity, sold_quantity,
			max_per_order, min_per_order,
			sale_starts_at, sale_ends_at,
			is_active, requires_approval, is_hidden, sales_channel,
			benefits, access_type, validation_rules,
			available_quantity, is_sold_out,
			created_at, updated_at
		FROM ticketing.ticket_types
		WHERE id = $1
	`

	var tt entities.TicketType
	var description *string
	var saleEndsAt *time.Time
	var benefitsJSON []byte
	var validationRulesJSON []byte

	err := r.db.QueryRow(ctx, query, id).Scan(
		&tt.ID, &tt.PublicID, &tt.EventID,
		&tt.Name, &description, &tt.TicketClass,
		&tt.BasePrice, &tt.Currency, &tt.TaxRate, &tt.ServiceFeeType, &tt.ServiceFeeValue,
		&tt.TotalQuantity, &tt.ReservedQuantity, &tt.SoldQuantity,
		&tt.MaxPerOrder, &tt.MinPerOrder,
		&tt.SaleStartsAt, &saleEndsAt,
		&tt.IsActive, &tt.RequiresApproval, &tt.IsHidden, &tt.SalesChannel,
		&benefitsJSON,
		&tt.AccessType,
		&validationRulesJSON,
		&tt.AvailableQuantity, &tt.IsSoldOut,
		&tt.CreatedAt, &tt.UpdatedAt,
	)

	if err != nil {
		return nil, r.handleError(err, "failed to get ticket type by ID")
	}

	if description != nil {
		tt.Description = description
	}
	if saleEndsAt != nil {
		tt.SaleEndsAt = saleEndsAt
	}

	if len(benefitsJSON) > 0 {
		if err := json.Unmarshal(benefitsJSON, &tt.Benefits); err != nil {
			log.Printf("⚠️ Error deserializando benefits: %v", err)
			tt.Benefits = []string{}
		}
	} else {
		tt.Benefits = []string{}
	}

	if len(validationRulesJSON) > 0 {
		var rules entities.ValidationRules
		if err := json.Unmarshal(validationRulesJSON, &rules); err != nil {
			log.Printf("⚠️ Error deserializando validationRules: %v", err)
		} else {
			tt.ValidationRules = &rules
		}
	}

	return &tt, nil
}

// FindByPublicID obtiene por UUID
func (r *TicketTypeRepository) FindByPublicID(ctx context.Context, publicID string) (*entities.TicketType, error) {
	log.Printf("🔍 FindByPublicID: %s", publicID)

	query := `
		SELECT 
			id, public_uuid, event_id, name, description, ticket_class,
			base_price, currency, tax_rate, service_fee_type, service_fee_value,
			total_quantity, reserved_quantity, sold_quantity,
			max_per_order, min_per_order,
			sale_starts_at, sale_ends_at,
			is_active, requires_approval, is_hidden, sales_channel,
			benefits, access_type, validation_rules,
			available_quantity, is_sold_out,
			created_at, updated_at
		FROM ticketing.ticket_types
		WHERE public_uuid = $1
	`

	var tt entities.TicketType
	var description *string
	var saleEndsAt *time.Time
	var benefitsJSON []byte
	var validationRulesJSON []byte

	err := r.db.QueryRow(ctx, query, publicID).Scan(
		&tt.ID, &tt.PublicID, &tt.EventID,
		&tt.Name, &description, &tt.TicketClass,
		&tt.BasePrice, &tt.Currency, &tt.TaxRate, &tt.ServiceFeeType, &tt.ServiceFeeValue,
		&tt.TotalQuantity, &tt.ReservedQuantity, &tt.SoldQuantity,
		&tt.MaxPerOrder, &tt.MinPerOrder,
		&tt.SaleStartsAt, &saleEndsAt,
		&tt.IsActive, &tt.RequiresApproval, &tt.IsHidden, &tt.SalesChannel,
		&benefitsJSON,
		&tt.AccessType,
		&validationRulesJSON,
		&tt.AvailableQuantity, &tt.IsSoldOut,
		&tt.CreatedAt, &tt.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrTicketNotFound
		}
		log.Printf("❌ Error en FindByPublicID: %v", err)
		return nil, r.handleError(err, "failed to get ticket type by public ID")
	}

	if description != nil {
		tt.Description = description
	}
	if saleEndsAt != nil {
		tt.SaleEndsAt = saleEndsAt
	}

	if len(benefitsJSON) > 0 {
		if err := json.Unmarshal(benefitsJSON, &tt.Benefits); err != nil {
			log.Printf("⚠️ Error deserializando benefits: %v", err)
			tt.Benefits = []string{}
		}
	} else {
		tt.Benefits = []string{}
	}

	if len(validationRulesJSON) > 0 {
		var rules entities.ValidationRules
		if err := json.Unmarshal(validationRulesJSON, &rules); err != nil {
			log.Printf("⚠️ Error deserializando validationRules: %v", err)
		} else {
			tt.ValidationRules = &rules
		}
	}

	log.Printf("✅ Ticket type encontrado: %s (ID: %d)", tt.Name, tt.ID)
	return &tt, nil
}

// Update actualiza un tipo de ticket
func (r *TicketTypeRepository) Update(ctx context.Context, ticketType *entities.TicketType) error {
	_, err := r.FindByID(ctx, ticketType.ID)
	if err != nil {
		return repository.ErrTicketNotFound
	}

	query := `
		UPDATE ticketing.ticket_types SET
			name = $1,
			description = $2,
			base_price = $3,
			currency = $4,
			tax_rate = $5,
			service_fee_type = $6,
			service_fee_value = $7,
			total_quantity = $8,
			max_per_order = $9,
			min_per_order = $10,
			sale_starts_at = $11,
			sale_ends_at = $12,
			is_active = $13,
			is_hidden = $14,
			benefits = $15,
			validation_rules = $16,
			updated_at = NOW()
		WHERE id = $17
		RETURNING updated_at
	`

	err = r.db.QueryRow(ctx, query,
		ticketType.Name,
		ticketType.Description,
		ticketType.BasePrice,
		ticketType.Currency,
		ticketType.TaxRate,
		ticketType.ServiceFeeType,
		ticketType.ServiceFeeValue,
		ticketType.TotalQuantity,
		ticketType.MaxPerOrder,
		ticketType.MinPerOrder,
		ticketType.SaleStartsAt,
		ticketType.SaleEndsAt,
		ticketType.IsActive,
		ticketType.IsHidden,
		ticketType.Benefits,
		ticketType.ValidationRules,
		ticketType.ID,
	).Scan(&ticketType.UpdatedAt)

	if err != nil {
		return r.handleError(err, "failed to update ticket type")
	}
	return nil
}

// Delete elimina permanentemente
func (r *TicketTypeRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM ticketing.ticket_types WHERE id = $1`
	cmdTag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return r.handleError(err, "failed to delete ticket type")
	}
	if cmdTag.RowsAffected() == 0 {
		return repository.ErrTicketNotFound
	}
	return nil
}

// SoftDelete desactiva un tipo de ticket
func (r *TicketTypeRepository) SoftDelete(ctx context.Context, publicID string) error {
	query := `UPDATE ticketing.ticket_types SET is_active = false, updated_at = NOW() WHERE public_uuid = $1`
	cmdTag, err := r.db.Exec(ctx, query, publicID)
	if err != nil {
		return r.handleError(err, "failed to soft delete ticket type")
	}
	if cmdTag.RowsAffected() == 0 {
		return repository.ErrTicketNotFound
	}
	return nil
}

// Exists verifica existencia por ID
func (r *TicketTypeRepository) Exists(ctx context.Context, id int64) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM ticketing.ticket_types WHERE id = $1)`
	err := r.db.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, r.handleError(err, "failed to check existence")
	}
	return exists, nil
}

// ============================================================================
// BÚSQUEDAS
// ============================================================================

// List lista con filtros y paginación
func (r *TicketTypeRepository) List(ctx context.Context, filter tickettypedto.TicketTypeFilter, pagination commondto.Pagination) ([]*entities.TicketType, int64, error) {
	where := []string{"1=1"}
	args := []interface{}{}
	argPos := 1

	if filter.EventID != nil {
		where = append(where, fmt.Sprintf("event_id = $%d", argPos))
		args = append(args, *filter.EventID)
		argPos++
	}
	if filter.IsActive != nil {
		where = append(where, fmt.Sprintf("is_active = $%d", argPos))
		args = append(args, *filter.IsActive)
		argPos++
	}
	if filter.IsSoldOut != nil {
		where = append(where, fmt.Sprintf("is_sold_out = $%d", argPos))
		args = append(args, *filter.IsSoldOut)
		argPos++
	}
	if filter.MinPrice != nil {
		where = append(where, fmt.Sprintf("base_price >= $%d", argPos))
		args = append(args, *filter.MinPrice)
		argPos++
	}
	if filter.MaxPrice != nil {
		where = append(where, fmt.Sprintf("base_price <= $%d", argPos))
		args = append(args, *filter.MaxPrice)
		argPos++
	}
	if filter.Currency != "" {
		where = append(where, fmt.Sprintf("currency = $%d", argPos))
		args = append(args, filter.Currency)
		argPos++
	}
	if filter.Search != "" {
		where = append(where, fmt.Sprintf("(name ILIKE $%d OR description ILIKE $%d)", argPos, argPos+1))
		args = append(args, "%"+filter.Search+"%", "%"+filter.Search+"%")
		argPos += 2
	}

	whereClause := strings.Join(where, " AND ")

	// Contar total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM ticketing.ticket_types WHERE %s", whereClause)
	var total int64
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, r.handleError(err, "failed to count ticket types")
	}

	// Obtener datos
	query := fmt.Sprintf(`
		SELECT 
			id, public_uuid, event_id, name, description, ticket_class,
			base_price, currency, tax_rate, service_fee_type, service_fee_value,
			total_quantity, reserved_quantity, sold_quantity,
			max_per_order, min_per_order,
			sale_starts_at, sale_ends_at,
			is_active, requires_approval, is_hidden, sales_channel,
			benefits, access_type, validation_rules,
			available_quantity, is_sold_out,
			created_at, updated_at
		FROM ticketing.ticket_types
		WHERE %s
		ORDER BY base_price
		LIMIT $%d OFFSET $%d
	`, whereClause, argPos, argPos+1)

	queryArgs := append(args, pagination.PageSize, (pagination.Page-1)*pagination.PageSize)

	rows, err := r.db.Query(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, r.handleError(err, "failed to list ticket types")
	}
	defer rows.Close()

	var types []*entities.TicketType
	for rows.Next() {
		var tt entities.TicketType
		var description *string
		var saleEndsAt *time.Time
		var benefitsJSON []byte
		var validationRulesJSON []byte

		err = rows.Scan(
			&tt.ID, &tt.PublicID, &tt.EventID,
			&tt.Name, &description, &tt.TicketClass,
			&tt.BasePrice, &tt.Currency, &tt.TaxRate, &tt.ServiceFeeType, &tt.ServiceFeeValue,
			&tt.TotalQuantity, &tt.ReservedQuantity, &tt.SoldQuantity,
			&tt.MaxPerOrder, &tt.MinPerOrder,
			&tt.SaleStartsAt, &saleEndsAt,
			&tt.IsActive, &tt.RequiresApproval, &tt.IsHidden, &tt.SalesChannel,
			&benefitsJSON,
			&tt.AccessType,
			&validationRulesJSON,
			&tt.AvailableQuantity, &tt.IsSoldOut,
			&tt.CreatedAt, &tt.UpdatedAt,
		)
		if err != nil {
			return nil, 0, r.handleError(err, "failed to scan ticket type row")
		}

		if description != nil {
			tt.Description = description
		}
		if saleEndsAt != nil {
			tt.SaleEndsAt = saleEndsAt
		}

		if len(benefitsJSON) > 0 {
			if err := json.Unmarshal(benefitsJSON, &tt.Benefits); err != nil {
				tt.Benefits = []string{}
			}
		} else {
			tt.Benefits = []string{}
		}

		if len(validationRulesJSON) > 0 {
			var rules entities.ValidationRules
			if err := json.Unmarshal(validationRulesJSON, &rules); err == nil {
				tt.ValidationRules = &rules
			}
		}

		types = append(types, &tt)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, r.handleError(err, "error iterating ticket type rows")
	}

	return types, total, nil
}

// FindByEvent obtiene tipos por evento
func (r *TicketTypeRepository) FindByEvent(ctx context.Context, eventID int64, activeOnly bool) ([]*entities.TicketType, error) {
	query := `
		SELECT 
			id, public_uuid, event_id, name, description, ticket_class,
			base_price, currency, tax_rate, service_fee_type, service_fee_value,
			total_quantity, reserved_quantity, sold_quantity,
			max_per_order, min_per_order,
			sale_starts_at, sale_ends_at,
			is_active, requires_approval, is_hidden, sales_channel,
			benefits, access_type, validation_rules,
			available_quantity, is_sold_out,
			created_at, updated_at
		FROM ticketing.ticket_types
		WHERE event_id = $1
	`
	if activeOnly {
		query += ` AND is_active = true`
	}
	query += ` ORDER BY base_price`

	rows, err := r.db.Query(ctx, query, eventID)
	if err != nil {
		return nil, r.handleError(err, "failed to get ticket types by event")
	}
	defer rows.Close()

	var types []*entities.TicketType
	for rows.Next() {
		var tt entities.TicketType
		var description *string
		var saleEndsAt *time.Time
		var benefitsJSON []byte
		var validationRulesJSON []byte

		err = rows.Scan(
			&tt.ID, &tt.PublicID, &tt.EventID,
			&tt.Name, &description, &tt.TicketClass,
			&tt.BasePrice, &tt.Currency, &tt.TaxRate, &tt.ServiceFeeType, &tt.ServiceFeeValue,
			&tt.TotalQuantity, &tt.ReservedQuantity, &tt.SoldQuantity,
			&tt.MaxPerOrder, &tt.MinPerOrder,
			&tt.SaleStartsAt, &saleEndsAt,
			&tt.IsActive, &tt.RequiresApproval, &tt.IsHidden, &tt.SalesChannel,
			&benefitsJSON,
			&tt.AccessType,
			&validationRulesJSON,
			&tt.AvailableQuantity, &tt.IsSoldOut,
			&tt.CreatedAt, &tt.UpdatedAt,
		)
		if err != nil {
			return nil, r.handleError(err, "failed to scan ticket type row")
		}

		if description != nil {
			tt.Description = description
		}
		if saleEndsAt != nil {
			tt.SaleEndsAt = saleEndsAt
		}

		if len(benefitsJSON) > 0 {
			json.Unmarshal(benefitsJSON, &tt.Benefits)
		}

		if len(validationRulesJSON) > 0 {
			var rules entities.ValidationRules
			if json.Unmarshal(validationRulesJSON, &rules) == nil {
				tt.ValidationRules = &rules
			}
		}

		types = append(types, &tt)
	}

	return types, nil
}

// FindByEventPublicID obtiene por UUID del evento
func (r *TicketTypeRepository) FindByEventPublicID(ctx context.Context, eventPublicID string) ([]*entities.TicketType, error) {
	query := `
		SELECT tt.*
		FROM ticketing.ticket_types tt
		JOIN ticketing.events e ON tt.event_id = e.id
		WHERE e.public_uuid = $1
		ORDER BY tt.base_price
	`

	rows, err := r.db.Query(ctx, query, eventPublicID)
	if err != nil {
		return nil, r.handleError(err, "failed to get ticket types by event public ID")
	}
	defer rows.Close()

	var types []*entities.TicketType
	for rows.Next() {
		var tt entities.TicketType
		var description *string
		var saleEndsAt *time.Time
		var benefitsJSON []byte
		var validationRulesJSON []byte

		err = rows.Scan(
			&tt.ID, &tt.PublicID, &tt.EventID,
			&tt.Name, &description, &tt.TicketClass,
			&tt.BasePrice, &tt.Currency, &tt.TaxRate, &tt.ServiceFeeType, &tt.ServiceFeeValue,
			&tt.TotalQuantity, &tt.ReservedQuantity, &tt.SoldQuantity,
			&tt.MaxPerOrder, &tt.MinPerOrder,
			&tt.SaleStartsAt, &saleEndsAt,
			&tt.IsActive, &tt.RequiresApproval, &tt.IsHidden, &tt.SalesChannel,
			&benefitsJSON,
			&tt.AccessType,
			&validationRulesJSON,
			&tt.AvailableQuantity, &tt.IsSoldOut,
			&tt.CreatedAt, &tt.UpdatedAt,
		)
		if err != nil {
			return nil, r.handleError(err, "failed to scan ticket type row")
		}

		if description != nil {
			tt.Description = description
		}
		if saleEndsAt != nil {
			tt.SaleEndsAt = saleEndsAt
		}

		if len(benefitsJSON) > 0 {
			json.Unmarshal(benefitsJSON, &tt.Benefits)
		}

		if len(validationRulesJSON) > 0 {
			var rules entities.ValidationRules
			if json.Unmarshal(validationRulesJSON, &rules) == nil {
				tt.ValidationRules = &rules
			}
		}

		types = append(types, &tt)
	}

	return types, nil
}

// FindAvailable obtiene tipos disponibles (con stock > 0)
func (r *TicketTypeRepository) FindAvailable(ctx context.Context, eventID int64) ([]*entities.TicketType, error) {
	query := `
		SELECT 
			id, public_uuid, event_id, name, description, ticket_class,
			base_price, currency, tax_rate, service_fee_type, service_fee_value,
			total_quantity, reserved_quantity, sold_quantity,
			max_per_order, min_per_order,
			sale_starts_at, sale_ends_at,
			is_active, requires_approval, is_hidden, sales_channel,
			benefits, access_type, validation_rules,
			available_quantity, is_sold_out,
			created_at, updated_at
		FROM ticketing.ticket_types
		WHERE event_id = $1
			AND is_active = true
			AND (total_quantity - sold_quantity - reserved_quantity) > 0
			AND (sale_starts_at IS NULL OR sale_starts_at <= NOW())
			AND (sale_ends_at IS NULL OR sale_ends_at >= NOW())
		ORDER BY base_price
	`

	rows, err := r.db.Query(ctx, query, eventID)
	if err != nil {
		return nil, r.handleError(err, "failed to find available ticket types")
	}
	defer rows.Close()

	var types []*entities.TicketType
	for rows.Next() {
		var tt entities.TicketType
		var description *string
		var saleEndsAt *time.Time
		var benefitsJSON []byte
		var validationRulesJSON []byte

		err = rows.Scan(
			&tt.ID, &tt.PublicID, &tt.EventID,
			&tt.Name, &description, &tt.TicketClass,
			&tt.BasePrice, &tt.Currency, &tt.TaxRate, &tt.ServiceFeeType, &tt.ServiceFeeValue,
			&tt.TotalQuantity, &tt.ReservedQuantity, &tt.SoldQuantity,
			&tt.MaxPerOrder, &tt.MinPerOrder,
			&tt.SaleStartsAt, &saleEndsAt,
			&tt.IsActive, &tt.RequiresApproval, &tt.IsHidden, &tt.SalesChannel,
			&benefitsJSON,
			&tt.AccessType,
			&validationRulesJSON,
			&tt.AvailableQuantity, &tt.IsSoldOut,
			&tt.CreatedAt, &tt.UpdatedAt,
		)
		if err != nil {
			return nil, r.handleError(err, "failed to scan ticket type row")
		}

		if description != nil {
			tt.Description = description
		}
		if saleEndsAt != nil {
			tt.SaleEndsAt = saleEndsAt
		}

		if len(benefitsJSON) > 0 {
			json.Unmarshal(benefitsJSON, &tt.Benefits)
		}

		if len(validationRulesJSON) > 0 {
			var rules entities.ValidationRules
			if json.Unmarshal(validationRulesJSON, &rules) == nil {
				tt.ValidationRules = &rules
			}
		}

		types = append(types, &tt)
	}

	return types, nil
}

// FindSoldOut obtiene tipos agotados
func (r *TicketTypeRepository) FindSoldOut(ctx context.Context, eventID int64) ([]*entities.TicketType, error) {
	query := `
		SELECT 
			id, public_uuid, event_id, name, description, ticket_class,
			base_price, currency, tax_rate, service_fee_type, service_fee_value,
			total_quantity, reserved_quantity, sold_quantity,
			max_per_order, min_per_order,
			sale_starts_at, sale_ends_at,
			is_active, requires_approval, is_hidden, sales_channel,
			benefits, access_type, validation_rules,
			available_quantity, is_sold_out,
			created_at, updated_at
		FROM ticketing.ticket_types
		WHERE event_id = $1 AND is_sold_out = true
		ORDER BY base_price
	`

	rows, err := r.db.Query(ctx, query, eventID)
	if err != nil {
		return nil, r.handleError(err, "failed to find sold out ticket types")
	}
	defer rows.Close()

	var types []*entities.TicketType
	for rows.Next() {
		var tt entities.TicketType
		var description *string
		var saleEndsAt *time.Time
		var benefitsJSON []byte
		var validationRulesJSON []byte

		err = rows.Scan(
			&tt.ID, &tt.PublicID, &tt.EventID,
			&tt.Name, &description, &tt.TicketClass,
			&tt.BasePrice, &tt.Currency, &tt.TaxRate, &tt.ServiceFeeType, &tt.ServiceFeeValue,
			&tt.TotalQuantity, &tt.ReservedQuantity, &tt.SoldQuantity,
			&tt.MaxPerOrder, &tt.MinPerOrder,
			&tt.SaleStartsAt, &saleEndsAt,
			&tt.IsActive, &tt.RequiresApproval, &tt.IsHidden, &tt.SalesChannel,
			&benefitsJSON,
			&tt.AccessType,
			&validationRulesJSON,
			&tt.AvailableQuantity, &tt.IsSoldOut,
			&tt.CreatedAt, &tt.UpdatedAt,
		)
		if err != nil {
			return nil, r.handleError(err, "failed to scan ticket type row")
		}

		if description != nil {
			tt.Description = description
		}
		if saleEndsAt != nil {
			tt.SaleEndsAt = saleEndsAt
		}

		if len(benefitsJSON) > 0 {
			json.Unmarshal(benefitsJSON, &tt.Benefits)
		}

		if len(validationRulesJSON) > 0 {
			var rules entities.ValidationRules
			if json.Unmarshal(validationRulesJSON, &rules) == nil {
				tt.ValidationRules = &rules
			}
		}

		types = append(types, &tt)
	}

	return types, nil
}

// ============================================================================
// OPERACIONES DE INVENTARIO
// ============================================================================

// UpdateQuantity actualiza la cantidad total
func (r *TicketTypeRepository) UpdateQuantity(ctx context.Context, ticketTypeID int64, quantity int) error {
	query := `
		UPDATE ticketing.ticket_types
		SET total_quantity = $1,
			updated_at = NOW()
		WHERE id = $2
		RETURNING id
	`
	var id int64
	err := r.db.QueryRow(ctx, query, quantity, ticketTypeID).Scan(&id)
	if err != nil {
		return r.handleError(err, "failed to update quantity")
	}
	return nil
}

// ReserveTickets reserva tickets
func (r *TicketTypeRepository) ReserveTickets(ctx context.Context, ticketTypeID int64, quantity int) error {
	query := `
		UPDATE ticketing.ticket_types
		SET reserved_quantity = reserved_quantity + $1,
			updated_at = NOW()
		WHERE id = $2 
		AND (total_quantity - sold_quantity - reserved_quantity) >= $1
		RETURNING id
	`
	var id int64
	err := r.db.QueryRow(ctx, query, quantity, ticketTypeID).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("not enough tickets available")
		}
		return r.handleError(err, "failed to reserve tickets")
	}
	return nil
}

// ReleaseReservation libera reservas
func (r *TicketTypeRepository) ReleaseReservation(ctx context.Context, ticketTypeID int64, quantity int) error {
	query := `
		UPDATE ticketing.ticket_types
		SET reserved_quantity = GREATEST(0, reserved_quantity - $1),
			updated_at = NOW()
		WHERE id = $2 AND reserved_quantity >= $1
		RETURNING id
	`
	var id int64
	err := r.db.QueryRow(ctx, query, quantity, ticketTypeID).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("not enough reserved tickets")
		}
		return r.handleError(err, "failed to release reservation")
	}
	return nil
}

// SellTickets vende tickets (convierte reservas en ventas)
func (r *TicketTypeRepository) SellTickets(ctx context.Context, ticketTypeID int64, quantity int) error {
	query := `
	    UPDATE ticketing.ticket_types
    SET reserved_quantity = reserved_quantity + $1,
        updated_at = NOW()
    WHERE id = $2 
    AND (total_quantity - sold_quantity - reserved_quantity) >= $1
    RETURNING id
`
	var id int64
	err := r.db.QueryRow(ctx, query, quantity, ticketTypeID).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("not enough reserved tickets to sell")
		}
		return r.handleError(err, "failed to sell tickets")
	}
	return nil
}

// CancelSoldTickets cancela tickets vendidos
func (r *TicketTypeRepository) CancelSoldTickets(ctx context.Context, ticketTypeID int64, quantity int) error {
	query := `
		UPDATE ticketing.ticket_types
		SET sold_quantity = GREATEST(0, sold_quantity - $1),
			available_quantity = total_quantity - GREATEST(0, sold_quantity - $1) - reserved_quantity,
			is_sold_out = (total_quantity - GREATEST(0, sold_quantity - $1) - reserved_quantity) <= 0,
			updated_at = NOW()
		WHERE id = $2 AND sold_quantity >= $1
		RETURNING id
	`
	var id int64
	err := r.db.QueryRow(ctx, query, quantity, ticketTypeID).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("not enough sold tickets to cancel")
		}
		return r.handleError(err, "failed to cancel sold tickets")
	}
	return nil
}

// RefundTickets reembolsa tickets vendidos
func (r *TicketTypeRepository) RefundTickets(ctx context.Context, ticketTypeID int64, quantity int) error {
	query := `
		UPDATE ticketing.ticket_types
		SET sold_quantity = GREATEST(0, sold_quantity - $1),
			available_quantity = total_quantity - GREATEST(0, sold_quantity - $1) - reserved_quantity,
			is_sold_out = (total_quantity - GREATEST(0, sold_quantity - $1) - reserved_quantity) <= 0,
			updated_at = NOW()
		WHERE id = $2 AND sold_quantity >= $1
		RETURNING id
	`
	var id int64
	err := r.db.QueryRow(ctx, query, quantity, ticketTypeID).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("not enough sold tickets to refund")
		}
		return r.handleError(err, "failed to refund tickets")
	}
	return nil
}

// CheckAvailability verifica disponibilidad
func (r *TicketTypeRepository) CheckAvailability(ctx context.Context, ticketTypeID int64, quantity int) (bool, error) {
	var available bool
	query := `
		SELECT (total_quantity - sold_quantity - reserved_quantity) >= $1
		FROM ticketing.ticket_types
		WHERE id = $2 AND is_active = true
	`
	err := r.db.QueryRow(ctx, query, quantity, ticketTypeID).Scan(&available)
	if err != nil {
		return false, r.handleError(err, "failed to check availability")
	}
	return available, nil
}

// GetAvailableQuantity obtiene cantidad disponible
func (r *TicketTypeRepository) GetAvailableQuantity(ctx context.Context, ticketTypeID int64) (int, error) {
	var quantity int
	query := `SELECT available_quantity FROM ticketing.ticket_types WHERE id = $1`
	err := r.db.QueryRow(ctx, query, ticketTypeID).Scan(&quantity)
	if err != nil {
		return 0, r.handleError(err, "failed to get available quantity")
	}
	return quantity, nil
}

// UpdateSaleDates actualiza fechas de venta
func (r *TicketTypeRepository) UpdateSaleDates(ctx context.Context, ticketTypeID int64, startsAt, endsAt string) error {
	var starts, ends *time.Time
	if startsAt != "" {
		t, err := time.Parse(time.RFC3339, startsAt)
		if err != nil {
			return fmt.Errorf("invalid start date format: %w", err)
		}
		starts = &t
	}
	if endsAt != "" {
		t, err := time.Parse(time.RFC3339, endsAt)
		if err != nil {
			return fmt.Errorf("invalid end date format: %w", err)
		}
		ends = &t
	}

	query := `
		UPDATE ticketing.ticket_types
		SET sale_starts_at = $1,
			sale_ends_at = $2,
			updated_at = NOW()
		WHERE id = $3
	`
	_, err := r.db.Exec(ctx, query, starts, ends, ticketTypeID)
	if err != nil {
		return r.handleError(err, "failed to update sale dates")
	}
	return nil
}

// UpdatePrice actualiza precio
func (r *TicketTypeRepository) UpdatePrice(ctx context.Context, ticketTypeID int64, price float64, currency string) error {
	query := `
		UPDATE ticketing.ticket_types
		SET base_price = $1,
			currency = $2,
			updated_at = NOW()
		WHERE id = $3
	`
	_, err := r.db.Exec(ctx, query, price, currency, ticketTypeID)
	if err != nil {
		return r.handleError(err, "failed to update price")
	}
	return nil
}

// UpdateStatus actualiza estado activo
func (r *TicketTypeRepository) UpdateStatus(ctx context.Context, ticketTypeID int64, active bool) error {
	query := `UPDATE ticketing.ticket_types SET is_active = $1, updated_at = NOW() WHERE id = $2`
	cmdTag, err := r.db.Exec(ctx, query, active, ticketTypeID)
	if err != nil {
		return r.handleError(err, "failed to update status")
	}
	if cmdTag.RowsAffected() == 0 {
		return repository.ErrTicketNotFound
	}
	return nil
}

// ============================================================================
// ESTADÍSTICAS
// ============================================================================

// CountSold cuenta tickets vendidos
func (r *TicketTypeRepository) CountSold(ctx context.Context, ticketTypeID int64) (int, error) {
	var count int
	query := `SELECT sold_quantity FROM ticketing.ticket_types WHERE id = $1`
	err := r.db.QueryRow(ctx, query, ticketTypeID).Scan(&count)
	if err != nil {
		return 0, r.handleError(err, "failed to count sold tickets")
	}
	return count, nil
}

// CountReserved cuenta tickets reservados
func (r *TicketTypeRepository) CountReserved(ctx context.Context, ticketTypeID int64) (int, error) {
	var count int
	query := `SELECT reserved_quantity FROM ticketing.ticket_types WHERE id = $1`
	err := r.db.QueryRow(ctx, query, ticketTypeID).Scan(&count)
	if err != nil {
		return 0, r.handleError(err, "failed to count reserved tickets")
	}
	return count, nil
}

// GetRevenue obtiene ingresos totales
func (r *TicketTypeRepository) GetRevenue(ctx context.Context, ticketTypeID int64) (float64, error) {
	var revenue float64
	query := `
		SELECT COALESCE(SUM(final_price), 0)
		FROM ticketing.tickets
		WHERE ticket_type_id = $1 AND status IN ('sold', 'checked_in')
	`
	err := r.db.QueryRow(ctx, query, ticketTypeID).Scan(&revenue)
	if err != nil {
		return 0, r.handleError(err, "failed to get revenue")
	}
	return revenue, nil
}

// GetSalesVelocity obtiene velocidad de ventas (tickets por día)
func (r *TicketTypeRepository) GetSalesVelocity(ctx context.Context, ticketTypeID int64) (float64, error) {
	var velocity float64
	query := `
		WITH first_sale AS (
			SELECT MIN(sold_at) as first_sale
			FROM ticketing.tickets
			WHERE ticket_type_id = $1 AND sold_at IS NOT NULL
		)
		SELECT 
			COALESCE(
				COUNT(*)::float / 
				EXTRACT(EPOCH FROM (NOW() - first_sale.first_sale)) / 86400,
				0
			) as velocity
		FROM ticketing.tickets, first_sale
		WHERE ticket_type_id = $1 AND sold_at IS NOT NULL
		GROUP BY first_sale.first_sale
	`
	err := r.db.QueryRow(ctx, query, ticketTypeID).Scan(&velocity)
	if err != nil {
		return 0, r.handleError(err, "failed to get sales velocity")
	}
	return velocity, nil
}

// GetStats obtiene estadísticas completas
func (r *TicketTypeRepository) GetStats(ctx context.Context, ticketTypeID int64) (*tickettypedto.TicketTypeStatsResponse, error) {
	query := `
        SELECT 
            total_quantity as total_tickets,
            reserved_quantity as reserved_tickets,
            sold_quantity as sold_tickets,
            total_quantity - sold_quantity - reserved_quantity as available_tickets,
            COALESCE(SUM(sold_quantity * base_price), 0) as total_revenue,
            COALESCE(AVG(base_price), 0) as avg_ticket_price,
            CASE 
                WHEN total_quantity > 0 
                THEN (sold_quantity::float / total_quantity::float) * 100 
                ELSE 0 
            END as sell_through_rate
        FROM ticketing.ticket_types
        WHERE id = $1
        GROUP BY id, total_quantity, reserved_quantity, sold_quantity, base_price
    `

	var stats tickettypedto.TicketTypeStatsResponse
	err := r.db.QueryRow(ctx, query, ticketTypeID).Scan(
		&stats.TotalTickets,
		&stats.ReservedTickets,
		&stats.SoldTickets,
		&stats.AvailableTickets,
		&stats.TotalRevenue,
		&stats.AvgTicketPrice,
		&stats.SellThroughRate,
	)
	if err != nil {
		return nil, r.handleError(err, "failed to get ticket type stats")
	}
	return &stats, nil
}

// GetEventTicketStats obtiene estadísticas de tickets para un evento
func (r *TicketTypeRepository) GetEventTicketStats(ctx context.Context, eventID int64) (*tickettypedto.EventTicketStats, error) {
	query := `
		SELECT 
			event_id,
			COUNT(*) as ticket_type_id,
			'' as ticket_type_name,
			SUM(total_quantity) as total_quantity,
			SUM(sold_quantity) as sold_quantity,
			SUM(reserved_quantity) as reserved_quantity,
			SUM(total_quantity - sold_quantity - reserved_quantity) as available_quantity,
			COALESCE(SUM(sold_quantity * base_price), 0) as revenue,
			CASE 
				WHEN SUM(total_quantity) > 0 
				THEN (SUM(sold_quantity)::float / SUM(total_quantity)::float) * 100
				ELSE 0
			END as sell_through_rate
		FROM ticketing.ticket_types
		WHERE event_id = $1
		GROUP BY event_id
	`

	var stats tickettypedto.EventTicketStats
	err := r.db.QueryRow(ctx, query, eventID).Scan(
		&stats.EventID,
		&stats.TicketTypeID,
		&stats.TicketTypeName,
		&stats.TotalQuantity,
		&stats.SoldQuantity,
		&stats.ReservedQuantity,
		&stats.AvailableQuantity,
		&stats.Revenue,
		&stats.SellThroughRate,
	)
	if err != nil {
		return nil, r.handleError(err, "failed to get event ticket stats")
	}
	return &stats, nil
}

// SellTicketsDirect vende tickets directamente sin reserva previa
func (r *TicketTypeRepository) SellTicketsDirect(ctx context.Context, ticketTypeID int64, quantity int) error {
	query := `
        UPDATE ticketing.ticket_types
        SET sold_quantity = sold_quantity + $1,
            updated_at = NOW()
        WHERE id = $2 
        AND (total_quantity - sold_quantity - reserved_quantity) >= $1
        RETURNING id
    `
	var id int64
	err := r.db.QueryRow(ctx, query, quantity, ticketTypeID).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("not enough tickets available to sell")
		}
		return r.handleError(err, "failed to sell tickets directly")
	}
	return nil
}

// ConfirmReservation confirma una reserva (la convierte en venta)
func (r *TicketTypeRepository) ConfirmReservation(ctx context.Context, ticketTypeID int64, quantity int) error {
	query := `
		UPDATE ticketing.ticket_types
		SET sold_quantity = sold_quantity + $1,
			reserved_quantity = reserved_quantity - $1,
			updated_at = NOW()
		WHERE id = $2 AND reserved_quantity >= $1
		RETURNING id
	`
	var id int64
	err := r.db.QueryRow(ctx, query, quantity, ticketTypeID).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("not enough reserved tickets to confirm")
		}
		return r.handleError(err, "failed to confirm reservation")
	}
	return nil
}
