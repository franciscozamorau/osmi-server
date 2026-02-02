package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/franciscozamorau/osmi-server/internal/api/dto"
	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
	"github.com/franciscozamorau/osmi-server/internal/domain/repository"
	"github.com/franciscozamorau/osmi-server/internal/domain/valueobjects"
	"github.com/franciscozamorau/osmi-server/internal/infrastructure/repositories/postgres/helpers/errors"
	"github.com/franciscozamorau/osmi-server/internal/infrastructure/repositories/postgres/helpers/query"
	"github.com/franciscozamorau/osmi-server/internal/infrastructure/repositories/postgres/helpers/scanner"
	"github.com/franciscozamorau/osmi-server/internal/infrastructure/repositories/postgres/helpers/types"
	"github.com/franciscozamorau/osmi-server/internal/infrastructure/repositories/postgres/helpers/utils"
	"github.com/franciscozamorau/osmi-server/internal/infrastructure/repositories/postgres/helpers/validations"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// customerRepository implementa repository.CustomerRepository usando helpers
type customerRepository struct {
	db         *pgxpool.Pool
	converter  *types.Converter
	scanner    *scanner.RowScanner
	errHandler *errors.PostgresErrorHandler
	validator  *errors.Validator
	logger     *utils.Logger
}

// NewCustomerRepository crea una nueva instancia con helpers
func NewCustomerRepository(db *pgxpool.Pool) repository.CustomerRepository {
	return &customerRepository{
		db:         db,
		converter:  types.NewConverter(),
		scanner:    scanner.NewRowScanner(),
		errHandler: errors.NewPostgresErrorHandler(),
		validator:  errors.NewValidator(),
		logger:     utils.NewLogger("customer-repository"),
	}
}

// Create implementa repository.CustomerRepository.Create usando helpers
func (r *customerRepository) Create(ctx context.Context, customer *entities.Customer) error {
	startTime := time.Now()

	// Validaciones usando helpers
	if err := r.validateCustomerForCreate(ctx, customer); err != nil {
		return err
	}

	// Validar email usando value object
	emailVO, err := valueobjects.NewEmail(customer.Email)
	if err != nil {
		return fmt.Errorf("invalid email: %w", err)
	}

	// Validar phone si está presente usando validations
	var phonePtr *string
	if customer.Phone != nil && *customer.Phone != "" {
		if !validations.IsValidPhone(*customer.Phone) {
			return fmt.Errorf("invalid phone: %s", *customer.Phone)
		}
		phonePtr = customer.Phone
	}

	// Generar public_uuid si no existe
	if customer.PublicID == "" {
		customer.PublicID = uuid.New().String()
	}

	// Validar UUID usando validations
	if !validations.IsValidUUID(customer.PublicID) {
		return fmt.Errorf("invalid public_id: must be a valid UUID")
	}

	// Serializar communication_preferences usando helper
	commPrefsJSON, err := r.marshalJSON(customer.CommunicationPreferences, "{}")
	if err != nil {
		return fmt.Errorf("failed to marshal communication preferences: %w", err)
	}

	query := `
		INSERT INTO crm.customers (
			public_uuid, user_id, full_name, email, phone,
			company_name, address_line1, address_line2,
			city, state, postal_code, country,
			tax_id, tax_id_type, tax_name, requires_invoice,
			communication_preferences,
			total_spent, total_orders, total_tickets, avg_order_value,
			first_order_at, last_order_at, last_purchase_at,
			is_active, is_vip, vip_since,
			customer_segment, lifetime_value
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12,
			$13, $14, $15, $16, $17, $18, $19, $20, $21, $22,
			$23, $24, $25, $26, $27, $28, $29
		)
		RETURNING id, created_at, updated_at
	`

	err = r.db.QueryRow(ctx, query,
		customer.PublicID,
		r.converter.Int64Ptr(customer.UserID),
		r.converter.Text(customer.FullName),
		emailVO.String(),
		r.converter.TextPtr(phonePtr),
		r.converter.TextPtr(customer.CompanyName),
		r.converter.TextPtr(customer.AddressLine1),
		r.converter.TextPtr(customer.AddressLine2),
		r.converter.TextPtr(customer.City),
		r.converter.TextPtr(customer.State),
		r.converter.TextPtr(customer.PostalCode),
		r.converter.TextPtr(customer.Country),
		r.converter.TextPtr(customer.TaxID),
		r.converter.TextPtr(customer.TaxIDType),
		r.converter.TextPtr(customer.TaxName),
		r.converter.BoolPtr(customer.RequiresInvoice),
		commPrefsJSON,
		r.converter.Float64Ptr(customer.TotalSpent),
		r.converter.Int32Ptr(customer.TotalOrders),
		r.converter.Int32Ptr(customer.TotalTickets),
		r.converter.Float64Ptr(customer.AvgOrderValue),
		r.converter.TimestampPtr(customer.FirstOrderAt),
		r.converter.TimestampPtr(customer.LastOrderAt),
		r.converter.TimestampPtr(customer.LastPurchaseAt),
		r.converter.BoolPtr(customer.IsActive),
		r.converter.BoolPtr(customer.IsVIP),
		r.converter.TimestampPtr(customer.VIPSince),
		r.converter.TextPtr(customer.CustomerSegment),
		r.converter.Float64Ptr(customer.LifetimeValue),
	).Scan(&customer.ID, &customer.CreatedAt, &customer.UpdatedAt)

	if err != nil {
		// Usar error handler para manejo consistente
		if r.errHandler.IsDuplicateKey(err) {
			constraint := r.errHandler.GetConstraintName(err)
			value := r.errHandler.GetDuplicateValue(err)

			if strings.Contains(strings.ToLower(constraint), "email") {
				return fmt.Errorf("email already exists: %s", customer.Email)
			} else if strings.Contains(strings.ToLower(constraint), "public_uuid") {
				return fmt.Errorf("public_uuid already exists: %s", customer.PublicID)
			} else {
				return r.errHandler.CreateUserFriendlyError(err, "customer")
			}
		}

		r.logger.DatabaseLogger("INSERT", "crm.customers", time.Since(startTime), 0, err, map[string]interface{}{
			"email":     utils.SafeEmailForLog(customer.Email),
			"full_name": utils.SafeStringForLog(customer.FullName),
			"public_id": customer.PublicID,
		})

		return r.errHandler.WrapError(err, "customer repository", "create customer")
	}

	r.logger.DatabaseLogger("INSERT", "crm.customers", time.Since(startTime), 1, nil, map[string]interface{}{
		"customer_id": customer.ID,
		"email":       utils.SafeEmailForLog(customer.Email),
	})

	return nil
}

