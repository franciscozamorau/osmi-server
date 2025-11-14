package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/franciscozamorau/osmi-server/internal/db"
	"github.com/franciscozamorau/osmi-server/internal/models"
	"github.com/google/uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// CustomerRepository maneja las operaciones de base de datos para clientes
type CustomerRepository struct{}

func NewCustomerRepository() *CustomerRepository {
	return &CustomerRepository{}
}

// Expresiones regulares para validación
var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z09.-]+\.[a-zA-Z]{2,}$`)
	phoneRegex = regexp.MustCompile(`^(\+?[1-9]\d{1,14}|[\d\s\(\)\.\-]+)$`)
)

// CreateCustomer crea un nuevo cliente en la base de datos
func (r *CustomerRepository) CreateCustomer(ctx context.Context, name, email, phone string) (int64, error) {
	// Validaciones
	if err := r.validateCustomerData(name, email, phone); err != nil {
		return 0, err
	}

	// Limpiar datos
	name = strings.TrimSpace(name)
	email = strings.TrimSpace(strings.ToLower(email))
	phone = strings.TrimSpace(phone)

	// Validar duplicados antes de insertar
	existing, err := r.GetCustomerByEmail(ctx, email)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return 0, fmt.Errorf("error checking existing customer: %v", err)
	}
	if existing != nil {
		return 0, fmt.Errorf("customer with email %s already exists", email)
	}

	// Generar public_id automáticamente
	publicID := uuid.New().String()

	query := `
		INSERT INTO customers (public_id, name, email, phone, is_verified) 
		VALUES ($1, $2, $3, $4, $5) 
		RETURNING id
	`

	var id int64
	err = db.Pool.QueryRow(ctx, query, publicID, name, email,
		toPgText(phone),
		false,
	).Scan(&id)

	if err != nil {
		if isDuplicateKeyError(err) {
			return 0, fmt.Errorf("customer with email %s already exists", email)
		}
		return 0, fmt.Errorf("error creating customer: %v", err)
	}

	log.Printf("Customer created: %s (ID: %d, PublicID: %s)", email, id, publicID)
	return id, nil
}

// GetCustomerByID obtiene un cliente por su ID
func (r *CustomerRepository) GetCustomerByID(ctx context.Context, id int64) (*models.Customer, error) {
	query := `
		SELECT id, public_id, user_id, name, email, phone, date_of_birth, 
		       address, preferences, loyalty_points, is_verified, 
		       verification_token, created_at, updated_at
		FROM customers 
		WHERE id = $1
	`

	var customer models.Customer
	var userID pgtype.Int4
	var dateOfBirth pgtype.Date

	err := db.Pool.QueryRow(ctx, query, id).Scan(
		&customer.ID,
		&customer.PublicID,
		&userID,
		&customer.Name,
		&customer.Email,
		&customer.Phone,
		&dateOfBirth,
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

	if userID.Valid {
		customer.UserID = userID
	}

	if dateOfBirth.Valid {
		customer.DateOfBirth = dateOfBirth
	}

	return &customer, nil
}

// GetCustomerByPublicID obtiene un cliente por su public_id
func (r *CustomerRepository) GetCustomerByPublicID(ctx context.Context, publicID string) (*models.Customer, error) {
	query := `
		SELECT id, public_id, user_id, name, email, phone, date_of_birth, 
		       address, preferences, loyalty_points, is_verified, 
		       verification_token, created_at, updated_at
		FROM customers 
		WHERE public_id = $1
	`

	var customer models.Customer
	var userID pgtype.Int4
	var dateOfBirth pgtype.Date

	err := db.Pool.QueryRow(ctx, query, publicID).Scan(
		&customer.ID,
		&customer.PublicID,
		&userID,
		&customer.Name,
		&customer.Email,
		&customer.Phone,
		&dateOfBirth,
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

	if userID.Valid {
		customer.UserID = userID
	}

	if dateOfBirth.Valid {
		customer.DateOfBirth = dateOfBirth
	}

	return &customer, nil
}

// GetCustomerByEmail obtiene un cliente por email
func (r *CustomerRepository) GetCustomerByEmail(ctx context.Context, email string) (*models.Customer, error) {
	query := `
		SELECT id, public_id, user_id, name, email, phone, date_of_birth, 
		       address, preferences, loyalty_points, is_verified, 
		       verification_token, created_at, updated_at
		FROM customers 
		WHERE email = $1
	`

	var customer models.Customer
	var userID pgtype.Int4
	var dateOfBirth pgtype.Date

	err := db.Pool.QueryRow(ctx, query, strings.TrimSpace(strings.ToLower(email))).Scan(
		&customer.ID,
		&customer.PublicID,
		&userID,
		&customer.Name,
		&customer.Email,
		&customer.Phone,
		&dateOfBirth,
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
			return nil, nil
		}
		return nil, fmt.Errorf("error getting customer by email: %v", err)
	}

	if userID.Valid {
		customer.UserID = userID
	}

	if dateOfBirth.Valid {
		customer.DateOfBirth = dateOfBirth
	}

	return &customer, nil
}

// UpdateCustomer actualiza un cliente existente
func (r *CustomerRepository) UpdateCustomer(ctx context.Context, id int64, name, email, phone string) error {
	if err := r.validateCustomerData(name, email, phone); err != nil {
		return err
	}

	name = strings.TrimSpace(name)
	email = strings.TrimSpace(strings.ToLower(email))
	phone = strings.TrimSpace(phone)

	query := `
		UPDATE customers 
		SET name = $1, email = $2, phone = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $4
	`

	result, err := db.Pool.Exec(ctx, query, name, email, toPgText(phone), id)
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

// UpdateCustomerWithUserID actualiza un cliente vinculándolo a un usuario
func (r *CustomerRepository) UpdateCustomerWithUserID(ctx context.Context, customerID int64, userID int64) error {
	query := `
		UPDATE customers 
		SET user_id = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`

	result, err := db.Pool.Exec(ctx, query, userID, customerID)
	if err != nil {
		return fmt.Errorf("error updating customer user_id: %v", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("customer not found with id: %d", customerID)
	}

	log.Printf("Customer %d linked to user %d", customerID, userID)
	return nil
}

// DeleteCustomer elimina un cliente (soft delete)
func (r *CustomerRepository) DeleteCustomer(ctx context.Context, id int64) error {
	query := `UPDATE customers SET is_verified = false, updated_at = CURRENT_TIMESTAMP WHERE id = $1`

	result, err := db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting customer: %v", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("customer not found with id: %d", id)
	}

	log.Printf("Customer soft deleted: ID %d", id)
	return nil
}

// ListCustomers lista todos los clientes con paginación
func (r *CustomerRepository) ListCustomers(ctx context.Context, limit, offset int) ([]models.Customer, error) {
	query := `
		SELECT id, public_id, name, email, phone, is_verified, created_at, updated_at
		FROM customers 
		WHERE is_verified = true
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
			&customer.IsVerified,
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
		SELECT id, public_id, name, email, phone, is_verified, created_at, updated_at
		FROM customers 
		WHERE name ILIKE $1 AND is_verified = true
		ORDER BY name
		LIMIT $2 OFFSET $3
	`

	rows, err := db.Pool.Query(ctx, query, "%"+strings.TrimSpace(name)+"%", limit, offset)
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
			&customer.IsVerified,
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
		SELECT id, public_id, name, email, phone, loyalty_points, created_at, updated_at
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
			&customer.LoyaltyPoints,
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
		"SELECT id FROM customers WHERE public_id = $1 AND is_verified = true",
		publicID).Scan(&customerID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, fmt.Errorf("customer not found with public_id: %s", publicID)
		}
		return 0, fmt.Errorf("error getting customer ID by public_id: %v", err)
	}
	return customerID, nil
}

// GetCustomerPublicIDByEmail obtiene el public_id de un cliente por su email
func (r *CustomerRepository) GetCustomerPublicIDByEmail(ctx context.Context, email string) (string, error) {
	var publicID string
	err := db.Pool.QueryRow(ctx,
		"SELECT public_id FROM customers WHERE email = $1 AND is_verified = true",
		strings.TrimSpace(strings.ToLower(email))).Scan(&publicID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", fmt.Errorf("customer not found with email: %s", email)
		}
		return "", fmt.Errorf("error getting customer public_id by email: %v", err)
	}
	return publicID, nil
}

// UpdateCustomerPreferences actualiza las preferencias de un cliente
func (r *CustomerRepository) UpdateCustomerPreferences(ctx context.Context, customerID int64, preferences map[string]interface{}) error {
	preferencesJSON, err := json.Marshal(preferences)
	if err != nil {
		return fmt.Errorf("error marshaling preferences: %v", err)
	}

	query := `UPDATE customers SET preferences = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`

	result, err := db.Pool.Exec(ctx, query, preferencesJSON, customerID)
	if err != nil {
		return fmt.Errorf("error updating customer preferences: %v", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("customer not found with id: %d", customerID)
	}

	log.Printf("Customer preferences updated: ID %d", customerID)
	return nil
}

// AddLoyaltyPoints agrega puntos de fidelidad a un cliente
func (r *CustomerRepository) AddLoyaltyPoints(ctx context.Context, customerID int64, points int32) error {
	query := `UPDATE customers SET loyalty_points = loyalty_points + $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`

	result, err := db.Pool.Exec(ctx, query, points, customerID)
	if err != nil {
		return fmt.Errorf("error adding loyalty points: %v", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("customer not found with id: %d", customerID)
	}

	log.Printf("Added %d loyalty points to customer ID %d", points, customerID)
	return nil
}

// validateCustomerData valida los datos del cliente
func (r *CustomerRepository) validateCustomerData(name, email, phone string) error {
	name = strings.TrimSpace(name)
	email = strings.TrimSpace(strings.ToLower(email))
	phone = strings.TrimSpace(phone)

	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}

	if len(name) > 255 {
		return fmt.Errorf("name cannot exceed 255 characters")
	}

	if email == "" {
		return fmt.Errorf("email cannot be empty")
	}

	if len(email) > 255 {
		return fmt.Errorf("email cannot exceed 255 characters")
	}

	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format: %s", email)
	}

	if phone != "" {
		if len(phone) > 50 {
			return fmt.Errorf("phone cannot exceed 50 characters")
		}

		digits := regexp.MustCompile(`\d`).FindAllString(phone, -1)
		if len(digits) < 6 {
			return fmt.Errorf("phone number must contain at least 6 digits")
		}

		if !phoneRegex.MatchString(phone) {
			return fmt.Errorf("invalid phone format. Use E.164 format: +1234567890 or standard phone format")
		}
	}

	return nil
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
