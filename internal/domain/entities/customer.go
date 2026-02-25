package entities

import "time"

// Customer representa un cliente en el sistema CRM
// Mapea exactamente la tabla crm.customers
type Customer struct {
	ID       int64  `json:"id" db:"id"`
	PublicID string `json:"public_id" db:"public_uuid"`
	UserID   *int64 `json:"user_id,omitempty" db:"user_id"`

	FullName string  `json:"full_name" db:"full_name"`
	Email    string  `json:"email" db:"email"`
	Phone    *string `json:"phone,omitempty" db:"phone"`

	CompanyName *string `json:"company_name,omitempty" db:"company_name"`

	AddressLine1 *string `json:"address_line1,omitempty" db:"address_line1"`
	AddressLine2 *string `json:"address_line2,omitempty" db:"address_line2"`
	City         *string `json:"city,omitempty" db:"city"`
	State        *string `json:"state,omitempty" db:"state"`
	PostalCode   *string `json:"postal_code,omitempty" db:"postal_code"`
	Country      *string `json:"country,omitempty" db:"country"`

	TaxID     *string `json:"tax_id,omitempty" db:"tax_id"`
	TaxIDType *string `json:"tax_id_type,omitempty" db:"tax_id_type"`
	TaxName   *string `json:"tax_name,omitempty" db:"tax_name"`

	RequiresInvoice bool `json:"requires_invoice" db:"requires_invoice"`

	// CORREGIDO: Añadido type:jsonb para campos JSONB
	CommunicationPreferences map[string]interface{} `json:"communication_preferences" db:"communication_preferences,type:jsonb"`

	// CORREGIDO: Los campos numéricos pueden ser float64 o int, pero en la BD son DECIMAL/INTEGER
	// Usamos tipos que coincidan exactamente con la BD
	TotalSpent    float64 `json:"total_spent" db:"total_spent"`         // DECIMAL(15,2)
	TotalOrders   int     `json:"total_orders" db:"total_orders"`       // INTEGER
	TotalTickets  int     `json:"total_tickets" db:"total_tickets"`     // INTEGER
	AvgOrderValue float64 `json:"avg_order_value" db:"avg_order_value"` // DECIMAL(10,2)

	FirstOrderAt   *time.Time `json:"first_order_at,omitempty" db:"first_order_at"`
	LastOrderAt    *time.Time `json:"last_order_at,omitempty" db:"last_order_at"`
	LastPurchaseAt *time.Time `json:"last_purchase_at,omitempty" db:"last_purchase_at"`

	IsActive bool       `json:"is_active" db:"is_active"`
	IsVIP    bool       `json:"is_vip" db:"is_vip"`
	VIPSince *time.Time `json:"vip_since,omitempty" db:"vip_since"`

	CustomerSegment string  `json:"customer_segment" db:"customer_segment"` // VARCHAR(50) con default 'new'
	LifetimeValue   float64 `json:"lifetime_value" db:"lifetime_value"`     // DECIMAL(15,2)

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Métodos de utilidad para Customer

// IsRegistered verifica si el cliente está asociado a un usuario registrado
func (c *Customer) IsRegistered() bool {
	return c.UserID != nil
}

// IsGuest verifica si el cliente es un invitado (no registrado)
func (c *Customer) IsGuest() bool {
	return c.UserID == nil
}

// HasCompleteAddress verifica si el cliente tiene dirección completa
func (c *Customer) HasCompleteAddress() bool {
	return c.AddressLine1 != nil &&
		c.City != nil &&
		c.State != nil &&
		c.PostalCode != nil &&
		c.Country != nil
}

// HasTaxInfo verifica si el cliente tiene información fiscal
func (c *Customer) HasTaxInfo() bool {
	return c.TaxID != nil && c.TaxIDType != nil && c.TaxName != nil
}

// GetFullAddress obtiene la dirección completa formateada
func (c *Customer) GetFullAddress() string {
	if !c.HasCompleteAddress() {
		return ""
	}

	address := *c.AddressLine1
	if c.AddressLine2 != nil && *c.AddressLine2 != "" {
		address += ", " + *c.AddressLine2
	}
	address += ", " + *c.City + ", " + *c.State + " " + *c.PostalCode + ", " + *c.Country

	return address
}

// UpdateStats actualiza las estadísticas del cliente basado en una nueva compra
func (c *Customer) UpdateStats(orderAmount float64, ticketCount int, orderTime time.Time) {
	// Actualizar totales
	c.TotalSpent += orderAmount
	c.TotalTickets += ticketCount
	c.TotalOrders++

	// Recalcular average order value
	if c.TotalOrders > 0 {
		c.AvgOrderValue = c.TotalSpent / float64(c.TotalOrders)
	}

	// Actualizar fechas
	if c.FirstOrderAt == nil {
		c.FirstOrderAt = &orderTime
	}
	c.LastOrderAt = &orderTime
	c.LastPurchaseAt = &orderTime

	// Actualizar segmento basado en actividad
	c.updateSegment()
}

// updateSegment actualiza el segmento del cliente basado en su actividad
func (c *Customer) updateSegment() {
	const (
		vipThreshold     = 10000.0 // $10,000 gastados
		regularThreshold = 1000.0  // $1,000 gastados
	)

	switch {
	case c.TotalSpent >= vipThreshold:
		c.CustomerSegment = "vip"
		if !c.IsVIP {
			c.IsVIP = true
			now := time.Now()
			c.VIPSince = &now
		}
	case c.TotalSpent >= regularThreshold:
		c.CustomerSegment = "regular"
	default:
		if c.TotalOrders == 0 {
			c.CustomerSegment = "new"
		} else {
			c.CustomerSegment = "occasional"
		}
	}
}

// GetCommunicationPreference obtiene una preferencia específica
func (c *Customer) GetCommunicationPreference(key string) bool {
	if c.CommunicationPreferences == nil {
		return false
	}
	val, ok := c.CommunicationPreferences[key]
	if !ok {
		return false
	}
	boolVal, ok := val.(bool)
	return ok && boolVal
}

// SetCommunicationPreference establece una preferencia de comunicación
func (c *Customer) SetCommunicationPreference(key string, value bool) {
	if c.CommunicationPreferences == nil {
		c.CommunicationPreferences = make(map[string]interface{})
	}
	c.CommunicationPreferences[key] = value
}
