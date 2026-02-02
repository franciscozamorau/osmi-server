package enums

type TicketStatus string

const (
	TicketStatusAvailable TicketStatus = "available"
	TicketStatusReserved  TicketStatus = "reserved"
	TicketStatusSold      TicketStatus = "sold"
	TicketStatusCheckedIn TicketStatus = "checked_in"
	TicketStatusCancelled TicketStatus = "cancelled"
	TicketStatusRefunded  TicketStatus = "refunded"
	TicketStatusExpired   TicketStatus = "expired"
)

func (ts TicketStatus) IsValid() bool {
	switch ts {
	case TicketStatusAvailable, TicketStatusReserved, TicketStatusSold,
		TicketStatusCheckedIn, TicketStatusCancelled, TicketStatusRefunded,
		TicketStatusExpired:
		return true
	}
	return false
}

func (ts TicketStatus) CanCheckIn() bool {
	return ts == TicketStatusSold
}

func (ts TicketStatus) CanTransfer() bool {
	return ts == TicketStatusSold
}

func (ts TicketStatus) CanRefund() bool {
	return ts == TicketStatusSold
}

func (ts TicketStatus) CanCancel() bool {
	return ts == TicketStatusAvailable || ts == TicketStatusReserved || ts == TicketStatusSold
}

func (ts TicketStatus) IsActive() bool {
	return ts == TicketStatusReserved || ts == TicketStatusSold || ts == TicketStatusCheckedIn
}

func (ts TicketStatus) String() string {
	return string(ts)
}

// ValidStatusTransitions define las transiciones permitidas
var ValidStatusTransitions = map[TicketStatus][]TicketStatus{
	TicketStatusAvailable: {TicketStatusReserved, TicketStatusSold, TicketStatusCancelled},
	TicketStatusReserved:  {TicketStatusSold, TicketStatusAvailable, TicketStatusCancelled},
	TicketStatusSold:      {TicketStatusCheckedIn, TicketStatusCancelled, TicketStatusRefunded},
	TicketStatusCheckedIn: {},
	TicketStatusCancelled: {},
	TicketStatusRefunded:  {},
	TicketStatusExpired:   {},
}

func CanTransition(from, to TicketStatus) bool {
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
