package response

import "time"

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
	Filters    RefundFilter     `json:"filters,omitempty"`
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
