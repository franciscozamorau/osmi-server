package enums

type EventStatus string

const (
	EventStatusDraft     EventStatus = "draft"
	EventStatusScheduled EventStatus = "scheduled"
	EventStatusPublished EventStatus = "published"
	EventStatusLive      EventStatus = "live"
	EventStatusCancelled EventStatus = "cancelled"
	EventStatusCompleted EventStatus = "completed"
	EventStatusSoldOut   EventStatus = "sold_out"
	EventStatusArchived  EventStatus = "archived"
)

func (es EventStatus) IsValid() bool {
	switch es {
	case EventStatusDraft, EventStatusScheduled, EventStatusPublished,
		EventStatusLive, EventStatusCancelled, EventStatusCompleted,
		EventStatusSoldOut, EventStatusArchived:
		return true
	}
	return false
}

func (es EventStatus) CanPublish() bool {
	return es == EventStatusDraft || es == EventStatusScheduled
}

func (es EventStatus) IsActive() bool {
	return es == EventStatusPublished || es == EventStatusLive || es == EventStatusScheduled
}

func (es EventStatus) IsEnded() bool {
	return es == EventStatusCompleted || es == EventStatusCancelled || es == EventStatusArchived
}

func (es EventStatus) String() string {
	return string(es)
}
