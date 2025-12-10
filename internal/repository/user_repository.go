// user_repository.go - COMPLETO Y CORREGIDO
package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/franciscozamorau/osmi-server/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// UserRepository maneja las operaciones de base de datos para usuarios
type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

// Valid roles (ACTUALIZADO con la nueva base de datos)
var validRoles = map[string]bool{
	"customer":  true,
	"organizer": true,
	"admin":     true,
	"guest":     true, // ✅ NUEVO ROL de la base de datos
}

// CreateUser crea un nuevo usuario en la base de datos (MEJORADO)
func (r *UserRepository) CreateUser(ctx context.Context, req *models.CreateUserRequest) (*models.User, error) {
	// Validaciones usando helpers
	if err := r.validateUserData(req.Username, req.Email, req.Password, req.Role); err != nil {
		return nil, err
	}

	// Limpiar y normalizar datos usando helpers
	username := strings.TrimSpace(req.Username)
	email := NormalizeEmail(req.Email) // ✅ USANDO HELPER
	role := strings.TrimSpace(strings.ToLower(req.Role))

	// Verificar si el usuario ya existe
	existing, err := r.GetUserByEmail(ctx, email)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("error checking existing user: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("user with email %s already exists", SafeStringForLog(email)) // ✅ USANDO HELPER
	}

	// Verificar si el username ya existe
	if username != "" {
		existingByUsername, err := r.GetUserByUsername(ctx, username)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("error checking existing username: %w", err)
		}
		if existingByUsername != nil {
			return nil, fmt.Errorf("username %s already exists", SafeStringForLog(username)) // ✅ USANDO HELPER
		}
	}

	// Generar public_id
	publicID := uuid.New().String()

	// Hash de la contraseña
	passwordHash, err := r.hashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("error hashing password: %w", err)
	}

	// Generar username si no se proporciona
	if username == "" {
		username = r.generateUsernameFromEmail(email)
	}

	query := `
		INSERT INTO users (
			public_id, username, email, password_hash, role, is_active,
			password_changed_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, public_id, username, email, role, is_active, 
		          last_login, failed_login_attempts, password_changed_at,
		          created_at, updated_at
	`

	var user models.User
	var dbLastLogin pgtype.Timestamp         // AGREGADO: variable para escanear
	var dbPasswordChangedAt pgtype.Timestamp // AGREGADO: variable para escanear

	err = r.db.QueryRow(ctx, query,
		publicID,
		username,
		email,
		passwordHash,
		role,
		true, // is_active por defecto
	).Scan(
		&user.ID,
		&user.PublicID,
		&user.Username,
		&user.Email,
		&user.Role,
		&user.IsActive,
		&dbLastLogin, // CAMBIADO: escanear a dbLastLogin
		&user.FailedLoginAttempts,
		&dbPasswordChangedAt, // CAMBIADO: escanear a dbPasswordChangedAt
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if IsDuplicateKeyError(err) { // ✅ USANDO HELPER
			return nil, fmt.Errorf("user with email %s or username %s already exists",
				SafeStringForLog(email), SafeStringForLog(username))
		}
		return nil, fmt.Errorf("error creating user: %w", err)
	}

	// Convertir pgtype a tipos nativos usando helpers
	user.LastLogin = ToTimeFromPgTimestamp(dbLastLogin)
	user.PasswordChangedAt = ToTimeFromPgTimestamp(dbPasswordChangedAt)

	log.Printf("User created: %s (ID: %d, PublicID: %s, Role: %s)",
		SafeStringForLog(email), user.ID, user.PublicID, role) // ✅ USANDO HELPER
	return &user, nil
}

// GetUserByID obtiene un usuario por su ID (MEJORADO)
func (r *UserRepository) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	query := `
		SELECT id, public_id, username, email, password_hash, role, is_active,
		       last_login, failed_login_attempts, password_changed_at,
		       created_at, updated_at
		FROM users 
		WHERE id = $1
	`

	var user models.User
	var dbLastLogin pgtype.Timestamp         // AGREGADO
	var dbPasswordChangedAt pgtype.Timestamp // AGREGADO

	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.PublicID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.IsActive,
		&dbLastLogin, // CAMBIADO
		&user.FailedLoginAttempts,
		&dbPasswordChangedAt, // CAMBIADO
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found with id: %d", id)
		}
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	// Convertir pgtype a tipos nativos usando helpers
	user.LastLogin = ToTimeFromPgTimestamp(dbLastLogin)
	user.PasswordChangedAt = ToTimeFromPgTimestamp(dbPasswordChangedAt)

	return &user, nil
}