// FindByID implementa repository.CustomerRepository.FindByID usando scanner
func (r *customerRepository) FindByID(ctx context.Context, id int64) (*entities.Customer, error) {
	startTime := time.Now()

	query := `
		SELECT 
			id, public_uuid, user_id, full_name, email, phone,
			company_name, address_line1, address_line2,
			city, state, postal_code, country,
			tax_id, tax_id_type, tax_name, requires_invoice,
			communication_preferences,
			total_spent, total_orders, total_tickets, avg_order_value,
			first_order_at, last_order_at, last_purchase_at,
			is_active, is_vip, vip_since,
			customer_segment, lifetime_value,
			created_at, updated_at
		FROM crm.customers
		WHERE id = $1 AND is_active = true
	`

	row := r.db.QueryRow(ctx, query, id)
	customer, err := r.scanCustomer(row)

	if err != nil {
		if err.Error() == "customer not found" {
			r.logger.Debug("Customer not found", map[string]interface{}{
				"customer_id": id,
			})
			return nil, err
		}

		r.logger.DatabaseLogger("SELECT", "crm.customers", time.Since(startTime), 0, err, map[string]interface{}{
			"customer_id": id,
		})

		return nil, r.errHandler.WrapError(err, "customer repository", "find customer by ID")
	}

	r.logger.DatabaseLogger("SELECT", "crm.customers", time.Since(startTime), 1, nil, map[string]interface{}{
		"customer_id": id,
	})

	return customer, nil
}

// FindByPublicID implementa repository.CustomerRepository.FindByPublicID
func (r *customerRepository) FindByPublicID(ctx context.Context, publicID string) (*entities.Customer, error) {
	startTime := time.Now()

	// Validar UUID usando helpers
	if !validations.IsValidUUID(publicID) {
		return nil, fmt.Errorf("invalid public_id: must be a valid UUID")
	}

	query := `
		SELECT 
			id, public_uuid, user_id, full_name, email, phone,
			company_name, address_line1, address_line2,
			city, state, postal_code, country,
			tax_id, tax_id_type, tax_name, requires_invoice,
			communication_preferences,
			total_spent, total_orders, total_tickets, avg_order_value,
			first_order_at, last_order_at, last_purchase_at,
			is_active, is_vip, vip_since,
			customer_segment, lifetime_value,
			created_at, updated_at
		FROM crm.customers
		WHERE public_uuid = $1 AND is_active = true
	`

	row := r.db.QueryRow(ctx, query, publicID)
	customer, err := r.scanCustomer(row)

	if err != nil {
		if err.Error() == "customer not found" {
			r.logger.Debug("Customer not found by public ID", map[string]interface{}{
				"public_id": publicID,
			})
			return nil, err
		}

		r.logger.DatabaseLogger("SELECT", "crm.customers", time.Since(startTime), 0, err, map[string]interface{}{
			"public_id": publicID,
		})

		return nil, r.errHandler.WrapError(err, "customer repository", "find customer by public ID")
	}

	r.logger.DatabaseLogger("SELECT", "crm.customers", time.Since(startTime), 1, nil, map[string]interface{}{
		"customer_id": customer.ID,
		"public_id":   publicID,
	})

	return customer, nil
}

// FindByEmail implementa repository.CustomerRepository.FindByEmail
func (r *customerRepository) FindByEmail(ctx context.Context, email string) (*entities.Customer, error) {
	startTime := time.Now()

	// Validar email usando helpers
	if !validations.IsValidEmail(email) {
		return nil, fmt.Errorf("invalid email: %s", email)
	}

	query := `
		SELECT 
			id, public_uuid, user_id, full_name, email, phone,
			company_name, address_line1, address_line2,
			city, state, postal_code, country,
			tax_id, tax_id_type, tax_name, requires_invoice,
			communication_preferences,
			total_spent, total_orders, total_tickets, avg_order_value,
			first_order_at, last_order_at, last_purchase_at,
			is_active, is_vip, vip_since,
			customer_segment, lifetime_value,
			created_at, updated_at
		FROM crm.customers
		WHERE email = $1 AND is_active = true
	`

	row := r.db.QueryRow(ctx, query, email)
	customer, err := r.scanCustomer(row)

	if err != nil {
		if err.Error() == "customer not found" {
			r.logger.Debug("Customer not found by email", map[string]interface{}{
				"email": utils.SafeEmailForLog(email),
			})
			return nil, err
		}

		r.logger.DatabaseLogger("SELECT", "crm.customers", time.Since(startTime), 0, err, map[string]interface{}{
			"email": utils.SafeEmailForLog(email),
		})

		return nil, r.errHandler.WrapError(err, "customer repository", "find customer by email")
	}

	r.logger.DatabaseLogger("SELECT", "crm.customers", time.Since(startTime), 1, nil, map[string]interface{}{
		"customer_id": customer.ID,
		"email":       utils.SafeEmailForLog(email),
	})

	return customer, nil
}

