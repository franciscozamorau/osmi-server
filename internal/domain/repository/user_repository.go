package repository

import (
	"context"
	"time"

	"github.com/franciscozamorau/osmi-server/internal/api/dto"
	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
)

// UserRepository define las operaciones CRUD para usuarios
type UserRepository interface {
	// CRUD básico
	Create(ctx context.Context, user *entities.User) error
	FindByID(ctx context.Context, id int64) (*entities.User, error)
	FindByPublicID(ctx context.Context, publicID string) (*entities.User, error)
	FindByEmail(ctx context.Context, email string) (*entities.User, error)
	FindByUsername(ctx context.Context, username string) (*entities.User, error)
	Update(ctx context.Context, user *entities.User) error
	Delete(ctx context.Context, id int64) error
	SoftDelete(ctx context.Context, publicID string) error

	// Búsquedas y listados
	List(ctx context.Context, filter dto.UserFilter, pagination dto.Pagination) ([]*entities.User, int64, error)
	FindByRole(ctx context.Context, role string, pagination dto.Pagination) ([]*entities.User, int64, error)
	Search(ctx context.Context, term string, limit int) ([]*entities.User, error)

	// Operaciones específicas
	UpdatePassword(ctx context.Context, userID int64, passwordHash string) error
	UpdateLastLogin(ctx context.Context, userID int64, ipAddress string) error
	IncrementFailedLoginAttempts(ctx context.Context, userID int64) error
	ResetFailedLoginAttempts(ctx context.Context, userID int64) error
	LockUser(ctx context.Context, userID int64, until time.Time) error
	UnlockUser(ctx context.Context, userID int64) error
	VerifyEmail(ctx context.Context, userID int64) error
	EnableMFA(ctx context.Context, userID int64, secret string) error
	DisableMFA(ctx context.Context, userID int64) error
	UpdatePreferences(ctx context.Context, userID int64, preferences map[string]interface{}) error

	// Estadísticas
	GetStats(ctx context.Context) (*dto.UserStatsResponse, error)
	CountActiveUsers(ctx context.Context) (int64, error)
	CountByRole(ctx context.Context, role string) (int64, error)
}
