package enums

type NotificationStatus string

const (
	NotificationStatusPending   NotificationStatus = "pending"
	NotificationStatusScheduled NotificationStatus = "scheduled"
	NotificationStatusSending   NotificationStatus = "sending"
	NotificationStatusSent      NotificationStatus = "sent"
	NotificationStatusDelivered NotificationStatus = "delivered"
	NotificationStatusFailed    NotificationStatus = "failed"
	NotificationStatusRetrying  NotificationStatus = "retrying"
	NotificationStatusCancelled NotificationStatus = "cancelled"
)

func (ns NotificationStatus) IsValid() bool {
	switch ns {
	case NotificationStatusPending, NotificationStatusScheduled, NotificationStatusSending,
		NotificationStatusSent, NotificationStatusDelivered, NotificationStatusFailed,
		NotificationStatusRetrying, NotificationStatusCancelled:
		return true
	}
	return false
}

func (ns NotificationStatus) IsInProgress() bool {
	return ns == NotificationStatusPending || ns == NotificationStatusScheduled ||
		ns == NotificationStatusSending || ns == NotificationStatusRetrying
}

func (ns NotificationStatus) IsFinal() bool {
	return ns == NotificationStatusDelivered || ns == NotificationStatusFailed ||
		ns == NotificationStatusCancelled
}

func (ns NotificationStatus) CanRetry() bool {
	return ns == NotificationStatusFailed
}

func (ns NotificationStatus) CanCancel() bool {
	return ns == NotificationStatusPending || ns == NotificationStatusScheduled
}

func (ns NotificationStatus) String() string {
	return string(ns)
}

// NotificationFlow define el flujo de estados de notificación
var NotificationFlow = map[NotificationStatus][]NotificationStatus{
	NotificationStatusPending:   {NotificationStatusSending, NotificationStatusCancelled},
	NotificationStatusScheduled: {NotificationStatusSending, NotificationStatusCancelled},
	NotificationStatusSending:   {NotificationStatusSent, NotificationStatusFailed},
	NotificationStatusSent:      {NotificationStatusDelivered, NotificationStatusFailed},
	NotificationStatusDelivered: {},
	NotificationStatusFailed:    {NotificationStatusRetrying, NotificationStatusCancelled},
	NotificationStatusRetrying:  {NotificationStatusSending, NotificationStatusFailed, NotificationStatusCancelled},
	NotificationStatusCancelled: {},
}

func CanTransitionNotification(from, to NotificationStatus) bool {
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