// GetUserByPublicID obtiene un usuario por su public_id (MEJORADO)
func (r *UserRepository) GetUserByPublicID(ctx context.Context, publicID string) (*models.User, error) {
	if !IsValidUUID(publicID) { // ✅ USANDO HELPER
		return nil, fmt.Errorf("invalid user ID format")
	}

	query := `
		SELECT id, public_id, username, email, password_hash, role, is_active,
		       last_login, failed_login_attempts, password_changed_at,
		       created_at, updated_at
		FROM users 
		WHERE public_id = $1
	`

	var user models.User
	var dbLastLogin pgtype.Timestamp         // AGREGADO
	var dbPasswordChangedAt pgtype.Timestamp // AGREGADO

	err := r.db.QueryRow(ctx, query, publicID).Scan(
		&user.ID,
		&user.PublicID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.IsActive,
		&dbLastLogin, // CAMBIADO
		&user.FailedLoginAttempts,
		&dbPasswordChangedAt, // CAMBIADO
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found with public_id: %s", publicID)
		}
		return nil, fmt.Errorf("error getting user by public_id: %w", err)
	}

	// Convertir pgtype a tipos nativos usando helpers
	user.LastLogin = ToTimeFromPgTimestamp(dbLastLogin)
	user.PasswordChangedAt = ToTimeFromPgTimestamp(dbPasswordChangedAt)

	return &user, nil
}

// GetUserByEmail obtiene un usuario por email (MEJORADO)
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	normalizedEmail := NormalizeEmail(email) // ✅ USANDO HELPER

	query := `
		SELECT id, public_id, username, email, password_hash, role, is_active,
		       last_login, failed_login_attempts, password_changed_at,
		       created_at, updated_at
		FROM users 
		WHERE email = $1
	`

	var user models.User
	var dbLastLogin pgtype.Timestamp         // AGREGADO
	var dbPasswordChangedAt pgtype.Timestamp // AGREGADO

	err := r.db.QueryRow(ctx, query, normalizedEmail).Scan(
		&user.ID,
		&user.PublicID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.IsActive,
		&dbLastLogin, // CAMBIADO
		&user.FailedLoginAttempts,
		&dbPasswordChangedAt, // CAMBIADO
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting user by email: %w", err)
	}

	// Convertir pgtype a tipos nativos usando helpers
	user.LastLogin = ToTimeFromPgTimestamp(dbLastLogin)
	user.PasswordChangedAt = ToTimeFromPgTimestamp(dbPasswordChangedAt)

	return &user, nil
}

// GetUserByUsername obtiene un usuario por username (MEJORADO)
func (r *UserRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	username = strings.TrimSpace(username)

	query := `
		SELECT id, public_id, username, email, password_hash, role, is_active,
		       last_login, failed_login_attempts, password_changed_at,
		       created_at, updated_at
		FROM users 
		WHERE username = $1
	`

	var user models.User
	var dbLastLogin pgtype.Timestamp         // AGREGADO
	var dbPasswordChangedAt pgtype.Timestamp // AGREGADO

	err := r.db.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.PublicID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.IsActive,
		&dbLastLogin, // CAMBIADO
		&user.FailedLoginAttempts,
		&dbPasswordChangedAt, // CAMBIADO
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting user by username: %w", err)
	}

	// Convertir pgtype a tipos nativos usando helpers
	user.LastLogin = ToTimeFromPgTimestamp(dbLastLogin)
	user.PasswordChangedAt = ToTimeFromPgTimestamp(dbPasswordChangedAt)

	return &user, nil
}

