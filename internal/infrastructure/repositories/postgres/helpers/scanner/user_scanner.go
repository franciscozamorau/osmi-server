package scanner

import (
	"database/sql"
	"fmt"

	"github.com/franciscozamorau/osmi-server/internal/domain/entities"

	"github.com/jackc/pgx/v5"
)

// UserScanner escanea resultados específicos de usuarios
type UserScanner struct {
	*RowScanner
}

// NewUserScanner crea un nuevo UserScanner
func NewUserScanner() *UserScanner {
	return &UserScanner{
		RowScanner: NewRowScanner(),
	}
}

// ScanUser escanea una fila completa a entidad User
func (us *UserScanner) ScanUser(row pgx.Row) (*entities.User, error) {
	var user entities.User
	var phone sql.NullString
	var username sql.NullString
	var firstName sql.NullString
	var lastName sql.NullString
	var fullName sql.NullString
	var avatarURL sql.NullString
	var dateOfBirth sql.NullTime
	var mfaSecret sql.NullString
	var lastLoginAt sql.NullTime
	var lastLoginIP sql.NullString
	var lockedUntil sql.NullTime
	var verifiedAt sql.NullTime
	var lastActiveAt sql.NullTime

	err := row.Scan(
		&user.ID,
		&user.PublicID,
		&user.Email,
		&phone,
		&username,
		&user.PasswordHash,
		&user.EmailVerified,
		&user.PhoneVerified,
		&verifiedAt,
		&firstName,
		&lastName,
		&fullName,
		&avatarURL,
		&dateOfBirth,
		&user.PreferredLanguage,
		&user.PreferredCurrency,
		&user.Timezone,
		&user.MFAEnabled,
		&mfaSecret,
		&lastLoginAt,
		&lastLoginIP,
		&user.FailedLoginAttempts,
		&lockedUntil,
		&user.IsActive,
		&user.IsStaff,
		&user.IsSuperuser,
		&lastActiveAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to scan user: %w", err)
	}

	// Convertir Null types a pointers
	user.Phone = us.ConvertSQLNullable(phone)
	user.Username = us.ConvertSQLNullable(username)
	user.FirstName = us.ConvertSQLNullable(firstName)
	user.LastName = us.ConvertSQLNullable(lastName)
	user.FullName = us.ConvertSQLNullable(fullName)
	user.AvatarURL = us.ConvertSQLNullable(avatarURL)
	user.DateOfBirth = us.ConvertSQLNullableTime(dateOfBirth)
	user.MFASecret = us.ConvertSQLNullable(mfaSecret)
	user.LastLoginAt = us.ConvertSQLNullableTime(lastLoginAt)
	user.LastLoginIP = us.ConvertSQLNullable(lastLoginIP)
	user.LockedUntil = us.ConvertSQLNullableTime(lockedUntil)
	user.VerifiedAt = us.ConvertSQLNullableTime(verifiedAt)
	user.LastActiveAt = us.ConvertSQLNullableTime(lastActiveAt)

	return &user, nil
}
