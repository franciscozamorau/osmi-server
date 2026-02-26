// internal/api/dto/response/organizer_response.go
package response

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
	TotalRevenue       float64 `json:"total_revenue"`
	AvgRevenuePerEvent float64 `json:"avg_revenue_per_event"`
	AvgTicketsPerEvent float64 `json:"avg_tickets_per_event"`
	SellOutRate        float64 `json:"sell_out_rate"`
	RepeatCustomerRate float64 `json:"repeat_customer_rate"`
	UpcomingEventCount int     `json:"upcoming_event_count"`
	PastEventCount     int     `json:"past_event_count"`
	CancellationRate   float64 `json:"cancellation_rate"`
}

// OrganizerFilter representa los filtros aplicados en la consulta
type OrganizerFilter struct {
	Search             *string `json:"search,omitempty"`
	IsVerified         *bool   `json:"is_verified,omitempty"`
	IsActive           *bool   `json:"is_active,omitempty"`
	VerificationStatus *string `json:"verification_status,omitempty"`
	Country            *string `json:"country,omitempty"`
	MinEvents          *int    `json:"min_events,omitempty"`
	MaxEvents          *int    `json:"max_events,omitempty"`
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