// FindByUserID implementa repository.CustomerRepository.FindByUserID
func (r *customerRepository) FindByUserID(ctx context.Context, userID int64) (*entities.Customer, error) {
	startTime := time.Now()

	query := `
		SELECT 
			id, public_uuid, user_id, full_name, email, phone,
			company_name, address_line1, address_line2,
			city, state, postal_code, country,
			tax_id, tax_id_type, tax_name, requires_invoice,
			communication_preferences,
			total_spent, total_orders, total_tickets, avg_order_value,
			first_order_at, last_order_at, last_purchase_at,
			is_active, is_vip, vip_since,
			customer_segment, lifetime_value,
			created_at, updated_at
		FROM crm.customers
		WHERE user_id = $1 AND is_active = true
	`

	row := r.db.QueryRow(ctx, query, userID)
	customer, err := r.scanCustomer(row)

	if err != nil {
		if err.Error() == "customer not found" {
			r.logger.Debug("Customer not found by user ID", map[string]interface{}{
				"user_id": userID,
			})
			return nil, err
		}

		r.logger.DatabaseLogger("SELECT", "crm.customers", time.Since(startTime), 0, err, map[string]interface{}{
			"user_id": userID,
		})

		return nil, r.errHandler.WrapError(err, "customer repository", "find customer by user ID")
	}

	r.logger.DatabaseLogger("SELECT", "crm.customers", time.Since(startTime), 1, nil, map[string]interface{}{
		"customer_id": customer.ID,
		"user_id":     userID,
	})

	return customer, nil
}

// Update implementa repository.CustomerRepository.Update usando helpers
func (r *customerRepository) Update(ctx context.Context, customer *entities.Customer) error {
	startTime := time.Now()

	// Validaciones
	if err := r.validateCustomerForUpdate(ctx, customer); err != nil {
		return err
	}

	// Validar email usando validations
	if !validations.IsValidEmail(customer.Email) {
		return fmt.Errorf("invalid email: %s", customer.Email)
	}

	// Validar phone si está presente
	var phonePtr *string
	if customer.Phone != nil && *customer.Phone != "" {
		if !validations.IsValidPhone(*customer.Phone) {
			return fmt.Errorf("invalid phone: %s", *customer.Phone)
		}
		phonePtr = customer.Phone
	}

	// Serializar communication_preferences usando helper
	commPrefsJSON, err := r.marshalJSON(customer.CommunicationPreferences, "{}")
	if err != nil {
		return fmt.Errorf("failed to marshal communication preferences: %w", err)
	}

	query := `
		UPDATE crm.customers SET
			user_id = $1,
			full_name = $2,
			email = $3,
			phone = $4,
			company_name = $5,
			address_line1 = $6,
			address_line2 = $7,
			city = $8,
			state = $9,
			postal_code = $10,
			country = $11,
			tax_id = $12,
			tax_id_type = $13,
			tax_name = $14,
			requires_invoice = $15,
			communication_preferences = $16,
			is_active = $17,
			is_vip = $18,
			vip_since = $19,
			customer_segment = $20,
			lifetime_value = $21,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $22
		RETURNING updated_at
	`

	err = r.db.QueryRow(ctx, query,
		r.converter.Int64Ptr(customer.UserID),
		r.converter.Text(customer.FullName),
		customer.Email,
		r.converter.TextPtr(phonePtr),
		r.converter.TextPtr(customer.CompanyName),
		r.converter.TextPtr(customer.AddressLine1),
		r.converter.TextPtr(customer.AddressLine2),
		r.converter.TextPtr(customer.City),
		r.converter.TextPtr(customer.State),
		r.converter.TextPtr(customer.PostalCode),
		r.converter.TextPtr(customer.Country),
		r.converter.TextPtr(customer.TaxID),
		r.converter.TextPtr(customer.TaxIDType),
		r.converter.TextPtr(customer.TaxName),
		r.converter.BoolPtr(customer.RequiresInvoice),
		commPrefsJSON,
		r.converter.BoolPtr(customer.IsActive),
		r.converter.BoolPtr(customer.IsVIP),
		r.converter.TimestampPtr(customer.VIPSince),
		r.converter.TextPtr(customer.CustomerSegment),
		r.converter.Float64Ptr(customer.LifetimeValue),
		customer.ID,
	).Scan(&customer.UpdatedAt)

	if err != nil {
		// Usar error handler
		if r.errHandler.IsDuplicateKey(err) {
			constraint := r.errHandler.GetConstraintName(err)
			if strings.Contains(strings.ToLower(constraint), "email") {
				return fmt.Errorf("email already exists: %s", customer.Email)
			}
		}

		r.logger.DatabaseLogger("UPDATE", "crm.customers", time.Since(startTime), 0, err, map[string]interface{}{
			"customer_id": customer.ID,
			"email":       utils.SafeEmailForLog(customer.Email),
		})

		return r.errHandler.WrapError(err, "customer repository", "update customer")
	}

	r.logger.DatabaseLogger("UPDATE", "crm.customers", time.Since(startTime), 1, nil, map[string]interface{}{
		"customer_id": customer.ID,
	})

	return nil
}

