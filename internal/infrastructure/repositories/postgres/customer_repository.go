package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
	"github.com/franciscozamorau/osmi-server/internal/domain/repository"
)

// CustomerRepository implementa la interfaz repository.CustomerRepository usando PostgreSQL
type CustomerRepository struct {
	db *sqlx.DB
}

// NewCustomerRepository crea una nueva instancia del repositorio
func NewCustomerRepository(db *sqlx.DB) *CustomerRepository {
	return &CustomerRepository{
		db: db,
	}
}

// handleError mapea errores de PostgreSQL a nuestros errores de dominio
func (r *CustomerRepository) handleError(err error, context string) error {
	if err == nil {
		return nil
	}

	if pqErr, ok := err.(*pq.Error); ok {
		switch pqErr.Code {
		case "23505": // Unique violation
			if strings.Contains(pqErr.Constraint, "customers_email_key") {
				return repository.ErrCustomerEmailExists
			}
			if strings.Contains(pqErr.Constraint, "customers_public_uuid_key") {
				return repository.ErrCustomerAlreadyLinked
			}
		case "23503": // Foreign key violation
			return fmt.Errorf("referenced user not found: %w", err)
		}
	}

	if errors.Is(err, sql.ErrNoRows) {
		return repository.ErrCustomerNotFound
	}

	return fmt.Errorf("%s: %w", context, err)
}

