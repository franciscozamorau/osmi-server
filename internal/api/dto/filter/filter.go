package filter

import "time"

// BaseFilter contiene filtros comunes para todas las consultas
type BaseFilter struct {
    IDs        []string   `json:"ids,omitempty"`
    DateFrom   *time.Time `json:"date_from,omitempty"`
    DateTo     *time.Time `json:"date_to,omitempty"`
    CreatedFrom *time.Time `json:"created_from,omitempty"`
    CreatedTo  *time.Time `json:"created_to,omitempty"`
    UpdatedFrom *time.Time `json:"updated_from,omitempty"`
    UpdatedTo  *time.Time `json:"updated_to,omitempty"`
    IsActive   *bool      `json:"is_active,omitempty"`
    Search     *string    `json:"search,omitempty"`
    SortBy     *string    `json:"sort_by,omitempty"`
    SortOrder  *string    `json:"sort_order,omitempty"`
}

// UserFilter filtros específicos para usuarios
type UserFilter struct {
    BaseFilter
    Email        *string    `json:"email,omitempty"`
    Username     *string    `json:"username,omitempty"`
    Role         *string    `json:"role,omitempty"`
    IsVerified   *bool      `json:"is_verified,omitempty"`
    HasMFA       *bool      `json:"has_mfa,omitempty"`
    LastLoginFrom *time.Time `json:"last_login_from,omitempty"`
    LastLoginTo  *time.Time `json:"last_login_to,omitempty"`
    Country      *string    `json:"country,omitempty"`
    Language     *string    `json:"language,omitempty"`
}

// CustomerFilter filtros específicos para clientes
type CustomerFilter struct {
    BaseFilter
    Email         *string    `json:"email,omitempty"`
    Phone         *string    `json:"phone,omitempty"`
    UserID        *string    `json:"user_id,omitempty"`
    CompanyName   *string    `json:"company_name,omitempty"`
    CustomerType  *string    `json:"customer_type,omitempty"`
    IsVIP         *bool      `json:"is_vip,omitempty"`
    Country       *string    `json:"country,omitempty"`
    City          *string    `json:"city,omitempty"`
    MinSpent      *float64   `json:"min_spent,omitempty"`
    MaxSpent      *float64   `json:"max_spent,omitempty"`
    MinOrders     *int       `json:"min_orders,omitempty"`
    MaxOrders     *int       `json:"max_orders,omitempty"`
    HasTaxID      *bool      `json:"has_tax_id,omitempty"`
    Segment       *string    `json:"segment,omitempty"`
}

// EventFilter filtros específicos para eventos
type EventFilter struct {
    BaseFilter
    OrganizerID   *string    `json:"organizer_id,omitempty"`
    VenueID       *string    `json:"venue_id,omitempty"`
    CategoryID    *string    `json:"category_id,omitempty"`
    EventType     *string    `json:"event_type,omitempty"`
    Status        *string    `json:"status,omitempty"`
    Visibility    *string    `json:"visibility,omitempty"`
    IsFeatured    *bool      `json:"is_featured,omitempty"`
    IsFree        *bool      `json:"is_free,omitempty"`
    City          *string    `json:"city,omitempty"`
    Country       *string    `json:"country,omitempty"`
    MinPrice      *float64   `json:"min_price,omitempty"`
    MaxPrice      *float64   `json:"max_price,omitempty"`
    HasTickets    *bool      `json:"has_tickets,omitempty"`
    OnlineOnly    *bool      `json:"online_only,omitempty"`
    Tags          []string   `json:"tags,omitempty"`
    AgeRestrictionFrom *int   `json:"age_restriction_from,omitempty"`
    AgeRestrictionTo   *int   `json:"age_restriction_to,omitempty"`
    StartsFrom    *time.Time `json:"starts_from,omitempty"`
    StartsTo      *time.Time `json:"starts_to,omitempty"`
    PublishedFrom *time.Time `json:"published_from,omitempty"`
    PublishedTo   *time.Time `json:"published_to,omitempty"`
}