// Delete implementa repository.CustomerRepository.Delete
func (r *customerRepository) Delete(ctx context.Context, id int64) error {
	startTime := time.Now()

	query := `DELETE FROM crm.customers WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		r.logger.DatabaseLogger("DELETE", "crm.customers", time.Since(startTime), 0, err, map[string]interface{}{
			"customer_id": id,
		})

		return r.errHandler.WrapError(err, "customer repository", "delete customer")
	}

	rowsAffected := result.RowsAffected()

	if rowsAffected == 0 {
		r.logger.Debug("Customer not found for deletion", map[string]interface{}{
			"customer_id": id,
		})
		return fmt.Errorf("customer not found")
	}

	r.logger.DatabaseLogger("DELETE", "crm.customers", time.Since(startTime), rowsAffected, nil, map[string]interface{}{
		"customer_id": id,
	})

	return nil
}

// List implementa repository.CustomerRepository.List usando query builder
func (r *customerRepository) List(ctx context.Context, filter dto.CustomerFilter, pagination dto.Pagination) ([]*entities.Customer, int64, error) {
	startTime := time.Now()

	// Usar query builder para construir la query
	qb := query.NewQueryBuilder(`
		SELECT 
			id, public_uuid, user_id, full_name, email, phone,
			company_name, address_line1, address_line2,
			city, state, postal_code, country,
			tax_id, tax_id_type, tax_name, requires_invoice,
			communication_preferences,
			total_spent, total_orders, total_tickets, avg_order_value,
			first_order_at, last_order_at, last_purchase_at,
			is_active, is_vip, vip_since,
			customer_segment, lifetime_value,
			created_at, updated_at
		FROM crm.customers
	`).Where("1=1", nil) // Condición inicial

	// Aplicar filtros
	if filter.IsActive != nil {
		qb.Where("is_active = ?", *filter.IsActive)
	}

	if filter.IsVIP != nil {
		qb.Where("is_vip = ?", *filter.IsVIP)
	}

	if filter.Country != "" {
		qb.Where("country = ?", filter.Country)
	}

	if filter.CustomerSegment != "" {
		qb.Where("customer_segment = ?", filter.CustomerSegment)
	}

	if filter.Search != "" {
		qb.Where("(email ILIKE ? OR full_name ILIKE ? OR company_name ILIKE ?)",
			"%"+filter.Search+"%", "%"+filter.Search+"%", "%"+filter.Search+"%")
	}

	if filter.DateFrom != "" {
		if dateFrom, err := utils.ParseDateFromString(filter.DateFrom); err == nil {
			qb.Where("created_at >= ?", dateFrom)
		}
	}

	if filter.DateTo != "" {
		if dateTo, err := utils.ParseDateFromString(filter.DateTo); err == nil {
			qb.Where("created_at <= ?", dateTo)
		}
	}

	// Ordenar
	qb.OrderBy("created_at", true) // DESC

	// Construir query de conteo
	countQuery, countArgs := qb.BuildCount()

	// Ejecutar count
	var total int64
	err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		r.logger.DatabaseLogger("SELECT", "crm.customers", time.Since(startTime), 0, err, map[string]interface{}{
			"operation": "count",
		})

		return nil, 0, r.errHandler.WrapError(err, "customer repository", "count customers")
	}

	// Aplicar paginación
	limit := pagination.PageSize
	if limit <= 0 {
		limit = 50
	}
	offset := (pagination.Page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	qb.Limit(limit).Offset(offset)

	// Construir query principal
	queryStr, args := qb.Build()

	// Ejecutar query principal
	rows, err := r.db.Query(ctx, queryStr, args...)
	if err != nil {
		r.logger.DatabaseLogger("SELECT", "crm.customers", time.Since(startTime), 0, err, map[string]interface{}{
			"operation": "list",
		})

		return nil, 0, r.errHandler.WrapError(err, "customer repository", "list customers")
	}
	defer rows.Close()

	// Usar scanner para procesar resultados
	customers := []*entities.Customer{}
	for rows.Next() {
		customer, err := r.scanCustomer(rows)
		if err != nil {
			r.logger.Error("Failed to scan customer row", err, map[string]interface{}{
				"operation": "list",
			})
			return nil, 0, fmt.Errorf("failed to scan customer: %w", err)
		}
		customers = append(customers, customer)
	}

	if err := rows.Err(); err != nil {
		r.logger.DatabaseLogger("SELECT", "crm.customers", time.Since(startTime), 0, err, map[string]interface{}{
			"operation": "list",
		})

		return nil, 0, r.errHandler.WrapError(err, "customer repository", "iterate customers")
	}

	r.logger.DatabaseLogger("SELECT", "crm.customers", time.Since(startTime), int64(len(customers)), nil, map[string]interface{}{
		"total":     total,
		"page":      pagination.Page,
		"page_size": pagination.PageSize,
		"found":     len(customers),
	})

	return customers, total, nil
}

// FindByName implementa repository.CustomerRepository.FindByName
func (r *customerRepository) FindByName(ctx context.Context, name string, limit int) ([]*entities.Customer, error) {
	startTime := time.Now()

	if limit <= 0 {
		limit = 20
	}

	// Usar query builder
	qb := query.NewQueryBuilder(`
		SELECT 
			id, public_uuid, user_id, full_name, email, phone,
			company_name, address_line1, address_line2,
			city, state, postal_code, country,
			tax_id, tax_id_type, tax_name, requires_invoice,
			communication_preferences,
			total_spent, total_orders, total_tickets, avg_order_value,
			first_order_at, last_order_at, last_purchase_at,
			is_active, is_vip, vip_since,
			customer_segment, lifetime_value,
			created_at, updated_at
		FROM crm.customers
	`).Where("is_active = true", nil).
		Where("full_name ILIKE ?", "%"+name+"%").
		OrderBy("created_at", true).
		Limit(limit)

	queryStr, args := qb.Build()

	rows, err := r.db.Query(ctx, queryStr, args...)
	if err != nil {
		r.logger.DatabaseLogger("SELECT", "crm.customers", time.Since(startTime), 0, err, map[string]interface{}{
			"operation": "find_by_name",
			"name":      utils.SafeStringForLog(name),
			"limit":     limit,
		})

		return nil, r.errHandler.WrapError(err, "customer repository", "find customers by name")
	}
	defer rows.Close()

	customers := []*entities.Customer{}
	for rows.Next() {
		customer, err := r.scanCustomer(rows)
		if err != nil {
			r.logger.Error("Failed to scan customer row", err, map[string]interface{}{
				"operation": "find_by_name",
			})
			return nil, fmt.Errorf("failed to scan customer: %w", err)
		}
		customers = append(customers, customer)
	}

	if err := rows.Err(); err != nil {
		r.logger.DatabaseLogger("SELECT", "crm.customers", time.Since(startTime), 0, err, map[string]interface{}{
			"operation": "find_by_name",
		})

		return nil, r.errHandler.WrapError(err, "customer repository", "iterate customers by name")
	}

	r.logger.DatabaseLogger("SELECT", "crm.customers", time.Since(startTime), int64(len(customers)), nil, map[string]interface{}{
		"name":  utils.SafeStringForLog(name),
		"limit": limit,
		"found": len(customers),
	})

	return customers, nil
}

// FindByType implementa repository.CustomerRepository.FindByType
func (r *customerRepository) FindByType(ctx context.Context, customerType string, pagination dto.Pagination) ([]*entities.Customer, int64, error) {
	filter := dto.CustomerFilter{CustomerSegment: customerType}
	return r.List(ctx, filter, pagination)
}

// FindByCountry implementa repository.CustomerRepository.FindByCountry
func (r *customerRepository) FindByCountry(ctx context.Context, country string, pagination dto.Pagination) ([]*entities.Customer, int64, error) {
	filter := dto.CustomerFilter{Country: country}
	return r.List(ctx, filter, pagination)
}

// Search implementa repository.CustomerRepository.Search usando query builder
func (r *customerRepository) Search(ctx context.Context, term string, limit int) ([]*entities.Customer, error) {
	startTime := time.Now()

	if limit <= 0 {
		limit = 20
	}

	// Usar query builder
	qb := query.NewQueryBuilder(`
		SELECT 
			id, public_uuid, user_id, full_name, email, phone,
			company_name, address_line1, address_line2,
			city, state, postal_code, country,
			tax_id, tax_id_type, tax_name, requires_invoice,
			communication_preferences,
			total_spent, total_orders, total_tickets, avg_order_value,
			first_order_at, last_order_at, last_purchase_at,
			is_active, is_vip, vip_since,
			customer_segment, lifetime_value,
			created_at, updated_at
		FROM crm.customers
	`).Where("is_active = true", nil)

	if term != "" {
		qb.Where("(email ILIKE ? OR full_name ILIKE ? OR company_name ILIKE ? OR tax_id ILIKE ?)",
			"%"+term+"%", "%"+term+"%", "%"+term+"%", "%"+term+"%")
	}

	// Ordenar por relevancia
	orderBy := `
		CASE 
			WHEN email ILIKE ? THEN 1
			WHEN full_name ILIKE ? THEN 2
			WHEN company_name ILIKE ? THEN 3
			ELSE 4
		END,
		created_at DESC
	`
	qb.OrderByRaw(orderBy)
	qb.Limit(limit)

	queryStr, args := qb.Build()

	rows, err := r.db.Query(ctx, queryStr, args...)
	if err != nil {
		r.logger.DatabaseLogger("SELECT", "crm.customers", time.Since(startTime), 0, err, map[string]interface{}{
			"operation": "search",
			"term":      utils.SafeStringForLog(term),
			"limit":     limit,
		})

		return nil, r.errHandler.WrapError(err, "customer repository", "search customers")
	}
	defer rows.Close()

	customers := []*entities.Customer{}
	for rows.Next() {
		customer, err := r.scanCustomer(rows)
		if err != nil {
			r.logger.Error("Failed to scan customer row during search", err, map[string]interface{}{
				"operation": "search",
			})
			return nil, fmt.Errorf("failed to scan customer: %w", err)
		}
		customers = append(customers, customer)
	}

	if err := rows.Err(); err != nil {
		r.logger.DatabaseLogger("SELECT", "crm.customers", time.Since(startTime), 0, err, map[string]interface{}{
			"operation": "search",
		})

		return nil, r.errHandler.WrapError(err, "customer repository", "iterate search results")
	}

	r.logger.DatabaseLogger("SELECT", "crm.customers", time.Since(startTime), int64(len(customers)), nil, map[string]interface{}{
		"term":  utils.SafeStringForLog(term),
		"limit": limit,
		"found": len(customers),
	})

	return customers, nil
}

// UpdateStats implementa repository.CustomerRepository.UpdateStats
func (r *customerRepository) UpdateStats(ctx context.Context, customerID int64, amount float64) error {
	startTime := time.Now()

	query := `
		UPDATE crm.customers 
		SET total_spent = total_spent + $1,
		    total_orders = total_orders + 1,
		    total_tickets = total_tickets + 1,
		    last_purchase_at = CURRENT_TIMESTAMP,
		    last_order_at = CURRENT_TIMESTAMP,
		    avg_order_value = total_spent / NULLIF(total_orders, 0),
		    lifetime_value = total_spent,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`

	result, err := r.db.Exec(ctx, query, amount, customerID)
	if err != nil {
		r.logger.DatabaseLogger("UPDATE", "crm.customers", time.Since(startTime), 0, err, map[string]interface{}{
			"customer_id": customerID,
			"amount":      amount,
			"operation":   "update_stats",
		})

		return r.errHandler.WrapError(err, "customer repository", "update customer stats")
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		r.logger.Debug("Customer not found for stats update", map[string]interface{}{
			"customer_id": customerID,
		})
		return fmt.Errorf("customer not found")
	}

	r.logger.DatabaseLogger("UPDATE", "crm.customers", time.Since(startTime), rowsAffected, nil, map[string]interface{}{
		"customer_id": customerID,
		"amount":      amount,
	})

	return nil
}

// UpdateLoyaltyPoints implementa repository.CustomerRepository.UpdateLoyaltyPoints
func (r *customerRepository) UpdateLoyaltyPoints(ctx context.Context, customerID int64, points int32) error {
	// Note: No hay campo loyalty_points en la tabla crm.customers
	// Por ahora, no implementado
	r.logger.Warn("UpdateLoyaltyPoints not implemented", map[string]interface{}{
		"customer_id": customerID,
		"points":      points,
	})
	return errors.New("not implemented: loyalty points not supported in current schema")
}

// UpdateVerification implementa repository.CustomerRepository.UpdateVerification
func (r *customerRepository) UpdateVerification(ctx context.Context, customerID int64, verified bool) error {
	// Note: No hay campo verification en la tabla crm.customers
	// Por ahora, no implementado
	r.logger.Warn("UpdateVerification not implemented", map[string]interface{}{
		"customer_id": customerID,
		"verified":    verified,
	})
	return errors.New("not implemented: verification not supported in current schema")
}

// UpdatePreferences implementa repository.CustomerRepository.UpdatePreferences
func (r *customerRepository) UpdatePreferences(ctx context.Context, customerID int64, preferences map[string]interface{}) error {
	startTime := time.Now()

	// Serializar preferences usando helper
	prefsJSON, err := r.marshalJSON(preferences, "{}")
	if err != nil {
		return fmt.Errorf("failed to marshal preferences: %w", err)
	}

	query := `
		UPDATE crm.customers 
		SET communication_preferences = $1,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`

	result, err := r.db.Exec(ctx, query, prefsJSON, customerID)
	if err != nil {
		r.logger.DatabaseLogger("UPDATE", "crm.customers", time.Since(startTime), 0, err, map[string]interface{}{
			"customer_id": customerID,
			"operation":   "update_preferences",
		})

		return r.errHandler.WrapError(err, "customer repository", "update customer preferences")
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		r.logger.Debug("Customer not found for preferences update", map[string]interface{}{
			"customer_id": customerID,
		})
		return fmt.Errorf("customer not found")
	}

	r.logger.DatabaseLogger("UPDATE", "crm.customers", time.Since(startTime), rowsAffected, nil, map[string]interface{}{
		"customer_id": customerID,
	})

	return nil
}

// SetVIP implementa repository.CustomerRepository.SetVIP
func (r *customerRepository) SetVIP(ctx context.Context, customerID int64, isVIP bool) error {
	startTime := time.Now()

	query := `
		UPDATE crm.customers 
		SET is_vip = $1,
		    vip_since = CASE WHEN $1 = true AND vip_since IS NULL THEN CURRENT_TIMESTAMP ELSE vip_since END,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`

	result, err := r.db.Exec(ctx, query, isVIP, customerID)
	if err != nil {
		r.logger.DatabaseLogger("UPDATE", "crm.customers", time.Since(startTime), 0, err, map[string]interface{}{
			"customer_id": customerID,
			"is_vip":      isVIP,
			"operation":   "set_vip",
		})

		return r.errHandler.WrapError(err, "customer repository", "set VIP status")
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		r.logger.Debug("Customer not found for VIP status update", map[string]interface{}{
			"customer_id": customerID,
		})
		return fmt.Errorf("customer not found")
	}

	r.logger.DatabaseLogger("UPDATE", "crm.customers", time.Since(startTime), rowsAffected, nil, map[string]interface{}{
		"customer_id": customerID,
		"is_vip":      isVIP,
	})

	return nil
}

// UpdateInvoiceSettings implementa repository.CustomerRepository.UpdateInvoiceSettings
func (r *customerRepository) UpdateInvoiceSettings(ctx context.Context, customerID int64, requiresInvoice bool, taxID, taxName string) error {
	startTime := time.Now()

	query := `
		UPDATE crm.customers 
		SET requires_invoice = $1,
		    tax_id = $2,
		    tax_name = $3,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $4
	`

	result, err := r.db.Exec(ctx, query, requiresInvoice, taxID, taxName, customerID)
	if err != nil {
		r.logger.DatabaseLogger("UPDATE", "crm.customers", time.Since(startTime), 0, err, map[string]interface{}{
			"customer_id":      customerID,
			"requires_invoice": requiresInvoice,
			"operation":        "update_invoice_settings",
		})

		return r.errHandler.WrapError(err, "customer repository", "update invoice settings")
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		r.logger.Debug("Customer not found for invoice settings update", map[string]interface{}{
			"customer_id": customerID,
		})
		return fmt.Errorf("customer not found")
	}

	r.logger.DatabaseLogger("UPDATE", "crm.customers", time.Since(startTime), rowsAffected, nil, map[string]interface{}{
		"customer_id":      customerID,
		"requires_invoice": requiresInvoice,
	})

	return nil
}

// GetStats implementa repository.CustomerRepository.GetStats
func (r *customerRepository) GetStats(ctx context.Context) (*dto.CustomerStatsResponse, error) {
	startTime := time.Now()

	query := `
		SELECT 
			COUNT(*) as total_customers,
			COUNT(CASE WHEN is_active = true THEN 1 END) as active_customers,
			COUNT(CASE WHEN is_vip = true THEN 1 END) as vip_customers,
			COUNT(CASE WHEN created_at >= CURRENT_DATE - INTERVAL '30 days' THEN 1 END) as new_last_30_days,
			SUM(total_spent) as total_revenue,
			AVG(lifetime_value) as avg_lifetime_value
		FROM crm.customers
	`

	var stats dto.CustomerStatsResponse
	err := r.db.QueryRow(ctx, query).Scan(
		&stats.TotalCustomers,
		&stats.ActiveCustomers,
		&stats.VIPCustomers,
		&stats.NewCustomersLast30Days,
		&stats.TotalRevenue,
		&stats.AvgLifetimeValue,
	)

	if err != nil {
		r.logger.DatabaseLogger("SELECT", "crm.customers", time.Since(startTime), 0, err, map[string]interface{}{
			"operation": "get_stats",
		})

		return nil, r.errHandler.WrapError(err, "customer repository", "get customer stats")
	}

	// Obtener estadísticas por país
	countryQuery := `
		SELECT country, COUNT(*) as count, SUM(total_spent) as revenue
		FROM crm.customers
		WHERE country IS NOT NULL
		GROUP BY country
		ORDER BY count DESC
		LIMIT 10
	`

	rows, err := r.db.Query(ctx, countryQuery)
	if err != nil {
		r.logger.Error("Failed to get country stats", err, nil)
		// Continuar sin estadísticas por país
		stats.TopCountries = []dto.CountryStats{}
	} else {
		defer rows.Close()

		topCountries := []dto.CountryStats{}
		for rows.Next() {
			var countryStat dto.CountryStats
			err := rows.Scan(&countryStat.Country, &countryStat.Count, &countryStat.Revenue)
			if err != nil {
				r.logger.Error("Failed to scan country stat", err, nil)
				continue
			}
			topCountries = append(topCountries, countryStat)
		}
		stats.TopCountries = topCountries
	}

	r.logger.DatabaseLogger("SELECT", "crm.customers", time.Since(startTime), 1, nil, map[string]interface{}{
		"operation": "get_stats",
	})

	return &stats, nil
}

// GetVIPCustomers implementa repository.CustomerRepository.GetVIPCustomers
func (r *customerRepository) GetVIPCustomers(ctx context.Context) ([]*entities.Customer, error) {
	startTime := time.Now()

	query := `
		SELECT 
			id, public_uuid, user_id, full_name, email, phone,
			company_name, address_line1, address_line2,
			city, state, postal_code, country,
			tax_id, tax_id_type, tax_name, requires_invoice,
			communication_preferences,
			total_spent, total_orders, total_tickets, avg_order_value,
			first_order_at, last_order_at, last_purchase_at,
			is_active, is_vip, vip_since,
			customer_segment, lifetime_value,
			created_at, updated_at
		FROM crm.customers
		WHERE is_vip = true AND is_active = true
		ORDER BY vip_since DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		r.logger.DatabaseLogger("SELECT", "crm.customers", time.Since(startTime), 0, err, map[string]interface{}{
			"operation": "get_vip_customers",
		})

		return nil, r.errHandler.WrapError(err, "customer repository", "get VIP customers")
	}
	defer rows.Close()

	customers := []*entities.Customer{}
	for rows.Next() {
		customer, err := r.scanCustomer(rows)
		if err != nil {
			r.logger.Error("Failed to scan customer row", err, map[string]interface{}{
				"operation": "get_vip_customers",
			})
			return nil, fmt.Errorf("failed to scan customer: %w", err)
		}
		customers = append(customers, customer)
	}

	if err := rows.Err(); err != nil {
		r.logger.DatabaseLogger("SELECT", "crm.customers", time.Since(startTime), 0, err, map[string]interface{}{
			"operation": "get_vip_customers",
		})

		return nil, r.errHandler.WrapError(err, "customer repository", "iterate VIP customers")
	}

	r.logger.DatabaseLogger("SELECT", "crm.customers", time.Since(startTime), int64(len(customers)), nil, map[string]interface{}{
		"operation": "get_vip_customers",
		"count":     len(customers),
	})

	return customers, nil
}

