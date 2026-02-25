package enums

// TicketStatus representa el estado de un ticket en el sistema
// Valores alineados con el DOMINIO ticketing.ticket_status de la base de datos
type TicketStatus string

const (
	// TicketStatusAvailable - Ticket disponible para la venta
	TicketStatusAvailable TicketStatus = "available"
	// TicketStatusReserved - Ticket reservado temporalmente
	TicketStatusReserved TicketStatus = "reserved"
	// TicketStatusSold - Ticket vendido
	TicketStatusSold TicketStatus = "sold"
	// TicketStatusCheckedIn - Ticket usado (check-in realizado)
	TicketStatusCheckedIn TicketStatus = "checked_in"
	// TicketStatusCancelled - Ticket cancelado
	TicketStatusCancelled TicketStatus = "cancelled"
	// TicketStatusRefunded - Ticket reembolsado
	TicketStatusRefunded TicketStatus = "refunded"
	// TicketStatusExpired - Ticket expirado
	TicketStatusExpired TicketStatus = "expired"
)

// IsValid verifica si el valor del enum es válido
func (ts TicketStatus) IsValid() bool {
	switch ts {
	case TicketStatusAvailable, TicketStatusReserved, TicketStatusSold,
		TicketStatusCheckedIn, TicketStatusCancelled, TicketStatusRefunded,
		TicketStatusExpired:
		return true
	}
	return false
}

// CanCheckIn verifica si el ticket puede ser marcado como usado
func (ts TicketStatus) CanCheckIn() bool {
	return ts == TicketStatusSold
}

// CanTransfer verifica si el ticket puede ser transferido
func (ts TicketStatus) CanTransfer() bool {
	return ts == TicketStatusSold
}

// CanRefund verifica si el ticket puede ser reembolsado
func (ts TicketStatus) CanRefund() bool {
	return ts == TicketStatusSold
}

// CanCancel verifica si el ticket puede ser cancelado
func (ts TicketStatus) CanCancel() bool {
	return ts == TicketStatusAvailable || ts == TicketStatusReserved || ts == TicketStatusSold
}

// IsActive verifica si el ticket está en un estado activo
func (ts TicketStatus) IsActive() bool {
	return ts == TicketStatusReserved || ts == TicketStatusSold || ts == TicketStatusCheckedIn
}

// String devuelve la representación string del estado
func (ts TicketStatus) String() string {
	return string(ts)
}

// ValidStatusTransitions define las transiciones permitidas entre estados
// Basado en la lógica de negocio del sistema
var ValidStatusTransitions = map[TicketStatus][]TicketStatus{
	TicketStatusAvailable: {TicketStatusReserved, TicketStatusSold, TicketStatusCancelled, TicketStatusExpired},
	TicketStatusReserved:  {TicketStatusSold, TicketStatusAvailable, TicketStatusCancelled, TicketStatusExpired},
	TicketStatusSold:      {TicketStatusCheckedIn, TicketStatusCancelled, TicketStatusRefunded},
	TicketStatusCheckedIn: {},
	TicketStatusCancelled: {},
	TicketStatusRefunded:  {},
	TicketStatusExpired:   {},
}

// CanTransition verifica si es posible transicionar de un estado a otro
func CanTransitionTicket(from, to TicketStatus) bool {
	if !from.IsValid() || !to.IsValid() {
		return false
	}

	allowed, exists := ValidStatusTransitions[from]
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
func GetAllStatuses() []TicketStatus {
	return []TicketStatus{
		TicketStatusAvailable,
		TicketStatusReserved,
		TicketStatusSold,
		TicketStatusCheckedIn,
		TicketStatusCancelled,
		TicketStatusRefunded,
		TicketStatusExpired,
	}
}

// GetActiveStatuses devuelve los estados que se consideran activos
func GetActiveStatuses() []TicketStatus {
	return []TicketStatus{
		TicketStatusReserved,
		TicketStatusSold,
		TicketStatusCheckedIn,
	}
}

// GetFinalStatuses devuelve los estados finales (no se puede transicionar desde ellos)
func GetFinalStatuses() []TicketStatus {
	return []TicketStatus{
		TicketStatusCheckedIn,
		TicketStatusCancelled,
		TicketStatusRefunded,
		TicketStatusExpired,
	}
}
