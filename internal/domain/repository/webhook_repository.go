package repository

import (
	"context"

	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
)

// WebhookRepository define operaciones para webhooks
type WebhookRepository interface {
	// CRUD básico
	Create(ctx context.Context, webhook *entities.Webhook) error
	FindByID(ctx context.Context, id int64) (*entities.Webhook, error)
	FindByPublicID(ctx context.Context, publicID string) (*entities.Webhook, error)
	Update(ctx context.Context, webhook *entities.Webhook) error
	Delete(ctx context.Context, id int64) error

	// Búsquedas
	List(ctx context.Context, activeOnly bool) ([]*entities.Webhook, error)
	ListByProvider(ctx context.Context, provider string) ([]*entities.Webhook, error)
	ListByEventType(ctx context.Context, eventType string) ([]*entities.Webhook, error)
	FindByTargetURL(ctx context.Context, targetURL string) ([]*entities.Webhook, error)

	// Operaciones específicas
	UpdateStatus(ctx context.Context, webhookID int64, active bool) error
	UpdateConfig(ctx context.Context, webhookID int64, config map[string]interface{}) error
	UpdateSecret(ctx context.Context, webhookID int64, secretToken string) error
	UpdateLastTriggered(ctx context.Context, webhookID int64) error
	RotateSecret(ctx context.Context, webhookID int64) (string, error)

	// Disparo de webhooks
	GetWebhooksForEvent(ctx context.Context, provider, eventType string) ([]*entities.Webhook, error)
	RecordDeliveryAttempt(ctx context.Context, webhookID int64, success bool, statusCode int, responseBody string) error

	// Validaciones
	ValidateSignature(ctx context.Context, webhookID int64, payload []byte, signature string) (bool, error)
	IsActive(ctx context.Context, webhookID int64) (bool, error)
	ShouldRetry(ctx context.Context, webhookID int64) (bool, error)

	// Estadísticas
	GetStats(ctx context.Context, webhookID int64) (*entities.WebhookStats, error)
	GetDeliveryStats(ctx context.Context, webhookID int64) (*entities.DeliveryStats, error)
	GetRecentDeliveries(ctx context.Context, webhookID int64, limit int) ([]*entities.DeliveryAttempt, error)
}
