package response

import "time"

type WebhookResponse struct {
	ID              string                 `json:"id"`
	Provider        string                 `json:"provider"`
	EventType       string                 `json:"event_type"`
	TargetURL       string                 `json:"target_url"`
	SecretToken     *string                `json:"secret_token,omitempty"`
	SignatureHeader *string                `json:"signature_header,omitempty"`
	IsActive        bool                   `json:"is_active"`
	Config          map[string]interface{} `json:"config,omitempty"`
	RetryConfig     *WebhookRetryConfig    `json:"retry_config,omitempty"`
	LastTriggeredAt *time.Time             `json:"last_triggered_at,omitempty"`
	LastResponse    *WebhookLastResponse   `json:"last_response,omitempty"`
	Stats           WebhookStats           `json:"stats"`
	Headers         map[string]string      `json:"headers,omitempty"`
	TimeoutSeconds  int                    `json:"timeout_seconds"`
	SuccessCodes    []int                  `json:"success_codes"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

type WebhookRetryConfig struct {
	MaxAttempts   int     `json:"max_attempts"`
	RetryDelay    int     `json:"retry_delay"`
	BackoffFactor float64 `json:"backoff_factor"`
}

type WebhookLastResponse struct {
	StatusCode int               `json:"status_code"`
	Body       *string           `json:"body,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
	DurationMs int64             `json:"duration_ms"`
	Timestamp  time.Time         `json:"timestamp"`
	Success    bool              `json:"success"`
	Error      *string           `json:"error,omitempty"`
}

type WebhookStats struct {
	TotalTriggers      int64      `json:"total_triggers"`
	SuccessfulTriggers int64      `json:"successful_triggers"`
	FailedTriggers     int64      `json:"failed_triggers"`
	LastTriggeredAt    *time.Time `json:"last_triggered_at,omitempty"`
	AvgResponseTime    float64    `json:"avg_response_time"`
	SuccessRate        float64    `json:"success_rate"`
	TotalRetries       int64      `json:"total_retries"`
	CurrentFailures    int        `json:"current_failures"`
	HealthStatus       string     `json:"health_status"` // healthy, warning, critical
}

type WebhookInfo struct {
	ID              string     `json:"id"`
	Provider        string     `json:"provider"`
	EventType       string     `json:"event_type"`
	TargetURL       string     `json:"target_url"`
	IsActive        bool       `json:"is_active"`
	LastTriggeredAt *time.Time `json:"last_triggered_at,omitempty"`
}

type WebhookListResponse struct {
	Webhooks   []WebhookResponse `json:"webhooks"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int               `json:"total_pages"`
	HasNext    bool              `json:"has_next"`
	HasPrev    bool              `json:"has_prev"`
	Summary    WebhookSummary    `json:"summary"`
	Filters    WebhookFilter     `json:"filters,omitempty"`
}

type WebhookSummary struct {
	TotalWebhooks      int                     `json:"total_webhooks"`
	ActiveWebhooks     int                     `json:"active_webhooks"`
	InactiveWebhooks   int                     `json:"inactive_webhooks"`
	TotalTriggers      int64                   `json:"total_triggers"`
	SuccessfulTriggers int64                   `json:"successful_triggers"`
	FailedTriggers     int64                   `json:"failed_triggers"`
	OverallSuccessRate float64                 `json:"overall_success_rate"`
	TopProviders       []WebhookProviderStats  `json:"top_providers"`
	TopEventTypes      []WebhookEventTypeStats `json:"top_event_types"`
	RecentFailures     int                     `json:"recent_failures"`
}

type WebhookProviderStats struct {
	Provider        string  `json:"provider"`
	Count           int     `json:"count"`
	SuccessRate     float64 `json:"success_rate"`
	AvgResponseTime float64 `json:"avg_response_time"`
}

type WebhookEventTypeStats struct {
	EventType   string  `json:"event_type"`
	Count       int     `json:"count"`
	Frequency   string  `json:"frequency"` // high, medium, low
	SuccessRate float64 `json:"success_rate"`
}

type WebhookTestResponse struct {
	WebhookID        string               `json:"webhook_id"`
	TestStatus       string               `json:"test_status"` // pending, sent, received, failed
	RequestSent      WebhookTestRequest   `json:"request_sent"`
	ResponseReceived *WebhookTestResponse `json:"response_received,omitempty"`
	DurationMs       int64                `json:"duration_ms"`
	Success          bool                 `json:"success"`
	Error            *string              `json:"error,omitempty"`
	Recommendations  []string             `json:"recommendations,omitempty"`
	Timestamp        time.Time            `json:"timestamp"`
}

type WebhookTestRequest struct {
	Method    string                 `json:"method"`
	URL       string                 `json:"url"`
	Headers   map[string]string      `json:"headers"`
	Body      map[string]interface{} `json:"body"`
	Signature *string                `json:"signature,omitempty"`
}

type WebhookTestResponse struct {
	StatusCode int                    `json:"status_code"`
	Headers    map[string]string      `json:"headers"`
	Body       map[string]interface{} `json:"body"`
	DurationMs int64                  `json:"duration_ms"`
}

type WebhookLogResponse struct {
	ID         string                 `json:"id"`
	WebhookID  string                 `json:"webhook_id"`
	EventType  string                 `json:"event_type"`
	Payload    map[string]interface{} `json:"payload"`
	Request    WebhookLogRequest      `json:"request"`
	Response   *WebhookLogResponse    `json:"response,omitempty"`
	Status     string                 `json:"status"`
	Attempt    int                    `json:"attempt"`
	Error      *string                `json:"error,omitempty"`
	DurationMs int64                  `json:"duration_ms"`
	CreatedAt  time.Time              `json:"created_at"`
}

type WebhookLogRequest struct {
	Method  string                 `json:"method"`
	URL     string                 `json:"url"`
	Headers map[string]string      `json:"headers"`
	Body    map[string]interface{} `json:"body"`
}

type WebhookLogResponse struct {
	StatusCode int                    `json:"status_code"`
	Headers    map[string]string      `json:"headers"`
	Body       map[string]interface{} `json:"body"`
}

type WebhookHealthResponse struct {
	WebhookID       string         `json:"webhook_id"`
	HealthStatus    string         `json:"health_status"` // healthy, degraded, offline
	LastCheck       time.Time      `json:"last_check"`
	Uptime          float64        `json:"uptime"`
	ResponseTime    float64        `json:"response_time"`
	FailureRate     float64        `json:"failure_rate"`
	Issues          []WebhookIssue `json:"issues,omitempty"`
	Recommendations []string       `json:"recommendations,omitempty"`
}

type WebhookIssue struct {
	Type        string    `json:"type"`     // timeout, connection_error, auth_error, etc.
	Severity    string    `json:"severity"` // critical, warning, info
	Description string    `json:"description"`
	FirstSeen   time.Time `json:"first_seen"`
	LastSeen    time.Time `json:"last_seen"`
	Occurrences int       `json:"occurrences"`
}
