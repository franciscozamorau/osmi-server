// internal/api/dto/webhook/filter.go
package webhook

type WebhookFilter struct {
	Provider    *string `json:"provider,omitempty" validate:"omitempty,max=50"`
	EventType   *string `json:"event_type,omitempty" validate:"omitempty,max=100"`
	IsActive    *bool   `json:"is_active,omitempty"`
	SearchQuery *string `json:"search_query,omitempty" validate:"omitempty,max=100"`
}
