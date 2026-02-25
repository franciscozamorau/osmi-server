package entities

import (
	"errors"
	"time"
)

// TicketType representa un tipo de ticket para un evento
// Mapea exactamente la tabla ticketing.ticket_types
type TicketType struct {
	ID       int64  `json:"id" db:"id"`
	PublicID string `json:"public_id" db:"public_uuid"`
	EventID  int64  `json:"event_id" db:"event_id"`

	Name        string  `json:"name" db:"name"`
	Description *string `json:"description,omitempty" db:"description"`
	TicketClass string  `json:"ticket_class" db:"ticket_class"`

	BasePrice       float64 `json:"base_price" db:"base_price"`
	Currency        string  `json:"currency" db:"currency"`
	TaxRate         float64 `json:"tax_rate" db:"tax_rate"`
	ServiceFeeType  string  `json:"service_fee_type" db:"service_fee_type"`
	ServiceFeeValue float64 `json:"service_fee_value" db:"service_fee_value"`

	// CORREGIDO: Usamos int en lugar de int32 para INTEGER en PostgreSQL
	TotalQuantity    int `json:"total_quantity" db:"total_quantity"`
	ReservedQuantity int `json:"reserved_quantity" db:"reserved_quantity"`
	SoldQuantity     int `json:"sold_quantity" db:"sold_quantity"`
	MaxPerOrder      int `json:"max_per_order" db:"max_per_order"`
	MinPerOrder      int `json:"min_per_order" db:"min_per_order"`

	SaleStartsAt time.Time  `json:"sale_starts_at" db:"sale_starts_at"`
	SaleEndsAt   *time.Time `json:"sale_ends_at,omitempty" db:"sale_ends_at"`

	IsActive         bool   `json:"is_active" db:"is_active"`
	RequiresApproval bool   `json:"requires_approval" db:"requires_approval"`
	IsHidden         bool   `json:"is_hidden" db:"is_hidden"`
	SalesChannel     string `json:"sales_channel" db:"sales_channel"`

	// CORREGIDO: benefits es JSONB
	Benefits   *[]string `json:"benefits,omitempty" db:"benefits,type:jsonb"`
	AccessType string    `json:"access_type" db:"access_type"`

	// CORREGIDO: validation_rules es JSONB con estructura específica
	ValidationRules *ValidationRules `json:"validation_rules,omitempty" db:"validation_rules,type:jsonb"`

	// CAMPOS GENERADOS (STORED) - No se actualizan directamente
	AvailableQuantity int  `json:"available_quantity" db:"available_quantity"`
	IsSoldOut         bool `json:"is_sold_out" db:"is_sold_out"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// ValidationRules representa las reglas de validación para el ticket
type ValidationRules struct {
	RequiresID         bool `json:"requires_id"`
	AgeRestriction     int  `json:"age_restriction"`
	RequiresMembership bool `json:"requires_membership"`
}

// Métodos de utilidad para TicketType

// IsAvailable verifica si hay tickets disponibles
func (tt *TicketType) IsAvailable() bool {
	return tt.IsActive &&
		!tt.IsSoldOut &&
		tt.AvailableQuantity > 0 &&
		tt.IsOnSale()
}

// IsOnSale verifica si el período de venta está activo
func (tt *TicketType) IsOnSale() bool {
	now := time.Now()

	// Verificar si ya empezó la venta
	if now.Before(tt.SaleStartsAt) {
		return false
	}

	// Verificar si ya terminó la venta
	if tt.SaleEndsAt != nil && now.After(*tt.SaleEndsAt) {
		return false
	}

	return true
}

// GetAvailableQuantity obtiene la cantidad disponible
func (tt *TicketType) GetAvailableQuantity() int {
	return tt.TotalQuantity - tt.SoldQuantity - tt.ReservedQuantity
}

// UpdateAvailableQuantity actualiza la cantidad disponible (útil para cálculos)
func (tt *TicketType) UpdateAvailableQuantity() {
	tt.AvailableQuantity = tt.GetAvailableQuantity()
	tt.IsSoldOut = tt.AvailableQuantity <= 0
}

// Reserve reserva una cantidad de tickets
func (tt *TicketType) Reserve(quantity int) error {
	if quantity <= 0 {
		return errors.New("quantity must be positive")
	}

	available := tt.GetAvailableQuantity()
	if available < quantity {
		return errors.New("insufficient available tickets")
	}

	tt.ReservedQuantity += quantity
	tt.UpdateAvailableQuantity()
	tt.UpdatedAt = time.Now()

	return nil
}

// Release libera una cantidad de tickets reservados
func (tt *TicketType) Release(quantity int) error {
	if quantity <= 0 {
		return errors.New("quantity must be positive")
	}

	if tt.ReservedQuantity < quantity {
		return errors.New("cannot release more than reserved")
	}

	tt.ReservedQuantity -= quantity
	tt.UpdateAvailableQuantity()
	tt.UpdatedAt = time.Now()

	return nil
}

// Sell vende una cantidad de tickets
func (tt *TicketType) Sell(quantity int) error {
	if quantity <= 0 {
		return errors.New("quantity must be positive")
	}

	available := tt.GetAvailableQuantity()
	if available < quantity {
		return errors.New("insufficient available tickets")
	}

	tt.SoldQuantity += quantity
	tt.UpdateAvailableQuantity()
	tt.UpdatedAt = time.Now()

	return nil
}

// GetFinalPrice calcula el precio final incluyendo fees
func (tt *TicketType) GetFinalPrice() float64 {
	finalPrice := tt.BasePrice

	// Aplicar service fee según el tipo
	switch tt.ServiceFeeType {
	case "percentage":
		finalPrice += tt.BasePrice * tt.ServiceFeeValue
	case "fixed":
		finalPrice += tt.ServiceFeeValue
	}

	// Aplicar impuestos
	finalPrice += finalPrice * tt.TaxRate

	return finalPrice
}

// GetBasePriceWithTax obtiene el precio base con impuestos
func (tt *TicketType) GetBasePriceWithTax() float64 {
	return tt.BasePrice * (1 + tt.TaxRate)
}

// ValidateOrderQuantity verifica si una cantidad es válida para ordenar
func (tt *TicketType) ValidateOrderQuantity(quantity int) error {
	if quantity < tt.MinPerOrder {
		return errors.New("quantity is below minimum per order")
	}
	if quantity > tt.MaxPerOrder {
		return errors.New("quantity exceeds maximum per order")
	}
	return nil
}

// AddBenefit añade un beneficio al ticket
func (tt *TicketType) AddBenefit(benefit string) {
	if tt.Benefits == nil {
		tt.Benefits = &[]string{}
	}

	// Verificar si ya existe
	for _, b := range *tt.Benefits {
		if b == benefit {
			return
		}
	}

	*tt.Benefits = append(*tt.Benefits, benefit)
	tt.UpdatedAt = time.Now()
}

// RemoveBenefit elimina un beneficio
func (tt *TicketType) RemoveBenefit(benefit string) {
	if tt.Benefits == nil {
		return
	}

	newBenefits := []string{}
	for _, b := range *tt.Benefits {
		if b != benefit {
			newBenefits = append(newBenefits, b)
		}
	}

	if len(newBenefits) == 0 {
		tt.Benefits = nil
	} else {
		*tt.Benefits = newBenefits
	}
	tt.UpdatedAt = time.Now()
}

// HasBenefit verifica si tiene un beneficio específico
func (tt *TicketType) HasBenefit(benefit string) bool {
	if tt.Benefits == nil {
		return false
	}

	for _, b := range *tt.Benefits {
		if b == benefit {
			return true
		}
	}
	return false
}

// SetValidationRules establece las reglas de validación
func (tt *TicketType) SetValidationRules(rules ValidationRules) {
	tt.ValidationRules = &rules
	tt.UpdatedAt = time.Now()
}

// GetValidationRules obtiene las reglas de validación
func (tt *TicketType) GetValidationRules() ValidationRules {
	if tt.ValidationRules == nil {
		return ValidationRules{
			RequiresID:         false,
			AgeRestriction:     0,
			RequiresMembership: false,
		}
	}
	return *tt.ValidationRules
}

// Validate verifica que el ticket type sea válido
func (tt *TicketType) Validate() error {
	if tt.EventID == 0 {
		return errors.New("event_id is required")
	}
	if tt.Name == "" {
		return errors.New("name is required")
	}
	if tt.BasePrice < 0 {
		return errors.New("base_price cannot be negative")
	}
	if tt.Currency == "" {
		return errors.New("currency is required")
	}
	if tt.TotalQuantity <= 0 {
		return errors.New("total_quantity must be greater than 0")
	}
	if tt.MaxPerOrder < tt.MinPerOrder {
		return errors.New("max_per_order cannot be less than min_per_order")
	}
	if tt.SaleEndsAt != nil && tt.SaleEndsAt.Before(tt.SaleStartsAt) {
		return errors.New("sale_ends_at cannot be before sale_starts_at")
	}
	return nil
}

// IsFree verifica si el ticket es gratuito
func (tt *TicketType) IsFree() bool {
	return tt.BasePrice == 0 && tt.ServiceFeeValue == 0
}

// HasAgeRestriction verifica si tiene restricción de edad
func (tt *TicketType) HasAgeRestriction() bool {
	return tt.ValidationRules != nil && tt.ValidationRules.AgeRestriction > 0
}
