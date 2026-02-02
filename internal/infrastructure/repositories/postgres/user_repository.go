package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/franciscozamorau/osmi-server/internal/api/dto"
	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
	"github.com/franciscozamorau/osmi-server/internal/domain/enums"
	"github.com/franciscozamorau/osmi-server/internal/domain/repository"
	"github.com/franciscozamorau/osmi-server/internal/domain/valueobjects"
	"github.com/franciscozamorau/osmi-server/internal/repositories/postgres/helpers/errors"
	"github.com/franciscozamorau/osmi-server/internal/repositories/postgres/helpers/query"
	"github.com/franciscozamorau/osmi-server/internal/repositories/postgres/helpers/scanner"
	"github.com/franciscozamorau/osmi-server/internal/repositories/postgres/helpers/types"
	"github.com/franciscozamorau/osmi-server/internal/repositories/postgres/helpers/utils"
	"github.com/franciscozamorau/osmi-server/internal/repositories/postgres/helpers/validations"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// userRepository implementa repository.UserRepository usando helpers
type userRepository struct {
	db         *pgxpool.Pool
	converter  *types.Converter
	userConv   *types.UserConverter
	scanner    *scanner.RowScanner
	userScan   *scanner.UserScanner
	errHandler *errors.PostgresErrorHandler
	validator  *errors.Validator
	logger     *utils.Logger
}

// NewUserRepository crea una nueva instancia con helpers
func NewUserRepository(db *pgxpool.Pool) repository.UserRepository {
	conv := types.NewConverter()

	return &userRepository{
		db:         db,
		converter:  conv,
		userConv:   types.NewUserConverter(),
		scanner:    scanner.NewRowScanner(),
		userScan:   scanner.NewUserScanner(),
		errHandler: errors.NewPostgresErrorHandler(),
		validator:  errors.NewValidator(),
		logger:     utils.NewLogger("user-repository"),
	}
}

