package models

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// User representa un usuario del sistema
type User struct {
	ID                  int64            `json:"id"`
	PublicID            string           `json:"public_id"`
	Username            string           `json:"username"`
	Email               string           `json:"email"`
	PasswordHash        string           `json:"password_hash"`
	Role                string           `json:"role"`
	IsActive            bool             `json:"is_active"`
	LastLogin           pgtype.Timestamp `json:"last_login"`
	FailedLoginAttempts int32            `json:"failed_login_attempts"`
	PasswordChangedAt   pgtype.Timestamp `json:"password_changed_at"`
	CreatedAt           time.Time        `json:"created_at"`
	UpdatedAt           time.Time        `json:"updated_at"`
}

// Customer representa un cliente en el sistema
type Customer struct {
	ID                int64       `json:"id"`
	PublicID          string      `json:"public_id"`
	UserID            pgtype.Int4 `json:"user_id"` // Relación con users
	Name              string      `json:"name"`
	Email             string      `json:"email"`
	Phone             pgtype.Text `json:"phone"`
	DateOfBirth       pgtype.Date `json:"date_of_birth"`
	Address           pgtype.Text `json:"address"`     // JSONB
	Preferences       pgtype.Text `json:"preferences"` // JSONB
	LoyaltyPoints     int32       `json:"loyalty_points"`
	IsVerified        bool        `json:"is_verified"`
	VerificationToken pgtype.Text `json:"verification_token"`
	CreatedAt         time.Time   `json:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at"`
}

// Ticket representa un ticket individual
type Ticket struct {
	ID                      int64            `json:"id"`
	PublicID                string           `json:"public_id"`
	CategoryID              int64            `json:"category_id"`
	TransactionID           pgtype.Int4      `json:"transaction_id"`
	EventID                 int64            `json:"event_id"`
	Code                    string           `json:"code"`
	Status                  string           `json:"status"`
	SeatNumber              pgtype.Text      `json:"seat_number"`
	QRCodeURL               pgtype.Text      `json:"qr_code_url"`
	Price                   float64          `json:"price"`
	UsedAt                  pgtype.Timestamp `json:"used_at"`
	TransferredFromTicketID pgtype.Int4      `json:"transferred_from_ticket_id"`
	CreatedAt               time.Time        `json:"created_at"`
	UpdatedAt               time.Time        `json:"updated_at"`
}

// Event representa un evento en el sistema
type Event struct {
	ID               int64       `json:"id"`
	PublicID         string      `json:"public_id"`
	OrganizerID      pgtype.Int4 `json:"organizer_id"`
	Name             string      `json:"name"`
	Description      pgtype.Text `json:"description"`
	ShortDescription pgtype.Text `json:"short_description"`
	StartDate        time.Time   `json:"start_date"`
	EndDate          time.Time   `json:"end_date"`
	Location         string      `json:"location"`
	VenueDetails     pgtype.Text `json:"venue_details"`
	Coordinates      pgtype.Text `json:"coordinates"` // POINT type como texto
	Category         pgtype.Text `json:"category"`
	Tags             []string    `json:"tags"` // ARRAY de strings
	IsActive         bool        `json:"is_active"`
	IsPublished      bool        `json:"is_published"`
	ImageURL         pgtype.Text `json:"image_url"`
	BannerURL        pgtype.Text `json:"banner_url"`
	MaxAttendees     pgtype.Int4 `json:"max_attendees"`
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at"`
}

// Category representa una categoría de tickets para un evento
type Category struct {
	ID                 int64            `json:"id"`
	PublicID           string           `json:"public_id"`
	EventID            int64            `json:"event_id"`
	Name               string           `json:"name"`
	Description        pgtype.Text      `json:"description"`
	Price              float64          `json:"price"`
	QuantityAvailable  int32            `json:"quantity_available"`
	QuantitySold       int32            `json:"quantity_sold"`
	MaxTicketsPerOrder int32            `json:"max_tickets_per_order"`
	SalesStart         time.Time        `json:"sales_start"`
	SalesEnd           pgtype.Timestamp `json:"sales_end"`
	IsActive           bool             `json:"is_active"`
	CreatedAt          time.Time        `json:"created_at"`
	UpdatedAt          time.Time        `json:"updated_at"`
}

// Promotion representa una promoción o descuento
type Promotion struct {
	ID                int64         `json:"id"`
	PublicID          string        `json:"public_id"`
	Code              string        `json:"code"`
	Description       pgtype.Text   `json:"description"`
	DiscountType      string        `json:"discount_type"`
	DiscountValue     float64       `json:"discount_value"`
	EventID           pgtype.Int4   `json:"event_id"`
	CategoryID        pgtype.Int4   `json:"category_id"`
	MinOrderAmount    pgtype.Float8 `json:"min_order_amount"`
	MaxDiscountAmount pgtype.Float8 `json:"max_discount_amount"`
	ValidFrom         time.Time     `json:"valid_from"`
	ValidTo           time.Time     `json:"valid_to"`
	UsageLimit        pgtype.Int4   `json:"usage_limit"`
	UsageCount        int32         `json:"usage_count"`
	IsActive          bool          `json:"is_active"`
	CreatedAt         time.Time     `json:"created_at"`
	UpdatedAt         time.Time     `json:"updated_at"`
}

// Transaction representa una transacción de pago
type Transaction struct {
	ID                  int64            `json:"id"`
	PublicID            string           `json:"public_id"`
	CustomerID          pgtype.Int4      `json:"customer_id"`
	PromotionID         pgtype.Int4      `json:"promotion_id"`
	Amount              float64          `json:"amount"`
	DiscountAmount      float64          `json:"discount_amount"`
	FinalAmount         float64          `json:"final_amount"`
	Currency            string           `json:"currency"`
	Status              string           `json:"status"`
	StripeSessionID     pgtype.Text      `json:"stripe_session_id"`
	StripePaymentIntent pgtype.Text      `json:"stripe_payment_intent_id"`
	PaymentMethod       pgtype.Text      `json:"payment_method"`
	CustomerEmail       string           `json:"customer_email"`
	CustomerName        string           `json:"customer_name"`
	ReceiptURL          pgtype.Text      `json:"receipt_url"`
	Metadata            pgtype.Text      `json:"metadata"` // JSONB
	ExpiresAt           pgtype.Timestamp `json:"expires_at"`
	CreatedAt           time.Time        `json:"created_at"`
	UpdatedAt           time.Time        `json:"updated_at"`
}
