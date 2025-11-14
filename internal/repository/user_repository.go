package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/franciscozamorau/osmi-server/internal/db"
	"github.com/franciscozamorau/osmi-server/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

// UserRepository maneja las operaciones de base de datos para usuarios
type UserRepository struct{}

func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

// Valid roles
var validRoles = map[string]bool{
	"customer":  true,
	"organizer": true,
	"admin":     true,
}

// CreateUser crea un nuevo usuario en la base de datos
func (r *UserRepository) CreateUser(ctx context.Context, name, email, password, role string) (int64, error) {
	// Validaciones
	if err := r.validateUserData(name, email, password, role); err != nil {
		return 0, err
	}

	// Limpiar y normalizar datos
	name = strings.TrimSpace(name)
	email = strings.TrimSpace(strings.ToLower(email))
	role = strings.TrimSpace(strings.ToLower(role))

	// Verificar si el usuario ya existe
	existing, err := r.GetUserByEmail(ctx, email)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return 0, fmt.Errorf("error checking existing user: %v", err)
	}
	if existing != nil {
		return 0, fmt.Errorf("user with email %s already exists", email)
	}

	// Generar public_id
	publicID := uuid.New().String()

	// Hash de la contraseña
	passwordHash, err := r.hashPassword(password)
	if err != nil {
		return 0, fmt.Errorf("error hashing password: %v", err)
	}

	// Asignar username basado en email si no se proporciona name
	username := r.generateUsername(name, email)

	query := `
		INSERT INTO users (
			public_id, username, email, password_hash, role, is_active,
			password_changed_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id
	`

	var id int64
	err = db.Pool.QueryRow(ctx, query,
		publicID,
		username,
		email,
		passwordHash,
		role,
		true, // is_active por defecto
	).Scan(&id)

	if err != nil {
		if isDuplicateKeyError(err) {
			return 0, fmt.Errorf("user with email %s already exists", email)
		}
		return 0, fmt.Errorf("error creating user: %v", err)
	}

	log.Printf("User created: %s (ID: %d, PublicID: %s, Role: %s)", email, id, publicID, role)
	return id, nil
}

// GetUserByID obtiene un usuario por su ID
func (r *UserRepository) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	query := `
		SELECT id, public_id, username, email, password_hash, role, is_active,
		       last_login, failed_login_attempts, password_changed_at,
		       created_at, updated_at
		FROM users 
		WHERE id = $1 AND is_active = true
	`

	var user models.User
	var lastLogin pgtype.Timestamp
	var passwordChangedAt pgtype.Timestamp

	err := db.Pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.PublicID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.IsActive,
		&lastLogin,
		&user.FailedLoginAttempts,
		&passwordChangedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found with id: %d", id)
		}
		return nil, fmt.Errorf("error getting user: %v", err)
	}

	// Asignar campos pgtype si son válidos
	if lastLogin.Valid {
		user.LastLogin = lastLogin
	}
	if passwordChangedAt.Valid {
		user.PasswordChangedAt = passwordChangedAt
	}

	return &user, nil
}

// GetUserByPublicID obtiene un usuario por su public_id
func (r *UserRepository) GetUserByPublicID(ctx context.Context, publicID string) (*models.User, error) {
	query := `
		SELECT id, public_id, username, email, password_hash, role, is_active,
		       last_login, failed_login_attempts, password_changed_at,
		       created_at, updated_at
		FROM users 
		WHERE public_id = $1 AND is_active = true
	`

	var user models.User
	var lastLogin pgtype.Timestamp
	var passwordChangedAt pgtype.Timestamp

	err := db.Pool.QueryRow(ctx, query, publicID).Scan(
		&user.ID,
		&user.PublicID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.IsActive,
		&lastLogin,
		&user.FailedLoginAttempts,
		&passwordChangedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found with public_id: %s", publicID)
		}
		return nil, fmt.Errorf("error getting user by public_id: %v", err)
	}

	// Asignar campos pgtype si son válidos
	if lastLogin.Valid {
		user.LastLogin = lastLogin
	}
	if passwordChangedAt.Valid {
		user.PasswordChangedAt = passwordChangedAt
	}

	return &user, nil
}

