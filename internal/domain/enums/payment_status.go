package enums

type PaymentStatus string

const (
	PaymentStatusPending    PaymentStatus = "pending"
	PaymentStatusProcessing PaymentStatus = "processing"
	PaymentStatusCompleted  PaymentStatus = "completed"
	PaymentStatusFailed     PaymentStatus = "failed"
	PaymentStatusRefunded   PaymentStatus = "refunded"
	PaymentStatusDisputed   PaymentStatus = "disputed"
	PaymentStatusChargeback PaymentStatus = "chargeback"
	PaymentStatusExpired    PaymentStatus = "expired"
)

func (ps PaymentStatus) IsValid() bool {
	switch ps {
	case PaymentStatusPending, PaymentStatusProcessing, PaymentStatusCompleted,
		PaymentStatusFailed, PaymentStatusRefunded, PaymentStatusDisputed,
		PaymentStatusChargeback, PaymentStatusExpired:
		return true
	}
	return false
}

func (ps PaymentStatus) IsSuccessful() bool {
	return ps == PaymentStatusCompleted
}

func (ps PaymentStatus) IsPending() bool {
	return ps == PaymentStatusPending || ps == PaymentStatusProcessing
}

func (ps PaymentStatus) IsFailed() bool {
	return ps == PaymentStatusFailed || ps == PaymentStatusExpired
}

func (ps PaymentStatus) CanRefund() bool {
	return ps == PaymentStatusCompleted
}

func (ps PaymentStatus) CanRetry() bool {
	return ps == PaymentStatusFailed
}

func (ps PaymentStatus) String() string {
	return string(ps)
}

// PaymentFlow define el flujo de estados de pago
var PaymentFlow = map[PaymentStatus][]PaymentStatus{
	PaymentStatusPending:    {PaymentStatusProcessing, PaymentStatusFailed, PaymentStatusExpired},
	PaymentStatusProcessing: {PaymentStatusCompleted, PaymentStatusFailed},
	PaymentStatusCompleted:  {PaymentStatusRefunded, PaymentStatusDisputed, PaymentStatusChargeback},
	PaymentStatusFailed:     {PaymentStatusPending}, // Retry
	PaymentStatusRefunded:   {},
	PaymentStatusDisputed:   {PaymentStatusCompleted, PaymentStatusChargeback},
	PaymentStatusChargeback: {},
	PaymentStatusExpired:    {},
}

func CanTransitionPayment(from, to PaymentStatus) bool {
	if !from.IsValid() || !to.IsValid() {
		return false
	}

	allowed, exists := PaymentFlow[from]
	if !exists {
		return false
	}

	for _, status := range allowed {
		if status == to {
			return true
		}
	}
	return false
}
