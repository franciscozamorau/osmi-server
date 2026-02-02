package request

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

type WebhookRetryConfig struct {
	MaxAttempts    int     `json:"max_attempts" validate:"min=1,max=10"`
	RetryDelay     int     `json:"retry_delay" validate:"min=1,max=3600"`
	BackoffFactor  float64 `json:"backoff_factor" validate:"min=1.0,max=5.0"`
	TimeoutSeconds int     `json:"timeout_seconds" validate:"min=1,max=60"`
	SuccessCodes   []int   `json:"success_codes" validate:"min=1"`
}

type WebhookFilter struct {
	Provider    *string `json:"provider,omitempty" validate:"omitempty,max=50"`
	EventType   *string `json:"event_type,omitempty" validate:"omitempty,max=100"`
	IsActive    *bool   `json:"is_active,omitempty"`
	SearchQuery *string `json:"search_query,omitempty" validate:"omitempty,max=100"`
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
