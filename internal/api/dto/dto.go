package dto

import (
	"time"
)

// ============================================================================
// REQUEST DTOs
// ============================================================================

type CreateCustomerRequest struct {
	Name         string `json:"name" validate:"required"`
	Email        string `json:"email" validate:"required,email"`
	Phone        string `json:"phone,omitempty"`
	UserID       string `json:"user_id,omitempty"`
	CustomerType string `json:"customer_type,omitempty"`
	Source       string `json:"source,omitempty"`
}

type UpdateCustomerRequest struct {
	Name         *string `json:"name,omitempty"`
	Email        *string `json:"email,omitempty"`
	Phone        *string `json:"phone,omitempty"`
	CompanyName  *string `json:"company_name,omitempty"`
	Address      *string `json:"address,omitempty"`
	IsVIP        *bool   `json:"is_vip,omitempty"`
	CustomerType *string `json:"customer_type,omitempty"`
}

type CustomerFilter struct {
	IsActive        *bool  `json:"is_active,omitempty"`
	IsVIP           *bool  `json:"is_vip,omitempty"`
	Country         string `json:"country,omitempty"`
	CustomerSegment string `json:"customer_segment,omitempty"`
	Search          string `json:"search,omitempty"`
	DateFrom        string `json:"date_from,omitempty"`
	DateTo          string `json:"date_to,omitempty"`
}

type CreateTicketRequest struct {
	EventID    string `json:"event_id" validate:"required"`
	CustomerID string `json:"customer_id" validate:"required"`
	CategoryID string `json:"category_id" validate:"required"`
	Quantity   int32  `json:"quantity" validate:"required,min=1,max=10"`
	UserID     string `json:"user_id,omitempty"`
}

type UpdateTicketStatusRequest struct {
	TicketID string `json:"ticket_id" validate:"required"`
	Status   string `json:"status" validate:"required,oneof=available reserved sold checked_in cancelled transferred refunded expired"`
	Reason   string `json:"reason,omitempty"`
}