// UpdateUser actualiza un usuario existente (MEJORADO)
func (r *UserRepository) UpdateUser(ctx context.Context, publicID string, username, email, role string) error {
	if !IsValidUUID(publicID) { // ✅ USANDO HELPER
		return fmt.Errorf("invalid user ID format")
	}

	// Validaciones
	if err := r.validateUserUpdateData(username, email, role); err != nil {
		return err
	}

	// Limpiar datos usando helpers
	username = strings.TrimSpace(username)
	email = NormalizeEmail(email) // ✅ USANDO HELPER
	role = strings.TrimSpace(strings.ToLower(role))

	query := `
		UPDATE users 
		SET username = $1, email = $2, role = $3, updated_at = CURRENT_TIMESTAMP
		WHERE public_id = $4
	`

	result, err := r.db.Exec(ctx, query, username, email, role, publicID)
	if err != nil {
		if IsDuplicateKeyError(err) { // ✅ USANDO HELPER
			return fmt.Errorf("user with email %s or username %s already exists",
				SafeStringForLog(email), SafeStringForLog(username))
		}
		return fmt.Errorf("error updating user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found with public_id: %s", publicID)
	}

	log.Printf("User updated: %s (PublicID: %s)", SafeStringForLog(email), publicID) // ✅ USANDO HELPER
	return nil
}

// UpdatePassword actualiza la contraseña de un usuario (MEJORADO)
func (r *UserRepository) UpdatePassword(ctx context.Context, publicID string, newPassword string) error {
	if !IsValidUUID(publicID) { // ✅ USANDO HELPER
		return fmt.Errorf("invalid user ID format")
	}

	if strings.TrimSpace(newPassword) == "" {
		return fmt.Errorf("password cannot be empty")
	}

	if len(newPassword) < 6 {
		return fmt.Errorf("password must be at least 6 characters long")
	}

	// Hash de la nueva contraseña
	passwordHash, err := r.hashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("error hashing password: %w", err)
	}

	query := `
		UPDATE users 
		SET password_hash = $1, password_changed_at = CURRENT_TIMESTAMP, 
		    updated_at = CURRENT_TIMESTAMP, failed_login_attempts = 0
		WHERE public_id = $2
	`

	result, err := r.db.Exec(ctx, query, passwordHash, publicID)
	if err != nil {
		return fmt.Errorf("error updating password: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found with public_id: %s", publicID)
	}

	log.Printf("Password updated for user: %s", publicID)
	return nil
}

