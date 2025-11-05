package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/franciscozamorau/osmi-server/internal/db"
	"github.com/franciscozamorau/osmi-server/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// CustomerRepository maneja las operaciones de base de datos para clientes
type CustomerRepository struct{}

func NewCustomerRepository() *CustomerRepository {
	return &CustomerRepository{}
}

// Expresiones regulares para validación
var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	phoneRegex = regexp.MustCompile(`^\+?[1-9]\d{1,14}$`) // E.164 format
)

// CreateCustomer crea un nuevo cliente en la base de datos
func (r *CustomerRepository) CreateCustomer(ctx context.Context, name, email, phone string) (int64, error) {
	// Validaciones
	if err := r.validateCustomerData(name, email, phone); err != nil {
		return 0, err
	}

	// Validar duplicados antes de insertar
	existing, err := r.GetCustomerByEmail(ctx, email)
	if err == nil && existing != nil {
		return 0, fmt.Errorf("customer with email %s already exists", email)
	}

	query := `
		INSERT INTO customers (name, email, phone) 
		VALUES ($1, $2, $3) 
		RETURNING id
	`

	var id int64
	err = db.Pool.QueryRow(ctx, query, name, email, phone).Scan(&id)
	if err != nil {
		// Verificar si es error de duplicado (aunque ya validamos)
		if isDuplicateKeyError(err) {
			return 0, fmt.Errorf("customer with email %s already exists", email)
		}
		return 0, fmt.Errorf("error creating customer: %v", err)
	}

	log.Printf("Customer created: %s (ID: %d)", email, id)
	return id, nil
}

// GetCustomerByID obtiene un cliente por su ID
func (r *CustomerRepository) GetCustomerByID(ctx context.Context, id int64) (*models.Customer, error) {
	query := `
		SELECT id, public_id, name, email, phone, date_of_birth, address, 
		       preferences, loyalty_points, is_verified, verification_token,
		       created_at, updated_at
		FROM customers 
		WHERE id = $1
	`

	var customer models.Customer
	err := db.Pool.QueryRow(ctx, query, id).Scan(
		&customer.ID,
		&customer.PublicID,
		&customer.Name,
		&customer.Email,
		&customer.Phone,
		&customer.DateOfBirth,
		&customer.Address,
		&customer.Preferences,
		&customer.LoyaltyPoints,
		&customer.IsVerified,
		&customer.VerificationToken,
		&customer.CreatedAt,
		&customer.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("customer not found with id: %d", id)
		}
		return nil, fmt.Errorf("error getting customer: %v", err)
	}

	return &customer, nil
}

// GetCustomerByPublicID obtiene un cliente por public_id
func (r *CustomerRepository) GetCustomerByPublicID(ctx context.Context, publicID string) (*models.Customer, error) {
	query := `
		SELECT id, public_id, name, email, phone, date_of_birth, address, 
		       preferences, loyalty_points, is_verified, verification_token,
		       created_at, updated_at
		FROM customers 
		WHERE public_id = $1
	`

	var customer models.Customer
	err := db.Pool.QueryRow(ctx, query, publicID).Scan(
		&customer.ID,
		&customer.PublicID,
		&customer.Name,
		&customer.Email,
		&customer.Phone,
		&customer.DateOfBirth,
		&customer.Address,
		&customer.Preferences,
		&customer.LoyaltyPoints,
		&customer.IsVerified,
		&customer.VerificationToken,
		&customer.CreatedAt,
		&customer.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("customer not found with public_id: %s", publicID)
		}
		return nil, fmt.Errorf("error getting customer by public_id: %v", err)
	}

	return &customer, nil
}

// GetCustomerByEmail obtiene un cliente por email
func (r *CustomerRepository) GetCustomerByEmail(ctx context.Context, email string) (*models.Customer, error) {
	query := `
		SELECT id, public_id, name, email, phone, date_of_birth, address, 
		       preferences, loyalty_points, is_verified, verification_token,
		       created_at, updated_at
		FROM customers 
		WHERE email = $1
	`

	var customer models.Customer
	err := db.Pool.QueryRow(ctx, query, email).Scan(
		&customer.ID,
		&customer.PublicID,
		&customer.Name,
		&customer.Email,
		&customer.Phone,
		&customer.DateOfBirth,
		&customer.Address,
		&customer.Preferences,
		&customer.LoyaltyPoints,
		&customer.IsVerified,
		&customer.VerificationToken,
		&customer.CreatedAt,
		&customer.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // No es error, simplemente no existe
		}
		return nil, fmt.Errorf("error getting customer by email: %v", err)
	}

	return &customer, nil
}