// Create implementa repository.UserRepository.Create usando helpers
func (r *userRepository) Create(ctx context.Context, user *entities.User) error {
	startTime := time.Now()

	// Validaciones usando helpers
	if !validations.IsValidEmail(user.Email) {
		return fmt.Errorf("invalid email: %s", user.Email)
	}

	if user.Phone != nil && *user.Phone != "" && !validations.IsValidPhone(*user.Phone) {
		return fmt.Errorf("invalid phone: %s", *user.Phone)
	}

	// Usar value objects existentes
	emailVO, err := valueobjects.NewEmail(user.Email)
	if err != nil {
		return fmt.Errorf("invalid email: %w", err)
	}

	// Validar phone si está presente
	var phonePtr *string
	if user.Phone != nil && *user.Phone != "" {
		phoneVO, err := valueobjects.NewPhone(*user.Phone)
		if err != nil {
			return fmt.Errorf("invalid phone: %w", err)
		}
		phone := phoneVO.String()
		phonePtr = &phone
	}

	// Generar public_uuid si no existe
	if user.PublicID == "" {
		user.PublicID = uuid.New().String()
	}

	// Validar role
	if !enums.UserRole(user.Role).IsValid() {
		return errors.New("invalid user role")
	}

	// Usar helpers para conversiones
	pgEmail := r.converter.Text(user.Email)
	pgPhone := r.converter.TextPtr(phonePtr)
	pgUsername := r.converter.TextPtr(user.Username)
	pgFirstName := r.converter.TextPtr(user.FirstName)
	pgLastName := r.converter.TextPtr(user.LastName)
	pgFullName := r.converter.TextPtr(user.FullName)
	pgAvatarURL := r.converter.TextPtr(user.AvatarURL)
	pgDateOfBirth := r.converter.DatePtr(user.DateOfBirth)
	pgMFASecret := r.converter.TextPtr(user.MFASecret)
	pgLastActiveAt := r.converter.TimestampPtr(user.LastActiveAt)

	query := `
		INSERT INTO auth.users (
			public_uuid, email, phone, username, password_hash,
			first_name, last_name, full_name, avatar_url,
			date_of_birth, preferred_language, preferred_currency,
			timezone, mfa_enabled, mfa_secret,
			is_active, is_staff, is_superuser,
			last_active_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12,
			$13, $14, $15, $16, $17, $18, $19
		)
		RETURNING id, created_at, updated_at
	`

	err = r.db.QueryRow(ctx, query,
		user.PublicID,
		emailVO.String(),
		phonePtr,
		user.Username,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
		user.FullName,
		user.AvatarURL,
		user.DateOfBirth,
		user.PreferredLanguage,
		user.PreferredCurrency,
		user.Timezone,
		user.MFAEnabled,
		user.MFASecret,
		user.IsActive,
		user.IsStaff,
		user.IsSuperuser,
		user.LastActiveAt,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		// Usar error handler para manejo consistente
		if r.errHandler.IsDuplicateKey(err) {
			constraint := r.errHandler.GetConstraintName(err)
			value := r.errHandler.GetDuplicateValue(err)

			if strings.Contains(strings.ToLower(constraint), "email") {
				return fmt.Errorf("email already exists: %s", user.Email)
			} else if strings.Contains(strings.ToLower(constraint), "username") {
				return fmt.Errorf("username already exists: %s", *user.Username)
			} else if strings.Contains(strings.ToLower(constraint), "public_uuid") {
				return fmt.Errorf("public_uuid already exists: %s", user.PublicID)
			} else {
				return r.errHandler.CreateUserFriendlyError(err, "user")
			}
		}

		// Log del error
		r.logger.DatabaseLogger("INSERT", "auth.users", time.Since(startTime), 1, err, map[string]interface{}{
			"email":     utils.SafeEmailForLog(user.Email),
			"public_id": user.PublicID,
		})

		return r.errHandler.WrapError(err, "user repository", "create user")
	}

	r.logger.DatabaseLogger("INSERT", "auth.users", time.Since(startTime), 1, nil, map[string]interface{}{
		"user_id": user.ID,
		"email":   utils.SafeEmailForLog(user.Email),
	})

	return nil
}

// FindByID implementa repository.UserRepository.FindByID usando scanner
func (r *userRepository) FindByID(ctx context.Context, id int64) (*entities.User, error) {
	startTime := time.Now()

	query := `
		SELECT 
			id, public_uuid, email, phone, username, password_hash,
			email_verified, phone_verified, verified_at,
			first_name, last_name, full_name, avatar_url, date_of_birth,
			preferred_language, preferred_currency, timezone,
			mfa_enabled, mfa_secret, last_login_at, last_login_ip,
			failed_login_attempts, locked_until,
			is_active, is_staff, is_superuser,
			last_active_at, created_at, updated_at
		FROM auth.users
		WHERE id = $1 AND is_active = true
	`

	row := r.db.QueryRow(ctx, query, id)
	user, err := r.userScan.ScanUser(row)

	if err != nil {
		if err.Error() == "user not found" {
			r.logger.Debug("User not found", map[string]interface{}{
				"user_id": id,
			})
			return nil, err
		}

		r.logger.DatabaseLogger("SELECT", "auth.users", time.Since(startTime), 0, err, map[string]interface{}{
			"user_id": id,
		})

		return nil, r.errHandler.WrapError(err, "user repository", "find user by ID")
	}

	r.logger.DatabaseLogger("SELECT", "auth.users", time.Since(startTime), 1, nil, map[string]interface{}{
		"user_id": id,
	})

	return user, nil
}

