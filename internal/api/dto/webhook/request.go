// internal/api/dto/webhook/request.go
package webhook

type CreateWebhookRequest struct {
	Provider    string                 `json:"provider" validate:"required,max=50"`
	EventType   string                 `json:"event_type" validate:"required,max=100"`
	TargetURL   string                 `json:"target_url" validate:"required,url,max=500"`
	SecretToken *string                `json:"secret_token,omitempty" validate:"omitempty,min=16,max=255"`
	Config      map[string]interface{} `json:"config,omitempty"`
	IsActive    bool                   `json:"is_active"`
	RetryConfig *WebhookRetryConfig    `json:"retry_config,omitempty"`
}

type UpdateWebhookRequest struct {
	TargetURL   *string                `json:"target_url,omitempty" validate:"omitempty,url,max=500"`
	SecretToken *string                `json:"secret_token,omitempty" validate:"omitempty,min=16,max=255"`
	Config      map[string]interface{} `json:"config,omitempty"`
	IsActive    *bool                  `json:"is_active,omitempty"`
	RetryConfig *WebhookRetryConfig    `json:"retry_config,omitempty"`
}

type WebhookTestRequest struct {
	WebhookID  string                 `json:"webhook_id" validate:"required,uuid4"`
	TestData   map[string]interface{} `json:"test_data,omitempty"`
	TestEvent  string                 `json:"test_event" validate:"required"`
	BypassAuth bool                   `json:"bypass_auth"`
}

type WebhookBatchUpdateRequest struct {
	WebhookIDs []string `json:"webhook_ids" validate:"required,min=1,max=50"`
	IsActive   bool     `json:"is_active"`
}