// TicketFilter filtros específicos para tickets
type TicketFilter struct {
    BaseFilter
    EventID       *string    `json:"event_id,omitempty"`
    CustomerID    *string    `json:"customer_id,omitempty"`
    TicketTypeID  *string    `json:"ticket_type_id,omitempty"`
    Status        *string    `json:"status,omitempty"`
    Code          *string    `json:"code,omitempty"`
    OrderID       *string    `json:"order_id,omitempty"`
    MinPrice      *float64   `json:"min_price,omitempty"`
    MaxPrice      *float64   `json:"max_price,omitempty"`
    Currency      *string    `json:"currency,omitempty"`
    IsCheckedIn   *bool      `json:"is_checked_in,omitempty"`
    IsTransferred *bool      `json:"is_transferred,omitempty"`
    IsRefunded    *bool      `json:"is_refunded,omitempty"`
    SoldFrom      *time.Time `json:"sold_from,omitempty"`
    SoldTo        *time.Time `json:"sold_to,omitempty"`
    CheckedInFrom *time.Time `json:"checked_in_from,omitempty"`
    CheckedInTo   *time.Time `json:"checked_in_to,omitempty"`
}

// OrderFilter filtros específicos para órdenes
type OrderFilter struct {
    BaseFilter
    CustomerID    *string    `json:"customer_id,omitempty"`
    CustomerEmail *string    `json:"customer_email,omitempty"`
    Status        *string    `json:"status,omitempty"`
    OrderType     *string    `json:"order_type,omitempty"`
    Currency      *string    `json:"currency,omitempty"`
    MinAmount     *float64   `json:"min_amount,omitempty"`
    MaxAmount     *float64   `json:"max_amount,omitempty"`
    PaymentMethod *string    `json:"payment_method,omitempty"`
    HasPromotion  *bool      `json:"has_promotion,omitempty"`
    RequiresInvoice *bool    `json:"requires_invoice,omitempty"`
    IsReservation *bool      `json:"is_reservation,omitempty"`
    PaidFrom      *time.Time `json:"paid_from,omitempty"`
    PaidTo        *time.Time `json:"paid_to,omitempty"`
}

// PaymentFilter filtros específicos para pagos
type PaymentFilter struct {
    BaseFilter
    OrderID        *string    `json:"order_id,omitempty"`
    ProviderID     *string    `json:"provider_id,omitempty"`
    Status         *string    `json:"status,omitempty"`
    PaymentMethod  *string    `json:"payment_method,omitempty"`
    Currency       *string    `json:"currency,omitempty"`
    MinAmount      *float64   `json:"min_amount,omitempty"`
    MaxAmount      *float64   `json:"max_amount,omitempty"`
    HasRetries     *bool      `json:"has_retries,omitempty"`
    HasError       *bool      `json:"has_error,omitempty"`
    ProcessedFrom  *time.Time `json:"processed_from,omitempty"`
    ProcessedTo    *time.Time `json:"processed_to,omitempty"`
    ProviderTransactionID *string `json:"provider_transaction_id,omitempty"`
}

// FilterOperators define operadores para filtros avanzados
type FilterOperators struct {
    Equal     interface{} `json:"eq,omitempty"`
    NotEqual  interface{} `json:"ne,omitempty"`
    GreaterThan interface{} `json:"gt,omitempty"`
    LessThan    interface{} `json:"lt,omitempty"`
    GreaterThanOrEqual interface{} `json:"gte,omitempty"`
    LessThanOrEqual    interface{} `json:"lte,omitempty"`
    In         []interface{} `json:"in,omitempty"`
    NotIn      []interface{} `json:"nin,omitempty"`
    Like       *string      `json:"like,omitempty"`
    ILike      *string      `json:"ilike,omitempty"`
    Between    []interface{} `json:"between,