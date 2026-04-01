// internal/api/dto/payment/request.go
package payment

type CreatePaymentRequest struct {
	OrderID              string                 `json:"order_id" validate:"required,uuid4"`
	Amount               float64                `json:"amount" validate:"required,min=0.01"`
	Currency             string                 `json:"currency" validate:"required,oneof=MXN USD EUR"`
	PaymentMethod        string                 `json:"payment_method" validate:"required"`
	PaymentProvider      string                 `json:"payment_provider" validate:"required"`
	PaymentMethodDetails map[string]interface{} `json:"payment_method_details,omitempty"`
	SaveCard             bool                   `json:"save_card,omitempty"`
}

type RetryPaymentRequest struct {
	PaymentID string `json:"payment_id" validate:"required,uuid4"`
}

type RefundPaymentRequest struct {
	PaymentID    string  `json:"payment_id" validate:"required,uuid4"`
	RefundAmount float64 `json:"refund_amount" validate:"required,min=0.01"`
	RefundReason string  `json:"refund_reason,omitempty" validate:"omitempty,max=100"`
	FullRefund   bool    `json:"full_refund,omitempty"`
}
