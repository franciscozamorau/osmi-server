package response

import "time"

type OrderResponse struct {
	ID                   string              `json:"id"`
	Customer             CustomerOrderInfo   `json:"customer"`
	Items                []OrderItemResponse `json:"items"`
	Subtotal             float64             `json:"subtotal"`
	TaxAmount            float64             `json:"tax_amount"`
	ServiceFeeAmount     float64             `json:"service_fee_amount"`
	DiscountAmount       float64             `json:"discount_amount"`
	TotalAmount          float64             `json:"total_amount"`
	Currency             string              `json:"currency"`
	Status               string              `json:"status"`
	OrderType            string              `json:"order_type"`
	IsReservation        bool                `json:"is_reservation"`
	ReservationExpiresAt time.Time           `json:"reservation_expires_at,omitempty"`
	PaymentMethod        string              `json:"payment_method,omitempty"`
	PaymentProvider      string              `json:"payment_provider,omitempty"`
	InvoiceRequired      bool                `json:"invoice_required"`
	InvoiceGenerated     bool                `json:"invoice_generated"`
	InvoiceNumber        string              `json:"invoice_number,omitempty"`
	PromotionCode        string              `json:"promotion_code,omitempty"`
	Notes                string              `json:"notes,omitempty"`
	IPAddress            string              `json:"ip_address,omitempty"`
	ExpiresAt            time.Time           `json:"expires_at,omitempty"`
	PaidAt               time.Time           `json:"paid_at,omitempty"`
	CancelledAt          time.Time           `json:"cancelled_at,omitempty"`
	RefundedAt           time.Time           `json:"refunded_at,omitempty"`
	CreatedAt            time.Time           `json:"created_at"`
	UpdatedAt            time.Time           `json:"updated_at"`
}

type OrderItemResponse struct {
	ID               string              `json:"id"`
	TicketType       TicketTypeBasicInfo `json:"ticket_type"`
	Quantity         int                 `json:"quantity"`
	UnitPrice        float64             `json:"unit_price"`
	TotalPrice       float64             `json:"total_price"`
	Currency         string              `json:"currency"`
	BasePrice        float64             `json:"base_price"`
	TaxAmount        float64             `json:"tax_amount"`
	ServiceFeeAmount float64             `json:"service_fee_amount"`
	DiscountAmount   float64             `json:"discount_amount"`
	TicketIDs        []string            `json:"ticket_ids"`
}

type OrderStatsResponse struct {
	TotalOrders       int              `json:"total_orders"`
	CompletedOrders   int              `json:"completed_orders"`
	PendingOrders     int              `json:"pending_orders"`
	FailedOrders      int              `json:"failed_orders"`
	TotalRevenue      float64          `json:"total_revenue"`
	AvgOrderValue     float64          `json:"avg_order_value"`
	ConversionRate    float64          `json:"conversion_rate"`
	ReservationRate   float64          `json:"reservation_rate"`
	TopPromotionCodes []PromotionStats `json:"top_promotion_codes,omitempty"`
}

type PromotionStats struct {
	Code          string  `json:"code"`
	UsageCount    int     `json:"usage_count"`
	TotalDiscount float64 `json:"total_discount"`
}

type OrderListResponse struct {
	Orders     []OrderResponse    `json:"orders"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalPages int                `json:"total_pages"`
	Stats      OrderStatsResponse `json:"stats"`
}

type CustomerOrderInfo struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
	Phone string `json:"phone,omitempty"`
}

type TicketTypeBasicInfo struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}
