// internal/api/dto/notification/filter.go
package notification

// NotificationFilter representa filtros para listar notificaciones
type NotificationFilter struct {
	Channel    string `json:"channel,omitempty"`
	Status     string `json:"status,omitempty"`
	Recipient  string `json:"recipient,omitempty"`
	DateFrom   string `json:"date_from,omitempty"`
	DateTo     string `json:"date_to,omitempty"`
	TemplateID *int64 `json:"template_id,omitempty"`
}