// VerifyPassword verifica si la contraseña proporcionada coincide con el hash almacenado (MEJORADO)
func (r *UserRepository) VerifyPassword(ctx context.Context, email, password string) (*models.User, error) {
	normalizedEmail := NormalizeEmail(email) // ✅ USANDO HELPER
	user, err := r.GetUserByEmail(ctx, normalizedEmail)
	if err != nil || user == nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if !user.IsActive {
		return nil, fmt.Errorf("user account is inactive")
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

// DeleteUser elimina un usuario (soft delete) (MEJORADO)
func (r *UserRepository) DeleteUser(ctx context.Context, publicID string) error {
	if !IsValidUUID(publicID) { // ✅ USANDO HELPER
		return fmt.Errorf("invalid user ID format")
	}

	query := `UPDATE users SET is_active = false, updated_at = CURRENT_TIMESTAMP WHERE public_id = $1`

	result, err := r.db.Exec(ctx, query, publicID)
	if err != nil {
		return fmt.Errorf("error deleting user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found with public_id: %s", publicID)
	}

	log.Printf("User soft deleted: %s", publicID)
	return nil
}

// ListUsers lista todos los usuarios con paginación (MEJORADO)
func (r *UserRepository) ListUsers(ctx context.Context, includeInactive bool, limit, offset int) ([]*models.User, error) {
	// Validar parámetros de paginación
	if limit <= 0 || limit > 100 {
		limit = 50 // límite por defecto
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT id, public_id, username, email, role, is_active, 
		       last_login, failed_login_attempts, created_at, updated_at
		FROM users 
		WHERE ($1 = true OR is_active = true)
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, includeInactive, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error listing users: %w", err)
	}
	defer rows.Close()

	return r.scanUsersFromRows(rows)
}

// GetUsersByRole lista usuarios por rol (MEJORADO)
func (r *UserRepository) GetUsersByRole(ctx context.Context, role string, includeInactive bool, limit, offset int) ([]*models.User, error) {
	role = strings.TrimSpace(strings.ToLower(role))
	if !r.isValidRole(role) {
		return nil, fmt.Errorf("invalid role: %s", role)
	}

	// Validar parámetros de paginación
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT id, public_id, username, email, role, is_active, 
		       last_login, failed_login_attempts, created_at, updated_at
		FROM users 
		WHERE role = $1 AND ($2 = true OR is_active = true)
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.Query(ctx, query, role, includeInactive, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error listing users by role: %w", err)
	}
	defer rows.Close()

	return r.scanUsersFromRows(rows)
}

// GetUserStats obtiene estadísticas de usuarios
func (r *UserRepository) GetUserStats(ctx context.Context) (*models.UserStats, error) {
	query := `
		SELECT 
			COUNT(*) as total_users,
			COUNT(CASE WHEN is_active = true THEN 1 END) as active_users,
			COUNT(CASE WHEN is_active = false THEN 1 END) as inactive_users,
			COUNT(CASE WHEN role = 'customer' THEN 1 END) as customer_users,
			COUNT(CASE WHEN role = 'organizer' THEN 1 END) as organizer_users,
			COUNT(CASE WHEN role = 'admin' THEN 1 END) as admin_users,
			COUNT(CASE WHEN role = 'guest' THEN 1 END) as guest_users,
			COUNT(CASE WHEN last_login >= CURRENT_DATE - INTERVAL '30 days' THEN 1 END) as active_last_30_days
		FROM users
	`

	var stats models.UserStats
	err := r.db.QueryRow(ctx, query).Scan(
		&stats.TotalUsers,
		&stats.ActiveUsers,
		&stats.InactiveUsers,
		&stats.CustomerUsers,
		&stats.OrganizerUsers,
		&stats.AdminUsers,
		&stats.GuestUsers,
		&stats.ActiveLast30Days,
	)

	if err != nil {
		return nil, fmt.Errorf("error getting user stats: %w", err)
	}

	return &stats, nil
}

// SearchUsers busca usuarios por término
func (r *UserRepository) SearchUsers(ctx context.Context, searchTerm string, limit int) ([]*models.User, error) {
	if limit <= 0 {
		limit = 20
	}

	query := `
		SELECT id, public_id, username, email, role, is_active, 
		       last_login, failed_login_attempts, created_at, updated_at
		FROM users 
		WHERE (username ILIKE $1 OR email ILIKE $1 OR public_id::text ILIKE $1)
		ORDER BY 
			CASE 
				WHEN username ILIKE $1 THEN 1
				WHEN email ILIKE $1 THEN 2
				ELSE 3
			END,
			created_at DESC
		LIMIT $2
	`

	searchPattern := "%" + strings.TrimSpace(searchTerm) + "%"
	rows, err := r.db.Query(ctx, query, searchPattern, limit)
	if err != nil {
		return nil, fmt.Errorf("error searching users: %w", err)
	}
	defer rows.Close()

	return r.scanUsersFromRows(rows)
}

// GetActiveUsersCount obtiene el número de usuarios activos (NUEVO MÉTODO)
func (r *UserRepository) GetActiveUsersCount(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE is_active = true").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error getting active users count: %w", err)
	}
	return count, nil
}

