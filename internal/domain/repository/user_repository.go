// internal/domain/repository/user_repository.go
package repository

import (
	"context"
	"errors"
	"time"

	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
	"github.com/franciscozamorau/osmi-server/internal/domain/enums"
)

// UserFilter encapsula TODOS los criterios de búsqueda para usuarios
type UserFilter struct {
	// Filtros por ID
	IDs       []int64
	PublicIDs []string
	Email     *string
	Username  *string

	// Filtros de texto
	SearchTerm *string // Busca en email, username, first_name, last_name
	FirstName  *string
	LastName   *string

	// Filtros de rol y estado
	Role          *enums.UserRole
	IsActive      *bool
	IsStaff       *bool
	IsSuperuser   *bool
	EmailVerified *bool
	PhoneVerified *bool
	MFAEnabled    *bool

	// Filtros de rango de fechas
	CreatedFrom   *time.Time
	CreatedTo     *time.Time
	LastLoginFrom *time.Time
	LastLoginTo   *time.Time

	// Paginación y ordenamiento
	Limit     int
	Offset    int
	SortBy    string // "created_at", "last_login_at", "email", "username"
	SortOrder string // "asc", "desc"
}

// Errores específicos del repositorio
var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserEmailExists    = errors.New("user email already exists")
	ErrUserUsernameExists = errors.New("username already exists")
	ErrUserLocked         = errors.New("user is locked")
)

type UserRepository interface {
	// --- Operaciones de Escritura ---
	Create(ctx context.Context, user *entities.User) error
	Update(ctx context.Context, user *entities.User) error
	Delete(ctx context.Context, id int64) error
	SoftDelete(ctx context.Context, publicID string) error

	// --- Operaciones de Lectura (Flexibles) ---
	Find(ctx context.Context, filter *UserFilter) ([]*entities.User, int64, error)

	// Atajos
	GetByID(ctx context.Context, id int64) (*entities.User, error)
	GetByPublicID(ctx context.Context, publicID string) (*entities.User, error)
	GetByEmail(ctx context.Context, email string) (*entities.User, error)
	GetByUsername(ctx context.Context, username string) (*entities.User, error)

	// --- Operaciones de Verificación ---
	Exists(ctx context.Context, id int64) (bool, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)

	// --- Operaciones de Autenticación ---
	UpdatePassword(ctx context.Context, userID int64, passwordHash string) error
	UpdateLastLogin(ctx context.Context, userID int64, ipAddress string) error
	IncrementFailedAttempts(ctx context.Context, userID int64) error
	ResetFailedAttempts(ctx context.Context, userID int64) error
	LockUser(ctx context.Context, userID int64, until time.Time) error
	UnlockUser(ctx context.Context, userID int64) error

	// --- Operaciones de Verificación ---
	VerifyEmail(ctx context.Context, userID int64) error
	VerifyPhone(ctx context.Context, userID int64) error

	// --- Operaciones MFA ---
	EnableMFA(ctx context.Context, userID int64, secret string) error
	DisableMFA(ctx context.Context, userID int64) error

	// --- Operaciones de Preferencias ---
	UpdatePreferences(ctx context.Context, userID int64, preferences map[string]interface{}) error

	// --- Estadísticas ---
	GetStats(ctx context.Context) (*UserStats, error)
	CountActive(ctx context.Context) (int64, error)
	CountByRole(ctx context.Context, role enums.UserRole) (int64, error)
}

// UserStats representa estadísticas agregadas de usuarios
type UserStats struct {
	TotalUsers         int64 `json:"total_users"`
	ActiveUsers        int64 `json:"active_users"`
	StaffUsers         int64 `json:"staff_users"`
	Superusers         int64 `json:"superusers"`
	EmailVerifiedUsers int64 `json:"email_verified_users"`
	PhoneVerifiedUsers int64 `json:"phone_verified_users"`
	MFAEnabledUsers    int64 `json:"mfa_enabled_users"`
	NewUsersLast7Days  int64 `json:"new_users_last_7_days"`
	NewUsersLast30Days int64 `json:"new_users_last_30_days"`
	ActiveLast7Days    int64 `json:"active_last_7_days"`
	ActiveLast30Days   int64 `json:"active_last_30_days"`
}
