package entities

import "time"

// NOTA: Estos structs NO mapean directamente a tablas de la base de datos.
// Son estructuras auxiliares para consultas, reportes y estadísticas.

// WebhookStats representa estadísticas agregadas de webhooks
// Usado para dashboards y reportes
type WebhookStats struct {
	TotalWebhooks        int64      `json:"total_webhooks"`
	ActiveWebhooks       int64      `json:"active_webhooks"`
	TotalDeliveries      int64      `json:"total_deliveries"`
	SuccessfulDeliveries int64      `json:"successful_deliveries"`
	FailedDeliveries     int64      `json:"failed_deliveries"`
	SuccessRate          float64    `json:"success_rate"`         // Calculado: successful / total * 100
	AvgResponseTime      float64    `json:"avg_response_time_ms"` // Tiempo promedio de respuesta en ms
	LastDeliveryAt       *time.Time `json:"last_delivery_at,omitempty"`
}

// DeliveryStats representa estadísticas por webhook individual
// Usado para monitoreo y análisis de rendimiento
type DeliveryStats struct {
	WebhookID     int64      `json:"webhook_id"`
	WebhookName   string     `json:"webhook_name"`
	EventType     string     `json:"event_type"`
	TargetURL     string     `json:"target_url"`
	TotalAttempts int64      `json:"total_attempts"`
	SuccessCount  int64      `json:"success_count"`
	FailureCount  int64      `json:"failure_count"`
	SuccessRate   float64    `json:"success_rate"`   // Calculado: success / total * 100
	AvgLatency    float64    `json:"avg_latency_ms"` // Latencia promedio en ms
	LastAttempt   *time.Time `json:"last_attempt,omitempty"`
}

// DeliveryAttempt representa un intento individual de entrega de webhook
// NOTA: Si necesitas persistir estos datos, deberías crear una tabla
// como integration.webhook_delivery_attempts
type DeliveryAttempt struct {
	ID             int64     `json:"id"`
	WebhookID      int64     `json:"webhook_id"`
	EventID        string    `json:"event_id"`        // ID único del evento (ej: UUID)
	Payload        string    `json:"payload"`         // Payload enviado (JSON)
	ResponseStatus int       `json:"response_status"` // Código HTTP de respuesta
	ResponseBody   string    `json:"response_body"`   // Cuerpo de la respuesta
	DurationMs     int       `json:"duration_ms"`     // Duración en milisegundos
	Success        bool      `json:"success"`         // Si fue exitoso (2xx)
	Error          string    `json:"error,omitempty"` // Mensaje de error si falló
	AttemptNumber  int       `json:"attempt_number"`  // Número de intento (1,2,3...)
	CreatedAt      time.Time `json:"created_at"`
}

// Métodos de utilidad para DeliveryAttempt

// IsSuccess verifica si el intento fue exitoso
func (da *DeliveryAttempt) IsSuccess() bool {
	return da.Success && da.ResponseStatus >= 200 && da.ResponseStatus < 300
}

// IsClientError verifica si fue un error del cliente (4xx)
func (da *DeliveryAttempt) IsClientError() bool {
	return da.ResponseStatus >= 400 && da.ResponseStatus < 500
}

// IsServerError verifica si fue un error del servidor (5xx)
func (da *DeliveryAttempt) IsServerError() bool {
	return da.ResponseStatus >= 500 && da.ResponseStatus < 600
}

// ShouldRetry determina si se debe reintentar basado en el código de estado
func (da *DeliveryAttempt) ShouldRetry() bool {
	// Reintentar solo en errores de servidor (5xx) o timeouts
	return da.IsServerError() || da.ResponseStatus == 0 // 0 indica timeout/error de conexión
}

// GetStatusCategory obtiene la categoría del código de estado
func (da *DeliveryAttempt) GetStatusCategory() string {
	switch {
	case da.ResponseStatus >= 200 && da.ResponseStatus < 300:
		return "success"
	case da.ResponseStatus >= 300 && da.ResponseStatus < 400:
		return "redirect"
	case da.ResponseStatus >= 400 && da.ResponseStatus < 500:
		return "client_error"
	case da.ResponseStatus >= 500 && da.ResponseStatus < 600:
		return "server_error"
	case da.ResponseStatus == 0:
		return "network_error"
	default:
		return "unknown"
	}
}

// Métodos de utilidad para WebhookStats

// CalculateSuccessRate calcula la tasa de éxito
func (ws *WebhookStats) CalculateSuccessRate() {
	if ws.TotalDeliveries == 0 {
		ws.SuccessRate = 0
		return
	}
	ws.SuccessRate = float64(ws.SuccessfulDeliveries) / float64(ws.TotalDeliveries) * 100
}

// Métodos de utilidad para DeliveryStats

// CalculateSuccessRate calcula la tasa de éxito
func (ds *DeliveryStats) CalculateSuccessRate() {
	if ds.TotalAttempts == 0 {
		ds.SuccessRate = 0
		return
	}
	ds.SuccessRate = float64(ds.SuccessCount) / float64(ds.TotalAttempts) * 100
}

// UpdateFromAttempt actualiza las estadísticas basado en un intento
func (ds *DeliveryStats) UpdateFromAttempt(attempt *DeliveryAttempt) {
	ds.TotalAttempts++

	if attempt.Success {
		ds.SuccessCount++
	} else {
		ds.FailureCount++
	}

	// Actualizar última fecha
	ds.LastAttempt = &attempt.CreatedAt

	// Recalcular tasa de éxito
	ds.CalculateSuccessRate()

	// Actualizar latencia promedio
	if ds.TotalAttempts == 1 {
		ds.AvgLatency = float64(attempt.DurationMs)
	} else {
		// Promedio móvil
		ds.AvgLatency = (ds.AvgLatency*float64(ds.TotalAttempts-1) + float64(attempt.DurationMs)) / float64(ds.TotalAttempts)
	}
}