// FindByPublicID implementa repository.UserRepository.FindByPublicID
func (r *userRepository) FindByPublicID(ctx context.Context, publicID string) (*entities.User, error) {
	startTime := time.Now()

	// Validar UUID usando helpers
	if !validations.IsValidUUID(publicID) {
		return nil, fmt.Errorf("invalid public_id format: %s", publicID)
	}

	query := `
		SELECT 
			id, public_uuid, email, phone, username, password_hash,
			email_verified, phone_verified, verified_at,
			first_name, last_name, full_name, avatar_url, date_of_birth,
			preferred_language, preferred_currency, timezone,
			mfa_enabled, mfa_secret, last_login_at, last_login_ip,
			failed_login_attempts, locked_until,
			is_active, is_staff, is_superuser,
			last_active_at, created_at, updated_at
		FROM auth.users
		WHERE public_uuid = $1 AND is_active = true
	`

	row := r.db.QueryRow(ctx, query, publicID)
	user, err := r.userScan.ScanUser(row)

	if err != nil {
		if err.Error() == "user not found" {
			r.logger.Debug("User not found by public ID", map[string]interface{}{
				"public_id": publicID,
			})
			return nil, err
		}

		r.logger.DatabaseLogger("SELECT", "auth.users", time.Since(startTime), 0, err, map[string]interface{}{
			"public_id": publicID,
		})

		return nil, r.errHandler.WrapError(err, "user repository", "find user by public ID")
	}

	r.logger.DatabaseLogger("SELECT", "auth.users", time.Since(startTime), 1, nil, map[string]interface{}{
		"user_id":   user.ID,
		"public_id": publicID,
	})

	return user, nil
}

// FindByEmail implementa repository.UserRepository.FindByEmail
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	startTime := time.Now()

	// Validar email usando helpers
	if !validations.IsValidEmail(email) {
		return nil, fmt.Errorf("invalid email: %s", email)
	}

	// Usar value object existente
	emailVO, err := valueobjects.NewEmail(email)
	if err != nil {
		return nil, fmt.Errorf("invalid email: %w", err)
	}

	query := `
		SELECT 
			id, public_uuid, email, phone, username, password_hash,
			email_verified, phone_verified, verified_at,
			first_name, last_name, full_name, avatar_url, date_of_birth,
			preferred_language, preferred_currency, timezone,
			mfa_enabled, mfa_secret, last_login_at, last_login_ip,
			failed_login_attempts, locked_until,
			is_active, is_staff, is_superuser,
			last_active_at, created_at, updated_at
		FROM auth.users
		WHERE email = $1 AND is_active = true
	`

	row := r.db.QueryRow(ctx, query, emailVO.String())
	user, err := r.userScan.ScanUser(row)

	if err != nil {
		if err.Error() == "user not found" {
			r.logger.Debug("User not found by email", map[string]interface{}{
				"email": utils.SafeEmailForLog(email),
			})
			return nil, err
		}

		r.logger.DatabaseLogger("SELECT", "auth.users", time.Since(startTime), 0, err, map[string]interface{}{
			"email": utils.SafeEmailForLog(email),
		})

		return nil, r.errHandler.WrapError(err, "user repository", "find user by email")
	}

	r.logger.DatabaseLogger("SELECT", "auth.users", time.Since(startTime), 1, nil, map[string]interface{}{
		"user_id": user.ID,
		"email":   utils.SafeEmailForLog(email),
	})

	return user, nil
}

// FindByUsername implementa repository.UserRepository.FindByUsername
func (r *userRepository) FindByUsername(ctx context.Context, username string) (*entities.User, error) {
	startTime := time.Now()

	if username == "" {
		return nil, errors.New("username cannot be empty")
	}

	// Validar username usando helpers
	if !validations.IsValidUsername(username) {
		return nil, fmt.Errorf("invalid username format: %s", username)
	}

	query := `
		SELECT 
			id, public_uuid, email, phone, username, password_hash,
			email_verified, phone_verified, verified_at,
			first_name, last_name, full_name, avatar_url, date_of_birth,
			preferred_language, preferred_currency, timezone,
			mfa_enabled, mfa_secret, last_login_at, last_login_ip,
			failed_login_attempts, locked_until,
			is_active, is_staff, is_superuser,
			last_active_at, created_at, updated_at
		FROM auth.users
		WHERE username = $1 AND is_active = true
	`

	row := r.db.QueryRow(ctx, query, strings.TrimSpace(username))
	user, err := r.userScan.ScanUser(row)

	if err != nil {
		if err.Error() == "user not found" {
			r.logger.Debug("User not found by username", map[string]interface{}{
				"username": utils.SafeStringForLog(username),
			})
			return nil, err
		}

		r.logger.DatabaseLogger("SELECT", "auth.users", time.Since(startTime), 0, err, map[string]interface{}{
			"username": utils.SafeStringForLog(username),
		})

		return nil, r.errHandler.WrapError(err, "user repository", "find user by username")
	}

	r.logger.DatabaseLogger("SELECT", "auth.users", time.Since(startTime), 1, nil, map[string]interface{}{
		"user_id":  user.ID,
		"username": utils.SafeStringForLog(username),
	})

	return user, nil
}

