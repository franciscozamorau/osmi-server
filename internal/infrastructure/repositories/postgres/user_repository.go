// internal/infrastructure/repositories/postgres/user_repository.go
package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
	"github.com/franciscozamorau/osmi-server/internal/domain/enums"
	"github.com/franciscozamorau/osmi-server/internal/domain/repository"
)

// UserRepository implementa la interfaz repository.UserRepository usando PostgreSQL
type UserRepository struct {
	db *pgxpool.Pool
}

// NewUserRepository crea una nueva instancia del repositorio
func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

// handleError mapea errores de PostgreSQL a nuestros errores de dominio
func (r *UserRepository) handleError(err error, context string) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return repository.ErrUserNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // Unique violation
			if strings.Contains(pgErr.ConstraintName, "users_email_key") {
				return repository.ErrUserEmailExists
			}
			if strings.Contains(pgErr.ConstraintName, "users_username_key") {
				return repository.ErrUserUsernameExists
			}
			if strings.Contains(pgErr.ConstraintName, "users_public_uuid_key") {
				return repository.ErrUserEmailExists // despues crear un error específico
			}
		}
	}

	return fmt.Errorf("%s: %w", context, err)
}

// Find busca usuarios según los criterios del filtro
func (r *UserRepository) Find(ctx context.Context, filter *repository.UserFilter) ([]*entities.User, int64, error) {
	baseQuery := `
		SELECT 
			id, public_uuid, email, phone, username, password_hash,
			first_name, last_name, full_name, avatar_url, date_of_birth,
			email_verified, phone_verified, verified_at,
			preferred_language, preferred_currency, timezone,
			mfa_enabled, mfa_secret, last_login_at, last_login_ip,
			failed_login_attempts, locked_until,
			is_active, is_staff, is_superuser, role,
			last_active_at, created_at, updated_at
		FROM auth.users
		WHERE 1=1
	`

	countQuery := `SELECT COUNT(*) FROM auth.users WHERE 1=1`

	var conditions []string
	args := pgx.NamedArgs{}
	argPos := 1

	if filter != nil {
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

		if filter.Email != nil {
			conditions = append(conditions, fmt.Sprintf("email = @email_%d", argPos))
			args[fmt.Sprintf("email_%d", argPos)] = *filter.Email
			argPos++
		}

		if filter.Username != nil {
			conditions = append(conditions, fmt.Sprintf("username = @username_%d", argPos))
			args[fmt.Sprintf("username_%d", argPos)] = *filter.Username
			argPos++
		}

		if filter.SearchTerm != nil && *filter.SearchTerm != "" {
			searchTerm := "%" + *filter.SearchTerm + "%"
			conditions = append(conditions, fmt.Sprintf(
				"(email ILIKE @search_%d OR username ILIKE @search_%d OR first_name ILIKE @search_%d OR last_name ILIKE @search_%d)",
				argPos, argPos, argPos, argPos,
			))
			args[fmt.Sprintf("search_%d", argPos)] = searchTerm
			argPos++
		}

		if filter.FirstName != nil {
			conditions = append(conditions, fmt.Sprintf("first_name ILIKE @first_%d", argPos))
			args[fmt.Sprintf("first_%d", argPos)] = "%" + *filter.FirstName + "%"
			argPos++
		}

		if filter.LastName != nil {
			conditions = append(conditions, fmt.Sprintf("last_name ILIKE @last_%d", argPos))
			args[fmt.Sprintf("last_%d", argPos)] = "%" + *filter.LastName + "%"
			argPos++
		}

		if filter.Role != nil {
			conditions = append(conditions, fmt.Sprintf("role = @role_%d", argPos))
			args[fmt.Sprintf("role_%d", argPos)] = filter.Role.String()
			argPos++
		}

		if filter.IsActive != nil {
			conditions = append(conditions, fmt.Sprintf("is_active = @active_%d", argPos))
			args[fmt.Sprintf("active_%d", argPos)] = *filter.IsActive
			argPos++
		}

		if filter.IsStaff != nil {
			conditions = append(conditions, fmt.Sprintf("is_staff = @staff_%d", argPos))
			args[fmt.Sprintf("staff_%d", argPos)] = *filter.IsStaff
			argPos++
		}

		if filter.IsSuperuser != nil {
			conditions = append(conditions, fmt.Sprintf("is_superuser = @super_%d", argPos))
			args[fmt.Sprintf("super_%d", argPos)] = *filter.IsSuperuser
			argPos++
		}

		if filter.EmailVerified != nil {
			conditions = append(conditions, fmt.Sprintf("email_verified = @email_ver_%d", argPos))
			args[fmt.Sprintf("email_ver_%d", argPos)] = *filter.EmailVerified
			argPos++
		}

		if filter.PhoneVerified != nil {
			conditions = append(conditions, fmt.Sprintf("phone_verified = @phone_ver_%d", argPos))
			args[fmt.Sprintf("phone_ver_%d", argPos)] = *filter.PhoneVerified
			argPos++
		}

		if filter.MFAEnabled != nil {
			conditions = append(conditions, fmt.Sprintf("mfa_enabled = @mfa_%d", argPos))
			args[fmt.Sprintf("mfa_%d", argPos)] = *filter.MFAEnabled
			argPos++
		}

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

		if filter.LastLoginFrom != nil {
			conditions = append(conditions, fmt.Sprintf("last_login_at >= @login_from_%d", argPos))
			args[fmt.Sprintf("login_from_%d", argPos)] = *filter.LastLoginFrom
			argPos++
		}

		if filter.LastLoginTo != nil {
			conditions = append(conditions, fmt.Sprintf("last_login_at <= @login_to_%d", argPos))
			args[fmt.Sprintf("login_to_%d", argPos)] = *filter.LastLoginTo
			argPos++
		}
	}

	if len(conditions) > 0 {
		whereClause := " AND " + strings.Join(conditions, " AND ")
		baseQuery += whereClause
		countQuery += whereClause
	}

	var total int64
	err := r.db.QueryRow(ctx, countQuery, args).Scan(&total)
	if err != nil {
		return nil, 0, r.handleError(err, "failed to count users")
	}

	if filter != nil {
		sortBy := "created_at"
		sortOrder := "DESC"
		if filter.SortBy != "" {
			allowedSortColumns := map[string]bool{
				"created_at":    true,
				"last_login_at": true,
				"email":         true,
				"username":      true,
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
			baseQuery += " LIMIT @limit"
			args["limit"] = filter.Limit
		}
		if filter.Offset > 0 {
			baseQuery += " OFFSET @offset"
			args["offset"] = filter.Offset
		}
	}

	rows, err := r.db.Query(ctx, baseQuery, args)
	if err != nil {
		return nil, 0, r.handleError(err, "failed to find users")
	}
	defer rows.Close()

	var users []*entities.User
	for rows.Next() {
		var user entities.User
		var phone, username, firstName, lastName, fullName, avatarURL *string
		var dateOfBirth, verifiedAt, lastLoginAt, lockedUntil, lastActiveAt *time.Time
		var lastLoginIP *string
		var mfaSecret *string

		err = rows.Scan(
			&user.ID, &user.PublicID, &user.Email, &phone, &username, &user.PasswordHash,
			&firstName, &lastName, &fullName, &avatarURL, &dateOfBirth,
			&user.EmailVerified, &user.PhoneVerified, &verifiedAt,
			&user.PreferredLanguage, &user.PreferredCurrency, &user.Timezone,
			&user.MFAEnabled, &mfaSecret, &lastLoginAt, &lastLoginIP,
			&user.FailedLoginAttempts, &lockedUntil,
			&user.IsActive, &user.IsStaff, &user.IsSuperuser, &user.Role,
			&lastActiveAt, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, 0, r.handleError(err, "failed to scan user row")
		}

		user.Phone = phone
		user.Username = username
		user.FirstName = firstName
		user.LastName = lastName
		user.FullName = fullName
		user.AvatarURL = avatarURL
		user.DateOfBirth = dateOfBirth
		user.VerifiedAt = verifiedAt
		user.LastLoginAt = lastLoginAt
		user.LastLoginIP = lastLoginIP
		user.LockedUntil = lockedUntil
		user.LastActiveAt = lastActiveAt
		user.MFASecret = mfaSecret

		users = append(users, &user)
	}

	return users, total, nil
}

// GetByID obtiene un usuario por su ID numérico
func (r *UserRepository) GetByID(ctx context.Context, id int64) (*entities.User, error) {
	filter := &repository.UserFilter{
		IDs:   []int64{id},
		Limit: 1,
	}

	users, _, err := r.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, repository.ErrUserNotFound
	}

	return users[0], nil
}

