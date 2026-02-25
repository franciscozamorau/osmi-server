package entities

import (
	"errors"
	"time"
)

// PaymentProvider proveedor de pagos
// Mapea exactamente la tabla billing.payment_providers
type PaymentProvider struct {
	// CORREGIDO: SMALLSERIAL en la BD -> int16 en Go
	ID   int16  `json:"id" db:"id"`
	Code string `json:"code" db:"code"`
	Name string `json:"name" db:"name"`

	// CAMPOS FALTANTES
	ProviderType string `json:"provider_type" db:"provider_type"` // gateway, method

	IsActive        bool `json:"is_active" db:"is_active"`
	IsOnline        bool `json:"is_online" db:"is_online"`
	SupportsRefunds bool `json:"supports_refunds" db:"supports_refunds"`

	MinAmount float64  `json:"min_amount" db:"min_amount"`
	MaxAmount *float64 `json:"max_amount,omitempty" db:"max_amount"`

	// CORREGIDO: Arrays en PostgreSQL -> []string en Go
	SupportedCurrencies []string `json:"supported_currencies" db:"supported_currencies,type:text[]"`
	SupportedCountries  []string `json:"supported_countries" db:"supported_countries,type:text[]"`

	// CORREGIDO: JSONB en la BD
	Config *map[string]interface{} `json:"config,omitempty" db:"config,type:jsonb"`

	// CORREGIDO: time.Time en lugar de string
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// ProviderStats estadísticas del proveedor de pagos
type ProviderStats struct {
	TotalTransactions   int64   `json:"total_transactions"`
	SuccessfulTx        int64   `json:"successful_tx"`
	FailedTx            int64   `json:"failed_tx"`
	SuccessRate         float64 `json:"success_rate"`
	TotalVolume         float64 `json:"total_volume"`
	AvgTransactionValue float64 `json:"avg_transaction_value"`
	AvgProcessingTime   float64 `json:"avg_processing_time_ms"`
}

// Métodos de utilidad para PaymentProvider

// IsGateway verifica si es un gateway de pago
func (p *PaymentProvider) IsGateway() bool {
	return p.ProviderType == "gateway"
}

// IsPaymentMethod verifica si es un método de pago directo
func (p *PaymentProvider) IsPaymentMethod() bool {
	return p.ProviderType == "method"
}

// SupportsCurrency verifica si soporta una moneda específica
func (p *PaymentProvider) SupportsCurrency(currency string) bool {
	if p.SupportedCurrencies == nil {
		return false
	}
	for _, c := range p.SupportedCurrencies {
		if c == currency {
			return true
		}
	}
	return false
}

// SupportsCountry verifica si soporta un país específico
func (p *PaymentProvider) SupportsCountry(country string) bool {
	if p.SupportedCountries == nil {
		return false
	}
	for _, c := range p.SupportedCountries {
		if c == country {
			return true
		}
	}
	return false
}

// IsAmountInRange verifica si un monto está dentro del rango soportado
func (p *PaymentProvider) IsAmountInRange(amount float64) bool {
	if amount < p.MinAmount {
		return false
	}
	if p.MaxAmount != nil && amount > *p.MaxAmount {
		return false
	}
	return true
}

// GetConfigValue obtiene un valor de configuración
func (p *PaymentProvider) GetConfigValue(key string) interface{} {
	if p.Config == nil {
		return nil
	}
	return (*p.Config)[key]
}

// SetConfigValue establece un valor de configuración
func (p *PaymentProvider) SetConfigValue(key string, value interface{}) {
	if p.Config == nil {
		p.Config = &map[string]interface{}{}
	}
	(*p.Config)[key] = value
}

// DeleteConfigValue elimina un valor de configuración
func (p *PaymentProvider) DeleteConfigValue(key string) {
	if p.Config == nil {
		return
	}
	delete(*p.Config, key)
	if len(*p.Config) == 0 {
		p.Config = nil
	}
}

// GetMinAmountFloat obtiene el monto mínimo como float64
func (p *PaymentProvider) GetMinAmountFloat() float64 {
	return p.MinAmount
}

// GetMaxAmountFloat obtiene el monto máximo como float64 (o 0 si es nil)
func (p *PaymentProvider) GetMaxAmountFloat() float64 {
	if p.MaxAmount == nil {
		return 0
	}
	return *p.MaxAmount
}

// AddSupportedCurrency añade una moneda soportada
func (p *PaymentProvider) AddSupportedCurrency(currency string) {
	if p.SupportedCurrencies == nil {
		p.SupportedCurrencies = []string{}
	}

	// Verificar si ya existe
	for _, c := range p.SupportedCurrencies {
		if c == currency {
			return
		}
	}

	p.SupportedCurrencies = append(p.SupportedCurrencies, currency)
}

// RemoveSupportedCurrency elimina una moneda soportada
func (p *PaymentProvider) RemoveSupportedCurrency(currency string) {
	if p.SupportedCurrencies == nil {
		return
	}

	newCurrencies := []string{}
	for _, c := range p.SupportedCurrencies {
		if c != currency {
			newCurrencies = append(newCurrencies, c)
		}
	}

	if len(newCurrencies) == 0 {
		p.SupportedCurrencies = nil
	} else {
		p.SupportedCurrencies = newCurrencies
	}
}

// AddSupportedCountry añade un país soportado
func (p *PaymentProvider) AddSupportedCountry(country string) {
	if p.SupportedCountries == nil {
		p.SupportedCountries = []string{}
	}

	// Verificar si ya existe
	for _, c := range p.SupportedCountries {
		if c == country {
			return
		}
	}

	p.SupportedCountries = append(p.SupportedCountries, country)
}

// RemoveSupportedCountry elimina un país soportado
func (p *PaymentProvider) RemoveSupportedCountry(country string) {
	if p.SupportedCountries == nil {
		return
	}

	newCountries := []string{}
	for _, c := range p.SupportedCountries {
		if c != country {
			newCountries = append(newCountries, c)
		}
	}

	if len(newCountries) == 0 {
		p.SupportedCountries = nil
	} else {
		p.SupportedCountries = newCountries
	}
}

// Validate verifica que el proveedor de pagos sea válido
func (p *PaymentProvider) Validate() error {
	if p.Code == "" {
		return errors.New("code is required")
	}
	if p.Name == "" {
		return errors.New("name is required")
	}
	if p.ProviderType == "" {
		return errors.New("provider_type is required")
	}
	if p.MinAmount < 0 {
		return errors.New("min_amount cannot be negative")
	}
	if p.MaxAmount != nil && *p.MaxAmount < p.MinAmount {
		return errors.New("max_amount cannot be less than min_amount")
	}
	if len(p.SupportedCurrencies) == 0 {
		return errors.New("at least one supported currency is required")
	}
	return nil
}

// Clone crea una copia del PaymentProvider
func (p *PaymentProvider) Clone() *PaymentProvider {
	clone := &PaymentProvider{
		ID:              p.ID,
		Code:            p.Code,
		Name:            p.Name,
		ProviderType:    p.ProviderType,
		IsActive:        p.IsActive,
		IsOnline:        p.IsOnline,
		SupportsRefunds: p.SupportsRefunds,
		MinAmount:       p.MinAmount,
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
	}

	// Clonar MaxAmount
	if p.MaxAmount != nil {
		maxAmount := *p.MaxAmount
		clone.MaxAmount = &maxAmount
	}

	// Clonar SupportedCurrencies
	if p.SupportedCurrencies != nil {
		clone.SupportedCurrencies = make([]string, len(p.SupportedCurrencies))
		copy(clone.SupportedCurrencies, p.SupportedCurrencies)
	}

	// Clonar SupportedCountries
	if p.SupportedCountries != nil {
		clone.SupportedCountries = make([]string, len(p.SupportedCountries))
		copy(clone.SupportedCountries, p.SupportedCountries)
	}

	// Clonar Config
	if p.Config != nil {
		config := make(map[string]interface{})
		for k, v := range *p.Config {
			config[k] = v
		}
		clone.Config = &config
	}

	return clone
}
