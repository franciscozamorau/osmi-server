package entities

import (
	"errors"
	"time"
)

// OrderItem item dentro de una orden
// Mapea exactamente la tabla billing.order_items
type OrderItem struct {
	ID           int64 `json:"id" db:"id"`
	OrderID      int64 `json:"order_id" db:"order_id"`
	TicketTypeID int64 `json:"ticket_type_id" db:"ticket_type_id"`

	Quantity   int     `json:"quantity" db:"quantity"`
	UnitPrice  float64 `json:"unit_price" db:"unit_price"`
	TotalPrice float64 `json:"total_price" db:"total_price"`
	Currency   string  `json:"currency" db:"currency"`

	BasePrice        float64 `json:"base_price" db:"base_price"`
	TaxAmount        float64 `json:"tax_amount" db:"tax_amount"`
	ServiceFeeAmount float64 `json:"service_fee_amount" db:"service_fee_amount"`
	DiscountAmount   float64 `json:"discount_amount" db:"discount_amount"`

	TicketIDs *[]int64                `json:"ticket_ids,omitempty" db:"ticket_ids,type:jsonb"`
	Metadata  *map[string]interface{} `json:"metadata,omitempty" db:"metadata,type:jsonb"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// CalculateTotals calcula los totales basado en quantity y precios
func (oi *OrderItem) CalculateTotals() {
	oi.TotalPrice = float64(oi.Quantity) * oi.UnitPrice
}

// GetSubtotal obtiene el subtotal (sin impuestos ni fees)
func (oi *OrderItem) GetSubtotal() float64 {
	return float64(oi.Quantity) * oi.BasePrice
}

// GetFinalUnitPrice obtiene el precio unitario final después de impuestos y fees
func (oi *OrderItem) GetFinalUnitPrice() float64 {
	return oi.BasePrice + oi.TaxAmount/float64(oi.Quantity) +
		oi.ServiceFeeAmount/float64(oi.Quantity) - oi.DiscountAmount/float64(oi.Quantity)
}

// Validate verifica que el OrderItem sea válido
func (oi *OrderItem) Validate() error {
	if oi.OrderID == 0 {
		return errors.New("order_id is required")
	}
	if oi.TicketTypeID == 0 {
		return errors.New("ticket_type_id is required")
	}
	if oi.Quantity <= 0 {
		return errors.New("quantity must be greater than 0")
	}
	if oi.UnitPrice < 0 {
		return errors.New("unit_price cannot be negative")
	}
	if oi.BasePrice < 0 {
		return errors.New("base_price cannot be negative")
	}
	if oi.TaxAmount < 0 {
		return errors.New("tax_amount cannot be negative")
	}
	if oi.ServiceFeeAmount < 0 {
		return errors.New("service_fee_amount cannot be negative")
	}
	if oi.DiscountAmount < 0 {
		return errors.New("discount_amount cannot be negative")
	}
	if oi.Currency == "" {
		return errors.New("currency is required")
	}

	// Verificar que discount_amount no exceda el subtotal
	if oi.DiscountAmount > oi.GetSubtotal() {
		return errors.New("discount_amount cannot exceed subtotal")
	}

	return nil
}

// AddTicketID añade un ID de ticket a la lista
func (oi *OrderItem) AddTicketID(ticketID int64) {
	if oi.TicketIDs == nil {
		oi.TicketIDs = &[]int64{}
	}

	// Verificar si ya existe
	for _, id := range *oi.TicketIDs {
		if id == ticketID {
			return
		}
	}

	*oi.TicketIDs = append(*oi.TicketIDs, ticketID)
}

// RemoveTicketID elimina un ID de ticket de la lista
func (oi *OrderItem) RemoveTicketID(ticketID int64) {
	if oi.TicketIDs == nil {
		return
	}

	newIDs := []int64{}
	for _, id := range *oi.TicketIDs {
		if id != ticketID {
			newIDs = append(newIDs, id)
		}
	}

	if len(newIDs) == 0 {
		oi.TicketIDs = nil
	} else {
		*oi.TicketIDs = newIDs
	}
}

// HasTicketID verifica si un ticket está en la lista
func (oi *OrderItem) HasTicketID(ticketID int64) bool {
	if oi.TicketIDs == nil {
		return false
	}

	for _, id := range *oi.TicketIDs {
		if id == ticketID {
			return true
		}
	}
	return false
}

// GetTicketCount obtiene el número de tickets en la lista
func (oi *OrderItem) GetTicketCount() int {
	if oi.TicketIDs == nil {
		return 0
	}
	return len(*oi.TicketIDs)
}

// SetMetadata establece un valor en metadata
func (oi *OrderItem) SetMetadata(key string, value interface{}) {
	if oi.Metadata == nil {
		oi.Metadata = &map[string]interface{}{}
	}
	(*oi.Metadata)[key] = value
}

// GetMetadata obtiene un valor de metadata
func (oi *OrderItem) GetMetadata(key string) interface{} {
	if oi.Metadata == nil {
		return nil
	}
	return (*oi.Metadata)[key]
}

// DeleteMetadata elimina una clave de metadata
func (oi *OrderItem) DeleteMetadata(key string) {
	if oi.Metadata == nil {
		return
	}
	delete(*oi.Metadata, key)
	if len(*oi.Metadata) == 0 {
		oi.Metadata = nil
	}
}

// Clone crea una copia del OrderItem
func (oi *OrderItem) Clone() *OrderItem {
	clone := &OrderItem{
		ID:               oi.ID,
		OrderID:          oi.OrderID,
		TicketTypeID:     oi.TicketTypeID,
		Quantity:         oi.Quantity,
		UnitPrice:        oi.UnitPrice,
		TotalPrice:       oi.TotalPrice,
		Currency:         oi.Currency,
		BasePrice:        oi.BasePrice,
		TaxAmount:        oi.TaxAmount,
		ServiceFeeAmount: oi.ServiceFeeAmount,
		DiscountAmount:   oi.DiscountAmount,
		CreatedAt:        oi.CreatedAt,
		UpdatedAt:        oi.UpdatedAt,
	}

	// Clonar TicketIDs
	if oi.TicketIDs != nil {
		ticketIDs := make([]int64, len(*oi.TicketIDs))
		copy(ticketIDs, *oi.TicketIDs)
		clone.TicketIDs = &ticketIDs
	}

	// Clonar Metadata
	if oi.Metadata != nil {
		metadata := make(map[string]interface{})
		for k, v := range *oi.Metadata {
			metadata[k] = v
		}
		clone.Metadata = &metadata
	}

	return clone
}