// GetByPublicID obtiene un usuario por su UUID público
func (r *UserRepository) GetByPublicID(ctx context.Context, publicID string) (*entities.User, error) {
	filter := &repository.UserFilter{
		PublicIDs: []string{publicID},
		Limit:     1,
	}

	users, _, err := r.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, repository.ErrUserNotFound
	}

	return users[0], nil
}

// GetByEmail obtiene un usuario por su email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	filter := &repository.UserFilter{
		Email: &email,
		Limit: 1,
	}

	users, _, err := r.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, repository.ErrUserNotFound
	}

	return users[0], nil
}

// GetByUsername obtiene un usuario por su username
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*entities.User, error) {
	filter := &repository.UserFilter{
		Username: &username,
		Limit:    1,
	}

	users, _, err := r.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, repository.ErrUserNotFound
	}

	return users[0], nil
}

// Create inserta un nuevo usuario
func (r *UserRepository) Create(ctx context.Context, user *entities.User) error {
	query := `
		INSERT INTO auth.users (
			public_uuid, email, phone, username, password_hash,
			first_name, last_name, full_name, avatar_url, date_of_birth,
			email_verified, phone_verified, verified_at,
			preferred_language, preferred_currency, timezone,
			mfa_enabled, mfa_secret, last_login_at, last_login_ip,
			failed_login_attempts, locked_until,
			is_active, is_staff, is_superuser, role,
			last_active_at, created_at, updated_at
		) VALUES (
			gen_random_uuid(), $1, $2, $3, $4,
			$5, $6, $7, $8, $9,
			$10, $11, $12, $13, $14, $15,
			$16, $17, $18, $19,
			$20, $21, $22, $23, $24, $25,
			$26, NOW(), NOW()
		)
		RETURNING id, public_uuid, created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query,
		user.Email, user.Phone, user.Username, user.PasswordHash,
		user.FirstName, user.LastName, user.FullName, user.AvatarURL, user.DateOfBirth,
		user.EmailVerified, user.PhoneVerified, user.VerifiedAt,
		user.PreferredLanguage, user.PreferredCurrency, user.Timezone,
		user.MFAEnabled, user.MFASecret, user.LastLoginAt, user.LastLoginIP,
		user.FailedLoginAttempts, user.LockedUntil,
		user.IsActive, user.IsStaff, user.IsSuperuser, user.Role,
		user.LastActiveAt,
	).Scan(&user.ID, &user.PublicID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return r.handleError(err, "failed to create user")
	}

	return nil
}

// Update actualiza un usuario existente
func (r *UserRepository) Update(ctx context.Context, user *entities.User) error {
	query := `
		UPDATE auth.users SET
			email = $1,
			phone = $2,
			username = $3,
			first_name = $4,
			last_name = $5,
			full_name = $6,
			avatar_url = $7,
			date_of_birth = $8,
			preferred_language = $9,
			preferred_currency = $10,
			timezone = $11,
			mfa_enabled = $12,
			mfa_secret = $13,
			is_active = $14,
			is_staff = $15,
			is_superuser = $16,
			role = $17,
			last_active_at = $18,
			updated_at = NOW()
		WHERE id = $19
		RETURNING updated_at
	`

	err := r.db.QueryRow(ctx, query,
		user.Email, user.Phone, user.Username,
		user.FirstName, user.LastName, user.FullName, user.AvatarURL, user.DateOfBirth,
		user.PreferredLanguage, user.PreferredCurrency, user.Timezone,
		user.MFAEnabled, user.MFASecret,
		user.IsActive, user.IsStaff, user.IsSuperuser, user.Role,
		user.LastActiveAt,
		user.ID,
	).Scan(&user.UpdatedAt)

	if err != nil {
		return r.handleError(err, "failed to update user")
	}

	return nil
}

// Delete elimina permanentemente un usuario
func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	cmdTag, err := r.db.Exec(ctx, `DELETE FROM auth.users WHERE id = $1`, id)
	if err != nil {
		return r.handleError(err, "failed to delete user")
	}

	if cmdTag.RowsAffected() == 0 {
		return repository.ErrUserNotFound
	}

	return nil
}

// SoftDelete desactiva un usuario (soft delete)
func (r *UserRepository) SoftDelete(ctx context.Context, publicID string) error {
	query := `
		UPDATE auth.users 
		SET is_active = false, updated_at = NOW()
		WHERE public_uuid = $1 AND is_active = true
	`
	cmdTag, err := r.db.Exec(ctx, query, publicID)
	if err != nil {
		return r.handleError(err, "failed to soft delete user")
	}

	if cmdTag.RowsAffected() == 0 {
		return repository.ErrUserNotFound
	}

	return nil
}

// Exists verifica si existe un usuario con el ID dado
func (r *UserRepository) Exists(ctx context.Context, id int64) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM auth.users WHERE id = $1)`, id).Scan(&exists)
	if err != nil {
		return false, r.handleError(err, "failed to check user existence")
	}
	return exists, nil
}