// UpdateLastLogin actualiza la última fecha de login (NUEVO MÉTODO)
func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID int64) error {
	query := `
		UPDATE users 
		SET last_login = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("error updating last login: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found with id: %d", userID)
	}

	return nil
}

// ActivateUser activa un usuario (NUEVO MÉTODO)
func (r *UserRepository) ActivateUser(ctx context.Context, publicID string) error {
	if !IsValidUUID(publicID) {
		return fmt.Errorf("invalid user ID format")
	}

	query := `UPDATE users SET is_active = true, updated_at = CURRENT_TIMESTAMP WHERE public_id = $1`

	result, err := r.db.Exec(ctx, query, publicID)
	if err != nil {
		return fmt.Errorf("error activating user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found with public_id: %s", publicID)
	}

	log.Printf("User activated: %s", publicID)
	return nil
}

// DeactivateUser desactiva un usuario (NUEVO MÉTODO)
func (r *UserRepository) DeactivateUser(ctx context.Context, publicID string) error {
	if !IsValidUUID(publicID) {
		return fmt.Errorf("invalid user ID format")
	}

	query := `UPDATE users SET is_active = false, updated_at = CURRENT_TIMESTAMP WHERE public_id = $1`

	result, err := r.db.Exec(ctx, query, publicID)
	if err != nil {
		return fmt.Errorf("error deactivating user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found with public_id: %s", publicID)
	}

	log.Printf("User deactivated: %s", publicID)
	return nil
}

// =============================================================================
// MÉTODOS PRIVADOS
// =============================================================================

// scanUsersFromRows escanea filas de usuarios (método helper reutilizable)
func (r *UserRepository) scanUsersFromRows(rows pgx.Rows) ([]*models.User, error) {
	var users []*models.User

	for rows.Next() {
		var user models.User
		var dbLastLogin pgtype.Timestamp // AGREGADO

		err := rows.Scan(
			&user.ID,
			&user.PublicID,
			&user.Username,
			&user.Email,
			&user.Role,
			&user.IsActive,
			&dbLastLogin, // CAMBIADO
			&user.FailedLoginAttempts,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning user: %w", err)
		}

		// Convertir pgtype a tipos nativos usando helpers
		user.LastLogin = ToTimeFromPgTimestamp(dbLastLogin)

		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
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

	_, err := r.db.Exec(ctx, query, userID)
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

	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		log.Printf("Error resetting failed login attempts for user %d: %v", userID, err)
		return err
	}

	return nil
}

// Helper functions específicas del UserRepository
func (r *UserRepository) validateUserData(username, email, password, role string) error {
	username = strings.TrimSpace(username)
	email = NormalizeEmail(email) // ✅ USANDO HELPER
	password = strings.TrimSpace(password)
	role = strings.TrimSpace(strings.ToLower(role))

	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	if len(username) > 100 {
		return fmt.Errorf("username cannot exceed 100 characters")
	}

	if email == "" {
		return fmt.Errorf("email cannot be empty")
	}

	if len(email) > 255 {
		return fmt.Errorf("email cannot exceed 255 characters")
	}

	if !IsValidEmail(email) { // ✅ USANDO HELPER
		return fmt.Errorf("invalid email format: %s", SafeStringForLog(email))
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
		return fmt.Errorf("invalid role. Must be: customer, organizer, admin, or guest")
	}

	return nil
}

func (r *UserRepository) validateUserUpdateData(username, email, role string) error {
	username = strings.TrimSpace(username)
	email = NormalizeEmail(email) // ✅ USANDO HELPER
	role = strings.TrimSpace(strings.ToLower(role))

	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	if len(username) > 100 {
		return fmt.Errorf("username cannot exceed 100 characters")
	}

	if email == "" {
		return fmt.Errorf("email cannot be empty")
	}

	if len(email) > 255 {
		return fmt.Errorf("email cannot exceed 255 characters")
	}

	if !IsValidEmail(email) { // ✅ USANDO HELPER
		return fmt.Errorf("invalid email format: %s", SafeStringForLog(email))
	}

	if role == "" {
		role = "customer"
	}

	if !r.isValidRole(role) {
		return fmt.Errorf("invalid role. Must be: customer, organizer, admin, or guest")
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

func (r *UserRepository) generateUsernameFromEmail(email string) string {
	// Usar la parte local del email como username
	parts := strings.Split(email, "@")
	if len(parts) > 0 {
		username := parts[0]
		// Limpiar username (solo letras, números, puntos y guiones bajos)
		cleanUsername := ""
		for _, char := range username {
			if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '.' || char == '_' {
				cleanUsername += string(char)
			}
		}

		if cleanUsername == "" {
			cleanUsername = "user"
		}

		// Asegurar longitud máxima
		if len(cleanUsername) > 100 {
			cleanUsername = cleanUsername[:100]
		}

		// Verificar si el username ya existe
		existing, _ := r.GetUserByUsername(context.Background(), cleanUsername)
		if existing == nil {
			return cleanUsername
		}

		// Si existe, agregar sufijo numérico
		return fmt.Sprintf("%s%d", cleanUsername, time.Now().Unix()%10000)
	}

	// Fallback: timestamp
	return fmt.Sprintf("user%d", time.Now().Unix())
}
