package entities

import (
	"errors"
	"time"
)

// User representa un usuario del sistema
// Mapea exactamente la tabla auth.users
type User struct {
	ID           int64   `json:"id" db:"id"`
	PublicID     string  `json:"public_id" db:"public_uuid"`
	Email        string  `json:"email" db:"email"`
	Phone        *string `json:"phone,omitempty" db:"phone"`
	Username     *string `json:"username,omitempty" db:"username"`
	PasswordHash string  `json:"-" db:"password_hash"`

	// NOTA: Role no existe en la tabla auth.users, se deriva de is_staff y is_superuser
	// Lo mantenemos para compatibilidad pero no tiene tag db
	Role string `json:"role,omitempty"`

	FirstName   *string    `json:"first_name,omitempty" db:"first_name"`
	LastName    *string    `json:"last_name,omitempty" db:"last_name"`
	FullName    *string    `json:"full_name,omitempty" db:"full_name"`
	AvatarURL   *string    `json:"avatar_url,omitempty" db:"avatar_url"`
	DateOfBirth *time.Time `json:"date_of_birth,omitempty" db:"date_of_birth"`

	EmailVerified bool       `json:"email_verified" db:"email_verified"`
	PhoneVerified bool       `json:"phone_verified" db:"phone_verified"`
	VerifiedAt    *time.Time `json:"verified_at,omitempty" db:"verified_at"`

	PreferredLanguage string `json:"preferred_language" db:"preferred_language"`
	PreferredCurrency string `json:"preferred_currency" db:"preferred_currency"`
	Timezone          string `json:"timezone" db:"timezone"`

	MFAEnabled  bool       `json:"mfa_enabled" db:"mfa_enabled"`
	MFASecret   *string    `json:"-" db:"mfa_secret"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
	LastLoginIP *string    `json:"last_login_ip,omitempty" db:"last_login_ip"`
	// CORREGIDO: failed_login_attempts es INTEGER, usamos int
	FailedLoginAttempts int        `json:"failed_login_attempts" db:"failed_login_attempts"`
	LockedUntil         *time.Time `json:"locked_until,omitempty" db:"locked_until"`

	IsActive    bool `json:"is_active" db:"is_active"`
	IsStaff     bool `json:"is_staff" db:"is_staff"`
	IsSuperuser bool `json:"is_superuser" db:"is_superuser"`

	LastActiveAt *time.Time `json:"last_active_at,omitempty" db:"last_active_at"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

// UserStats para estadísticas
// NOTA: Esta tabla no existe en la BD, es para uso interno
type UserStats struct {
	UserID           int64      `json:"user_id"`
	TotalLogins      int64      `json:"total_logins"`
	FailedLogins     int64      `json:"failed_logins"`
	TicketsPurchased int64      `json:"tickets_purchased"`
	TotalSpent       float64    `json:"total_spent"`
	LastLogin        *time.Time `json:"last_login,omitempty"`
	LastActive       *time.Time `json:"last_active,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// UserPublic para datos públicos
type UserPublic struct {
	ID        string    `json:"id"`
	PublicID  string    `json:"public_id"`
	Email     string    `json:"email"`
	FirstName *string   `json:"first_name,omitempty"`
	LastName  *string   `json:"last_name,omitempty"`
	AvatarURL *string   `json:"avatar_url,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// Métodos de utilidad para User

// IsLocked verifica si la cuenta está bloqueada
func (u *User) IsLocked() bool {
	if u.LockedUntil == nil {
		return false
	}
	return time.Now().Before(*u.LockedUntil)
}

// CanLogin verifica si el usuario puede iniciar sesión
func (u *User) CanLogin() bool {
	return u.IsActive && !u.IsLocked()
}

// IsAdmin verifica si el usuario es administrador
func (u *User) IsAdmin() bool {
	return u.IsSuperuser
}

// IsStaffUser verifica si el usuario es staff
func (u *User) IsStaffUser() bool {
	return u.IsStaff || u.IsSuperuser
}

// GetRole obtiene el rol del usuario basado en sus permisos
func (u *User) GetRole() string {
	switch {
	case u.IsSuperuser:
		return "admin"
	case u.IsStaff:
		return "staff"
	default:
		return "user"
	}
}

// GetDisplayName obtiene el nombre para mostrar
func (u *User) GetDisplayName() string {
	if u.FullName != nil && *u.FullName != "" {
		return *u.FullName
	}
	if u.FirstName != nil && u.LastName != nil {
		return *u.FirstName + " " + *u.LastName
	}
	if u.FirstName != nil {
		return *u.FirstName
	}
	if u.Username != nil {
		return *u.Username
	}
	return u.Email
}

// IsVerified verifica si el usuario está verificado
func (u *User) IsVerified() bool {
	return u.EmailVerified && (u.Phone == nil || u.PhoneVerified)
}

// RecordLogin registra un inicio de sesión exitoso
func (u *User) RecordLogin(ip string) {
	now := time.Now()
	u.LastLoginAt = &now
	u.LastLoginIP = &ip
	u.FailedLoginAttempts = 0
	u.LastActiveAt = &now
	u.UpdatedAt = now
}

// RecordFailedLogin registra un intento fallido de inicio de sesión
func (u *User) RecordFailedLogin(maxAttempts int, lockDuration time.Duration) {
	u.FailedLoginAttempts++

	if u.FailedLoginAttempts >= maxAttempts {
		lockedUntil := time.Now().Add(lockDuration)
		u.LockedUntil = &lockedUntil
	}

	u.UpdatedAt = time.Now()
}

// Unlock desbloquea la cuenta
func (u *User) Unlock() {
	u.LockedUntil = nil
	u.FailedLoginAttempts = 0
	u.UpdatedAt = time.Now()
}

// Verify marca el usuario como verificado
func (u *User) Verify() {
	now := time.Now()
	u.EmailVerified = true
	u.VerifiedAt = &now
	u.UpdatedAt = now
}

// VerifyPhone marca el teléfono como verificado
func (u *User) VerifyPhone() {
	now := time.Now()
	u.PhoneVerified = true
	if u.VerifiedAt == nil {
		u.VerifiedAt = &now
	}
	u.UpdatedAt = now
}

// EnableMFA habilita MFA
func (u *User) EnableMFA(secret string) {
	u.MFAEnabled = true
	u.MFASecret = &secret
	u.UpdatedAt = time.Now()
}

// DisableMFA deshabilita MFA
func (u *User) DisableMFA() {
	u.MFAEnabled = false
	u.MFASecret = nil
	u.UpdatedAt = time.Now()
}

// UpdateLastActive actualiza la última actividad
func (u *User) UpdateLastActive() {
	now := time.Now()
	u.LastActiveAt = &now
	u.UpdatedAt = now
}

// Validate verifica que el usuario sea válido
func (u *User) Validate() error {
	if u.Email == "" {
		return errors.New("email is required")
	}
	if u.PasswordHash == "" {
		return errors.New("password_hash is required")
	}
	if u.FailedLoginAttempts < 0 {
		return errors.New("failed_login_attempts cannot be negative")
	}
	return nil
}

// ToPublic convierte a UserPublic
func (u *User) ToPublic() *UserPublic {
	return &UserPublic{
		ID:        u.PublicID,
		PublicID:  u.PublicID,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		AvatarURL: u.AvatarURL,
		CreatedAt: u.CreatedAt,
	}
}
