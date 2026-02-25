package enums

// EventStatus representa el estado de un evento en el sistema
// Valores alineados con el DOMINIO ticketing.event_status de la base de datos
type EventStatus string

const (
	// EventStatusDraft - Evento en borrador, solo visible para organizadores
	EventStatusDraft EventStatus = "draft"
	// EventStatusScheduled - Evento programado, aún no publicado
	EventStatusScheduled EventStatus = "scheduled"
	// EventStatusPublished - Evento publicado y visible al público
	EventStatusPublished EventStatus = "published"
	// EventStatusLive - Evento en vivo (actualmente ocurriendo)
	EventStatusLive EventStatus = "live"
	// EventStatusCancelled - Evento cancelado
	EventStatusCancelled EventStatus = "cancelled"
	// EventStatusCompleted - Evento finalizado
	EventStatusCompleted EventStatus = "completed"
	// EventStatusSoldOut - Evento con todas las entradas agotadas
	EventStatusSoldOut EventStatus = "sold_out"
	// EventStatusArchived - Evento archivado (oculto pero conservado)
	EventStatusArchived EventStatus = "archived"
)

// IsValid verifica si el valor del enum es válido
func (es EventStatus) IsValid() bool {
	switch es {
	case EventStatusDraft, EventStatusScheduled, EventStatusPublished,
		EventStatusLive, EventStatusCancelled, EventStatusCompleted,
		EventStatusSoldOut, EventStatusArchived:
		return true
	}
	return false
}

// CanPublish indica si el evento puede ser publicado
func (es EventStatus) CanPublish() bool {
	return es == EventStatusDraft || es == EventStatusScheduled
}

// CanCancel indica si el evento puede ser cancelado
func (es EventStatus) CanCancel() bool {
	return es == EventStatusScheduled || es == EventStatusPublished ||
		es == EventStatusLive || es == EventStatusSoldOut
}

// CanUpdate indica si el evento puede ser actualizado
func (es EventStatus) CanUpdate() bool {
	return es == EventStatusDraft || es == EventStatusScheduled || es == EventStatusPublished
}

// IsVisible indica si el evento es visible para el público
func (es EventStatus) IsVisible() bool {
	return es == EventStatusPublished || es == EventStatusLive ||
		es == EventStatusSoldOut || es == EventStatusScheduled
}

// IsActive indica si el evento está activo (puede recibir ventas)
func (es EventStatus) IsActive() bool {
	return es == EventStatusPublished || es == EventStatusLive ||
		es == EventStatusScheduled || es == EventStatusSoldOut
}

// IsEnded indica si el evento ha terminado
func (es EventStatus) IsEnded() bool {
	return es == EventStatusCompleted || es == EventStatusCancelled || es == EventStatusArchived
}

// IsSellable indica si se pueden vender tickets para este evento
func (es EventStatus) IsSellable() bool {
	return es == EventStatusPublished || es == EventStatusLive || es == EventStatusScheduled
}

// CanTransitionTo verifica si es posible transicionar a otro estado
func (es EventStatus) CanTransitionTo(to EventStatus) bool {
	if !es.IsValid() || !to.IsValid() {
		return false
	}

	// Definir transiciones válidas
	switch es {
	case EventStatusDraft:
		return to == EventStatusScheduled || to == EventStatusPublished ||
			to == EventStatusCancelled || to == EventStatusArchived
	case EventStatusScheduled:
		return to == EventStatusPublished || to == EventStatusLive ||
			to == EventStatusCancelled || to == EventStatusSoldOut
	case EventStatusPublished:
		return to == EventStatusLive || to == EventStatusSoldOut ||
			to == EventStatusCancelled || to == EventStatusCompleted
	case EventStatusLive:
		return to == EventStatusCompleted || to == EventStatusCancelled ||
			to == EventStatusSoldOut
	case EventStatusSoldOut:
		return to == EventStatusLive || to == EventStatusCompleted ||
			to == EventStatusCancelled
	case EventStatusCompleted:
		return to == EventStatusArchived
	case EventStatusCancelled:
		return to == EventStatusArchived
	case EventStatusArchived:
		return false
	default:
		return false
	}
}

// GetNextStatuses devuelve los posibles siguientes estados
func (es EventStatus) GetNextStatuses() []EventStatus {
	switch es {
	case EventStatusDraft:
		return []EventStatus{EventStatusScheduled, EventStatusPublished, EventStatusCancelled, EventStatusArchived}
	case EventStatusScheduled:
		return []EventStatus{EventStatusPublished, EventStatusLive, EventStatusCancelled, EventStatusSoldOut}
	case EventStatusPublished:
		return []EventStatus{EventStatusLive, EventStatusSoldOut, EventStatusCancelled, EventStatusCompleted}
	case EventStatusLive:
		return []EventStatus{EventStatusCompleted, EventStatusCancelled, EventStatusSoldOut}
	case EventStatusSoldOut:
		return []EventStatus{EventStatusLive, EventStatusCompleted, EventStatusCancelled}
	case EventStatusCompleted:
		return []EventStatus{EventStatusArchived}
	case EventStatusCancelled:
		return []EventStatus{EventStatusArchived}
	case EventStatusArchived:
		return []EventStatus{}
	default:
		return []EventStatus{}
	}
}

// String devuelve la representación string del estado
func (es EventStatus) String() string {
	return string(es)
}

// GetAllStatuses devuelve todos los estados posibles
func GetAllEventStatuses() []EventStatus {
	return []EventStatus{
		EventStatusDraft,
		EventStatusScheduled,
		EventStatusPublished,
		EventStatusLive,
		EventStatusCancelled,
		EventStatusCompleted,
		EventStatusSoldOut,
		EventStatusArchived,
	}
}

// GetActiveStatuses devuelve los estados activos (con ventas posibles)
func GetActiveEventStatuses() []EventStatus {
	return []EventStatus{
		EventStatusPublished,
		EventStatusLive,
		EventStatusScheduled,
		EventStatusSoldOut,
	}
}

// GetPreSaleStatuses devuelve los estados previos a la venta
func GetPreSaleEventStatuses() []EventStatus {
	return []EventStatus{
		EventStatusDraft,
		EventStatusScheduled,
	}
}

// GetPostSaleStatuses devuelve los estados posteriores a la venta
func GetPostSaleEventStatuses() []EventStatus {
	return []EventStatus{
		EventStatusCompleted,
		EventStatusCancelled,
		EventStatusArchived,
	}
}

// MarshalJSON implementa la interfaz json.Marshaler
func (es EventStatus) MarshalJSON() ([]byte, error) {
	return []byte(`"` + string(es) + `"`), nil
}

// UnmarshalJSON implementa la interfaz json.Unmarshaler
func (es *EventStatus) UnmarshalJSON(data []byte) error {
	// Remover comillas
	str := string(data)
	if len(str) >= 2 {
		str = str[1 : len(str)-1]
	}

	status := EventStatus(str)
	if !status.IsValid() {
		return &InvalidEventStatusError{Status: str}
	}

	*es = status
	return nil
}

// InvalidEventStatusError error para valores inválidos
type InvalidEventStatusError struct {
	Status string
}

func (e *InvalidEventStatusError) Error() string {
	return "invalid event status: " + e.Status
}
