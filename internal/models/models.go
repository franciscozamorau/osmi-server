package models

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type Customer struct {
	ID                int64       `json:"id"`
	PublicID          string      `json:"public_id"`
	Name              string      `json:"name"`
	Email             string      `json:"email"`
	Phone             pgtype.Text `json:"phone"`
	DateOfBirth       pgtype.Date `json:"date_of_birth"`
	Address           pgtype.Text `json:"address"` // JSONB se maneja como Text
	Preferences       pgtype.Text `json:"preferences"`
	LoyaltyPoints     int32       `json:"loyalty_points"`
	IsVerified        bool        `json:"is_verified"`
	VerificationToken pgtype.Text `json:"verification_token"`
	CreatedAt         time.Time   `json:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at"`
}

type Ticket struct {
	ID                      int64            `json:"id"`
	PublicID                string           `json:"public_id"`
	CategoryID              int64            `json:"category_id"`
	EventID                 int64            `json:"event_id"`
	UserID                  pgtype.Text      `json:"user_id"` // ✅ CORREGIDO: Cambiado de CustomerID int64 a UserID string (UUID)
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

type Event struct {
	ID               int64       `json:"id"`
	PublicID         string      `json:"public_id"`
	Name             string      `json:"name"`
	Description      pgtype.Text `json:"description"`
	ShortDescription pgtype.Text `json:"short_description"`
	StartDate        time.Time   `json:"start_date"`
	EndDate          time.Time   `json:"end_date"`
	Location         string      `json:"location"`
	VenueDetails     pgtype.Text `json:"venue_details"`
	Category         pgtype.Text `json:"category"`
	Tags             pgtype.Text `json:"tags"` // Array se maneja como Text
	IsActive         bool        `json:"is_active"`
	IsPublished      bool        `json:"is_published"`
	ImageURL         pgtype.Text `json:"image_url"`
	BannerURL        pgtype.Text `json:"banner_url"`
	MaxAttendees     pgtype.Int4 `json:"max_attendees"`
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at"`
}
