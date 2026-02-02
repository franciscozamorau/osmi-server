package repository

import (
	"context"

	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
)

// CountryConfigRepository define operaciones para configuración por país
type CountryConfigRepository interface {
	// CRUD básico
	Create(ctx context.Context, config *entities.CountryConfig) error
	FindByID(ctx context.Context, id int64) (*entities.CountryConfig, error)
	FindByCountryCode(ctx context.Context, countryCode string) (*entities.CountryConfig, error)
	Update(ctx context.Context, config *entities.CountryConfig) error
	Delete(ctx context.Context, id int64) error

	// Búsquedas
	List(ctx context.Context, activeOnly bool) ([]*entities.CountryConfig, error)
	ListByTaxSystem(ctx context.Context, taxSystem string) ([]*entities.CountryConfig, error)

	// Operaciones específicas
	UpdateTaxRate(ctx context.Context, countryCode string, taxRate float64) error
	UpdateTaxSystem(ctx context.Context, countryCode string, taxSystem string) error
	UpdateInvoiceSettings(ctx context.Context, countryCode string, requiresInvoice bool, format string) error
	UpdateCountrySettings(ctx context.Context, countryCode string, settings map[string]interface{}) error
	ActivateCountry(ctx context.Context, countryCode string) error
	DeactivateCountry(ctx context.Context, countryCode string) error

	// Validaciones fiscales
	ValidateTaxID(ctx context.Context, countryCode, taxID string) (bool, error)
	GetTaxIDRegex(ctx context.Context, countryCode string) (string, error)
	GetTaxIDType(ctx context.Context, countryCode, taxID string) (string, error)
	IsVATReverseCharge(ctx context.Context, countryCode string) (bool, error)
	IsInvoiceRequired(ctx context.Context, countryCode string) (bool, error)

	// Configuración específica por país
	GetMXCFDISettings(ctx context.Context) (*entities.MXCFDISettings, error)
	GetUSSettings(ctx context.Context) (*entities.USSettings, error)
	GetEUSettings(ctx context.Context) (*entities.EUSettings, error)

	// Consultas
	GetDefaultTaxRate(ctx context.Context, countryCode string) (float64, error)
	IsTaxInclusive(ctx context.Context, countryCode string) (bool, error)
	GetInvoiceFormat(ctx context.Context, countryCode string) (string, error)
	GetSupportedPaymentMethods(ctx context.Context, countryCode string) ([]string, error)
}
