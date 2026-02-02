package entities

import "time"

type Payment struct {
	ID      int64 `json:"id" db:"id"`
	OrderID int64 `json:"order_id" db:"order_id"`

	ProviderID            int32   `json:"provider_id" db:"provider_id"`
	ProviderTransactionID *string `json:"provider_transaction_id,omitempty" db:"provider_transaction_id"`
	ProviderSessionID     *string `json:"provider_session_id,omitempty" db:"provider_session_id"`

	Amount       float64 `json:"amount" db:"amount"`
	Currency     string  `json:"currency" db:"currency"`
	ExchangeRate float64 `json:"exchange_rate" db:"exchange_rate"`

	Status               string  `json:"status" db:"status"`
	PaymentMethod        *string `json:"payment_method,omitempty" db:"payment_method"`
	PaymentMethodDetails *string `json:"payment_method_details,omitempty" db:"payment_method_details"`

	Attempts    int32      `json:"attempts" db:"attempts"`
	MaxAttempts int32      `json:"max_attempts" db:"max_attempts"`
	NextRetryAt *time.Time `json:"next_retry_at,omitempty" db:"next_retry_at"`
	LastError   *string    `json:"last_error,omitempty" db:"last_error"`
	ErrorCode   *string    `json:"error_code,omitempty" db:"error_code"`

	IPAddress *string `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent *string `json:"user_agent,omitempty" db:"user_agent"`

	ProcessedAt *time.Time `json:"processed_at,omitempty" db:"processed_at"`
	RefundedAt  *time.Time `json:"refunded_at,omitempty" db:"refunded_at"`
	CancelledAt *time.Time `json:"cancelled_at,omitempty" db:"cancelled_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}
