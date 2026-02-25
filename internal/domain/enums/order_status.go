package enums

// OrderStatus representa el estado de una orden en el sistema de facturación
// Valores alineados con el DOMINIO billing.payment_status de la base de datos
type OrderStatus string

const (
	// OrderStatusPending - Orden pendiente de pago
	OrderStatusPending OrderStatus = "pending"
	// OrderStatusProcessing - Orden en proceso de pago
	OrderStatusProcessing OrderStatus = "processing"
	// OrderStatusCompleted - Orden completada exitosamente
	OrderStatusCompleted OrderStatus = "completed"
	// OrderStatusFailed - Orden fallida
	OrderStatusFailed OrderStatus = "failed"
	// OrderStatusRefunded - Orden reembolsada
	OrderStatusRefunded OrderStatus = "refunded"
	// OrderStatusDisputed - Orden en disputa
	OrderStatusDisputed OrderStatus = "disputed"
	// OrderStatusChargeback - Orden con chargeback
	OrderStatusChargeback OrderStatus = "chargeback"
	// OrderStatusExpired - Orden expirada
	OrderStatusExpired OrderStatus = "expired"
)

// IsValid verifica si el valor del enum es válido según el dominio de la BD
func (os OrderStatus) IsValid() bool {
	switch os {
	case OrderStatusPending, OrderStatusProcessing, OrderStatusCompleted,
		OrderStatusFailed, OrderStatusRefunded, OrderStatusDisputed,
		OrderStatusChargeback, OrderStatusExpired:
		return true
	}
	return false
}

// IsActive indica si la orden está activa (pendiente o en proceso)
func (os OrderStatus) IsActive() bool {
	return os == OrderStatusPending || os == OrderStatusProcessing
}

// IsCompleted indica si la orden está completada
func (os OrderStatus) IsCompleted() bool {
	return os == OrderStatusCompleted
}

// IsFailed indica si la orden falló
func (os OrderStatus) IsFailed() bool {
	return os == OrderStatusFailed
}

// IsRefunded indica si la orden fue reembolsada
func (os OrderStatus) IsRefunded() bool {
	return os == OrderStatusRefunded
}

// IsDisputed indica si la orden está en disputa
func (os OrderStatus) IsDisputed() bool {
	return os == OrderStatusDisputed
}

// IsChargeback indica si la orden tiene chargeback
func (os OrderStatus) IsChargeback() bool {
	return os == OrderStatusChargeback
}

// IsExpired indica si la orden expiró
func (os OrderStatus) IsExpired() bool {
	return os == OrderStatusExpired
}

// IsFinal indica si la orden está en un estado final
func (os OrderStatus) IsFinal() bool {
	return os == OrderStatusCompleted || os == OrderStatusFailed ||
		os == OrderStatusRefunded || os == OrderStatusExpired ||
		os == OrderStatusDisputed || os == OrderStatusChargeback
}

// IsProblematic indica si la orden tiene un problema (disputa, chargeback, fallo)
func (os OrderStatus) IsProblematic() bool {
	return os == OrderStatusDisputed || os == OrderStatusChargeback || os == OrderStatusFailed
}

// CanCancel indica si la orden puede ser cancelada
func (os OrderStatus) CanCancel() bool {
	return os == OrderStatusPending
}

// CanRefund indica si la orden puede ser reembolsada
func (os OrderStatus) CanRefund() bool {
	return os == OrderStatusCompleted
}

// CanProcess indica si la orden puede ser procesada
func (os OrderStatus) CanProcess() bool {
	return os == OrderStatusPending
}

// CanRetry indica si se puede reintentar la orden
func (os OrderStatus) CanRetry() bool {
	return os == OrderStatusFailed
}

// String devuelve la representación string del estado
func (os OrderStatus) String() string {
	return string(os)
}

// OrderFlow define las transiciones válidas entre estados
var OrderFlow = map[OrderStatus][]OrderStatus{
	OrderStatusPending:    {OrderStatusProcessing, OrderStatusFailed, OrderStatusExpired},
	OrderStatusProcessing: {OrderStatusCompleted, OrderStatusFailed, OrderStatusDisputed, OrderStatusChargeback},
	OrderStatusCompleted:  {OrderStatusRefunded},
	OrderStatusFailed:     {OrderStatusPending}, // Reintento
	OrderStatusRefunded:   {},
	OrderStatusDisputed:   {OrderStatusChargeback, OrderStatusCompleted, OrderStatusRefunded},
	OrderStatusChargeback: {OrderStatusRefunded},
	OrderStatusExpired:    {},
}

// CanTransition verifica si es posible transicionar de un estado a otro
func CanTransitionOrder(from, to OrderStatus) bool {
	if !from.IsValid() || !to.IsValid() {
		return false
	}

	allowed, exists := OrderFlow[from]
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
func GetAllOrderStatuses() []OrderStatus {
	return []OrderStatus{
		OrderStatusPending,
		OrderStatusProcessing,
		OrderStatusCompleted,
		OrderStatusFailed,
		OrderStatusRefunded,
		OrderStatusDisputed,
		OrderStatusChargeback,
		OrderStatusExpired,
	}
}

// GetActiveStatuses devuelve los estados activos
func GetActiveOrderStatuses() []OrderStatus {
	return []OrderStatus{
		OrderStatusPending,
		OrderStatusProcessing,
	}
}

// GetFinalStatuses devuelve los estados finales
func GetFinalOrderStatuses() []OrderStatus {
	return []OrderStatus{
		OrderStatusCompleted,
		OrderStatusFailed,
		OrderStatusRefunded,
		OrderStatusDisputed,
		OrderStatusChargeback,
		OrderStatusExpired,
	}
}

// GetSuccessfulStatuses devuelve los estados exitosos
func GetSuccessfulOrderStatuses() []OrderStatus {
	return []OrderStatus{
		OrderStatusCompleted,
	}
}

// GetFailedStatuses devuelve los estados fallidos
func GetFailedOrderStatuses() []OrderStatus {
	return []OrderStatus{
		OrderStatusFailed,
		OrderStatusDisputed,
		OrderStatusChargeback,
		OrderStatusExpired,
	}
}

// MarshalJSON implementa la interfaz json.Marshaler
func (os OrderStatus) MarshalJSON() ([]byte, error) {
	return []byte(`"` + string(os) + `"`), nil
}

// UnmarshalJSON implementa la interfaz json.Unmarshaler
func (os *OrderStatus) UnmarshalJSON(data []byte) error {
	// Remover comillas
	str := string(data)
	if len(str) >= 2 {
		str = str[1 : len(str)-1]
	}

	status := OrderStatus(str)
	if !status.IsValid() {
		return &InvalidOrderStatusError{Status: str}
	}

	*os = status
	return nil
}

// InvalidOrderStatusError error para valores inválidos
type InvalidOrderStatusError struct {
	Status string
}

func (e *InvalidOrderStatusError) Error() string {
	return "invalid order status: " + e.Status
}