// Update implementa repository.UserRepository.Update usando helpers
func (r *userRepository) Update(ctx context.Context, user *entities.User) error {
	startTime := time.Now()

	// Validaciones
	if !validations.IsValidEmail(user.Email) {
		return fmt.Errorf("invalid email: %s", user.Email)
	}

	if user.Phone != nil && *user.Phone != "" && !validations.IsValidPhone(*user.Phone) {
		return fmt.Errorf("invalid phone: %s", *user.Phone)
	}

	// Usar value objects existentes
	emailVO, err := valueobjects.NewEmail(user.Email)
	if err != nil {
		return fmt.Errorf("invalid email: %w", err)
	}

	// Validar phone si está presente
	var phonePtr *string
	if user.Phone != nil && *user.Phone != "" {
		phoneVO, err := valueobjects.NewPhone(*user.Phone)
		if err != nil {
			return fmt.Errorf("invalid phone: %w", err)
		}
		phone := phoneVO.String()
		phonePtr = &phone
	}

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
			last_active_at = $17,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $18
		RETURNING updated_at
	`

	err = r.db.QueryRow(ctx, query,
		emailVO.String(),
		phonePtr,
		user.Username,
		user.FirstName,
		user.LastName,
		user.FullName,
		user.AvatarURL,
		user.DateOfBirth,
		user.PreferredLanguage,
		user.PreferredCurrency,
		user.Timezone,
		user.MFAEnabled,
		user.MFASecret,
		user.IsActive,
		user.IsStaff,
		user.IsSuperuser,
		user.LastActiveAt,
		user.ID,
	).Scan(&user.UpdatedAt)

	if err != nil {
		// Usar error handler
		if r.errHandler.IsDuplicateKey(err) {
			constraint := r.errHandler.GetConstraintName(err)

			if strings.Contains(strings.ToLower(constraint), "email") {
				return fmt.Errorf("email already exists: %s", user.Email)
			} else if strings.Contains(strings.ToLower(constraint), "username") {
				return fmt.Errorf("username already exists: %s", *user.Username)
			}
		}

		r.logger.DatabaseLogger("UPDATE", "auth.users", time.Since(startTime), 0, err, map[string]interface{}{
			"user_id": user.ID,
			"email":   utils.SafeEmailForLog(user.Email),
		})

		return r.errHandler.WrapError(err, "user repository", "update user")
	}

	r.logger.DatabaseLogger("UPDATE", "auth.users", time.Since(startTime), 1, nil, map[string]interface{}{
		"user_id": user.ID,
	})

	return nil
}

// Delete implementa repository.UserRepository.Delete
func (r *userRepository) Delete(ctx context.Context, id int64) error {
	startTime := time.Now()

	query := `DELETE FROM auth.users WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		r.logger.DatabaseLogger("DELETE", "auth.users", time.Since(startTime), 0, err, map[string]interface{}{
			"user_id": id,
		})

		return r.errHandler.WrapError(err, "user repository", "delete user")
	}

	rowsAffected := result.RowsAffected()

	if rowsAffected == 0 {
		r.logger.Debug("User not found for deletion", map[string]interface{}{
			"user_id": id,
		})
		return errors.New("user not found")
	}

	r.logger.DatabaseLogger("DELETE", "auth.users", time.Since(startTime), rowsAffected, nil, map[string]interface{}{
		"user_id": id,
	})

	return nil
}

