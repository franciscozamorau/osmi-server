// internal/api/dto/response/common_types.go
package response

import "time"

// ============================================================================
// TIPOS DE INFORMACIÓN BÁSICA (usados en múltiples responses)
// ============================================================================

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

// CategoryInfo representa información básica de una categoría
type CategoryInfo struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Icon        *string `json:"icon,omitempty"`
	ColorHex    string  `json:"color_hex"`
	TotalEvents int     `json:"total_events"`
	IsActive    bool    `json:"is_active"`
	IsFeatured  bool    `json:"is_featured,omitempty"`
}

// OrganizerInfo representa información básica de un organizador
type OrganizerInfo struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Slug            string   `json:"slug"`
	LogoURL         *string  `json:"logo_url,omitempty"`
	IsVerified      bool     `json:"is_verified"`
	TotalEvents     int      `json:"total_events"`
	OrganizerRating *float64 `json:"organizer_rating,omitempty"`
	RatingCount     int      `json:"rating_count"`
}

// VenueInfo representa información básica de un venue
type VenueInfo struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Slug       string   `json:"slug"`
	City       string   `json:"city"`
	Country    string   `json:"country"`
	VenueType  string   `json:"venue_type"`
	Capacity   *int     `json:"capacity,omitempty"`
	IsVerified bool     `json:"is_verified"`
	Rating     *float64 `json:"rating,omitempty"`
}

// UserInfo representa información básica de un usuario
type UserInfo struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	Username  *string `json:"username,omitempty"`
	FullName  *string `json:"full_name,omitempty"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}

// CustomerInfo representa información básica de un cliente
type CustomerInfo struct {
	ID       string  `json:"id"`
	FullName string  `json:"full_name"`
	Email    string  `json:"email"`
	Phone    *string `json:"phone,omitempty"`
	IsVIP    bool    `json:"is_vip"`
}

// ============================================================================
// TIPOS DE PAGINACIÓN (comunes para listas)
// ============================================================================

// PaginationInfo representa información de paginación
type PaginationInfo struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// ============================================================================
// TIPOS DE UBICACIÓN (comunes)
// ============================================================================

// GeoLocation representa coordenadas geográficas
type GeoLocation struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// Address representa una dirección física
type Address struct {
	AddressLine1 string  `json:"address_line1"`
	AddressLine2 *string `json:"address_line2,omitempty"`
	City         string  `json:"city"`
	State        *string `json:"state,omitempty"`
	PostalCode   *string `json:"postal_code,omitempty"`
	Country      string  `json:"country"`
	FullAddress  string  `json:"full_address,omitempty"`
}

// MapBounds representa límites geográficos para mapas
type MapBounds struct {
	NorthEast GeoLocation `json:"north_east"`
	SouthWest GeoLocation `json:"south_west"`
}

// OrderInfo representa información básica de una orden
type OrderInfo struct {
	ID          string    `json:"id"`
	OrderNumber string    `json:"order_number"`
	TotalAmount float64   `json:"total_amount"`
	Currency    string    `json:"currency"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

// PaymentFilter representa filtros para búsqueda de pagos
type PaymentFilter struct {
	OrderID    *string  `json:"order_id,omitempty"`
	ProviderID *string  `json:"provider_id,omitempty"`
	Status     *string  `json:"status,omitempty"`
	DateFrom   *string  `json:"date_from,omitempty"`
	DateTo     *string  `json:"date_to,omitempty"`
	MinAmount  *float64 `json:"min_amount,omitempty"`
	MaxAmount  *float64 `json:"max_amount,omitempty"`
	Currency   *string  `json:"currency,omitempty"`
	Search     *string  `json:"search,omitempty"`
}

// WebhookFilter representa filtros para búsqueda de webhooks
type WebhookFilter struct {
	Provider     *string `json:"provider,omitempty"`
	EventType    *string `json:"event_type,omitempty"`
	IsActive     *bool   `json:"is_active,omitempty"`
	Search       *string `json:"search,omitempty"`
	DateFrom     *string `json:"date_from,omitempty"`
	DateTo       *string `json:"date_to,omitempty"`
	HealthStatus *string `json:"health_status,omitempty"`
}
