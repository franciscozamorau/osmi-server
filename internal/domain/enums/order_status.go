package enums

type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusConfirmed  OrderStatus = "confirmed"
	OrderStatusPaid       OrderStatus = "paid"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusCompleted  OrderStatus = "completed"
	OrderStatusCancelled  OrderStatus = "cancelled"
	OrderStatusRefunded   OrderStatus = "refunded"
	OrderStatusExpired    OrderStatus = "expired"
	OrderStatusFailed     OrderStatus = "failed"
)

func (os OrderStatus) IsValid() bool {
	switch os {
	case OrderStatusPending, OrderStatusConfirmed, OrderStatusPaid,
		OrderStatusProcessing, OrderStatusCompleted, OrderStatusCancelled,
		OrderStatusRefunded, OrderStatusExpired, OrderStatusFailed:
		return true
	}
	return false
}

func (os OrderStatus) IsActive() bool {
	return os == OrderStatusPending || os == OrderStatusConfirmed ||
		os == OrderStatusPaid || os == OrderStatusProcessing
}

func (os OrderStatus) IsCompleted() bool {
	return os == OrderStatusCompleted
}

func (os OrderStatus) CanCancel() bool {
	return os == OrderStatusPending || os == OrderStatusConfirmed
}

func (os OrderStatus) CanRefund() bool {
	return os == OrderStatusPaid || os == OrderStatusCompleted
}

func (os OrderStatus) CanProcess() bool {
	return os == OrderStatusPaid
}

func (os OrderStatus) String() string {
	return string(os)
}

// OrderFlow define el flujo de estados de orden
var OrderFlow = map[OrderStatus][]OrderStatus{
	OrderStatusPending:    {OrderStatusConfirmed, OrderStatusCancelled, OrderStatusExpired, OrderStatusFailed},
	OrderStatusConfirmed:  {OrderStatusPaid, OrderStatusCancelled, OrderStatusFailed},
	OrderStatusPaid:       {OrderStatusProcessing, OrderStatusRefunded, OrderStatusFailed},
	OrderStatusProcessing: {OrderStatusCompleted, OrderStatusFailed, OrderStatusCancelled},
	OrderStatusCompleted:  {OrderStatusRefunded},
	OrderStatusCancelled:  {},
	OrderStatusRefunded:   {},
	OrderStatusExpired:    {},
	OrderStatusFailed:     {},
}

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