// ExistsByEmail verifica si existe un usuario con el email dado
func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM auth.users WHERE email = $1)`, email).Scan(&exists)
	if err != nil {
		return false, r.handleError(err, "failed to check email existence")
	}
	return exists, nil
}

// ExistsByUsername verifica si existe un usuario con el username dado
func (r *UserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM auth.users WHERE username = $1)`, username).Scan(&exists)
	if err != nil {
		return false, r.handleError(err, "failed to check username existence")
	}
	return exists, nil
}

// UpdatePassword actualiza la contraseña del usuario
func (r *UserRepository) UpdatePassword(ctx context.Context, userID int64, passwordHash string) error {
	cmdTag, err := r.db.Exec(ctx, `
		UPDATE auth.users 
		SET password_hash = $1, updated_at = NOW()
		WHERE id = $2
	`, passwordHash, userID)
	if err != nil {
		return r.handleError(err, "failed to update password")
	}

	if cmdTag.RowsAffected() == 0 {
		return repository.ErrUserNotFound
	}

	return nil
}

// UpdateLastLogin actualiza la información del último login
func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID int64, ipAddress string) error {
	cmdTag, err := r.db.Exec(ctx, `
		UPDATE auth.users 
		SET last_login_at = NOW(),
			last_login_ip = $1,
			last_active_at = NOW(),
			failed_login_attempts = 0,
			updated_at = NOW()
		WHERE id = $2
	`, ipAddress, userID)
	if err != nil {
		return r.handleError(err, "failed to update last login")
	}

	if cmdTag.RowsAffected() == 0 {
		return repository.ErrUserNotFound
	}

	return nil
}

