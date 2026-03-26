// internal/infrastructure/repositories/postgres/customer_repository.go
package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
	"github.com/franciscozamorau/osmi-server/internal/domain/repository"
)

// CustomerRepository implementa la interfaz repository.CustomerRepository usando PostgreSQL
type CustomerRepository struct {
	db *pgxpool.Pool
}

// NewCustomerRepository crea una nueva instancia del repositorio
func NewCustomerRepository(db *pgxpool.Pool) *CustomerRepository {
	return &CustomerRepository{
		db: db,
	}
}

// handleError mapea errores de PostgreSQL a nuestros errores de dominio
func (r *CustomerRepository) handleError(err error, context string) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return repository.ErrCustomerNotFound
	}

	// Verificar si es un error de PostgreSQL con código
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // Unique violation
			if strings.Contains(pgErr.ConstraintName, "customers_email_key") {
				return repository.ErrCustomerEmailExists
			}
			if strings.Contains(pgErr.ConstraintName, "customers_public_uuid_key") {
				return repository.ErrCustomerAlreadyLinked
			}
		case "23503": // Foreign key violation
			return fmt.Errorf("referenced user not found: %w", err)
		}
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
	args := pgx.NamedArgs{}
	argPos := 1

	if filter != nil {
		// Filtros por ID
		if len(filter.IDs) > 0 {
			conditions = append(conditions, fmt.Sprintf("id = ANY(@id_%d)", argPos))
			args[fmt.Sprintf("id_%d", argPos)] = filter.IDs
			argPos++
		}

		if len(filter.PublicIDs) > 0 {
			conditions = append(conditions, fmt.Sprintf("public_uuid = ANY(@public_%d)", argPos))
			args[fmt.Sprintf("public_%d", argPos)] = filter.PublicIDs
			argPos++
		}

		if filter.UserID != nil {
			conditions = append(conditions, fmt.Sprintf("user_id = @user_%d", argPos))
			args[fmt.Sprintf("user_%d", argPos)] = *filter.UserID
			argPos++
		}

		if filter.Email != nil {
			conditions = append(conditions, fmt.Sprintf("email = @email_%d", argPos))
			args[fmt.Sprintf("email_%d", argPos)] = *filter.Email
			argPos++
		}

		// Filtros de texto
		if filter.SearchTerm != nil && *filter.SearchTerm != "" {
			searchTerm := "%" + *filter.SearchTerm + "%"
			conditions = append(conditions, fmt.Sprintf(
				"(full_name ILIKE @search_%d OR email ILIKE @search_%d OR company_name ILIKE @search_%d OR tax_id ILIKE @search_%d)",
				argPos, argPos, argPos, argPos,
			))
			args[fmt.Sprintf("search_%d", argPos)] = searchTerm
			argPos++
		}

		if filter.FullName != nil {
			conditions = append(conditions, fmt.Sprintf("full_name ILIKE @fullname_%d", argPos))
			args[fmt.Sprintf("fullname_%d", argPos)] = "%" + *filter.FullName + "%"
			argPos++
		}

		if filter.CompanyName != nil {
			conditions = append(conditions, fmt.Sprintf("company_name ILIKE @company_%d", argPos))
			args[fmt.Sprintf("company_%d", argPos)] = "%" + *filter.CompanyName + "%"
			argPos++
		}

		if filter.Country != nil {
			conditions = append(conditions, fmt.Sprintf("country = @country_%d", argPos))
			args[fmt.Sprintf("country_%d", argPos)] = *filter.Country
			argPos++
		}

		if filter.City != nil {
			conditions = append(conditions, fmt.Sprintf("city ILIKE @city_%d", argPos))
			args[fmt.Sprintf("city_%d", argPos)] = "%" + *filter.City + "%"
			argPos++
		}

		if filter.IsActive != nil {
			conditions = append(conditions, fmt.Sprintf("is_active = @active_%d", argPos))
			args[fmt.Sprintf("active_%d", argPos)] = *filter.IsActive
			argPos++
		}

		if filter.IsVIP != nil {
			conditions = append(conditions, fmt.Sprintf("is_vip = @vip_%d", argPos))
			args[fmt.Sprintf("vip_%d", argPos)] = *filter.IsVIP
			argPos++
		}

		if filter.RequiresInvoice != nil {
			conditions = append(conditions, fmt.Sprintf("requires_invoice = @invoice_%d", argPos))
			args[fmt.Sprintf("invoice_%d", argPos)] = *filter.RequiresInvoice
			argPos++
		}

		if filter.CustomerSegment != nil {
			conditions = append(conditions, fmt.Sprintf("customer_segment = @segment_%d", argPos))
			args[fmt.Sprintf("segment_%d", argPos)] = *filter.CustomerSegment
			argPos++
		}

		// Filtros de fechas
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

		if filter.LastPurchaseFrom != nil {
			conditions = append(conditions, fmt.Sprintf("last_purchase_at >= @purchase_from_%d", argPos))
			args[fmt.Sprintf("purchase_from_%d", argPos)] = *filter.LastPurchaseFrom
			argPos++
		}

		if filter.LastPurchaseTo != nil {
			conditions = append(conditions, fmt.Sprintf("last_purchase_at <= @purchase_to_%d", argPos))
			args[fmt.Sprintf("purchase_to_%d", argPos)] = *filter.LastPurchaseTo
			argPos++
		}

		if filter.MinTotalSpent != nil {
			conditions = append(conditions, fmt.Sprintf("total_spent >= @min_spent_%d", argPos))
			args[fmt.Sprintf("min_spent_%d", argPos)] = *filter.MinTotalSpent
			argPos++
		}

		if filter.MaxTotalSpent != nil {
			conditions = append(conditions, fmt.Sprintf("total_spent <= @max_spent_%d", argPos))
			args[fmt.Sprintf("max_spent_%d", argPos)] = *filter.MaxTotalSpent
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
	err := r.db.QueryRow(ctx, countQuery, args).Scan(&total)
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

		// Establecer límite
		limit := filter.Limit
		if limit <= 0 {
			limit = 20
		}

		if limit > 0 {
			baseQuery += " LIMIT @limit"
			args["limit"] = limit
		}

		if filter.Offset > 0 {
			baseQuery += " OFFSET @offset"
			args["offset"] = filter.Offset
		}
	} else {
		baseQuery += " ORDER BY created_at DESC LIMIT 20"
	}

	// Ejecutar query
	rows, err := r.db.Query(ctx, baseQuery, args)
	if err != nil {
		return nil, 0, r.handleError(err, "failed to find customers")
	}
	defer rows.Close()

	var customers []*entities.Customer
	for rows.Next() {
		var customer entities.Customer
		var commPrefsJSON []byte
		var userID *int64
		var phone *string
		var companyName *string
		var addressLine1 *string
		var addressLine2 *string
		var city *string
		var state *string
		var postalCode *string
		var country *string
		var taxID *string
		var taxIDType *string
		var taxName *string
		var firstOrderAt *time.Time
		var lastOrderAt *time.Time
		var lastPurchaseAt *time.Time
		var vipSince *time.Time

		err = rows.Scan(
			&customer.ID, &customer.PublicID, &userID,
			&customer.FullName, &customer.Email, &phone,
			&companyName, &addressLine1, &addressLine2,
			&city, &state, &postalCode, &country,
			&taxID, &taxIDType, &taxName, &customer.RequiresInvoice,
			&commPrefsJSON,
			&customer.TotalSpent, &customer.TotalOrders, &customer.TotalTickets, &customer.AvgOrderValue,
			&firstOrderAt, &lastOrderAt, &lastPurchaseAt,
			&customer.IsActive, &customer.IsVIP, &vipSince,
			&customer.CustomerSegment, &customer.LifetimeValue,
			&customer.CreatedAt, &customer.UpdatedAt,
		)
		if err != nil {
			return nil, 0, r.handleError(err, "failed to scan customer row")
		}

		// Asignar campos NULL
		customer.UserID = userID
		customer.Phone = phone
		customer.CompanyName = companyName
		customer.AddressLine1 = addressLine1
		customer.AddressLine2 = addressLine2
		customer.City = city
		customer.State = state
		customer.PostalCode = postalCode
		customer.Country = country
		customer.TaxID = taxID
		customer.TaxIDType = taxIDType
		customer.TaxName = taxName
		customer.FirstOrderAt = firstOrderAt
		customer.LastOrderAt = lastOrderAt
		customer.LastPurchaseAt = lastPurchaseAt
		customer.VIPSince = vipSince

		// Deserializar JSON
		if len(commPrefsJSON) > 0 {
			json.Unmarshal(commPrefsJSON, &customer.CommunicationPreferences)
		}

		customers = append(customers, &customer)
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

	err = r.db.QueryRow(ctx, query,
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

	err = r.db.QueryRow(ctx, query,
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
	cmdTag, err := r.db.Exec(ctx, `DELETE FROM crm.customers WHERE id = $1`, id)
	if err != nil {
		return r.handleError(err, "failed to delete customer")
	}

	if cmdTag.RowsAffected() == 0 {
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
	cmdTag, err := r.db.Exec(ctx, query, publicID)
	if err != nil {
		return r.handleError(err, "failed to soft delete customer")
	}

	if cmdTag.RowsAffected() == 0 {
		return repository.ErrCustomerNotFound
	}

	return nil
}

// Exists verifica si existe un cliente con el ID dado
func (r *CustomerRepository) Exists(ctx context.Context, id int64) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM crm.customers WHERE id = $1)`
	err := r.db.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, r.handleError(err, "failed to check customer existence")
	}
	return exists, nil
}

// ExistsByEmail verifica si existe un cliente con el email dado
func (r *CustomerRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM crm.customers WHERE email = $1)`
	err := r.db.QueryRow(ctx, query, email).Scan(&exists)
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
	cmdTag, err := r.db.Exec(ctx, query, amount, customerID)
	if err != nil {
		return r.handleError(err, "failed to update customer stats")
	}

	if cmdTag.RowsAffected() == 0 {
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
	cmdTag, err := r.db.Exec(ctx, query, isVIP, customerID)
	if err != nil {
		return r.handleError(err, "failed to set VIP status")
	}

	if cmdTag.RowsAffected() == 0 {
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
	cmdTag, err := r.db.Exec(ctx, query, prefsJSON, customerID)
	if err != nil {
		return r.handleError(err, "failed to update preferences")
	}

	if cmdTag.RowsAffected() == 0 {
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
	cmdTag, err := r.db.Exec(ctx, query, requiresInvoice, taxID, taxName, customerID)
	if err != nil {
		return r.handleError(err, "failed to update invoice settings")
	}

	if cmdTag.RowsAffected() == 0 {
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
	err := r.db.QueryRow(ctx, query).Scan(
		&stats.TotalCustomers,
		&stats.ActiveCustomers,
		&stats.VIPCustomers,
		&stats.NewCustomersLast30Days,
		&stats.TotalRevenue,
		&stats.AvgLifetimeValue,
	)
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

	rows, err := r.db.Query(ctx, countryQuery)
	if err != nil {
		stats.TopCountries = []repository.CountryStat{}
		return &stats, nil
	}
	defer rows.Close()

	var topCountries []repository.CountryStat
	for rows.Next() {
		var cs repository.CountryStat
		err = rows.Scan(&cs.Country, &cs.Count, &cs.Revenue)
		if err != nil {
			continue
		}
		topCountries = append(topCountries, cs)
	}
	stats.TopCountries = topCountries

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