// GetUserByEmail obtiene un usuario por email
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, public_id, username, email, password_hash, role, is_active,
		       last_login, failed_login_attempts, password_changed_at,
		       created_at, updated_at
		FROM users 
		WHERE email = $1 AND is_active = true
	`

	var user models.User
	var lastLogin pgtype.Timestamp
	var passwordChangedAt pgtype.Timestamp

	err := db.Pool.QueryRow(ctx, query, strings.TrimSpace(strings.ToLower(email))).Scan(
		&user.ID,
		&user.PublicID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.IsActive,
		&lastLogin,
		&user.FailedLoginAttempts,
		&passwordChangedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting user by email: %v", err)
	}

	// Asignar campos pgtype si son válidos
	if lastLogin.Valid {
		user.LastLogin = lastLogin
	}
	if passwordChangedAt.Valid {
		user.PasswordChangedAt = passwordChangedAt
	}

	return &user, nil
}

// GetUserByUsername obtiene un usuario por username
func (r *UserRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `
		SELECT id, public_id, username, email, password_hash, role, is_active,
		       last_login, failed_login_attempts, password_changed_at,
		       created_at, updated_at
		FROM users 
		WHERE username = $1 AND is_active = true
	`

	var user models.User
	var lastLogin pgtype.Timestamp
	var passwordChangedAt pgtype.Timestamp

	err := db.Pool.QueryRow(ctx, query, strings.TrimSpace(username)).Scan(
		&user.ID,
		&user.PublicID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.IsActive,
		&lastLogin,
		&user.FailedLoginAttempts,
		&passwordChangedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting user by username: %v", err)
	}

	// Asignar campos pgtype si son válidos
	if lastLogin.Valid {
		user.LastLogin = lastLogin
	}
	if passwordChangedAt.Valid {
		user.PasswordChangedAt = passwordChangedAt
	}

	return &user, nil
}

// UpdateUser actualiza un usuario existente
func (r *UserRepository) UpdateUser(ctx context.Context, id int64, name, email, role string) error {
	// Validaciones
	if err := r.validateUserUpdateData(name, email, role); err != nil {
		return err
	}

	// Limpiar datos
	name = strings.TrimSpace(name)
	email = strings.TrimSpace(strings.ToLower(email))
	role = strings.TrimSpace(strings.ToLower(role))

	// Generar username si name cambió
	username := r.generateUsername(name, email)

	query := `
		UPDATE users 
		SET username = $1, email = $2, role = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $4 AND is_active = true
	`

	result, err := db.Pool.Exec(ctx, query, username, email, role, id)
	if err != nil {
		if isDuplicateKeyError(err) {
			return fmt.Errorf("user with email %s already exists", email)
		}
		return fmt.Errorf("error updating user: %v", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found with id: %d", id)
	}

	log.Printf("User updated: %s (ID: %d)", email, id)
	return nil
}

// UpdatePassword actualiza la contraseña de un usuario
func (r *UserRepository) UpdatePassword(ctx context.Context, id int64, newPassword string) error {
	if strings.TrimSpace(newPassword) == "" {
		return fmt.Errorf("password cannot be empty")
	}

	if len(newPassword) < 6 {
		return fmt.Errorf("password must be at least 6 characters long")
	}

	// Hash de la nueva contraseña
	passwordHash, err := r.hashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("error hashing password: %v", err)
	}

	query := `
		UPDATE users 
		SET password_hash = $1, password_changed_at = CURRENT_TIMESTAMP, 
		    updated_at = CURRENT_TIMESTAMP, failed_login_attempts = 0
		WHERE id = $2 AND is_active = true
	`

	result, err := db.Pool.Exec(ctx, query, passwordHash, id)
	if err != nil {
		return fmt.Errorf("error updating password: %v", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found with id: %d", id)
	}

	log.Printf("Password updated for user ID: %d", id)
	return nil
}

// VerifyPassword verifica si la contraseña proporcionada coincide con el hash almacenado
func (r *UserRepository) VerifyPassword(ctx context.Context, email, password string) (*models.User, error) {
	user, err := r.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		// Incrementar contador de intentos fallidos
		r.incrementFailedLoginAttempts(ctx, user.ID)
		return nil, fmt.Errorf("invalid credentials")
	}

	// Reiniciar contador de intentos fallidos y actualizar last_login
	err = r.resetFailedLoginAttempts(ctx, user.ID)
	if err != nil {
		log.Printf("Error resetting failed login attempts for user %d: %v", user.ID, err)
	}

	return user, nil
}

// DeleteUser elimina un usuario (soft delete)
func (r *UserRepository) DeleteUser(ctx context.Context, id int64) error {
	query := `UPDATE users SET is_active = false, updated_at = CURRENT_TIMESTAMP WHERE id = $1`

	result, err := db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting user: %v", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found with id: %d", id)
	}

	log.Printf("User soft deleted: ID %d", id)
	return nil
}

// ListUsers lista todos los usuarios con paginación
func (r *UserRepository) ListUsers(ctx context.Context, limit, offset int) ([]*models.User, error) {
	query := `
		SELECT id, public_id, username, email, role, is_active, 
		       last_login, failed_login_attempts, created_at, updated_at
		FROM users 
		WHERE is_active = true
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := db.Pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error listing users: %v", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		var lastLogin pgtype.Timestamp

		err := rows.Scan(
			&user.ID,
			&user.PublicID,
			&user.Username,
			&user.Email,
			&user.Role,
			&user.IsActive,
			&lastLogin,
			&user.FailedLoginAttempts,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning user: %v", err)
		}

		if lastLogin.Valid {
			user.LastLogin = lastLogin
		}

		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %v", err)
	}

	return users, nil
}