// CountByType implementa repository.CustomerRepository.CountByType
func (r *customerRepository) CountByType(ctx context.Context, customerType string) (int64, error) {
	startTime := time.Now()

	query := `SELECT COUNT(*) FROM crm.customers WHERE customer_segment = $1 AND is_active = true`

	var count int64
	err := r.db.QueryRow(ctx, query, customerType).Scan(&count)
	if err != nil {
		r.logger.DatabaseLogger("SELECT", "crm.customers", time.Since(startTime), 0, err, map[string]interface{}{
			"operation":     "count_by_type",
			"customer_type": customerType,
		})

		return 0, r.errHandler.WrapError(err, "customer repository", "count customers by type")
	}

	r.logger.DatabaseLogger("SELECT", "crm.customers", time.Since(startTime), 1, nil, map[string]interface{}{
		"operation":     "count_by_type",
		"customer_type": customerType,
		"count":         count,
	})

	return count, nil
}

// GetTotalSpent implementa repository.CustomerRepository.GetTotalSpent
func (r *customerRepository) GetTotalSpent(ctx context.Context, customerID int64) (float64, error) {
	startTime := time.Now()

	query := `SELECT total_spent FROM crm.customers WHERE id = $1`

	var totalSpent float64
	err := r.db.QueryRow(ctx, query, customerID).Scan(&totalSpent)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Debug("Customer not found for total spent", map[string]interface{}{
				"customer_id": customerID,
			})
			return 0, fmt.Errorf("customer not found")
		}

		r.logger.DatabaseLogger("SELECT", "crm.customers", time.Since(startTime), 0, err, map[string]interface{}{
			"customer_id": customerID,
			"operation":   "get_total_spent",
		})

		return 0, r.errHandler.WrapError(err, "customer repository", "get total spent")
	}

	r.logger.DatabaseLogger("SELECT", "crm.customers", time.Since(startTime), 1, nil, map[string]interface{}{
		"customer_id": customerID,
	})

	return totalSpent, nil
}