// SoftDelete implementa repository.UserRepository.SoftDelete
func (r *userRepository) SoftDelete(ctx context.Context, publicID string) error {
	startTime := time.Now()

	// Validar UUID usando helpers
	if !validations.IsValidUUID(publicID) {
		return fmt.Errorf("invalid public_id format: %s", publicID)
	}

	query := `
		UPDATE auth.users 
		SET is_active = false, updated_at = CURRENT_TIMESTAMP
		WHERE public_uuid = $1 AND is_active = true
		RETURNING id
	`

	var id int64
	err := r.db.QueryRow(ctx, query, publicID).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.logger.Debug("User not found or already inactive", map[string]interface{}{
				"public_id": publicID,
			})
			return errors.New("user not found or already inactive")
		}

		r.logger.DatabaseLogger("UPDATE", "auth.users", time.Since(startTime), 0, err, map[string]interface{}{
			"public_id": publicID,
		})

		return r.errHandler.WrapError(err, "user repository", "soft delete user")
	}

	r.logger.DatabaseLogger("UPDATE", "auth.users", time.Since(startTime), 1, nil, map[string]interface{}{
		"user_id":   id,
		"public_id": publicID,
	})

	return nil
}

// List implementa repository.UserRepository.List usando query builder
func (r *userRepository) List(ctx context.Context, filter dto.UserFilter, pagination dto.Pagination) ([]*entities.User, int64, error) {
	startTime := time.Now()

	// Usar query builder para construir la query
	qb := query.NewQueryBuilder(`
		SELECT 
			id, public_uuid, email, phone, username, password_hash,
			email_verified, phone_verified, verified_at,
			first_name, last_name, full_name, avatar_url, date_of_birth,
			preferred_language, preferred_currency, timezone,
			mfa_enabled, mfa_secret, last_login_at, last_login_ip,
			failed_login_attempts, locked_until,
			is_active, is_staff, is_superuser,
			last_active_at, created_at, updated_at
		FROM auth.users
	`).Where("1=1", nil) // Condición inicial

	// Aplicar filtros
	if filter.IsActive != nil {
		qb.Where("is_active = ?", *filter.IsActive)
	}

	if filter.Search != "" {
		qb.Where("(email ILIKE ? OR username ILIKE ? OR first_name ILIKE ? OR last_name ILIKE ?)",
			"%"+filter.Search+"%", "%"+filter.Search+"%", "%"+filter.Search+"%", "%"+filter.Search+"%")
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
		r.logger.DatabaseLogger("SELECT", "auth.users", time.Since(startTime), 0, err, map[string]interface{}{
			"operation": "count",
		})

		return nil, 0, r.errHandler.WrapError(err, "user repository", "count users")
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
		r.logger.DatabaseLogger("SELECT", "auth.users", time.Since(startTime), 0, err, map[string]interface{}{
			"operation": "list",
		})

		return nil, 0, r.errHandler.WrapError(err, "user repository", "list users")
	}
	defer rows.Close()

	// Usar scanner para procesar resultados
	users := []*entities.User{}
	for rows.Next() {
		user, err := r.userScan.ScanUser(rows)
		if err != nil {
			r.logger.Error("Failed to scan user row", err, map[string]interface{}{
				"operation": "list",
			})
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		r.logger.DatabaseLogger("SELECT", "auth.users", time.Since(startTime), 0, err, map[string]interface{}{
			"operation": "list",
		})

		return nil, 0, r.errHandler.WrapError(err, "user repository", "iterate users")
	}

	r.logger.DatabaseLogger("SELECT", "auth.users", time.Since(startTime), int64(len(users)), nil, map[string]interface{}{
		"total":     total,
		"page":      pagination.Page,
		"page_size": pagination.PageSize,
	})

	return users, total, nil
}

// Search implementa repository.UserRepository.Search usando query builder
func (r *userRepository) Search(ctx context.Context, term string, limit int) ([]*entities.User, error) {
	startTime := time.Now()

	if limit <= 0 {
		limit = 20
	}

	// Usar query builder
	qb := query.NewQueryBuilder(`
		SELECT 
			id, public_uuid, email, phone, username, password_hash,
			email_verified, phone_verified, verified_at,
			first_name, last_name, full_name, avatar_url, date_of_birth,
			preferred_language, preferred_currency, timezone,
			mfa_enabled, mfa_secret, last_login_at, last_login_ip,
			failed_login_attempts, locked_until,
			is_active, is_staff, is_superuser,
			last_active_at, created_at, updated_at
		FROM auth.users
	`).Where("is_active = true", nil)

	if term != "" {
		qb.Where("(email ILIKE ? OR username ILIKE ? OR first_name ILIKE ? OR last_name ILIKE ?)",
			"%"+term+"%", "%"+term+"%", "%"+term+"%", "%"+term+"%")
	}

	// Ordenar por relevancia
	orderBy := `
		CASE 
			WHEN email ILIKE ? THEN 1
			WHEN username ILIKE ? THEN 2
			WHEN first_name ILIKE ? THEN 3
			ELSE 4
		END,
		created_at DESC
	`
	qb.OrderByRaw(orderBy)
	qb.Limit(limit)

	queryStr, args := qb.Build()

	rows, err := r.db.Query(ctx, queryStr, args...)
	if err != nil {
		r.logger.DatabaseLogger("SELECT", "auth.users", time.Since(startTime), 0, err, map[string]interface{}{
			"operation": "search",
			"term":      utils.SafeStringForLog(term),
		})

		return nil, r.errHandler.WrapError(err, "user repository", "search users")
	}
	defer rows.Close()

	users := []*entities.User{}
	for rows.Next() {
		user, err := r.userScan.ScanUser(rows)
		if err != nil {
			r.logger.Error("Failed to scan user row during search", err, map[string]interface{}{
				"operation": "search",
			})
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		r.logger.DatabaseLogger("SELECT", "auth.users", time.Since(startTime), 0, err, map[string]interface{}{
			"operation": "search",
		})

		return nil, r.errHandler.WrapError(err, "user repository", "iterate search results")
	}

	r.logger.DatabaseLogger("SELECT", "auth.users", time.Since(startTime), int64(len(users)), nil, map[string]interface{}{
		"term":  utils.SafeStringForLog(term),
		"limit": limit,
		"found": len(users),
	})

	return users, nil
}

// Métodos restantes (UpdatePassword, UpdateLastLogin, etc.) mantienen estructura similar
// Voy a mostrar algunos ejemplos y puedes aplicar el patrón a los demás:

// UpdatePassword implementa repository.UserRepository.UpdatePassword
func (r *userRepository) UpdatePassword(ctx context.Context, userID int64, passwordHash string) error {
	startTime := time.Now()

	query := `
		UPDATE auth.users 
		SET password_hash = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2 AND is_active = true
	`

	result, err := r.db.Exec(ctx, query, passwordHash, userID)
	if err != nil {
		r.logger.DatabaseLogger("UPDATE", "auth.users", time.Since(startTime), 0, err, map[string]interface{}{
			"user_id":   userID,
			"operation": "update_password",
		})

		return r.errHandler.WrapError(err, "user repository", "update password")
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		r.logger.Debug("User not found or inactive for password update", map[string]interface{}{
			"user_id": userID,
		})
		return errors.New("user not found or inactive")
	}

	r.logger.DatabaseLogger("UPDATE", "auth.users", time.Since(startTime), rowsAffected, nil, map[string]interface{}{
		"user_id": userID,
	})

	return nil
}

// UpdateLastLogin implementa repository.UserRepository.UpdateLastLogin
func (r *userRepository) UpdateLastLogin(ctx context.Context, userID int64, ipAddress string) error {
	startTime := time.Now()

	query := `
		UPDATE auth.users 
		SET last_login_at = CURRENT_TIMESTAMP, 
		    last_login_ip = $1,
		    last_active_at = CURRENT_TIMESTAMP,
		    updated_at = CURRENT_TIMESTAMP,
		    failed_login_attempts = 0
		WHERE id = $2
	`

	result, err := r.db.Exec(ctx, query, ipAddress, userID)
	if err != nil {
		r.logger.DatabaseLogger("UPDATE", "auth.users", time.Since(startTime), 0, err, map[string]interface{}{
			"user_id":   userID,
			"operation": "update_last_login",
		})

		return r.errHandler.WrapError(err, "user repository", "update last login")
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		r.logger.Debug("User not found for last login update", map[string]interface{}{
			"user_id": userID,
		})
		return errors.New("user not found")
	}

	r.logger.DatabaseLogger("UPDATE", "auth.users", time.Since(startTime), rowsAffected, nil, map[string]interface{}{
		"user_id":    userID,
		"ip_address": utils.SafeStringForLog(ipAddress),
	})

	return nil
}

// GetStats implementa repository.UserRepository.GetStats
func (r *userRepository) GetStats(ctx context.Context) (*dto.UserStatsResponse, error) {
	startTime := time.Now()

	query := `
		SELECT 
			COUNT(*) as total_users,
			COUNT(CASE WHEN is_active = true THEN 1 END) as active_users,
			COUNT(CASE WHEN email_verified = true THEN 1 END) as verified_users,
			COUNT(CASE WHEN mfa_enabled = true THEN 1 END) as mfa_users,
			COUNT(CASE WHEN last_login_at >= CURRENT_DATE - INTERVAL '30 days' THEN 1 END) as active_last_30_days,
			COUNT(CASE WHEN created_at >= CURRENT_DATE - INTERVAL '30 days' THEN 1 END) as new_last_30_days
		FROM auth.users
	`

	var stats dto.UserStatsResponse
	err := r.db.QueryRow(ctx, query).Scan(
		&stats.TotalUsers,
		&stats.ActiveUsers,
		&stats.VerifiedUsers,
		&stats.MFAUsers,
		&stats.ActiveLast30Days,
		&stats.NewLast30Days,
	)

	if err != nil {
		r.logger.DatabaseLogger("SELECT", "auth.users", time.Since(startTime), 0, err, map[string]interface{}{
			"operation": "get_stats",
		})

		return nil, r.errHandler.WrapError(err, "user repository", "get user stats")
	}

	r.logger.DatabaseLogger("SELECT", "auth.users", time.Since(startTime), 1, nil, map[string]interface{}{
		"operation": "get_stats",
	})

	return &stats, nil
}

// CountActiveUsers implementa repository.UserRepository.CountActiveUsers
func (r *userRepository) CountActiveUsers(ctx context.Context) (int64, error) {
	startTime := time.Now()

	query := `SELECT COUNT(*) FROM auth.users WHERE is_active = true`

	var count int64
	err := r.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		r.logger.DatabaseLogger("SELECT", "auth.users", time.Since(startTime), 0, err, map[string]interface{}{
			"operation": "count_active",
		})

		return 0, r.errHandler.WrapError(err, "user repository", "count active users")
	}

	r.logger.DatabaseLogger("SELECT", "auth.users", time.Since(startTime), 1, nil, map[string]interface{}{
		"operation": "count_active",
		"count":     count,
	})

	return count, nil
}

// NOTA: Los métodos restantes (IncrementFailedLoginAttempts, ResetFailedLoginAttempts,
// LockUser, UnlockUser, VerifyEmail, EnableMFA, DisableMFA, UpdatePreferences,
// FindByRole, CountByRole) siguen el MISMO patrón de:
// 1. Log de inicio
// 2. Ejecución con manejo de errores usando errHandler
// 3. Log de resultado
// 4. Retorno apropiado