// UpdateCustomer actualiza un cliente existente
func (r *CustomerRepository) UpdateCustomer(ctx context.Context, id int64, name, email, phone string) error {
	// Validaciones
	if err := r.validateCustomerData(name, email, phone); err != nil {
		return err
	}

	query := `
		UPDATE customers 
		SET name = $1, email = $2, phone = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $4
	`

	result, err := db.Pool.Exec(ctx, query, name, email, phone, id)
	if err != nil {
		if isDuplicateKeyError(err) {
			return fmt.Errorf("customer with email %s already exists", email)
		}
		return fmt.Errorf("error updating customer: %v", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("customer not found with id: %d", id)
	}

	log.Printf("Customer updated: %s (ID: %d)", email, id)
	return nil
}

// DeleteCustomer elimina un cliente
func (r *CustomerRepository) DeleteCustomer(ctx context.Context, id int64) error {
	query := `DELETE FROM customers WHERE id = $1`

	result, err := db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting customer: %v", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("customer not found with id: %d", id)
	}

	log.Printf("Customer deleted: ID %d", id)
	return nil
}

// ListCustomers lista todos los clientes con paginación
func (r *CustomerRepository) ListCustomers(ctx context.Context, limit, offset int) ([]models.Customer, error) {
	query := `
		SELECT id, public_id, name, email, phone, created_at, updated_at
		FROM customers 
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := db.Pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error listing customers: %v", err)
	}
	defer rows.Close()

	var customers []models.Customer
	for rows.Next() {
		var customer models.Customer
		err := rows.Scan(
			&customer.ID,
			&customer.PublicID,
			&customer.Name,
			&customer.Email,
			&customer.Phone,
			&customer.CreatedAt,
			&customer.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning customer: %v", err)
		}
		customers = append(customers, customer)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating customers: %v", err)
	}

	return customers, nil
}

// FindCustomersByName busca clientes por nombre (búsqueda parcial)
func (r *CustomerRepository) FindCustomersByName(ctx context.Context, name string, limit, offset int) ([]models.Customer, error) {
	query := `
		SELECT id, public_id, name, email, phone, created_at, updated_at
		FROM customers 
		WHERE name ILIKE $1
		ORDER BY name
		LIMIT $2 OFFSET $3
	`

	rows, err := db.Pool.Query(ctx, query, "%"+name+"%", limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error finding customers by name: %v", err)
	}
	defer rows.Close()

	var customers []models.Customer
	for rows.Next() {
		var customer models.Customer
		err := rows.Scan(
			&customer.ID,
			&customer.PublicID,
			&customer.Name,
			&customer.Email,
			&customer.Phone,
			&customer.CreatedAt,
			&customer.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning customer: %v", err)
		}
		customers = append(customers, customer)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating customers: %v", err)
	}

	return customers, nil
}

// ListVerifiedCustomers lista solo clientes verificados
func (r *CustomerRepository) ListVerifiedCustomers(ctx context.Context, limit, offset int) ([]models.Customer, error) {
	query := `
		SELECT id, public_id, name, email, phone, created_at, updated_at
		FROM customers 
		WHERE is_verified = true
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := db.Pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error listing verified customers: %v", err)
	}
	defer rows.Close()

	var customers []models.Customer
	for rows.Next() {
		var customer models.Customer
		err := rows.Scan(
			&customer.ID,
			&customer.PublicID,
			&customer.Name,
			&customer.Email,
			&customer.Phone,
			&customer.CreatedAt,
			&customer.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning customer: %v", err)
		}
		customers = append(customers, customer)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating customers: %v", err)
	}

	return customers, nil
}

// GetCustomerIDByPublicID obtiene el ID interno de un cliente por su public_id
func (r *CustomerRepository) GetCustomerIDByPublicID(ctx context.Context, publicID string) (int64, error) {
	var customerID int64
	err := db.Pool.QueryRow(ctx,
		"SELECT id FROM customers WHERE public_id = $1",
		publicID).Scan(&customerID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, fmt.Errorf("customer not found with public_id: %s", publicID)
		}
		return 0, fmt.Errorf("error getting customer ID by public_id: %v", err)
	}
	return customerID, nil
}

// validateCustomerData valida los datos del cliente
func (r *CustomerRepository) validateCustomerData(name, email, phone string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("name cannot be empty")
	}

	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format: %s", email)
	}

	if phone != "" && !phoneRegex.MatchString(phone) {
		return fmt.Errorf("invalid phone format. Use E.164 format: +1234567890")
	}

	return nil
}

// isDuplicateKeyError verifica si el error es por violación de unique constraint
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}

	// Fallback: verificar el mensaje de error
	errorMsg := err.Error()
	return strings.Contains(errorMsg, "duplicate key") ||
		strings.Contains(errorMsg, "already exists") ||
		strings.Contains(errorMsg, "unique constraint")
}

// Funciones globales (para compatibilidad con código existente)

func CreateCustomer(ctx context.Context, name, email string) (int, error) {
	repo := NewCustomerRepository()
	id, err := repo.CreateCustomer(ctx, name, email, "")
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func GetCustomerByID(ctx context.Context, id int) (*models.Customer, error) {
	repo := NewCustomerRepository()
	return repo.GetCustomerByID(ctx, int64(id))
}
