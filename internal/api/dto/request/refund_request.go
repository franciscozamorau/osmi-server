package request

type CreateRefundRequest struct {
	PaymentID        string  `json:"payment_id" validate:"required,uuid4"`
	OrderID          string  `json:"order_id" validate:"required,uuid4"`
	RefundAmount     float64 `json:"refund_amount" validate:"required,min=0.01"`
	RefundReason     string  `json:"refund_reason" validate:"required,max=100"`
	ReasonDetails    *string `json:"reason_details,omitempty" validate:"omitempty,max=1000"`
	PartialRefund    bool    `json:"partial_refund"`
	RefundToSource   bool    `json:"refund_to_source" validate:"required"`
	CustomerNotified bool    `json:"customer_notified"`
	MerchantComment  *string `json:"merchant_comment,omitempty" validate:"omitempty,max=500"`
}

type UpdateRefundRequest struct {
	Status           *string `json:"status,omitempty" validate:"omitempty,oneof=pending processing completed failed cancelled"`
	ProviderRefundID *string `json:"provider_refund_id,omitempty" validate:"omitempty,max=255"`
	ProcessorNotes   *string `json:"processor_notes,omitempty" validate:"omitempty,max=1000"`
	FailureReason    *string `json:"failure_reason,omitempty" validate:"omitempty,max=500"`
	ApprovedBy       *string `json:"approved_by,omitempty" validate:"omitempty,uuid4"`
}

type RefundFilter struct {
	PaymentID     *string  `json:"payment_id,omitempty" validate:"omitempty,uuid4"`
	OrderID       *string  `json:"order_id,omitempty" validate:"omitempty,uuid4"`
	CustomerID    *string  `json:"customer_id,omitempty" validate:"omitempty,uuid4"`
	Status        *string  `json:"status,omitempty" validate:"omitempty,oneof=pending processing completed failed cancelled"`
	RefundReason  *string  `json:"refund_reason,omitempty" validate:"omitempty,max=100"`
	DateFrom      *string  `json:"date_from,omitempty" validate:"omitempty,date"`
	DateTo        *string  `json:"date_to,omitempty" validate:"omitempty,date"`
	MinAmount     *float64 `json:"min_amount,omitempty" validate:"omitempty,min=0"`
	MaxAmount     *float64 `json:"max_amount,omitempty" validate:"omitempty,min=0"`
	PartialOnly   *bool    `json:"partial_only,omitempty"`
	HasProviderID *bool    `json:"has_provider_id,omitempty"`
}

type RefundApprovalRequest struct {
	RefundID    string  `json:"refund_id" validate:"required,uuid4"`
	Approve     bool    `json:"approve"`
	Notes       *string `json:"notes,omitempty" validate:"omitempty,max=500"`
	AutoProcess bool    `json:"auto_process"`
}

type RefundBatchRequest struct {
	RefundIDs   []string `json:"refund_ids" validate:"required,min=1,max=100"`
	BatchReason string   `json:"batch_reason" validate:"required,max=200"`
	Priority    int      `json:"priority" validate:"min=1,max=10"`
}
