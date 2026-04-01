// internal/api/dto/organizer/response.go
package organizer

import "time"

// OrganizerAddress representa la dirección del organizador
type OrganizerAddress struct {
	AddressLine1 string  `json:"address_line1"`
	AddressLine2 *string `json:"address_line2,omitempty"`
	City         string  `json:"city"`
	State        *string `json:"state,omitempty"`
	PostalCode   *string `json:"postal_code,omitempty"`
}

// OrganizerStats representa estadísticas del organizador
type OrganizerStats struct {
	TotalRevenue       float64
	AvgRevenuePerEvent float64
	AvgTicketsPerEvent float64
	SellOutRate        float64
	RepeatCustomerRate float64
	UpcomingEventCount int
	PastEventCount     int
	CancellationRate   float64
}

// VerificationDocument representa un documento de verificación
type VerificationDocument struct {
	DocumentType string     `json:"document_type"`
	DocumentURL  string     `json:"document_url"`
	UploadedAt   time.Time  `json:"uploaded_at"`
	Status       string     `json:"status"`
	ReviewedAt   *time.Time `json:"reviewed_at,omitempty"`
	Reviewer     *string    `json:"reviewer,omitempty"`
	Notes        *string    `json:"notes,omitempty"`
}

// OrganizerResponse representa la respuesta completa de un organizador
type OrganizerResponse struct {
	ID                 string            `json:"id"`
	Name               string            `json:"name"`
	Slug               string            `json:"slug"`
	Description        *string           `json:"description,omitempty"`
	LogoURL            *string           `json:"logo_url,omitempty"`
	LegalName          *string           `json:"legal_name,omitempty"`
	TaxID              *string           `json:"tax_id,omitempty"`
	TaxIDType          *string           `json:"tax_id_type,omitempty"`
	Country            *string           `json:"country,omitempty"`
	ContactEmail       string            `json:"contact_email"`
	ContactPhone       *string           `json:"contact_phone,omitempty"`
	Address            *OrganizerAddress `json:"address,omitempty"`
	IsVerified         bool              `json:"is_verified"`
	IsActive           bool              `json:"is_active"`
	VerificationStatus string            `json:"verification_status"`
	TotalEvents        int               `json:"total_events"`
	TotalTicketsSold   int64             `json:"total_tickets_sold"`
	OrganizerRating    *float64          `json:"organizer_rating,omitempty"`
	RatingCount        int               `json:"rating_count"`
	SocialLinks        map[string]string `json:"social_links,omitempty"`
	UpcomingEvents     []EventInfo       `json:"upcoming_events,omitempty"`
	Stats              OrganizerStats    `json:"stats"`
	CreatedAt          time.Time         `json:"created_at"`
	UpdatedAt          time.Time         `json:"updated_at"`
}

// OrganizerListResponse representa una lista paginada de organizadores
type OrganizerListResponse struct {
	Organizers []OrganizerResponse `json:"organizers"`
	Total      int64               `json:"total"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"page_size"`
	TotalPages int                 `json:"total_pages"`
	HasNext    bool                `json:"has_next"`
	HasPrev    bool                `json:"has_prev"`
	Filters    *OrganizerFilter    `json:"filters,omitempty"`
}

// OrganizerVerificationResponse representa el estado de verificación de un organizador
type OrganizerVerificationResponse struct {
	OrganizerID        string                 `json:"organizer_id"`
	VerificationStatus string                 `json:"verification_status"`
	VerifiedAt         *time.Time             `json:"verified_at,omitempty"`
	VerifiedBy         *string                `json:"verified_by,omitempty"`
	RejectionReason    *string                `json:"rejection_reason,omitempty"`
	Documents          []VerificationDocument `json:"documents,omitempty"`
	NextReviewDate     *time.Time             `json:"next_review_date,omitempty"`
}

// EventInfo representa información básica de un evento
type EventInfo struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	Location    string    `json:"location"`
	CoverImage  *string   `json:"cover_image,omitempty"`
	Status      string    `json:"status"`
	TicketsSold int64     `json:"tickets_sold"`
}
