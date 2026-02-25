package enums

import "time"

// NotificationStatus representa el estado de una notificación en el sistema
// Valores utilizados en la tabla notifications.messages
type NotificationStatus string

const (
	// NotificationStatusPending - Notificación pendiente de procesar
	NotificationStatusPending NotificationStatus = "pending"
	// NotificationStatusScheduled - Notificación programada para futuro
	NotificationStatusScheduled NotificationStatus = "scheduled"
	// NotificationStatusSending - Notificación en proceso de envío
	NotificationStatusSending NotificationStatus = "sending"
	// NotificationStatusSent - Notificación enviada al proveedor
	NotificationStatusSent NotificationStatus = "sent"
	// NotificationStatusDelivered - Notificación entregada al destinatario
	NotificationStatusDelivered NotificationStatus = "delivered"
	// NotificationStatusFailed - Notificación fallida
	NotificationStatusFailed NotificationStatus = "failed"
	// NotificationStatusRetrying - Reintentando envío después de fallo
	NotificationStatusRetrying NotificationStatus = "retrying"
	// NotificationStatusCancelled - Notificación cancelada
	NotificationStatusCancelled NotificationStatus = "cancelled"
)

// IsValid verifica si el valor del enum es válido
func (ns NotificationStatus) IsValid() bool {
	switch ns {
	case NotificationStatusPending, NotificationStatusScheduled, NotificationStatusSending,
		NotificationStatusSent, NotificationStatusDelivered, NotificationStatusFailed,
		NotificationStatusRetrying, NotificationStatusCancelled:
		return true
	}
	return false
}

// IsInProgress indica si la notificación está en proceso
func (ns NotificationStatus) IsInProgress() bool {
	return ns == NotificationStatusPending || ns == NotificationStatusScheduled ||
		ns == NotificationStatusSending || ns == NotificationStatusRetrying
}

// IsFinal indica si la notificación ha alcanzado un estado final
func (ns NotificationStatus) IsFinal() bool {
	return ns == NotificationStatusDelivered || ns == NotificationStatusFailed ||
		ns == NotificationStatusCancelled
}

// IsSuccess indica si la notificación fue exitosa
func (ns NotificationStatus) IsSuccess() bool {
	return ns == NotificationStatusDelivered
}

// IsFailure indica si la notificación falló definitivamente
func (ns NotificationStatus) IsFailure() bool {
	return ns == NotificationStatusFailed
}

// CanRetry indica si se puede reintentar la notificación
func (ns NotificationStatus) CanRetry() bool {
	return ns == NotificationStatusFailed
}

// CanCancel indica si se puede cancelar la notificación
func (ns NotificationStatus) CanCancel() bool {
	return ns == NotificationStatusPending || ns == NotificationStatusScheduled
}

// CanSchedule indica si se puede programar la notificación
func (ns NotificationStatus) CanSchedule() bool {
	return ns == NotificationStatusPending
}

// ShouldUpdateCounters indica si se deben actualizar contadores de métricas
func (ns NotificationStatus) ShouldUpdateCounters() bool {
	return ns == NotificationStatusSent || ns == NotificationStatusDelivered || ns == NotificationStatusFailed
}

// GetNextRetryDelay calcula el delay para el próximo reintento basado en el número de intento
func (ns NotificationStatus) GetNextRetryDelay(attempt int, baseDelay time.Duration) time.Duration {
	if !ns.CanRetry() {
		return 0
	}

	// Backoff exponencial: baseDelay * 2^(attempt-1)
	delay := baseDelay
	for i := 1; i < attempt; i++ {
		delay *= 2
		if delay > 24*time.Hour {
			return 24 * time.Hour
		}
	}
	return delay
}

// String devuelve la representación string del estado
func (ns NotificationStatus) String() string {
	return string(ns)
}

// NotificationFlow define las transiciones válidas entre estados
var NotificationFlow = map[NotificationStatus][]NotificationStatus{
	NotificationStatusPending:   {NotificationStatusSending, NotificationStatusScheduled, NotificationStatusCancelled},
	NotificationStatusScheduled: {NotificationStatusSending, NotificationStatusCancelled},
	NotificationStatusSending:   {NotificationStatusSent, NotificationStatusFailed},
	NotificationStatusSent:      {NotificationStatusDelivered, NotificationStatusFailed},
	NotificationStatusDelivered: {},
	NotificationStatusFailed:    {NotificationStatusRetrying, NotificationStatusCancelled},
	NotificationStatusRetrying:  {NotificationStatusSending, NotificationStatusFailed, NotificationStatusCancelled},
	NotificationStatusCancelled: {},
}

// CanTransition verifica si es posible transicionar de un estado a otro
func CanTransition(from, to NotificationStatus) bool {
	if !from.IsValid() || !to.IsValid() {
		return false
	}

	allowed, exists := NotificationFlow[from]
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
func GetAllNotificationStatuses() []NotificationStatus {
	return []NotificationStatus{
		NotificationStatusPending,
		NotificationStatusScheduled,
		NotificationStatusSending,
		NotificationStatusSent,
		NotificationStatusDelivered,
		NotificationStatusFailed,
		NotificationStatusRetrying,
		NotificationStatusCancelled,
	}
}

// GetActiveStatuses devuelve los estados que requieren seguimiento
func GetActiveNotificationStatuses() []NotificationStatus {
	return []NotificationStatus{
		NotificationStatusPending,
		NotificationStatusScheduled,
		NotificationStatusSending,
		NotificationStatusRetrying,
	}
}

// GetFinalStatuses devuelve los estados finales
func GetFinalNotificationStatuses() []NotificationStatus {
	return []NotificationStatus{
		NotificationStatusDelivered,
		NotificationStatusFailed,
		NotificationStatusCancelled,
	}
}

// GetRetryableStatuses devuelve los estados desde los que se puede reintentar
func GetRetryableNotificationStatuses() []NotificationStatus {
	return []NotificationStatus{
		NotificationStatusFailed,
	}
}

// MarshalJSON implementa la interfaz json.Marshaler
func (ns NotificationStatus) MarshalJSON() ([]byte, error) {
	return []byte(`"` + string(ns) + `"`), nil
}

// UnmarshalJSON implementa la interfaz json.Unmarshaler
func (ns *NotificationStatus) UnmarshalJSON(data []byte) error {
	// Remover comillas
	str := string(data)
	if len(str) >= 2 {
		str = str[1 : len(str)-1]
	}

	status := NotificationStatus(str)
	if !status.IsValid() {
		return &InvalidNotificationStatusError{Status: str}
	}

	*ns = status
	return nil
}

// InvalidNotificationStatusError error para valores inválidos
type InvalidNotificationStatusError struct {
	Status string
}

func (e *InvalidNotificationStatusError) Error() string {
	return "invalid notification status: " + e.Status
}