// GetUsersByRole lista usuarios por rol
func (r *UserRepository) GetUsersByRole(ctx context.Context, role string, limit, offset int) ([]*models.User, error) {
	role = strings.TrimSpace(strings.ToLower(role))
	if !r.isValidRole(role) {
		return nil, fmt.Errorf("invalid role: %s", role)
	}

	query := `
		SELECT id, public_id, username, email, role, is_active, 
		       last_login, failed_login_attempts, created_at, updated_at
		FROM users 
		WHERE role = $1 AND is_active = true
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := db.Pool.Query(ctx, query, role, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error listing users by role: %v", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		var lastLogin pgtype.Timestamp

		err := rows.Scan(
			&user.ID,
			&user.PublicID,
			&user.Username,
			&user.Email,
			&user.Role,
			&user.IsActive,
			&lastLogin,
			&user.FailedLoginAttempts,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning user: %v", err)
		}

		if lastLogin.Valid {
			user.LastLogin = lastLogin
		}

		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %v", err)
	}

	return users, nil
}

// incrementFailedLoginAttempts incrementa el contador de intentos fallidos
func (r *UserRepository) incrementFailedLoginAttempts(ctx context.Context, userID int64) error {
	query := `
		UPDATE users 
		SET failed_login_attempts = failed_login_attempts + 1, 
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	_, err := db.Pool.Exec(ctx, query, userID)
	if err != nil {
		log.Printf("Error incrementing failed login attempts for user %d: %v", userID, err)
		return err
	}

	return nil
}

// resetFailedLoginAttempts reinicia el contador de intentos fallidos y actualiza last_login
func (r *UserRepository) resetFailedLoginAttempts(ctx context.Context, userID int64) error {
	query := `
		UPDATE users 
		SET failed_login_attempts = 0, last_login = CURRENT_TIMESTAMP,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	_, err := db.Pool.Exec(ctx, query, userID)
	if err != nil {
		log.Printf("Error resetting failed login attempts for user %d: %v", userID, err)
		return err
	}

	return nil
}

// Helper functions específicas del UserRepository

func (r *UserRepository) validateUserData(name, email, password, role string) error {
	name = strings.TrimSpace(name)
	email = strings.TrimSpace(strings.ToLower(email))
	password = strings.TrimSpace(password)
	role = strings.TrimSpace(strings.ToLower(role))

	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}

	if len(name) > 100 {
		return fmt.Errorf("name cannot exceed 100 characters")
	}

	if email == "" {
		return fmt.Errorf("email cannot be empty")
	}

	if len(email) > 255 {
		return fmt.Errorf("email cannot exceed 255 characters")
	}

	if !isValidEmail(email) {
		return fmt.Errorf("invalid email format: %s", email)
	}

	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	if len(password) < 6 {
		return fmt.Errorf("password must be at least 6 characters long")
	}

	if role == "" {
		role = "customer"
	}

	if !r.isValidRole(role) {
		return fmt.Errorf("invalid role. Must be: customer, organizer, or admin")
	}

	return nil
}

func (r *UserRepository) validateUserUpdateData(name, email, role string) error {
	name = strings.TrimSpace(name)
	email = strings.TrimSpace(strings.ToLower(email))
	role = strings.TrimSpace(strings.ToLower(role))

	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}

	if len(name) > 100 {
		return fmt.Errorf("name cannot exceed 100 characters")
	}

	if email == "" {
		return fmt.Errorf("email cannot be empty")
	}

	if len(email) > 255 {
		return fmt.Errorf("email cannot exceed 255 characters")
	}

	if !isValidEmail(email) {
		return fmt.Errorf("invalid email format: %s", email)
	}

	if role == "" {
		role = "customer"
	}

	if !r.isValidRole(role) {
		return fmt.Errorf("invalid role. Must be: customer, organizer, or admin")
	}

	return nil
}

func (r *UserRepository) isValidRole(role string) bool {
	return validRoles[role]
}

func (r *UserRepository) hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

func (r *UserRepository) generateUsername(name, email string) string {
	// Usar el nombre si está disponible, de lo contrario usar parte del email
	if name != "" {
		// Convertir nombre a formato username (minusculas, sin espacios)
		username := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(name), " ", "."))
		// Asegurar que no exceda la longitud máxima
		if len(username) > 100 {
			username = username[:100]
		}
		return username
	}

	// Usar la parte local del email como username
	parts := strings.Split(email, "@")
	if len(parts) > 0 {
		username := parts[0]
		if len(username) > 100 {
			username = username[:100]
		}
		return username
	}

	// Fallback: timestamp
	return fmt.Sprintf("user%d", time.Now().Unix())
}
