// internal/api/dto/refund/response.go
package refund

import "time"

// PaymentInfo representa información resumida de un pago
type PaymentInfo struct {
	ID            string     `json:"id"`
	Amount        float64    `json:"amount"`
	Currency      string     `json:"currency"`
	Status        string     `json:"status"`
	PaymentMethod string     `json:"payment_method"`
	ProcessedAt   *time.Time `json:"processed_at,omitempty"`
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

// CustomerInfo representa información básica de un cliente
type CustomerInfo struct {
	ID       string  `json:"id"`
	FullName string  `json:"full_name"`
	Email    string  `json:"email"`
	Phone    *string `json:"phone,omitempty"`
	IsVIP    bool    `json:"is_vip"`
}

// UserInfo representa información básica de un usuario
type UserInfo struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username,omitempty"`
	FullName string `json:"full_name,omitempty"`
}

// RefundResponse representa la respuesta completa de un reembolso
type RefundResponse struct {
	ID                 string                 `json:"id"`
	PaymentID          string                 `json:"payment_id"`
	Payment            *PaymentInfo           `json:"payment,omitempty"`
	OrderID            string                 `json:"order_id"`
	Order              *OrderInfo             `json:"order,omitempty"`
	Customer           *CustomerInfo          `json:"customer,omitempty"`
	RefundReason       string                 `json:"refund_reason"`
	ReasonDetails      *string                `json:"reason_details,omitempty"`
	RefundAmount       float64                `json:"refund_amount"`
	Currency           string                 `json:"currency"`
	Status             string                 `json:"status"`
	ProviderRefundID   *string                `json:"provider_refund_id,omitempty"`
	ProviderResponse   map[string]interface{} `json:"provider_response,omitempty"`
	RequestedBy        *UserInfo              `json:"requested_by,omitempty"`
	ApprovedBy         *UserInfo              `json:"approved_by,omitempty"`
	Processor          *UserInfo              `json:"processor,omitempty"`
	RequestedAt        time.Time              `json:"requested_at"`
	ApprovedAt         *time.Time             `json:"approved_at,omitempty"`
	ProcessedAt        *time.Time             `json:"processed_at,omitempty"`
	CompletedAt        *time.Time             `json:"completed_at,omitempty"`
	CancelledAt        *time.Time             `json:"cancelled_at,omitempty"`
	FailureReason      *string                `json:"failure_reason,omitempty"`
	ProcessorNotes     *string                `json:"processor_notes,omitempty"`
	MerchantComment    *string                `json:"merchant_comment,omitempty"`
	CustomerNotified   bool                   `json:"customer_notified"`
	NotificationSentAt *time.Time             `json:"notification_sent_at,omitempty"`
	PartialRefund      bool                   `json:"partial_refund"`
	RefundToSource     bool                   `json:"refund_to_source"`
	EstimatedArrival   *time.Time             `json:"estimated_arrival,omitempty"`
	ActualArrival      *time.Time             `json:"actual_arrival,omitempty"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
}

type RefundInfo struct {
	ID           string     `json:"id"`
	RefundAmount float64    `json:"refund_amount"`
	Currency     string     `json:"currency"`
	Status       string     `json:"status"`
	RefundReason string     `json:"refund_reason"`
	RequestedAt  time.Time  `json:"requested_at"`
	ProcessedAt  *time.Time `json:"processed_at,omitempty"`
}

type RefundListResponse struct {
	Refunds    []RefundResponse `json:"refunds"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	TotalPages int              `json:"total_pages"`
	HasNext    bool             `json:"has_next"`
	HasPrev    bool             `json:"has_prev"`
	Summary    RefundSummary    `json:"summary"`
}

type RefundSummary struct {
	TotalRefunds       int                 `json:"total_refunds"`
	TotalAmount        float64             `json:"total_amount"`
	PendingAmount      float64             `json:"pending_amount"`
	ProcessedAmount    float64             `json:"processed_amount"`
	FailedAmount       float64             `json:"failed_amount"`
	PendingCount       int                 `json:"pending_count"`
	ProcessedCount     int                 `json:"processed_count"`
	FailedCount        int                 `json:"failed_count"`
	AvgProcessingTime  float64             `json:"avg_processing_time"`
	TopReasons         []RefundReasonStats `json:"top_reasons"`
	SuccessRate        float64             `json:"success_rate"`
	AvgRefundAmount    float64             `json:"avg_refund_amount"`
	PartialRefundCount int                 `json:"partial_refund_count"`
}

type RefundReasonStats struct {
	Reason      string  `json:"reason"`
	Count       int     `json:"count"`
	TotalAmount float64 `json:"total_amount"`
	Percentage  float64 `json:"percentage"`
}

type RefundProcessingResponse struct {
	RefundID            string     `json:"refund_id"`
	Status              string     `json:"status"`
	ProviderRefundID    *string    `json:"provider_refund_id,omitempty"`
	EstimatedCompletion *time.Time `json:"estimated_completion,omitempty"`
	NextSteps           []string   `json:"next_steps,omitempty"`
	RequiresApproval    bool       `json:"requires_approval"`
	ApprovalRequiredBy  *UserInfo  `json:"approval_required_by,omitempty"`
}

type RefundBatchResponse struct {
	BatchID      string              `json:"batch_id"`
	TotalRefunds int                 `json:"total_refunds"`
	TotalAmount  float64             `json:"total_amount"`
	Status       string              `json:"status"`
	Results      []RefundBatchResult `json:"results"`
	StartedAt    time.Time           `json:"started_at"`
	CompletedAt  *time.Time          `json:"completed_at,omitempty"`
	FailedCount  int                 `json:"failed_count"`
	SuccessCount int                 `json:"success_count"`
}

type RefundBatchResult struct {
	RefundID         string  `json:"refund_id"`
	Status           string  `json:"status"`
	Success          bool    `json:"success"`
	ErrorMessage     *string `json:"error_message,omitempty"`
	ProviderRefundID *string `json:"provider_refund_id,omitempty"`
}

// ============================================================================
// TIPOS ADICIONALES PARA REPOSITORIOS
// ============================================================================

// RefundStatsResponse - estadísticas de reembolsos
type RefundStatsResponse struct {
	TotalRefunds     int64   `json:"total_refunds"`
	CompletedRefunds int64   `json:"completed_refunds"`
	PendingRefunds   int64   `json:"pending_refunds"`
	FailedRefunds    int64   `json:"failed_refunds"`
	TotalAmount      float64 `json:"total_amount"`
	AvgRefundAmount  float64 `json:"avg_refund_amount"`
	RefundRate       float64 `json:"refund_rate"`
}

// ProcessingTimeStats - estadísticas de tiempo de procesamiento
type ProcessingTimeStats struct {
	AvgProcessingHours    float64 `json:"avg_processing_hours"`
	MinProcessingHours    float64 `json:"min_processing_hours"`
	MaxProcessingHours    float64 `json:"max_processing_hours"`
	MedianProcessingHours float64 `json:"median_processing_hours"`
}