// IncrementFailedAttempts incrementa el contador de intentos fallidos
func (r *UserRepository) IncrementFailedAttempts(ctx context.Context, userID int64) error {
	_, err := r.db.Exec(ctx, `
		UPDATE auth.users 
		SET failed_login_attempts = failed_login_attempts + 1,
			updated_at = NOW()
		WHERE id = $1
	`, userID)
	if err != nil {
		return r.handleError(err, "failed to increment failed attempts")
	}
	return nil
}

// ResetFailedAttempts resetea el contador de intentos fallidos
func (r *UserRepository) ResetFailedAttempts(ctx context.Context, userID int64) error {
	_, err := r.db.Exec(ctx, `
		UPDATE auth.users 
		SET failed_login_attempts = 0,
			updated_at = NOW()
		WHERE id = $1
	`, userID)
	if err != nil {
		return r.handleError(err, "failed to reset failed attempts")
	}
	return nil
}

// LockUser bloquea un usuario hasta una fecha específica
func (r *UserRepository) LockUser(ctx context.Context, userID int64, until time.Time) error {
	cmdTag, err := r.db.Exec(ctx, `
		UPDATE auth.users 
		SET locked_until = $1,
			updated_at = NOW()
		WHERE id = $2
	`, until, userID)
	if err != nil {
		return r.handleError(err, "failed to lock user")
	}

	if cmdTag.RowsAffected() == 0 {
		return repository.ErrUserNotFound
	}

	return nil
}

// UnlockUser desbloquea un usuario
func (r *UserRepository) UnlockUser(ctx context.Context, userID int64) error {
	cmdTag, err := r.db.Exec(ctx, `
		UPDATE auth.users 
		SET locked_until = NULL,
			failed_login_attempts = 0,
			updated_at = NOW()
		WHERE id = $1
	`, userID)
	if err != nil {
		return r.handleError(err, "failed to unlock user")
	}

	if cmdTag.RowsAffected() == 0 {
		return repository.ErrUserNotFound
	}

	return nil
}

// VerifyEmail marca el email como verificado
func (r *UserRepository) VerifyEmail(ctx context.Context, userID int64) error {
	now := time.Now()
	cmdTag, err := r.db.Exec(ctx, `
		UPDATE auth.users 
		SET email_verified = true,
			verified_at = $1,
			updated_at = NOW()
		WHERE id = $2
	`, now, userID)
	if err != nil {
		return r.handleError(err, "failed to verify email")
	}

	if cmdTag.RowsAffected() == 0 {
		return repository.ErrUserNotFound
	}

	return nil
}

// VerifyPhone marca el teléfono como verificado
func (r *UserRepository) VerifyPhone(ctx context.Context, userID int64) error {
	cmdTag, err := r.db.Exec(ctx, `
		UPDATE auth.users 
		SET phone_verified = true,
			updated_at = NOW()
		WHERE id = $1
	`, userID)
	if err != nil {
		return r.handleError(err, "failed to verify phone")
	}

	if cmdTag.RowsAffected() == 0 {
		return repository.ErrUserNotFound
	}

	return nil
}