// Find busca clientes según los criterios del filtro
func (r *CustomerRepository) Find(ctx context.Context, filter *repository.CustomerFilter) ([]*entities.Customer, int64, error) {
	baseQuery := `
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
		WHERE 1=1
	`

	countQuery := `SELECT COUNT(*) FROM crm.customers WHERE 1=1`

	var conditions []string
	var args []interface{}
	argPos := 1

	if filter != nil {
		// Filtros por ID
		if len(filter.IDs) > 0 {
			conditions = append(conditions, fmt.Sprintf("id = ANY($%d)", argPos))
			args = append(args, pq.Array(filter.IDs))
			argPos++
		}

		if len(filter.PublicIDs) > 0 {
			conditions = append(conditions, fmt.Sprintf("public_uuid = ANY($%d)", argPos))
			args = append(args, pq.Array(filter.PublicIDs))
			argPos++
		}

		if filter.UserID != nil {
			conditions = append(conditions, fmt.Sprintf("user_id = $%d", argPos))
			args = append(args, *filter.UserID)
			argPos++
		}

		if filter.Email != nil {
			conditions = append(conditions, fmt.Sprintf("email = $%d", argPos))
			args = append(args, *filter.Email)
			argPos++
		}

		// Filtros de texto
		if filter.SearchTerm != nil && *filter.SearchTerm != "" {
			searchTerm := "%" + *filter.SearchTerm + "%"
			conditions = append(conditions, fmt.Sprintf(
				"(full_name ILIKE $%d OR email ILIKE $%d OR company_name ILIKE $%d OR tax_id ILIKE $%d)",
				argPos, argPos, argPos, argPos,
			))
			args = append(args, searchTerm, searchTerm, searchTerm, searchTerm)
			argPos += 4
		}

		if filter.FullName != nil {
			conditions = append(conditions, fmt.Sprintf("full_name ILIKE $%d", argPos))
			args = append(args, "%"+*filter.FullName+"%")
			argPos++
		}

		if filter.CompanyName != nil {
			conditions = append(conditions, fmt.Sprintf("company_name ILIKE $%d", argPos))
			args = append(args, "%"+*filter.CompanyName+"%")
			argPos++
		}

		if filter.Country != nil {
			conditions = append(conditions, fmt.Sprintf("country = $%d", argPos))
			args = append(args, *filter.Country)
			argPos++
		}

		if filter.City != nil {
			conditions = append(conditions, fmt.Sprintf("city ILIKE $%d", argPos))
			args = append(args, "%"+*filter.City+"%")
			argPos++
		}

		// Filtros booleanos
		if filter.IsActive != nil {
			conditions = append(conditions, fmt.Sprintf("is_active = $%d", argPos))
			args = append(args, *filter.IsActive)
			argPos++
		}

		if filter.IsVIP != nil {
			conditions = append(conditions, fmt.Sprintf("is_vip = $%d", argPos))
			args = append(args, *filter.IsVIP)
			argPos++
		}

		if filter.RequiresInvoice != nil {
			conditions = append(conditions, fmt.Sprintf("requires_invoice = $%d", argPos))
			args = append(args, *filter.RequiresInvoice)
			argPos++
		}

		if filter.CustomerSegment != nil {
			conditions = append(conditions, fmt.Sprintf("customer_segment = $%d", argPos))
			args = append(args, *filter.CustomerSegment)
			argPos++
		}

		// Filtros de fechas
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

		if filter.LastPurchaseFrom != nil {
			conditions = append(conditions, fmt.Sprintf("last_purchase_at >= $%d", argPos))
			args = append(args, *filter.LastPurchaseFrom)
			argPos++
		}

		if filter.LastPurchaseTo != nil {
			conditions = append(conditions, fmt.Sprintf("last_purchase_at <= $%d", argPos))
			args = append(args, *filter.LastPurchaseTo)
			argPos++
		}

		// Filtros de estadísticas
		if filter.MinTotalSpent != nil {
			conditions = append(conditions, fmt.Sprintf("total_spent >= $%d", argPos))
			args = append(args, *filter.MinTotalSpent)
			argPos++
		}

		if filter.MaxTotalSpent != nil {
			conditions = append(conditions, fmt.Sprintf("total_spent <= $%d", argPos))
			args = append(args, *filter.MaxTotalSpent)
			argPos++
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
		return nil, 0, r.handleError(err, "failed to count customers")
	}

	// Añadir ordenamiento y paginación
	if filter != nil {
		sortBy := "created_at"
		sortOrder := "DESC"
		if filter.SortBy != "" {
			allowedSortColumns := map[string]bool{
				"created_at":       true,
				"total_spent":      true,
				"total_orders":     true,
				"last_purchase_at": true,
				"full_name":        true,
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
	var customers []*entities.Customer
	err = r.db.SelectContext(ctx, &customers, baseQuery, args...)
	if err != nil {
		return nil, 0, r.handleError(err, "failed to find customers")
	}

	return customers, total, nil
}

// GetByID obtiene un cliente por su ID numérico
func (r *CustomerRepository) GetByID(ctx context.Context, id int64) (*entities.Customer, error) {
	filter := &repository.CustomerFilter{
		IDs:   []int64{id},
		Limit: 1,
	}

	customers, _, err := r.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	if len(customers) == 0 {
		return nil, repository.ErrCustomerNotFound
	}

	return customers[0], nil
}

// GetByPublicID obtiene un cliente por su UUID público
func (r *CustomerRepository) GetByPublicID(ctx context.Context, publicID string) (*entities.Customer, error) {
	filter := &repository.CustomerFilter{
		PublicIDs: []string{publicID},
		Limit:     1,
	}

	customers, _, err := r.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	if len(customers) == 0 {
		return nil, repository.ErrCustomerNotFound
	}

	return customers[0], nil
}

// GetByEmail obtiene un cliente por su email
func (r *CustomerRepository) GetByEmail(ctx context.Context, email string) (*entities.Customer, error) {
	filter := &repository.CustomerFilter{
		Email: &email,
		Limit: 1,
	}

	customers, _, err := r.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	if len(customers) == 0 {
		return nil, repository.ErrCustomerNotFound
	}

	return customers[0], nil
}

// GetByUserID obtiene un cliente por su ID de usuario asociado
func (r *CustomerRepository) GetByUserID(ctx context.Context, userID int64) (*entities.Customer, error) {
	filter := &repository.CustomerFilter{
		UserID: &userID,
		Limit:  1,
	}

	customers, _, err := r.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	if len(customers) == 0 {
		return nil, repository.ErrCustomerNotFound
	}

	return customers[0], nil
}

// Create inserta un nuevo cliente
func (r *CustomerRepository) Create(ctx context.Context, customer *entities.Customer) error {
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
			customer_segment, lifetime_value,
			created_at, updated_at
		) VALUES (
			gen_random_uuid(), $1, $2, $3, $4,
			$5, $6, $7, $8, $9, $10, $11,
			$12, $13, $14, $15,
			$16,
			$17, $18, $19, $20,
			$21, $22, $23,
			$24, $25, $26,
			$27, $28,
			NOW(), NOW()
		)
		RETURNING id, public_uuid, created_at, updated_at
	`

	// Convertir preferencias a JSON
	prefsJSON, err := json.Marshal(customer.CommunicationPreferences)
	if err != nil {
		return fmt.Errorf("failed to marshal communication preferences: %w", err)
	}

	err = r.db.QueryRowContext(
		ctx, query,
		customer.UserID, customer.FullName, customer.Email, customer.Phone,
		customer.CompanyName, customer.AddressLine1, customer.AddressLine2,
		customer.City, customer.State, customer.PostalCode, customer.Country,
		customer.TaxID, customer.TaxIDType, customer.TaxName, customer.RequiresInvoice,
		prefsJSON,
		customer.TotalSpent, customer.TotalOrders, customer.TotalTickets, customer.AvgOrderValue,
		customer.FirstOrderAt, customer.LastOrderAt, customer.LastPurchaseAt,
		customer.IsActive, customer.IsVIP, customer.VIPSince,
		customer.CustomerSegment, customer.LifetimeValue,
	).Scan(&customer.ID, &customer.PublicID, &customer.CreatedAt, &customer.UpdatedAt)

	if err != nil {
		return r.handleError(err, "failed to create customer")
	}

	return nil
}

// Update actualiza un cliente existente
func (r *CustomerRepository) Update(ctx context.Context, customer *entities.Customer) error {
	exists, err := r.Exists(ctx, customer.ID)
	if err != nil {
		return err
	}
	if !exists {
		return repository.ErrCustomerNotFound
	}

	prefsJSON, err := json.Marshal(customer.CommunicationPreferences)
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
			updated_at = NOW()
		WHERE id = $22
		RETURNING updated_at
	`

	err = r.db.QueryRowContext(
		ctx, query,
		customer.UserID, customer.FullName, customer.Email, customer.Phone,
		customer.CompanyName, customer.AddressLine1, customer.AddressLine2,
		customer.City, customer.State, customer.PostalCode, customer.Country,
		customer.TaxID, customer.TaxIDType, customer.TaxName, customer.RequiresInvoice,
		prefsJSON,
		customer.IsActive, customer.IsVIP, customer.VIPSince,
		customer.CustomerSegment, customer.LifetimeValue,
		customer.ID,
	).Scan(&customer.UpdatedAt)

	if err != nil {
		return r.handleError(err, "failed to update customer")
	}

	return nil
}

// Delete elimina permanentemente un cliente
func (r *CustomerRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM crm.customers WHERE id = $1`, id)
	if err != nil {
		return r.handleError(err, "failed to delete customer")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return repository.ErrCustomerNotFound
	}

	return nil
}

// SoftDelete desactiva un cliente (soft delete)
func (r *CustomerRepository) SoftDelete(ctx context.Context, publicID string) error {
	query := `
		UPDATE crm.customers 
		SET is_active = false, updated_at = NOW()
		WHERE public_uuid = $1 AND is_active = true
	`
	result, err := r.db.ExecContext(ctx, query, publicID)
	if err != nil {
		return r.handleError(err, "failed to soft delete customer")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return repository.ErrCustomerNotFound
	}

	return nil
}

// Exists verifica si existe un cliente con el ID dado
func (r *CustomerRepository) Exists(ctx context.Context, id int64) (bool, error) {
	var exists bool
	err := r.db.GetContext(ctx, &exists, `SELECT EXISTS(SELECT 1 FROM crm.customers WHERE id = $1)`, id)
	if err != nil {
		return false, r.handleError(err, "failed to check customer existence")
	}
	return exists, nil
}

// ExistsByEmail verifica si existe un cliente con el email dado
func (r *CustomerRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.db.GetContext(ctx, &exists, `SELECT EXISTS(SELECT 1 FROM crm.customers WHERE email = $1)`, email)
	if err != nil {
		return false, r.handleError(err, "failed to check email existence")
	}
	return exists, nil
}

// UpdateStats actualiza las estadísticas del cliente después de una compra
func (r *CustomerRepository) UpdateStats(ctx context.Context, customerID int64, amount float64) error {
	query := `
		UPDATE crm.customers 
		SET total_spent = total_spent + $1,
			total_orders = total_orders + 1,
			total_tickets = total_tickets + 1,
			last_purchase_at = NOW(),
			last_order_at = NOW(),
			avg_order_value = (total_spent + $1) / NULLIF(total_orders + 1, 0),
			lifetime_value = total_spent + $1,
			updated_at = NOW()
		WHERE id = $2
	`
	result, err := r.db.ExecContext(ctx, query, amount, customerID)
	if err != nil {
		return r.handleError(err, "failed to update customer stats")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return repository.ErrCustomerNotFound
	}

	return nil
}

// UpdateLoyaltyPoints actualiza los puntos de lealtad del cliente
func (r *CustomerRepository) UpdateLoyaltyPoints(ctx context.Context, customerID int64, points int32) error {
	// Por ahora no implementado
	return nil
}

// SetVIP establece o quita el estado VIP del cliente
func (r *CustomerRepository) SetVIP(ctx context.Context, customerID int64, isVIP bool) error {
	query := `
		UPDATE crm.customers 
		SET is_vip = $1,
			vip_since = CASE WHEN $1 = true AND vip_since IS NULL THEN NOW() ELSE vip_since END,
			updated_at = NOW()
		WHERE id = $2
	`
	result, err := r.db.ExecContext(ctx, query, isVIP, customerID)
	if err != nil {
		return r.handleError(err, "failed to set VIP status")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return repository.ErrCustomerNotFound
	}

	return nil
}

// UpdatePreferences actualiza las preferencias de comunicación del cliente
func (r *CustomerRepository) UpdatePreferences(ctx context.Context, customerID int64, preferences map[string]interface{}) error {
	prefsJSON, err := json.Marshal(preferences)
	if err != nil {
		return fmt.Errorf("failed to marshal preferences: %w", err)
	}

	query := `
		UPDATE crm.customers 
		SET communication_preferences = $1,
			updated_at = NOW()
		WHERE id = $2
	`
	result, err := r.db.ExecContext(ctx, query, prefsJSON, customerID)
	if err != nil {
		return r.handleError(err, "failed to update preferences")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return repository.ErrCustomerNotFound
	}

	return nil
}

// UpdateInvoiceSettings actualiza la configuración de facturación del cliente
func (r *CustomerRepository) UpdateInvoiceSettings(ctx context.Context, customerID int64, requiresInvoice bool, taxID, taxName string) error {
	query := `
		UPDATE crm.customers 
		SET requires_invoice = $1,
			tax_id = $2,
			tax_name = $3,
			updated_at = NOW()
		WHERE id = $4
	`
	result, err := r.db.ExecContext(ctx, query, requiresInvoice, taxID, taxName, customerID)
	if err != nil {
		return r.handleError(err, "failed to update invoice settings")
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return repository.ErrCustomerNotFound
	}

	return nil
}

// GetStats obtiene estadísticas agregadas de clientes
func (r *CustomerRepository) GetStats(ctx context.Context) (*repository.CustomerStats, error) {
	query := `
		SELECT 
			COUNT(*) as total_customers,
			COUNT(CASE WHEN is_active = true THEN 1 END) as active_customers,
			COUNT(CASE WHEN is_vip = true THEN 1 END) as vip_customers,
			COUNT(CASE WHEN created_at >= NOW() - INTERVAL '30 days' THEN 1 END) as new_customers_last_30_days,
			COALESCE(SUM(total_spent), 0) as total_revenue,
			COALESCE(AVG(lifetime_value), 0) as avg_lifetime_value
		FROM crm.customers
	`

	var stats repository.CustomerStats
	err := r.db.GetContext(ctx, &stats, query)
	if err != nil {
		return nil, r.handleError(err, "failed to get customer stats")
	}

	// Obtener top países
	countryQuery := `
		SELECT 
			COALESCE(country, 'Unknown') as country,
			COUNT(*) as count,
			COALESCE(SUM(total_spent), 0) as revenue
		FROM crm.customers
		GROUP BY country
		ORDER BY count DESC
		LIMIT 10
	`

	var topCountries []repository.CountryStat
	err = r.db.SelectContext(ctx, &topCountries, countryQuery)
	if err != nil {
		// No fallamos si esto no funciona, solo devolvemos vacío
		stats.TopCountries = []repository.CountryStat{}
	} else {
		stats.TopCountries = topCountries
	}

	return &stats, nil
}

// GetVIPCustomers obtiene todos los clientes VIP activos
func (r *CustomerRepository) GetVIPCustomers(ctx context.Context) ([]*entities.Customer, error) {
	filter := &repository.CustomerFilter{
		IsVIP:     boolPtr(true),
		IsActive:  boolPtr(true),
		SortBy:    "vip_since",
		SortOrder: "DESC",
	}

	customers, _, err := r.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	return customers, nil
}

// boolPtr es una función auxiliar para crear un puntero a bool
func boolPtr(b bool) *bool {
	return &b
}
