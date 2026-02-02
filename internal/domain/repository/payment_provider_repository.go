package repository

import (
	"context"

	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
)

// PaymentProviderRepository define operaciones para proveedores de pago
type PaymentProviderRepository interface {
	// CRUD básico
	Create(ctx context.Context, provider *entities.PaymentProvider) error
	FindByID(ctx context.Context, id int64) (*entities.PaymentProvider, error)
	FindByCode(ctx context.Context, code string) (*entities.PaymentProvider, error)
	Update(ctx context.Context, provider *entities.PaymentProvider) error
	Delete(ctx context.Context, id int64) error

	// Búsquedas
	List(ctx context.Context, activeOnly bool) ([]*entities.PaymentProvider, error)
	ListByCountry(ctx context.Context, countryCode string) ([]*entities.PaymentProvider, error)
	ListByCurrency(ctx context.Context, currency string) ([]*entities.PaymentProvider, error)
	ListByType(ctx context.Context, providerType string) ([]*entities.PaymentProvider, error)

	// Operaciones específicas
	UpdateStatus(ctx context.Context, providerID int64, active bool) error
	UpdateConfig(ctx context.Context, providerID int64, config map[string]interface{}) error
	AddSupportedCurrency(ctx context.Context, providerID int64, currency string) error
	RemoveSupportedCurrency(ctx context.Context, providerID int64, currency string) error
	AddSupportedCountry(ctx context.Context, providerID int64, countryCode string) error
	RemoveSupportedCountry(ctx context.Context, providerID int64, countryCode string) error
	UpdateLimits(ctx context.Context, providerID int64, minAmount, maxAmount float64) error
	TestConnection(ctx context.Context, providerID int64) (bool, error)

	// Verificaciones
	IsCurrencySupported(ctx context.Context, providerID int64, currency string) (bool, error)
	IsCountrySupported(ctx context.Context, providerID int64, countryCode string) (bool, error)
	IsAmountInRange(ctx context.Context, providerID int64, amount float64) (bool, error)
	SupportsRefunds(ctx context.Context, providerID int64) (bool, error)
	IsOnline(ctx context.Context, providerID int64) (bool, error)

	// Estadísticas
	GetProviderStats(ctx context.Context, providerID int64) (*entities.ProviderStats, error)
	CountTransactions(ctx context.Context, providerID int64) (int64, error)
	GetTotalProcessed(ctx context.Context, providerID int64) (float64, error)
	GetSuccessRate(ctx context.Context, providerID int64) (float64, error)
}