// GetPurchaseHistory implementa repository.CustomerRepository.GetPurchaseHistory
func (r *customerRepository) GetPurchaseHistory(ctx context.Context, customerID int64, limit int) ([]*dto.PurchaseRecord, error) {
	startTime := time.Now()

	if limit <= 0 {
		limit = 10
	}

	query := `
		SELECT 
			o.public_uuid as order_id,
			o.total_amount,
			o.currency,
			o.created_at as purchase_date,
			o.status,
			COUNT(oi.id) as items_count
		FROM billing.orders o
		LEFT JOIN billing.order_items oi ON o.id = oi.order_id
		WHERE o.customer_id = $1
		GROUP BY o.id, o.public_uuid, o.total_amount, o.currency, o.created_at, o.status
		ORDER BY o.created_at DESC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, customerID, limit)
	if err != nil {
		r.logger.DatabaseLogger("SELECT", "billing.orders", time.Since(startTime), 0, err, map[string]interface{}{
			"customer_id": customerID,
			"limit":       limit,
			"operation":   "get_purchase_history",
		})

		return nil, r.errHandler.WrapError(err, "customer repository", "get purchase history")
	}
	defer rows.Close()

	purchases := []*dto.PurchaseRecord{}
	for rows.Next() {
		var purchase dto.PurchaseRecord
		err := rows.Scan(
			&purchase.OrderID,
			&purchase.Amount,
			&purchase.Currency,
			&purchase.PurchaseDate,
			&purchase.Status,
			&purchase.ItemsCount,
		)
		if err != nil {
			r.logger.Error("Failed to scan purchase record", err, map[string]interface{}{
				"customer_id": customerID,
			})
			return nil, fmt.Errorf("failed to scan purchase record: %w", err)
		}
		purchases = append(purchases, &purchase)
	}

	if err := rows.Err(); err != nil {
		r.logger.DatabaseLogger("SELECT", "billing.orders", time.Since(startTime), 0, err, map[string]interface{}{
			"customer_id": customerID,
		})

		return nil, r.errHandler.WrapError(err, "customer repository", "iterate purchase records")
	}

	r.logger.DatabaseLogger("SELECT", "billing.orders", time.Since(startTime), int64(len(purchases)), nil, map[string]interface{}{
		"customer_id": customerID,
		"count":       len(purchases),
	})

	return purchases, nil
}

// =============================================================================
// MÉTODOS PRIVADOS CON HELPERS
// =============================================================================

// scanCustomer escanea una fila de cliente usando scanner
func (r *customerRepository) scanCustomer(row interface{}) (*entities.Customer, error) {
	var customer entities.Customer
	var commPrefsJSON []byte

	// Escanear los valores básicos usando scanner
	values, err := r.scanner.ScanRowToMap(row, []string{
		"id", "public_uuid", "user_id", "full_name", "email", "phone",
		"company_name", "address_line1", "address_line2",
		"city", "state", "postal_code", "country",
		"tax_id", "tax_id_type", "tax_name", "requires_invoice",
		"communication_preferences",
		"total_spent", "total_orders", "total_tickets", "avg_order_value",
		"first_order_at", "last_order_at", "last_purchase_at",
		"is_active", "is_vip", "vip_since",
		"customer_segment", "lifetime_value",
		"created_at", "updated_at",
	})

	if err != nil {
		if err.Error() == "customer not found" {
			return nil, err
		}
		return nil, fmt.Errorf("failed to scan customer row: %w", err)
	}

	// Mapear valores a la estructura customer
	// Esto es simplificado - en implementación real usarías reflection o mapping manual

	// Parsear communication_preferences
	if commPrefsJSONStr, ok := values["communication_preferences"].(string); ok && commPrefsJSONStr != "" {
		var commPrefs map[string]interface{}
		if err := json.Unmarshal([]byte(commPrefsJSONStr), &commPrefs); err != nil {
			r.logger.Error("Failed to unmarshal communication preferences", err)
			customer.CommunicationPreferences = make(map[string]interface{})
		} else {
			customer.CommunicationPreferences = commPrefs
		}
	} else {
		customer.CommunicationPreferences = make(map[string]interface{})
	}

	// Convertir valores usando converter donde sea necesario
	customer.ID = values["id"].(int64)
	customer.PublicID = values["public_uuid"].(string)
	customer.FullName = values["full_name"].(string)
	customer.Email = values["email"].(string)

	// Manejar valores nullable
	if phone, ok := values["phone"].(string); ok && phone != "" {
		customer.Phone = &phone
	}
	if userID, ok := values["user_id"].(int64); ok && userID > 0 {
		customer.UserID = &userID
	}

	return &customer, nil
}

// validateCustomerForCreate valida un cliente para creación
func (r *customerRepository) validateCustomerForCreate(ctx context.Context, customer *entities.Customer) error {
	// Usar validator
	r.validator.Required("email", customer.Email).
		Required("full_name", customer.FullName).
		Email("email", customer.Email).
		MaxLength("full_name", customer.FullName, 200).
		Custom("email", validations.IsValidEmail(customer.Email), "invalid email format")

	if customer.Phone != nil && *customer.Phone != "" {
		r.validator.Custom("phone", validations.IsValidPhone(*customer.Phone), "invalid phone format")
	}

	if validationErr := r.validator.Validate(); validationErr != nil {
		return validationErr
	}

	// Validar emails duplicados
	var exists bool
	err := r.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM crm.customers WHERE email = $1)", customer.Email).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check email uniqueness: %w", err)
	}
	if exists {
		return fmt.Errorf("email already exists: %s", customer.Email)
	}

	return nil
}

// validateCustomerForUpdate valida un cliente para actualización
func (r *customerRepository) validateCustomerForUpdate(ctx context.Context, customer *entities.Customer) error {
	// Validar que el cliente exista
	var exists bool
	err := r.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM crm.customers WHERE id = $1)", customer.ID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to validate customer existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("customer not found with ID: %d", customer.ID)
	}

	// Usar validator para validaciones generales
	return r.validateCustomerForCreate(ctx, customer)
}

// marshalJSON serializa JSON con valor por defecto usando helpers
func (r *customerRepository) marshalJSON(data interface{}, defaultValue string) ([]byte, error) {
	if data == nil {
		return []byte(defaultValue), nil
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	// Si está vacío, usar valor por defecto
	if len(jsonBytes) == 0 || string(jsonBytes) == "null" {
		return []byte(defaultValue), nil
	}

	return jsonBytes, nil
}
