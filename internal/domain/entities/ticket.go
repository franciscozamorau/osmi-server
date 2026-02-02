package entities

import "time"

type Ticket struct {
	ID           int64  `json:"id" db:"id"`
	PublicID     string `json:"public_id" db:"public_uuid"`
	TicketTypeID int64  `json:"ticket_type_id" db:"ticket_type_id"`
	EventID      int64  `json:"event_id" db:"event_id"`
	CustomerID   *int64 `json:"customer_id,omitempty" db:"customer_id"`
	OrderID      *int64 `json:"order_id,omitempty" db:"order_id"`

	Code       string  `json:"code" db:"code"`
	SecretHash string  `json:"-" db:"secret_hash"`
	QRCodeData *string `json:"qr_code_data,omitempty" db:"qr_code_data"`

	Status string `json:"status" db:"status"`

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

	TransferToken   *string    `json:"transfer_token,omitempty" db:"transfer_token"`
	TransferredFrom *int64     `json:"transferred_from,omitempty" db:"transferred_from"`
	TransferredAt   *time.Time `json:"transferred_at,omitempty" db:"transferred_at"`

	ValidationCount int32      `json:"validation_count" db:"validation_count"`
	LastValidatedAt *time.Time `json:"last_validated_at,omitempty" db:"last_validated_at"`

	SoldAt      *time.Time `json:"sold_at,omitempty" db:"sold_at"`
	CancelledAt *time.Time `json:"cancelled_at,omitempty" db:"cancelled_at"`
	RefundedAt  *time.Time `json:"refunded_at,omitempty" db:"refunded_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}
