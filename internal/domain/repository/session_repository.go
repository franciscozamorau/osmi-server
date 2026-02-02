package repository

import (
	"context"

	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
)

// SessionRepository define operaciones para sesiones de usuario
type SessionRepository interface {
	// CRUD básico
	Create(ctx context.Context, session *entities.Session) error
	FindByID(ctx context.Context, id int64) (*entities.Session, error)
	FindBySessionID(ctx context.Context, sessionID string) (*entities.Session, error)
	FindByRefreshToken(ctx context.Context, refreshTokenHash string) (*entities.Session, error)
	Update(ctx context.Context, session *entities.Session) error
	Delete(ctx context.Context, id int64) error

	// Búsquedas
	FindByUser(ctx context.Context, userID int64, activeOnly bool) ([]*entities.Session, error)
	FindExpired(ctx context.Context) ([]*entities.Session, error)
	FindByDevice(ctx context.Context, userID int64, deviceInfo string) (*entities.Session, error)

	// Operaciones específicas
	Invalidate(ctx context.Context, sessionID string) error
	InvalidateAllForUser(ctx context.Context, userID int64) error
	InvalidateAllExceptCurrent(ctx context.Context, userID int64, currentSessionID string) error
	Refresh(ctx context.Context, sessionID string, newRefreshTokenHash string, expiresAt string) error
	UpdateActivity(ctx context.Context, sessionID string) error
	UpdateDeviceInfo(ctx context.Context, sessionID string, deviceInfo map[string]interface{}) error

	// Limpieza
	CleanExpiredSessions(ctx context.Context) (int64, error)
	CleanInactiveSessions(ctx context.Context, days int) (int64, error)

	// Verificaciones
	IsValid(ctx context.Context, sessionID string) (bool, error)
	CountActiveSessions(ctx context.Context, userID int64) (int64, error)
	GetLastActivity(ctx context.Context, sessionID string) (string, error)
}
