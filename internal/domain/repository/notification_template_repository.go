package repository

import (
	"context"

	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
)

// NotificationTemplateRepository define operaciones para plantillas de notificación
type NotificationTemplateRepository interface {
	// CRUD básico
	Create(ctx context.Context, template *entities.NotificationTemplate) error
	FindByID(ctx context.Context, id int64) (*entities.NotificationTemplate, error)
	FindByCode(ctx context.Context, code string) (*entities.NotificationTemplate, error)
	Update(ctx context.Context, template *entities.NotificationTemplate) error
	Delete(ctx context.Context, id int64) error

	// Búsquedas
	List(ctx context.Context, activeOnly bool) ([]*entities.NotificationTemplate, error)
	ListByChannel(ctx context.Context, channel string) ([]*entities.NotificationTemplate, error)
	ListByCategory(ctx context.Context, category string) ([]*entities.NotificationTemplate, error)
	Search(ctx context.Context, term string) ([]*entities.NotificationTemplate, error)

	// Operaciones específicas
	UpdateStatus(ctx context.Context, templateID int64, active bool) error
	UpdateContent(ctx context.Context, templateID int64, subjectTranslations, bodyTranslations map[string]string) error
	UpdateVariables(ctx context.Context, templateID int64, variables []string) error
	UpdatePriority(ctx context.Context, templateID int64, priority int) error
	AddTag(ctx context.Context, templateID int64, tag string) error
	RemoveTag(ctx context.Context, templateID int64, tag string) error
	SetTags(ctx context.Context, templateID int64, tags []string) error

	// Rendering
	RenderTemplate(ctx context.Context, templateCode, language string, data map[string]interface{}) (subject, body string, err error)
	GetAvailableVariables(ctx context.Context, templateCode string) ([]string, error)

	// Validaciones
	ValidateVariables(ctx context.Context, templateCode string, data map[string]interface{}) ([]string, error)
	IsActive(ctx context.Context, templateCode string) (bool, error)
	SupportsLanguage(ctx context.Context, templateCode, language string) (bool, error)
	SupportsChannel(ctx context.Context, templateCode, channel string) (bool, error)

	// Estadísticas
	GetUsageStats(ctx context.Context, templateCode string) (*entities.TemplateUsageStats, error)
	GetMostUsedTemplates(ctx context.Context, limit int) ([]*entities.TemplateUsage, error)
}
