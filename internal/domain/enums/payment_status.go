package enums

// PaymentStatus representa el estado de un pago en el sistema de facturación
// Valores alineados con el DOMINIO billing.payment_status de la base de datos
type PaymentStatus string

const (
	// PaymentStatusPending - Pago pendiente de procesar
	PaymentStatusPending PaymentStatus = "pending"
	// PaymentStatusProcessing - Pago en proceso
	PaymentStatusProcessing PaymentStatus = "processing"
	// PaymentStatusCompleted - Pago completado exitosamente
	PaymentStatusCompleted PaymentStatus = "completed"
	// PaymentStatusFailed - Pago fallido
	PaymentStatusFailed PaymentStatus = "failed"
	// PaymentStatusRefunded - Pago reembolsado
	PaymentStatusRefunded PaymentStatus = "refunded"
	// PaymentStatusDisputed - Pago en disputa
	PaymentStatusDisputed PaymentStatus = "disputed"
	// PaymentStatusChargeback - Pago con chargeback
	PaymentStatusChargeback PaymentStatus = "chargeback"
	// PaymentStatusExpired - Pago expirado
	PaymentStatusExpired PaymentStatus = "expired"
)

// IsValid verifica si el valor del enum es válido según el dominio de la BD
func (ps PaymentStatus) IsValid() bool {
	switch ps {
	case PaymentStatusPending, PaymentStatusProcessing, PaymentStatusCompleted,
		PaymentStatusFailed, PaymentStatusRefunded, PaymentStatusDisputed,
		PaymentStatusChargeback, PaymentStatusExpired:
		return true
	}
	return false
}

// IsSuccessful indica si el pago fue exitoso
func (ps PaymentStatus) IsSuccessful() bool {
	return ps == PaymentStatusCompleted
}

// IsPending indica si el pago está pendiente
func (ps PaymentStatus) IsPending() bool {
	return ps == PaymentStatusPending || ps == PaymentStatusProcessing
}

// IsFailed indica si el pago falló
func (ps PaymentStatus) IsFailed() bool {
	return ps == PaymentStatusFailed || ps == PaymentStatusExpired
}

// IsRefunded indica si el pago fue reembolsado
func (ps PaymentStatus) IsRefunded() bool {
	return ps == PaymentStatusRefunded
}

// IsDisputed indica si el pago está en disputa
func (ps PaymentStatus) IsDisputed() bool {
	return ps == PaymentStatusDisputed
}

// IsChargeback indica si el pago tiene chargeback
func (ps PaymentStatus) IsChargeback() bool {
	return ps == PaymentStatusChargeback
}

// IsExpired indica si el pago expiró
func (ps PaymentStatus) IsExpired() bool {
	return ps == PaymentStatusExpired
}

// IsFinal indica si el pago está en un estado final
func (ps PaymentStatus) IsFinal() bool {
	return ps == PaymentStatusCompleted || ps == PaymentStatusFailed ||
		ps == PaymentStatusRefunded || ps == PaymentStatusExpired ||
		ps == PaymentStatusDisputed || ps == PaymentStatusChargeback
}

// IsProblematic indica si el pago tiene un problema
func (ps PaymentStatus) IsProblematic() bool {
	return ps == PaymentStatusFailed || ps == PaymentStatusDisputed ||
		ps == PaymentStatusChargeback || ps == PaymentStatusExpired
}

// CanRefund indica si el pago puede ser reembolsado
func (ps PaymentStatus) CanRefund() bool {
	return ps == PaymentStatusCompleted
}

// CanRetry indica si se puede reintentar el pago
func (ps PaymentStatus) CanRetry() bool {
	return ps == PaymentStatusFailed
}

// CanDispute indica si se puede disputar el pago
func (ps PaymentStatus) CanDispute() bool {
	return ps == PaymentStatusCompleted
}

// String devuelve la representación string del estado
func (ps PaymentStatus) String() string {
	return string(ps)
}

// PaymentFlow define las transiciones válidas entre estados
var PaymentFlow = map[PaymentStatus][]PaymentStatus{
	PaymentStatusPending:    {PaymentStatusProcessing, PaymentStatusFailed, PaymentStatusExpired},
	PaymentStatusProcessing: {PaymentStatusCompleted, PaymentStatusFailed, PaymentStatusDisputed, PaymentStatusChargeback},
	PaymentStatusCompleted:  {PaymentStatusRefunded, PaymentStatusDisputed, PaymentStatusChargeback},
	PaymentStatusFailed:     {PaymentStatusPending}, // Reintento
	PaymentStatusRefunded:   {},
	PaymentStatusDisputed:   {PaymentStatusCompleted, PaymentStatusChargeback, PaymentStatusRefunded},
	PaymentStatusChargeback: {PaymentStatusRefunded},
	PaymentStatusExpired:    {},
}

// CanTransition verifica si es posible transicionar de un estado a otro
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

// GetAllStatuses devuelve todos los estados posibles
func GetAllPaymentStatuses() []PaymentStatus {
	return []PaymentStatus{
		PaymentStatusPending,
		PaymentStatusProcessing,
		PaymentStatusCompleted,
		PaymentStatusFailed,
		PaymentStatusRefunded,
		PaymentStatusDisputed,
		PaymentStatusChargeback,
		PaymentStatusExpired,
	}
}

// GetActiveStatuses devuelve los estados activos (no finales)
func GetActivePaymentStatuses() []PaymentStatus {
	return []PaymentStatus{
		PaymentStatusPending,
		PaymentStatusProcessing,
	}
}

// GetFinalStatuses devuelve los estados finales
func GetFinalPaymentStatuses() []PaymentStatus {
	return []PaymentStatus{
		PaymentStatusCompleted,
		PaymentStatusFailed,
		PaymentStatusRefunded,
		PaymentStatusDisputed,
		PaymentStatusChargeback,
		PaymentStatusExpired,
	}
}

// GetSuccessfulStatuses devuelve los estados exitosos
func GetSuccessfulPaymentStatuses() []PaymentStatus {
	return []PaymentStatus{
		PaymentStatusCompleted,
	}
}

// GetFailedStatuses devuelve los estados fallidos
func GetFailedPaymentStatuses() []PaymentStatus {
	return []PaymentStatus{
		PaymentStatusFailed,
		PaymentStatusDisputed,
		PaymentStatusChargeback,
		PaymentStatusExpired,
	}
}

// GetRequiresActionStatuses devuelve los estados que requieren acción
func GetRequiresActionPaymentStatuses() []PaymentStatus {
	return []PaymentStatus{
		PaymentStatusDisputed,
		PaymentStatusChargeback,
		PaymentStatusFailed,
	}
}

// MarshalJSON implementa la interfaz json.Marshaler
func (ps PaymentStatus) MarshalJSON() ([]byte, error) {
	return []byte(`"` + string(ps) + `"`), nil
}

// UnmarshalJSON implementa la interfaz json.Unmarshaler
func (ps *PaymentStatus) UnmarshalJSON(data []byte) error {
	// Remover comillas
	str := string(data)
	if len(str) >= 2 {
		str = str[1 : len(str)-1]
	}

	status := PaymentStatus(str)
	if !status.IsValid() {
		return &InvalidPaymentStatusError{Status: str}
	}

	*ps = status
	return nil
}

// InvalidPaymentStatusError error para valores inválidos
type InvalidPaymentStatusError struct {
	Status string
}

func (e *InvalidPaymentStatusError) Error() string {
	return "invalid payment status: " + e.Status
}
