package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

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

	// Auditoría estructurada
	auditCtx := &AuditContext{
		UserID:    "system", // En un caso real, obtener del contexto
		IPAddress: "127.0.0.1",
		UserAgent: "osmi-server",
	}
	r.auditCustomerChange(ctx, "INSERT", nil, id, auditCtx)

	log.Printf("✅ Customer created: %s (ID: %d)", email, id)
	return id, nil
}

// GetCustomerByID obtiene un cliente por su ID
func (r *CustomerRepository) GetCustomerByID(ctx context.Context, id int) (*models.Customer, error) {
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
			return nil, fmt.Errorf("customer not found with email: %s", email)
		}
		return nil, fmt.Errorf("error getting customer by email: %v", err)
	}

	return &customer, nil
}

// UpdateCustomer actualiza un cliente existente
func (r *CustomerRepository) UpdateCustomer(ctx context.Context, id int, name, email, phone string) error {
	// Validaciones
	if err := r.validateCustomerData(name, email, phone); err != nil {
		return err
	}

	// Obtener datos antiguos para auditoría
	oldCustomer, err := r.GetCustomerByID(ctx, id)
	if err != nil {
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

	// Auditoría estructurada
	auditCtx := &AuditContext{
		UserID:    "system",
		IPAddress: "127.0.0.1",
		UserAgent: "osmi-server",
	}
	r.auditCustomerChange(ctx, "UPDATE", oldCustomer, int64(id), auditCtx)

	log.Printf("✅ Customer updated: %s (ID: %d)", email, id)
	return nil
}

// DeleteCustomer elimina un cliente
func (r *CustomerRepository) DeleteCustomer(ctx context.Context, id int) error {
	// Obtener datos para auditoría antes de eliminar
	oldCustomer, err := r.GetCustomerByID(ctx, id)
	if err != nil {
		return err
	}

	query := `DELETE FROM customers WHERE id = $1`

	result, err := db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting customer: %v", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("customer not found with id: %d", id)
	}

	// Auditoría estructurada
	auditCtx := &AuditContext{
		UserID:    "system",
		IPAddress: "127.0.0.1",
		UserAgent: "osmi-server",
	}
	r.auditCustomerChange(ctx, "DELETE", oldCustomer, int64(id), auditCtx)

	log.Printf("✅ Customer deleted: ID %d", id)
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

// Helper types and functions

type AuditContext struct {
	UserID    string
	IPAddress string
	UserAgent string
	Metadata  map[string]interface{}
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

// auditCustomerChange realiza auditoría estructurada de cambios
func (r *CustomerRepository) auditCustomerChange(ctx context.Context, operation string, oldCustomer *models.Customer, customerID int64, auditCtx *AuditContext) {
	auditPayload := map[string]interface{}{
		"operation":   operation,
		"customer_id": customerID,
		"timestamp":   time.Now().UTC(),
		"context":     auditCtx,
	}

	if oldCustomer != nil {
		auditPayload["old_data"] = map[string]interface{}{
			"name":  oldCustomer.Name,
			"email": oldCustomer.Email,
			"phone": oldCustomer.Phone.String,
		}
	}

	// En producción, enviar a servicio de auditoría externo
	// Ej: r.auditService.Send(auditPayload)

	log.Printf("📝 Customer audit - %s: %+v", operation, auditPayload)
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
	return repo.GetCustomerByID(ctx, id)
}
