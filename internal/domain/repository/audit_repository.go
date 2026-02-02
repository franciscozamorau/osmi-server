package repository

import (
	"context"

	"github.com/franciscozamorau/osmi-server/internal/api/dto"
	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
)

// AuditRepository define operaciones para auditoría
type AuditRepository interface {
	// Registro
	LogDataChange(ctx context.Context, change *entities.DataChange) error
	LogSecurityEvent(ctx context.Context, event *entities.SecurityLog) error

	// Búsquedas
	GetDataChanges(ctx context.Context, filter dto.AuditFilter, pagination dto.Pagination) ([]*entities.DataChange, int64, error)
	GetSecurityLogs(ctx context.Context, filter dto.AuditFilter, pagination dto.Pagination) ([]*entities.SecurityLog, int64, error)
	GetChangesForRecord(ctx context.Context, tableName string, recordID int64, limit int) ([]*entities.DataChange, error)
	GetChangesByUser(ctx context.Context, userID int64, pagination dto.Pagination) ([]*entities.DataChange, int64, error)
	GetSecurityEventsByUser(ctx context.Context, userID int64, pagination dto.Pagination) ([]*entities.SecurityLog, int64, error)
	GetChangesByTable(ctx context.Context, tableName string, pagination dto.Pagination) ([]*entities.DataChange, int64, error)
	SearchDataChanges(ctx context.Context, term string, pagination dto.Pagination) ([]*entities.DataChange, int64, error)
	SearchSecurityLogs(ctx context.Context, term string, pagination dto.Pagination) ([]*entities.SecurityLog, int64, error)

	// Consultas específicas
	GetLastChangeForRecord(ctx context.Context, tableName string, recordID int64) (*entities.DataChange, error)
	GetChangesInPeriod(ctx context.Context, startDate, endDate string) ([]*entities.DataChange, error)
	GetSecurityEventsInPeriod(ctx context.Context, startDate, endDate string) ([]*entities.SecurityLog, error)
	GetHighSeverityEvents(ctx context.Context, days int) ([]*entities.SecurityLog, error)
	GetFailedLoginAttempts(ctx context.Context, userID int64, hours int) ([]*entities.SecurityLog, error)

	// Limpieza
	CleanOldAuditLogs(ctx context.Context, retentionDays int) (int64, error)
	ArchiveAuditLogs(ctx context.Context, archiveBefore string) (int64, error)

	// Estadísticas
	GetAuditStats(ctx context.Context) (*dto.AuditStatsResponse, error)
	GetActivityTimeline(ctx context.Context, days int) ([]*dto.ActivityPoint, error)
	GetMostActiveTables(ctx context.Context, limit int) ([]*dto.TableActivity, error)
	GetMostActiveUsers(ctx context.Context, limit int) ([]*dto.UserActivity, error)
	GetSecurityEventDistribution(ctx context.Context) (*dto.SecurityEventDistribution, error)
	GetDataChangeFrequency(ctx context.Context, period string) ([]*dto.ChangeFrequency, error)
}
