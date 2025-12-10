// models.go - COMPLETO Y CORREGIDO
package models

import (
	"time"
)

// User representa un usuario del sistema
type User struct {
	ID                  int64      `json:"id"`
	PublicID            string     `json:"public_id"`
	Username            string     `json:"username"`
	Email               string     `json:"email"`
	PasswordHash        string     `json:"-"` // ⬅️ OMITIDO en JSON
	Role                string     `json:"role"`
	IsActive            bool       `json:"is_active"`
	LastLogin           *time.Time `json:"last_login,omitempty"`
	FailedLoginAttempts int32      `json:"failed_login_attempts"`
	PasswordChangedAt   *time.Time `json:"password_changed_at,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// Customer representa un cliente en el sistema
type Customer struct {
	ID                int64      `json:"id"`
	PublicID          string     `json:"public_id"`
	UserID            *int64     `json:"user_id,omitempty"`
	Name              string     `json:"name"`
	Email             string     `json:"email"`
	Phone             *string    `json:"phone,omitempty"`
	DateOfBirth       *time.Time `json:"date_of_birth,omitempty"`
	Address           *string    `json:"address,omitempty"`
	Preferences       *string    `json:"preferences,omitempty"`
	LoyaltyPoints     int32      `json:"loyalty_points"`
	IsVerified        bool       `json:"is_verified"`
	VerificationToken *string    `json:"-"`
	CustomerType      string     `json:"customer_type"`
	Source            string     `json:"source"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// Ticket representa un ticket individual
type Ticket struct {
	ID                      int64      `json:"id"`
	PublicID                string     `json:"public_id"`
	CategoryID              int64      `json:"category_id"`
	TransactionID           *int64     `json:"transaction_id,omitempty"`
	EventID                 int64      `json:"event_id"`
	CustomerID              int64      `json:"customer_id"`
	UserID                  *int64     `json:"user_id,omitempty"`
	Code                    string     `json:"code"`
	Status                  string     `json:"status"`
	SeatNumber              *string    `json:"seat_number,omitempty"`
	QRCodeURL               *string    `json:"qr_code_url,omitempty"`
	Price                   float64    `json:"price"`
	UsedAt                  *time.Time `json:"used_at,omitempty"`
	TransferredFromTicketID *int64     `json:"transferred_from_ticket_id,omitempty"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"`
}

// Event representa un evento en el sistema
type Event struct {
	ID               int64     `json:"id"`
	PublicID         string    `json:"public_id"`
	OrganizerID      *int64    `json:"organizer_id,omitempty"`
	Name             string    `json:"name"`
	Description      *string   `json:"description,omitempty"`
	ShortDescription *string   `json:"short_description,omitempty"`
	StartDate        time.Time `json:"start_date"`
	EndDate          time.Time `json:"end_date"`
	Location         string    `json:"location"`
	VenueDetails     *string   `json:"venue_details,omitempty"`
	Coordinates      *string   `json:"coordinates,omitempty"`
	Category         *string   `json:"category,omitempty"`
	Tags             []string  `json:"tags"`
	IsActive         bool      `json:"is_active"`
	IsPublished      bool      `json:"is_published"`
	ImageURL         *string   `json:"image_url,omitempty"`
	BannerURL        *string   `json:"banner_url,omitempty"`
	MaxAttendees     *int32    `json:"max_attendees,omitempty"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// Category representa una categoría de tickets
type Category struct {
	ID                 int64      `json:"id"`
	PublicID           string     `json:"public_id"`
	EventID            int64      `json:"event_id"`
	EventPublicID      string     `json:"event_public_id,omitempty"`
	Name               string     `json:"name"`
	Description        *string    `json:"description,omitempty"`
	Price              float64    `json:"price"`
	QuantityAvailable  int32      `json:"quantity_available"`
	QuantitySold       int32      `json:"quantity_sold"`
	MaxTicketsPerOrder int32      `json:"max_tickets_per_order"`
	SalesStart         time.Time  `json:"sales_start"`
	SalesEnd           *time.Time `json:"sales_end,omitempty"`
	IsActive           bool       `json:"is_active"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// Promotion representa una promoción o descuento
type Promotion struct {
	ID                int64     `json:"id"`
	PublicID          string    `json:"public_id"`
	Code              string    `json:"code"`
	Description       *string   `json:"description,omitempty"`
	DiscountType      string    `json:"discount_type"`
	DiscountValue     float64   `json:"discount_value"`
	EventID           *int64    `json:"event_id,omitempty"`
	CategoryID        *int64    `json:"category_id,omitempty"`
	MinOrderAmount    *float64  `json:"min_order_amount,omitempty"`
	MaxDiscountAmount *float64  `json:"max_discount_amount,omitempty"`
	ValidFrom         time.Time `json:"valid_from"`
	ValidTo           time.Time `json:"valid_to"`
	UsageLimit        *int32    `json:"usage_limit,omitempty"`
	UsageCount        int32     `json:"usage_count"`
	IsActive          bool      `json:"is_active"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// Transaction representa una transacción de pago
type Transaction struct {
	ID                    int64      `json:"id"`
	PublicID              string     `json:"public_id"`
	CustomerID            *int64     `json:"customer_id,omitempty"`
	PromotionID           *int64     `json:"promotion_id,omitempty"`
	Amount                float64    `json:"amount"`
	DiscountAmount        float64    `json:"discount_amount"`
	FinalAmount           float64    `json:"final_amount"`
	Currency              string     `json:"currency"`
	Status                string     `json:"status"`
	StripeSessionID       *string    `json:"stripe_session_id,omitempty"`
	StripePaymentIntentID *string    `json:"stripe_payment_intent_id,omitempty"`
	PaymentMethod         *string    `json:"payment_method,omitempty"`
	CustomerEmail         string     `json:"customer_email"`
	CustomerName          string     `json:"customer_name"`
	ReceiptURL            *string    `json:"receipt_url,omitempty"`
	Metadata              *string    `json:"metadata,omitempty"`
	ExpiresAt             *time.Time `json:"expires_at,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

// CategoryBenefit representa un beneficio de categoría
type CategoryBenefit struct {
	ID           int64     `json:"id"`
	PublicID     string    `json:"public_id"`
	CategoryID   int64     `json:"category_id"`
	Name         string    `json:"name"`
	Description  *string   `json:"description,omitempty"`
	IconURL      *string   `json:"icon_url,omitempty"`
	DisplayOrder int32     `json:"display_order"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
}

// =============================================
// REQUEST MODELS - MEJORAS APLICADAS
// =============================================

// CreateUserRequest representa la solicitud para crear un usuario
type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Role     string `json:"role" validate:"required,oneof=admin customer organizer guest"`
}

// CreateCustomerRequest representa la solicitud para crear un cliente
type CreateCustomerRequest struct {
	Name         string `json:"name" validate:"required"`
	Email        string `json:"email" validate:"required,email"`
	Phone        string `json:"phone,omitempty"`
	UserID       string `json:"user_id,omitempty"`       // UUID del usuario (opcional)
	CustomerType string `json:"customer_type,omitempty"` // registered, guest, corporate
	Source       string `json:"source,omitempty"`        // web, app, etc.
}

// CreateTicketRequest representa la solicitud para crear un ticket
type CreateTicketRequest struct {
	EventID    string `json:"event_id" validate:"required"`    // UUID del evento
	CustomerID string `json:"customer_id" validate:"required"` // UUID del cliente (OBLIGATORIO)
	UserID     string `json:"user_id,omitempty"`               // UUID del usuario (OPCIONAL)
	CategoryID string `json:"category_id" validate:"required"` // UUID de categoría
	Quantity   int32  `json:"quantity" validate:"required,min=1,max=10"`
}

// UpdateTicketStatusRequest representa la solicitud para actualizar estado de ticket
type UpdateTicketStatusRequest struct {
	TicketID string `json:"ticket_id" validate:"required"`
	Status   string `json:"status" validate:"required,oneof=available reserved sold used cancelled transferred refunded"`
}

// =============================================
// RESPONSE MODELS - MEJORAS APLICADAS
// =============================================

// TicketWithDetails representa un ticket con información completa
type TicketWithDetails struct {
	TicketID          string     `json:"ticket_id"`
	Code              string     `json:"code"`
	Status            string     `json:"status"`
	SeatNumber        string     `json:"seat_number,omitempty"`
	Price             float64    `json:"price"`
	CreatedAt         time.Time  `json:"created_at"`
	UsedAt            *time.Time `json:"used_at,omitempty"`
	CategoryID        string     `json:"category_id"`
	CategoryName      string     `json:"category_name"`
	EventID           string     `json:"event_id"`
	EventName         string     `json:"event_name"`
	StartDate         time.Time  `json:"start_date"`
	Location          string     `json:"location"`
	CustomerID        string     `json:"customer_id"`
	CustomerName      string     `json:"customer_name"`
	CustomerEmail     string     `json:"customer_email"`
	CustomerType      string     `json:"customer_type"`
	UserID            *string    `json:"user_id,omitempty"`
	UserName          *string    `json:"user_name,omitempty"`
	UserRole          *string    `json:"user_role,omitempty"`
	TransactionID     *string    `json:"transaction_id,omitempty"`
	TransactionStatus *string    `json:"transaction_status,omitempty"`
}

// UserStats representa estadísticas de usuarios
type UserStats struct {
	TotalUsers       int64 `json:"total_users"`
	ActiveUsers      int64 `json:"active_users"`
	InactiveUsers    int64 `json:"inactive_users"`
	CustomerUsers    int64 `json:"customer_users"`
	OrganizerUsers   int64 `json:"organizer_users"`
	AdminUsers       int64 `json:"admin_users"`
	GuestUsers       int64 `json:"guest_users"`
	ActiveLast30Days int64 `json:"active_last_30_days"`
}

// CustomerStats representa estadísticas de clientes
type CustomerStats struct {
	TotalCustomers      int64 `json:"total_customers"`
	VerifiedCustomers   int64 `json:"verified_customers"`
	GuestCustomers      int64 `json:"guest_customers"`
	CorporateCustomers  int64 `json:"corporate_customers"`
	RegisteredCustomers int64 `json:"registered_customers"`
}

// EventStats representa estadísticas de eventos
type EventStats struct {
	TotalEvents     int64 `json:"total_events"`
	ActiveEvents    int64 `json:"active_events"`
	PublishedEvents int64 `json:"published_events"`
	DraftEvents     int64 `json:"draft_events"`
	SoldOutEvents   int64 `json:"sold_out_events"`
	CancelledEvents int64 `json:"cancelled_events"`
	CompletedEvents int64 `json:"completed_events"`
}

// HealthCheck representa el estado del servicio
type HealthCheck struct {
	Status    string    `json:"status"`
	Service   string    `json:"service"`
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
}
