// internal/domain/entities/ticket.go - VERSIÓN CORREGIDA
package entities

import (
	"errors"
	"time"
)

// Ticket representa un ticket individual
// Mapea exactamente la tabla ticketing.tickets
type Ticket struct {
	ID           int64  `json:"id" db:"id"`
	PublicID     string `json:"public_id" db:"public_uuid"`
	TicketTypeID int64  `json:"ticket_type_id" db:"ticket_type_id"`
	EventID      int64  `json:"event_id" db:"event_id"`
	CustomerID   *int64 `json:"customer_id,omitempty" db:"customer_id"`
	OrderID      *int64 `json:"order_id,omitempty" db:"order_id"`

	Code       string  `json:"code" db:"code"`
	SecretHash string  `json:"-" db:"secret_hash"` // Nunca se expone en JSON
	QRCodeData *string `json:"qr_code_data,omitempty" db:"qr_code_data"`

	Status string `json:"status" db:"status"` // available, reserved, sold, checked_in, cancelled, refunded, expired

	FinalPrice float64 `json:"final_price" db:"final_price"`
	Currency   string  `json:"currency" db:"currency"`
	TaxAmount  float64 `json:"tax_amount" db:"tax_amount"`

	AttendeeName  *string `json:"attendee_name,omitempty" db:"attendee_name"`
	AttendeeEmail *string `json:"attendee_email,omitempty" db:"attendee_email"`
	AttendeePhone *string `json:"attendee_phone,omitempty" db:"attendee_phone"`

	CheckedInAt     *time.Time `json:"checked_in_at,omitempty" db:"checked_in_at"`
	CheckedInBy     *int64     `json:"checked_in_by,omitempty" db:"checked_in_by"`
	CheckinMethod   *string    `json:"checkin_method,omitempty" db:"checkin_method"`
	CheckinLocation *string    `json:"checkin_location,omitempty" db:"checkin_location"`

	ReservedAt           *time.Time `json:"reserved_at,omitempty" db:"reserved_at"`
	ReservedBy           *int64     `json:"reserved_by,omitempty" db:"reserved_by"`
	ReservationExpiresAt *time.Time `json:"reservation_expires_at,omitempty" db:"reservation_expires_at"`

	TransferToken *string `json:"transfer_token,omitempty" db:"transfer_token"`
	// CORREGIDO: transferred_from hace referencia a customer_id, no a ticket_id
	TransferredFrom *int64     `json:"transferred_from,omitempty" db:"transferred_from"`
	TransferredAt   *time.Time `json:"transferred_at,omitempty" db:"transferred_at"`

	// CORREGIDO: validation_count es INTEGER, usamos int
	ValidationCount int        `json:"validation_count" db:"validation_count"`
	LastValidatedAt *time.Time `json:"last_validated_at,omitempty" db:"last_validated_at"`

	SoldAt      *time.Time `json:"sold_at,omitempty" db:"sold_at"`
	CancelledAt *time.Time `json:"cancelled_at,omitempty" db:"cancelled_at"`
	RefundedAt  *time.Time `json:"refunded_at,omitempty" db:"refunded_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// Métodos de utilidad para Ticket

// IsAvailable verifica si el ticket está disponible
func (t *Ticket) IsAvailable() bool {
	return t.Status == "available"
}

// IsReserved verifica si el ticket está reservado
func (t *Ticket) IsReserved() bool {
	return t.Status == "reserved"
}

// IsSold verifica si el ticket está vendido
func (t *Ticket) IsSold() bool {
	return t.Status == "sold"
}

// IsCheckedIn verifica si el ticket ha sido usado (check-in)
func (t *Ticket) IsCheckedIn() bool {
	return t.Status == "checked_in" || t.CheckedInAt != nil
}

// IsCancelled verifica si el ticket está cancelado
func (t *Ticket) IsCancelled() bool {
	return t.Status == "cancelled" || t.CancelledAt != nil
}

// IsRefunded verifica si el ticket está reembolsado
func (t *Ticket) IsRefunded() bool {
	return t.Status == "refunded" || t.RefundedAt != nil
}

// IsExpired verifica si el ticket ha expirado
func (t *Ticket) IsExpired() bool {
	return t.Status == "expired"
}

// IsActive verifica si el ticket está activo (disponible para usar)
func (t *Ticket) IsActive() bool {
	return t.IsSold() && !t.IsCheckedIn() && !t.IsCancelled() && !t.IsRefunded() && !t.IsExpired()
}

// CanBeCheckedIn verifica si el ticket puede ser marcado como usado
func (t *Ticket) CanBeCheckedIn() bool {
	return t.IsSold() && !t.IsCheckedIn() && !t.IsCancelled() && !t.IsRefunded()
}

// CanBeCancelled verifica si el ticket puede ser cancelado
func (t *Ticket) CanBeCancelled() bool {
	return (t.IsAvailable() || t.IsReserved() || t.IsSold()) && !t.IsCheckedIn() && !t.IsCancelled() && !t.IsRefunded()
}

// CanBeRefunded verifica si el ticket puede ser reembolsado
func (t *Ticket) CanBeRefunded() bool {
	return t.IsSold() && !t.IsCheckedIn() && !t.IsRefunded()
}

// CanBeTransferred verifica si el ticket puede ser transferido
func (t *Ticket) CanBeTransferred() bool {
	return t.IsSold() && !t.IsCheckedIn() && !t.IsCancelled() && !t.IsRefunded()
}

// MarkAsSold marca el ticket como vendido
func (t *Ticket) MarkAsSold(customerID int64, orderID int64, finalPrice float64, currency string, taxAmount float64) {
	now := time.Now()
	t.Status = "sold"
	t.CustomerID = &customerID
	t.OrderID = &orderID
	t.FinalPrice = finalPrice
	t.Currency = currency
	t.TaxAmount = taxAmount
	t.SoldAt = &now
	t.UpdatedAt = now

	// Limpiar campos de reserva si existían
	t.ReservedAt = nil
	t.ReservedBy = nil
	t.ReservationExpiresAt = nil
}

// MarkAsReserved marca el ticket como reservado
func (t *Ticket) MarkAsReserved(reservedBy int64, expiresAt time.Time) {
	now := time.Now()
	t.Status = "reserved"
	t.ReservedAt = &now
	t.ReservedBy = &reservedBy
	t.ReservationExpiresAt = &expiresAt
	t.UpdatedAt = now
}

// MarkAsCheckedIn marca el ticket como usado (check-in)
func (t *Ticket) MarkAsCheckedIn(checkedInBy int64, method string, location string) {
	now := time.Now()
	t.Status = "checked_in"
	t.CheckedInAt = &now
	t.CheckedInBy = &checkedInBy
	t.CheckinMethod = &method
	t.CheckinLocation = &location
	t.UpdatedAt = now
}

// MarkAsCancelled marca el ticket como cancelado
func (t *Ticket) MarkAsCancelled() {
	now := time.Now()
	t.Status = "cancelled"
	t.CancelledAt = &now
	t.UpdatedAt = now
}

// MarkAsRefunded marca el ticket como reembolsado
func (t *Ticket) MarkAsRefunded() {
	now := time.Now()
	t.Status = "refunded"
	t.RefundedAt = &now
	t.UpdatedAt = now
}

// MarkAsExpired marca el ticket como expirado
func (t *Ticket) MarkAsExpired() {
	now := time.Now()
	t.Status = "expired"
	t.UpdatedAt = now
}

// Transfer transfiere el ticket a otro cliente
func (t *Ticket) Transfer(fromCustomerID int64, toCustomerID int64, transferToken string) {
	now := time.Now()
	t.TransferredFrom = &fromCustomerID
	t.CustomerID = &toCustomerID
	t.TransferToken = &transferToken
	t.TransferredAt = &now
	t.UpdatedAt = now
}

// IncrementValidation incrementa el contador de validaciones
func (t *Ticket) IncrementValidation() {
	t.ValidationCount++
	now := time.Now()
	t.LastValidatedAt = &now
	t.UpdatedAt = now
}

// IsReservationExpired verifica si la reserva ha expirado
func (t *Ticket) IsReservationExpired() bool {
	if !t.IsReserved() || t.ReservationExpiresAt == nil {
		return false
	}
	return time.Now().After(*t.ReservationExpiresAt)
}

// GetTimeUntilExpiry obtiene el tiempo hasta que expire la reserva
func (t *Ticket) GetTimeUntilExpiry() time.Duration {
	if !t.IsReserved() || t.ReservationExpiresAt == nil {
		return 0
	}
	return t.ReservationExpiresAt.Sub(time.Now())
}

// Validate verifica que el ticket sea válido
func (t *Ticket) Validate() error {
	if t.TicketTypeID == 0 {
		return errors.New("ticket_type_id is required")
	}
	if t.EventID == 0 {
		return errors.New("event_id is required")
	}
	if t.Code == "" {
		return errors.New("code is required")
	}
	if t.SecretHash == "" {
		return errors.New("secret_hash is required")
	}
	if t.Status == "" {
		return errors.New("status is required")
	}
	if t.FinalPrice < 0 {
		return errors.New("final_price cannot be negative")
	}
	if t.Currency == "" {
		return errors.New("currency is required")
	}
	if t.TaxAmount < 0 {
		return errors.New("tax_amount cannot be negative")
	}

	// Validar reglas específicas según estado
	if t.IsReserved() && t.ReservationExpiresAt == nil {
		return errors.New("reserved tickets must have reservation_expires_at")
	}

	return nil
}

// GetAttendeeInfo obtiene la información del asistente
func (t *Ticket) GetAttendeeInfo() map[string]interface{} {
	info := make(map[string]interface{})

	if t.AttendeeName != nil {
		info["name"] = *t.AttendeeName
	}
	if t.AttendeeEmail != nil {
		info["email"] = *t.AttendeeEmail
	}
	if t.AttendeePhone != nil {
		info["phone"] = *t.AttendeePhone
	}

	return info
}

// SetAttendeeInfo establece la información del asistente
func (t *Ticket) SetAttendeeInfo(name, email, phone string) {
	if name != "" {
		t.AttendeeName = &name
	}
	if email != "" {
		t.AttendeeEmail = &email
	}
	if phone != "" {
		t.AttendeePhone = &phone
	}
	t.UpdatedAt = time.Now()
}