type ReserveTicketRequest struct {
	TicketID  string    `json:"ticket_id" validate:"required"`
	UserID    string    `json:"user_id" validate:"required"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

type CheckInTicketRequest struct {
	TicketID  string `json:"ticket_id" validate:"required"`
	CheckedBy string `json:"checked_by" validate:"required"`
	Method    string `json:"method,omitempty"`
	Location  string `json:"location,omitempty"`
}

type TransferTicketRequest struct {
	TicketID       string `json:"ticket_id" validate:"required"`
	FromCustomerID string `json:"from_customer_id" validate:"required"`
	ToCustomerID   string `json:"to_customer_id" validate:"required"`
	Token          string `json:"token,omitempty"`
}

type UpdateTicketRequest struct {
	AttendeeName  *string `json:"attendee_name,omitempty"`
	AttendeeEmail *string `json:"attendee_email,omitempty"`
	AttendeePhone *string `json:"attendee_phone,omitempty"`
	Status        *string `json:"status,omitempty"`
}

type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Role     string `json:"role" validate:"required,oneof=admin customer organizer guest staff"`
}

type CreateEventRequest struct {
	Name                string    `json:"name" validate:"required"`
	Slug                string    `json:"slug"`
	Description         string    `json:"description"`
	ShortDescription    string    `json:"short_description"`
	OrganizerID         string    `json:"organizer_id" validate:"required"`
	VenueID             string    `json:"venue_id"`
	PrimaryCategoryID   string    `json:"primary_category_id"`
	CategoryIDs         []string  `json:"category_ids"`
	StartsAt            time.Time `json:"starts_at" validate:"required"`
	EndsAt              time.Time `json:"ends_at" validate:"required,gtfield=StartsAt"`
	DoorsOpenAt         time.Time `json:"doors_open_at"`
	DoorsCloseAt        time.Time `json:"doors_close_at"`
	Timezone            string    `json:"timezone" validate:"required"`
	EventType           string    `json:"event_type"`
	CoverImageURL       string    `json:"cover_image_url"`
	BannerImageURL      string    `json:"banner_image_url"`
	VenueName           string    `json:"venue_name"`
	AddressFull         string    `json:"address_full"`
	City                string    `json:"city"`
	State               string    `json:"state"`
	Country             string    `json:"country"`
	Status              string    `json:"status"`
	Visibility          string    `json:"visibility"`
	IsFeatured          bool      `json:"is_featured"`
	IsFree              bool      `json:"is_free"`
	MaxAttendees        int32     `json:"max_attendees"`
	MinAttendees        int32     `json:"min_attendees"`
	Tags                string    `json:"tags"`
	AgeRestriction      int32     `json:"age_restriction"`
	RequiresApproval    bool      `json:"requires_approval"`
	AllowReservations   bool      `json:"allow_reservations"`
	ReservationDuration int32     `json:"reservation_duration_minutes"`
}

type UpdateEventRequest struct {
	Name             *string    `json:"name,omitempty"`
	Slug             *string    `json:"slug,omitempty"`
	Description      *string    `json:"description,omitempty"`
	ShortDescription *string    `json:"short_description,omitempty"`
	VenueID          *string    `json:"venue_id,omitempty"`
	StartsAt         *time.Time `json:"starts_at,omitempty"`
	EndsAt           *time.Time `json:"ends_at,omitempty"`
	Timezone         *string    `json:"timezone,omitempty"`
	Status           *string    `json:"status,omitempty"`
	Visibility       *string    `json:"visibility,omitempty"`
	IsFeatured       *bool      `json:"is_featured,omitempty"`
	IsPublished      *bool      `json:"is_published,omitempty"`
	MaxAttendees     *int32     `json:"max_attendees,omitempty"`
	AgeRestriction   *int32     `json:"age_restriction,omitempty"`
}

type CreateCategoryRequest struct {
	Name            string   `json:"name" validate:"required"`
	Slug            string   `json:"slug"`
	Description     string   `json:"description"`
	Icon            string   `json:"icon"`
	ColorHex        string   `json:"color_hex"`
	ParentID        string   `json:"parent_id,omitempty"`
	IsActive        bool     `json:"is_active"`
	IsFeatured      bool     `json:"is_featured"`
	SortOrder       int      `json:"sort_order"`
	MetaTitle       string   `json:"meta_title"`
	MetaDescription string   `json:"meta_description"`
	Benefits        []string `json:"benefits,omitempty"`
}

type UpdateCategoryRequest struct {
	Name            *string  `json:"name,omitempty"`
	Slug            *string  `json:"slug,omitempty"`
	Description     *string  `json:"description,omitempty"`
	Icon            *string  `json:"icon,omitempty"`
	ColorHex        *string  `json:"color_hex,omitempty"`
	ParentID        *string  `json:"parent_id,omitempty"`
	IsActive        *bool    `json:"is_active,omitempty"`
	IsFeatured      *bool    `json:"is_featured,omitempty"`
	SortOrder       *int     `json:"sort_order,omitempty"`
	MetaTitle       *string  `json:"meta_title,omitempty"`
	MetaDescription *string  `json:"meta_description,omitempty"`
	Benefits        []string `json:"benefits,omitempty"`
}

type CategoryFilter struct {
	IsActive   *bool  `json:"is_active,omitempty"`
	IsFeatured *bool  `json:"is_featured,omitempty"`
	EventID    int64  `json:"event_id,omitempty"`
	ParentID   *int64 `json:"parent_id,omitempty"`
	Level      *int32 `json:"level,omitempty"`
	Search     string `json:"search,omitempty"`
}

type Pagination struct {
	Page     int `json:"page" form:"page" query:"page"`
	PageSize int `json:"page_size" form:"page_size" query:"page_size"`
}

// ============================================================================
// RESPONSE DTOs
// ============================================================================

type CustomerResponse struct {
	ID              string    `json:"id"`
	UserID          *string   `json:"user_id,omitempty"`
	FullName        string    `json:"full_name"`
	Email           string    `json:"email"`
	Phone           *string   `json:"phone,omitempty"`
	CompanyName     *string   `json:"company_name,omitempty"`
	IsVIP           bool      `json:"is_vip"`
	CustomerSegment string    `json:"customer_segment"`
	TotalSpent      float64   `json:"total_spent"`
	TotalOrders     int32     `json:"total_orders"`
	TotalTickets    int32     `json:"total_tickets"`
	CreatedAt       time.Time `json:"created_at"`
}

type CustomerStatsResponse struct {
	TotalCustomers         int64          `json:"total_customers"`
	ActiveCustomers        int64          `json:"active_customers"`
	VIPCustomers           int64          `json:"vip_customers"`
	NewCustomersLast30Days int64          `json:"new_customers_last_30_days"`
	TotalRevenue           float64        `json:"total_revenue"`
	AvgLifetimeValue       float64        `json:"avg_lifetime_value"`
	TopCountries           []CountryStats `json:"top_countries,omitempty"`
}

type CountryStats struct {
	Country string  `json:"country"`
	Count   int64   `json:"count"`
	Revenue float64 `json:"revenue"`
}

type PurchaseRecord struct {
	OrderID      string    `json:"order_id"`
	Amount       float64   `json:"amount"`
	Currency     string    `json:"currency"`
	PurchaseDate time.Time `json:"purchase_date"`
	Status       string    `json:"status"`
	ItemsCount   int64     `json:"items_count"`
}

type TicketResponse struct {
	ID           string    `json:"id"`
	PublicID     string    `json:"public_id"`
	TicketTypeID string    `json:"ticket_type_id"`
	EventID      string    `json:"event_id"`
	Code         string    `json:"code"`
	Status       string    `json:"status"`
	FinalPrice   float64   `json:"final_price"`
	Currency     string    `json:"currency"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type TicketListResponse struct {
	Tickets    []TicketResponse `json:"tickets"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	TotalPages int              `json:"total_pages"`
}

type UserResponse struct {
	ID                string    `json:"id"`
	Email             string    `json:"email"`
	Username          *string   `json:"username,omitempty"`
	FirstName         *string   `json:"first_name,omitempty"`
	LastName          *string   `json:"last_name,omitempty"`
	FullName          *string   `json:"full_name,omitempty"`
	AvatarURL         *string   `json:"avatar_url,omitempty"`
	EmailVerified     bool      `json:"email_verified"`
	PhoneVerified     bool      `json:"phone_verified"`
	PreferredLanguage string    `json:"preferred_language"`
	PreferredCurrency string    `json:"preferred_currency"`
	Timezone          string    `json:"timezone"`
	MFAEnabled        bool      `json:"mfa_enabled"`
	IsActive          bool      `json:"is_active"`
	Role              string    `json:"role"`
	CreatedAt         time.Time `json:"created_at"`
}

type UserStatsResponse struct {
	TotalUsers         int64 `json:"total_users"`
	ActiveUsers        int64 `json:"active_users"`
	InactiveUsers      int64 `json:"inactive_users"`
	StaffUsers         int64 `json:"staff_users"`
	Superusers         int64 `json:"superusers"`
	EmailVerifiedUsers int64 `json:"email_verified_users"`
	PhoneVerifiedUsers int64 `json:"phone_verified_users"`
	MFAEnabledUsers    int64 `json:"mfa_enabled_users"`
	NewUsersLast7Days  int64 `json:"new_users_last_7_days"`
	NewUsersLast30Days int64 `json:"new_users_last_30_days"`
	ActiveLast7Days    int64 `json:"active_last_7_days"`
	ActiveLast30Days   int64 `json:"active_last_30_days"`
}

// ============================================================================
// TIPOS PARA MÓDULOS ESPECIALIZADOS
// ============================================================================

// API Calls
type APICallFilter struct {
	Provider    string `json:"provider,omitempty"`
	Method      string `json:"method,omitempty"`
	Success     *bool  `json:"success,omitempty"`
	DateFrom    string `json:"date_from,omitempty"`
	DateTo      string `json:"date_to,omitempty"`
	MinDuration *int   `json:"min_duration,omitempty"`
	MaxDuration *int   `json:"max_duration,omitempty"`
}

type RetryStats struct {
	TotalCalls      int64   `json:"total_calls"`
	SuccessfulCalls int64   `json:"successful_calls"`
	FailedCalls     int64   `json:"failed_calls"`
	RetriedCalls    int64   `json:"retried_calls"`
	AvgRetries      float64 `json:"avg_retries"`
	MaxRetries      int     `json:"max_retries"`
}

type APICallStatsResponse struct {
	TotalCalls    int64   `json:"total_calls"`
	SuccessRate   float64 `json:"success_rate"`
	AvgResponseMs float64 `json:"avg_response_ms"`
	P95ResponseMs float64 `json:"p95_response_ms"`
	P99ResponseMs float64 `json:"p99_response_ms"`
}

type ProviderAPICallStats struct {
	Provider      string  `json:"provider"`
	CallCount     int64   `json:"call_count"`
	SuccessRate   float64 `json:"success_rate"`
	AvgResponseMs float64 `json:"avg_response_ms"`
}

type EndpointStats struct {
	Endpoint      string  `json:"endpoint"`
	Method        string  `json:"method"`
	CallCount     int64   `json:"call_count"`
	SuccessRate   float64 `json:"success_rate"`
	AvgResponseMs float64 `json:"avg_response_ms"`
}

type ErrorFrequency struct {
	ErrorMessage string `json:"error_message"`
	Count        int64  `json:"count"`
	LastOccurred string `json:"last_occurred"`
}

type UsagePeak struct {
	Hour      int   `json:"hour"`
	CallCount int64 `json:"call_count"`
}

// Auditoría
type AuditFilter struct {
	EventType string `json:"event_type,omitempty"`
	Severity  string `json:"severity,omitempty"`
	UserID    *int64 `json:"user_id,omitempty"`
	DateFrom  string `json:"date_from,omitempty"`
	DateTo    string `json:"date_to,omitempty"`
	Search    string `json:"search,omitempty"`
}

type AuditStatsResponse struct {
	TotalEvents       int64   `json:"total_events"`
	EventsLast24Hours int64   `json:"events_last_24_hours"`
	AvgEventsPerHour  float64 `json:"avg_events_per_hour"`
	HighSeverityCount int64   `json:"high_severity_count"`
}

type ActivityPoint struct {
	Hour  int   `json:"hour"`
	Count int64 `json:"count"`
}

type TableActivity struct {
	TableName string `json:"table_name"`
	Reads     int64  `json:"reads"`
	Writes    int64  `json:"writes"`
	Deletes   int64  `json:"deletes"`
}

type UserActivity struct {
	UserID       int64  `json:"user_id"`
	UserName     string `json:"user_name"`
	EventCount   int64  `json:"event_count"`
	LastActivity string `json:"last_activity"`
}

type SecurityEventDistribution struct {
	EventType  string  `json:"event_type"`
	Count      int64   `json:"count"`
	Percentage float64 `json:"percentage"`
}

type ChangeFrequency struct {
	TableName   string `json:"table_name"`
	ChangeCount int64  `json:"change_count"`
	LastChange  string `json:"last_change"`
}

// Categorías
type CategoryStatsResponse struct {
	TotalCapacity int64   `json:"total_capacity"`
	Sold          int64   `json:"sold"`
	Available     int64   `json:"available"`
	TotalRevenue  float64 `json:"total_revenue"`
	SellRate      float64 `json:"sell_rate"`
}

type CategoryGlobalStats struct {
	TotalCategories       int64   `json:"total_categories"`
	ActiveCategories      int64   `json:"active_categories"`
	TotalTicketsSold      int64   `json:"total_tickets_sold"`
	TotalRevenue          float64 `json:"total_revenue"`
	AvgTicketsPerCategory float64 `json:"avg_tickets_per_category"`
}

type PopularCategory struct {
	CategoryID   int64   `json:"category_id"`
	CategoryName string  `json:"category_name"`
	EventCount   int64   `json:"event_count"`
	TicketSales  int64   `json:"ticket_sales"`
	Revenue      float64 `json:"revenue"`
}

// Eventos
type EventFilter struct {
	Name        string `json:"name,omitempty"`
	OrganizerID *int64 `json:"organizer_id,omitempty"`
	CategoryID  *int64 `json:"category_id,omitempty"`
	Status      string `json:"status,omitempty"`
	DateFrom    string `json:"date_from,omitempty"`
	DateTo      string `json:"date_to,omitempty"`
	City        string `json:"city,omitempty"`
	Country     string `json:"country,omitempty"`
	IsFeatured  *bool  `json:"is_featured,omitempty"`
	IsFree      *bool  `json:"is_free,omitempty"`
	Search      string `json:"search,omitempty"`
}

type EventStatsResponse struct {
	TicketsSold      int64   `json:"tickets_sold"`
	TicketsAvailable int64   `json:"tickets_available"`
	TotalRevenue     float64 `json:"total_revenue"`
	AvgTicketPrice   float64 `json:"avg_ticket_price"`
	CheckInRate      float64 `json:"check_in_rate"`
}

type EventGlobalStats struct {
	TotalEvents        int64   `json:"total_events"`
	ActiveEvents       int64   `json:"active_events"`
	TotalTicketsSold   int64   `json:"total_tickets_sold"`
	TotalRevenue       float64 `json:"total_revenue"`
	AvgTicketsPerEvent float64 `json:"avg_tickets_per_event"`
	UpcomingEvents     int64   `json:"upcoming_events"`
}

type PopularEvent struct {
	EventID     int64   `json:"event_id"`
	EventName   string  `json:"event_name"`
	TicketsSold int64   `json:"tickets_sold"`
	Revenue     float64 `json:"revenue"`
	Rating      float64 `json:"rating"`
}

// Órdenes
type OrderFilter struct {
	CustomerID *int64   `json:"customer_id,omitempty"`
	Status     string   `json:"status,omitempty"`
	DateFrom   string   `json:"date_from,omitempty"`
	DateTo     string   `json:"date_to,omitempty"`
	MinAmount  *float64 `json:"min_amount,omitempty"`
	MaxAmount  *float64 `json:"max_amount,omitempty"`
	Currency   string   `json:"currency,omitempty"`
	Search     string   `json:"search,omitempty"`
}

type OrderStatsResponse struct {
	TotalOrders     int64   `json:"total_orders"`
	CompletedOrders int64   `json:"completed_orders"`
	PendingOrders   int64   `json:"pending_orders"`
	CancelledOrders int64   `json:"cancelled_orders"`
	TotalRevenue    float64 `json:"total_revenue"`
	AvgOrderValue   float64 `json:"avg_order_value"`
}

type RevenueTrend struct {
	Period     string  `json:"period"`
	Revenue    float64 `json:"revenue"`
	OrderCount int64   `json:"order_count"`
}

type TopCustomer struct {
	CustomerID   int64   `json:"customer_id"`
	CustomerName string  `json:"customer_name"`
	OrderCount   int64   `json:"order_count"`
	TotalSpent   float64 `json:"total_spent"`
}

type OrderTotals struct {
	Subtotal       float64 `json:"subtotal"`
	TaxAmount      float64 `json:"tax_amount"`
	ServiceFee     float64 `json:"service_fee"`
	DiscountAmount float64 `json:"discount_amount"`
	TotalAmount    float64 `json:"total_amount"`
}

type CustomerOrderStats struct {
	CustomerID    int64   `json:"customer_id"`
	CustomerName  string  `json:"customer_name"`
	OrderCount    int64   `json:"order_count"`
	TotalSpent    float64 `json:"total_spent"`
	AvgOrderValue float64 `json:"avg_order_value"`
}

type EventOrderStats struct {
	EventID      int64   `json:"event_id"`
	EventName    string  `json:"event_name"`
	OrderCount   int64   `json:"order_count"`
	TicketsSold  int64   `json:"tickets_sold"`
	TotalRevenue float64 `json:"total_revenue"`
}

type DailyRevenue struct {
	Date          string  `json:"date"`
	Revenue       float64 `json:"revenue"`
	OrderCount    int64   `json:"order_count"`
	AvgOrderValue float64 `json:"avg_order_value"`
}

// Organizadores
type OrganizerFilter struct {
	Name               string `json:"name,omitempty"`
	IsVerified         *bool  `json:"is_verified,omitempty"`
	IsActive           *bool  `json:"is_active,omitempty"`
	VerificationStatus string `json:"verification_status,omitempty"`
	Country            string `json:"country,omitempty"`
	Search             string `json:"search,omitempty"`
	DateFrom           string `json:"date_from,omitempty"`
	DateTo             string `json:"date_to,omitempty"`
}

type OrganizerStatsResponse struct {
	TotalEvents      int64   `json:"total_events"`
	TotalTicketsSold int64   `json:"total_tickets_sold"`
	TotalRevenue     float64 `json:"total_revenue"`
	AvgRating        float64 `json:"avg_rating"`
	SellOutRate      float64 `json:"sell_out_rate"`
}

type OrganizerGlobalStats struct {
	TotalOrganizers    int64   `json:"total_organizers"`
	VerifiedOrganizers int64   `json:"verified_organizers"`
	ActiveOrganizers   int64   `json:"active_organizers"`
	AvgEventsPerOrg    float64 `json:"avg_events_per_org"`
	TotalRevenue       float64 `json:"total_revenue"`
}

type TopOrganizer struct {
	OrganizerID   int64   `json:"organizer_id"`
	OrganizerName string  `json:"organizer_name"`
	EventCount    int64   `json:"event_count"`
	TicketsSold   int64   `json:"tickets_sold"`
	Revenue       float64 `json:"revenue"`
	Rating        float64 `json:"rating"`
}

// Pagos
type PaymentFilter struct {
	OrderID    *int64   `json:"order_id,omitempty"`
	ProviderID *int64   `json:"provider_id,omitempty"`
	Status     string   `json:"status,omitempty"`
	DateFrom   string   `json:"date_from,omitempty"`
	DateTo     string   `json:"date_to,omitempty"`
	MinAmount  *float64 `json:"min_amount,omitempty"`
	MaxAmount  *float64 `json:"max_amount,omitempty"`
	Currency   string   `json:"currency,omitempty"`
}

type PaymentStatsResponse struct {
	TotalPayments      int64   `json:"total_payments"`
	SuccessfulPayments int64   `json:"successful_payments"`
	FailedPayments     int64   `json:"failed_payments"`
	TotalVolume        float64 `json:"total_volume"`
	AvgPaymentValue    float64 `json:"avg_payment_value"`
	SuccessRate        float64 `json:"success_rate"`
}

type ProviderStats struct {
	ProviderID        int64   `json:"provider_id"`
	ProviderName      string  `json:"provider_name"`
	TransactionCount  int64   `json:"transaction_count"`
	TotalVolume       float64 `json:"total_volume"`
	SuccessRate       float64 `json:"success_rate"`
	AvgProcessingTime float64 `json:"avg_processing_time_ms"`
}

type DailyVolume struct {
	Date         string  `json:"date"`
	PaymentCount int64   `json:"payment_count"`
	TotalVolume  float64 `json:"total_volume"`
	AvgPayment   float64 `json:"avg_payment"`
}

// Reembolsos
type RefundFilter struct {
	OrderID   *int64   `json:"order_id,omitempty"`
	PaymentID *int64   `json:"payment_id,omitempty"`
	Status    string   `json:"status,omitempty"`
	DateFrom  string   `json:"date_from,omitempty"`
	DateTo    string   `json:"date_to,omitempty"`
	MinAmount *float64 `json:"min_amount,omitempty"`
	MaxAmount *float64 `json:"max_amount,omitempty"`
	Reason    string   `json:"reason,omitempty"`
}

type RefundStatsResponse struct {
	TotalRefunds     int64   `json:"total_refunds"`
	CompletedRefunds int64   `json:"completed_refunds"`
	PendingRefunds   int64   `json:"pending_refunds"`
	FailedRefunds    int64   `json:"failed_refunds"`
	TotalAmount      float64 `json:"total_amount"`
	AvgRefundAmount  float64 `json:"avg_refund_amount"`
	RefundRate       float64 `json:"refund_rate"`
}

type RefundReasonStats struct {
	Reason      string  `json:"reason"`
	Count       int64   `json:"count"`
	Percentage  float64 `json:"percentage"`
	TotalAmount float64 `json:"total_amount"`
}

type ProcessingTimeStats struct {
	AvgProcessingHours    float64 `json:"avg_processing_hours"`
	MinProcessingHours    float64 `json:"min_processing_hours"`
	MaxProcessingHours    float64 `json:"max_processing_hours"`
	MedianProcessingHours float64 `json:"median_processing_hours"`
}

// Tickets
type TicketFilter struct {
	EventID      *int64   `json:"event_id,omitempty"`
	CustomerID   *int64   `json:"customer_id,omitempty"`
	OrderID      *int64   `json:"order_id,omitempty"`
	Status       string   `json:"status,omitempty"`
	TicketTypeID *int64   `json:"ticket_type_id,omitempty"`
	DateFrom     string   `json:"date_from,omitempty"`
	DateTo       string   `json:"date_to,omitempty"`
	MinPrice     *float64 `json:"min_price,omitempty"`
	MaxPrice     *float64 `json:"max_price,omitempty"`
	Code         string   `json:"code,omitempty"`
	Search       string   `json:"search,omitempty"`
}

type TicketStatsResponse struct {
	TotalTickets     int64   `json:"total_tickets"`
	AvailableTickets int64   `json:"available_tickets"`
	SoldTickets      int64   `json:"sold_tickets"`
	ReservedTickets  int64   `json:"reserved_tickets"`
	CheckedInTickets int64   `json:"checked_in_tickets"`
	CancelledTickets int64   `json:"cancelled_tickets"`
	RefundedTickets  int64   `json:"refunded_tickets"`
	TotalRevenue     float64 `json:"total_revenue"`
	AvgTicketPrice   float64 `json:"avg_ticket_price"`
	CheckInRate      float64 `json:"check_in_rate"`
}

// Ticket Types
type TicketTypeFilter struct {
	EventID   *int64   `json:"event_id,omitempty"`
	Name      string   `json:"name,omitempty"`
	IsActive  *bool    `json:"is_active,omitempty"`
	IsSoldOut *bool    `json:"is_sold_out,omitempty"`
	MinPrice  *float64 `json:"min_price,omitempty"`
	MaxPrice  *float64 `json:"max_price,omitempty"`
	Currency  string   `json:"currency,omitempty"`
	Search    string   `json:"search,omitempty"`
}

type TicketTypeStatsResponse struct {
	TotalTypes      int64   `json:"total_types"`
	ActiveTypes     int64   `json:"active_types"`
	SoldOutTypes    int64   `json:"sold_out_types"`
	TotalCapacity   int64   `json:"total_capacity"`
	TicketsSold     int64   `json:"tickets_sold"`
	TicketsReserved int64   `json:"tickets_reserved"`
	TotalRevenue    float64 `json:"total_revenue"`
	AvgPrice        float64 `json:"avg_price"`
}

type EventTicketStats struct {
	EventID           int64   `json:"event_id"`
	EventName         string  `json:"event_name"`
	TicketTypeID      int64   `json:"ticket_type_id"`
	TicketTypeName    string  `json:"ticket_type_name"`
	TotalQuantity     int64   `json:"total_quantity"`
	SoldQuantity      int64   `json:"sold_quantity"`
	ReservedQuantity  int64   `json:"reserved_quantity"`
	AvailableQuantity int64   `json:"available_quantity"`
	Revenue           float64 `json:"revenue"`
	SellThroughRate   float64 `json:"sell_through_rate"`
}

// Usuarios
type UserFilter struct {
	Email       string `json:"email,omitempty"`
	Username    string `json:"username,omitempty"`
	IsActive    *bool  `json:"is_active,omitempty"`
	IsStaff     *bool  `json:"is_staff,omitempty"`
	IsSuperuser *bool  `json:"is_superuser,omitempty"`
	DateFrom    string `json:"date_from,omitempty"`
	DateTo      string `json:"date_to,omitempty"`
	Search      string `json:"search,omitempty"`
	Role        string `json:"role,omitempty"`
}

// Venues
type VenueFilter struct {
	Name        string `json:"name,omitempty"`
	City        string `json:"city,omitempty"`
	State       string `json:"state,omitempty"`
	Country     string `json:"country,omitempty"`
	VenueType   string `json:"venue_type,omitempty"`
	IsActive    *bool  `json:"is_active,omitempty"`
	MinCapacity *int   `json:"min_capacity,omitempty"`
	MaxCapacity *int   `json:"max_capacity,omitempty"`
	Search      string `json:"search,omitempty"`
}

type VenueStatsResponse struct {
	TotalVenues      int64   `json:"total_venues"`
	ActiveVenues     int64   `json:"active_venues"`
	AvgCapacity      float64 `json:"avg_capacity"`
	TotalEvents      int64   `json:"total_events"`
	TotalTicketsSold int64   `json:"total_tickets_sold"`
}

type PopularVenue struct {
	VenueID     int64  `json:"venue_id"`
	VenueName   string `json:"venue_name"`
	City        string `json:"city"`
	Country     string `json:"country"`
	EventCount  int64  `json:"event_count"`
	TicketsSold int64  `json:"tickets_sold"`
}

type VenueEvent struct {
	EventID       int64  `json:"event_id"`
	EventName     string `json:"event_name"`
	EventSlug     string `json:"event_slug"`
	StartDate     string `json:"start_date"`
	EndDate       string `json:"end_date"`
	TicketsSold   int64  `json:"tickets_sold"`
	TotalCapacity int64  `json:"total_capacity"`
}

// Facturación
type InvoiceFilter struct {
	CustomerID *int64   `json:"customer_id,omitempty"`
	Status     string   `json:"status,omitempty"`
	DateFrom   string   `json:"date_from,omitempty"`
	DateTo     string   `json:"date_to,omitempty"`
	MinAmount  *float64 `json:"min_amount,omitempty"`
	MaxAmount  *float64 `json:"max_amount,omitempty"`
	Search     string   `json:"search,omitempty"`
}

type MonthlyInvoiceReport struct {
	Month           string  `json:"month"`
	InvoiceCount    int64   `json:"invoice_count"`
	TotalAmount     float64 `json:"total_amount"`
	PaidInvoices    int64   `json:"paid_invoices"`
	PendingInvoices int64   `json:"pending_invoices"`
}

type InvoiceHistory struct {
	InvoiceID    string  `json:"invoice_id"`
	InvoiceDate  string  `json:"invoice_date"`
	CustomerName string  `json:"customer_name"`
	Amount       float64 `json:"amount"`
	Status       string  `json:"status"`
	PaidDate     *string `json:"paid_date,omitempty"`
}

type TaxSummary struct {
	TaxType       string  `json:"tax_type"`
	TaxRate       float64 `json:"tax_rate"`
	TaxAmount     float64 `json:"tax_amount"`
	TaxableAmount float64 `json:"taxable_amount"`
}

type InvoiceStatsResponse struct {
	TotalInvoices    int64   `json:"total_invoices"`
	PaidInvoices     int64   `json:"paid_invoices"`
	PendingInvoices  int64   `json:"pending_invoices"`
	OverdueInvoices  int64   `json:"overdue_invoices"`
	TotalRevenue     float64 `json:"total_revenue"`
	AvgInvoiceAmount float64 `json:"avg_invoice_amount"`
}

type RevenueByPeriod struct {
	Period       string  `json:"period"`
	Revenue      float64 `json:"revenue"`
	InvoiceCount int64   `json:"invoice_count"`
}

type PaymentTermsStats struct {
	PaymentTerm  string  `json:"payment_term"`
	InvoiceCount int64   `json:"invoice_count"`
	AvgDaysToPay float64 `json:"avg_days_to_pay"`
}

// Notificaciones
type NotificationFilter struct {
	Channel    string `json:"channel,omitempty"`
	Status     string `json:"status,omitempty"`
	Recipient  string `json:"recipient,omitempty"`
	DateFrom   string `json:"date_from,omitempty"`
	DateTo     string `json:"date_to,omitempty"`
	TemplateID *int64 `json:"template_id,omitempty"`
}

type NotificationStatsResponse struct {
	TotalNotifications  int64   `json:"total_notifications"`
	SentNotifications   int64   `json:"sent_notifications"`
	FailedNotifications int64   `json:"failed_notifications"`
	DeliveryRate        float64 `json:"delivery_rate"`
	OpenRate            float64 `json:"open_rate,omitempty"`
	ClickRate           float64 `json:"click_rate,omitempty"`
}

type FailureReasonStats struct {
	Reason       string `json:"reason"`
	Count        int64  `json:"count"`
	LastOccurred string `json:"last_occurred"`
}

// ============================================================================
// TIPOS COMUNES
// ============================================================================

type HealthCheck struct {
	Status    string    `json:"status"`
	Service   string    `json:"service"`
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
}