// EnableMFA habilita la autenticación de dos factores
func (r *UserRepository) EnableMFA(ctx context.Context, userID int64, secret string) error {
	cmdTag, err := r.db.Exec(ctx, `
		UPDATE auth.users 
		SET mfa_enabled = true,
			mfa_secret = $1,
			updated_at = NOW()
		WHERE id = $2
	`, secret, userID)
	if err != nil {
		return r.handleError(err, "failed to enable MFA")
	}

	if cmdTag.RowsAffected() == 0 {
		return repository.ErrUserNotFound
	}

	return nil
}

// DisableMFA deshabilita la autenticación de dos factores
func (r *UserRepository) DisableMFA(ctx context.Context, userID int64) error {
	cmdTag, err := r.db.Exec(ctx, `
		UPDATE auth.users 
		SET mfa_enabled = false,
			mfa_secret = NULL,
			updated_at = NOW()
		WHERE id = $1
	`, userID)
	if err != nil {
		return r.handleError(err, "failed to disable MFA")
	}

	if cmdTag.RowsAffected() == 0 {
		return repository.ErrUserNotFound
	}

	return nil
}

// UpdatePreferences actualiza las preferencias del usuario
func (r *UserRepository) UpdatePreferences(ctx context.Context, userID int64, preferences map[string]interface{}) error {
	lang, _ := preferences["language"].(string)
	currency, _ := preferences["currency"].(string)
	timezone, _ := preferences["timezone"].(string)

	cmdTag, err := r.db.Exec(ctx, `
		UPDATE auth.users 
		SET preferred_language = COALESCE($1, preferred_language),
			preferred_currency = COALESCE($2, preferred_currency),
			timezone = COALESCE($3, timezone),
			updated_at = NOW()
		WHERE id = $4
	`, lang, currency, timezone, userID)
	if err != nil {
		return r.handleError(err, "failed to update preferences")
	}

	if cmdTag.RowsAffected() == 0 {
		return repository.ErrUserNotFound
	}

	return nil
}

// GetStats obtiene estadísticas agregadas de usuarios
func (r *UserRepository) GetStats(ctx context.Context) (*repository.UserStats, error) {
	query := `
		SELECT 
			COUNT(*) as total_users,
			COUNT(CASE WHEN is_active = true THEN 1 END) as active_users,
			COUNT(CASE WHEN is_staff = true THEN 1 END) as staff_users,
			COUNT(CASE WHEN is_superuser = true THEN 1 END) as superusers,
			COUNT(CASE WHEN email_verified = true THEN 1 END) as email_verified_users,
			COUNT(CASE WHEN phone_verified = true THEN 1 END) as phone_verified_users,
			COUNT(CASE WHEN mfa_enabled = true THEN 1 END) as mfa_enabled_users,
			COUNT(CASE WHEN created_at >= NOW() - INTERVAL '7 days' THEN 1 END) as new_users_last_7_days,
			COUNT(CASE WHEN created_at >= NOW() - INTERVAL '30 days' THEN 1 END) as new_users_last_30_days,
			COUNT(CASE WHEN last_login_at >= NOW() - INTERVAL '7 days' THEN 1 END) as active_last_7_days,
			COUNT(CASE WHEN last_login_at >= NOW() - INTERVAL '30 days' THEN 1 END) as active_last_30_days
		FROM auth.users
	`

	var stats repository.UserStats
	err := r.db.QueryRow(ctx, query).Scan(
		&stats.TotalUsers,
		&stats.ActiveUsers,
		&stats.StaffUsers,
		&stats.Superusers,
		&stats.EmailVerifiedUsers,
		&stats.PhoneVerifiedUsers,
		&stats.MFAEnabledUsers,
		&stats.NewUsersLast7Days,
		&stats.NewUsersLast30Days,
		&stats.ActiveLast7Days,
		&stats.ActiveLast30Days,
	)
	if err != nil {
		return nil, r.handleError(err, "failed to get user stats")
	}

	return &stats, nil
}

// CountActive cuenta usuarios activos
func (r *UserRepository) CountActive(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM auth.users WHERE is_active = true`).Scan(&count)
	if err != nil {
		return 0, r.handleError(err, "failed to count active users")
	}
	return count, nil
}

// CountByRole cuenta usuarios por rol
func (r *UserRepository) CountByRole(ctx context.Context, role enums.UserRole) (int64, error) {
	var count int64
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM auth.users WHERE role = $1 AND is_active = true`, role.String()).Scan(&count)
	if err != nil {
		return 0, r.handleError(err, "failed to count users by role")
	}
	return count, nil
}
