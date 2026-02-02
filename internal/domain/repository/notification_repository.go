package repository

import (
	"context"

	"github.com/franciscozamorau/osmi-server/internal/api/dto"
	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
)

// NotificationRepository define operaciones para notificaciones
type NotificationRepository interface {
	// CRUD básico
	Create(ctx context.Context, notification *entities.Notification) error
	FindByID(ctx context.Context, id int64) (*entities.Notification, error)
	Update(ctx context.Context, notification *entities.Notification) error
	Delete(ctx context.Context, id int64) error

	// Búsquedas
	List(ctx context.Context, filter dto.NotificationFilter, pagination dto.Pagination) ([]*entities.Notification, int64, error)
	FindByRecipient(ctx context.Context, recipientType, recipientID string, pagination dto.Pagination) ([]*entities.Notification, int64, error)
	FindByTemplate(ctx context.Context, templateID int64, pagination dto.Pagination) ([]*entities.Notification, int64, error)
	FindByStatus(ctx context.Context, status string, pagination dto.Pagination) ([]*entities.Notification, int64, error)
	FindByChannel(ctx context.Context, channel string, pagination dto.Pagination) ([]*entities.Notification, int64, error)
	FindScheduled(ctx context.Context) ([]*entities.Notification, error)
	FindFailed(ctx context.Context, maxAttempts int) ([]*entities.Notification, error)
	FindRetryable(ctx context.Context) ([]*entities.Notification, error)

	// Operaciones específicas
	UpdateStatus(ctx context.Context, notificationID int64, status string) error
	MarkAsSent(ctx context.Context, notificationID int64, sentAt string, providerMessageID string) error
	MarkAsDelivered(ctx context.Context, notificationID int64, deliveredAt string) error
	MarkAsFailed(ctx context.Context, notificationID int64, errorMessage, errorCode string) error
	IncrementAttempts(ctx context.Context, notificationID int64) error
	SetNextRetry(ctx context.Context, notificationID int64, nextRetryAt string) error
	AddErrorToHistory(ctx context.Context, notificationID int64, errorMessage, errorCode string) error
	RecordOpen(ctx context.Context, notificationID int64) error
	RecordClick(ctx context.Context, notificationID int64) error
	UpdateProviderResponse(ctx context.Context, notificationID int64, response map[string]interface{}) error

	// Envío masivo
	CreateBulk(ctx context.Context, notifications []*entities.Notification) error
	UpdateBulkStatus(ctx context.Context, notificationIDs []int64, status string) error

	// Limpieza
	CleanOldNotifications(ctx context.Context, days int) (int64, error)
	CleanFailedNotifications(ctx context.Context, maxAgeDays int) (int64, error)

	// Estadísticas
	GetStats(ctx context.Context, filter dto.NotificationFilter) (*dto.NotificationStatsResponse, error)
	GetDeliveryRate(ctx context.Context, channel string, period string) (float64, error)
	GetOpenRate(ctx context.Context, channel string, period string) (float64, error)
	GetClickRate(ctx context.Context, channel string, period string) (float64, error)
	GetAverageDeliveryTime(ctx context.Context, channel string) (float64, error)
	GetFailureReasons(ctx context.Context, period string) ([]*dto.FailureReasonStats, error)
}
