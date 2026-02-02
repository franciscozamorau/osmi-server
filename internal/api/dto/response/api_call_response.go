package dto

import "time"

type APICallResponse struct {
	ID              int64             `json:"id"`
	Provider        string            `json:"provider"`
	Endpoint        string            `json:"endpoint"`
	Method          string            `json:"method"`
	RequestHeaders  map[string]string `json:"request_headers,omitempty"`
	ResponseStatus  int32             `json:"response_status"`
	ResponseHeaders map[string]string `json:"response_headers,omitempty"`
	ResponseTimeMs  int32             `json:"response_time_ms"`
	RetryCount      int32             `json:"retry_count"`
	Success         bool              `json:"success"`
	ErrorMessage    string            `json:"error_message,omitempty"`
	UserID          string            `json:"user_id,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
}

type APICallStatsResponse struct {
	TotalCalls      int64           `json:"total_calls"`
	SuccessCalls    int64           `json:"success_calls"`
	FailedCalls     int64           `json:"failed_calls"`
	SuccessRate     float64         `json:"success_rate"`
	AvgResponseTime float64         `json:"avg_response_time"`
	MaxResponseTime int32           `json:"max_response_time"`
	MinResponseTime int32           `json:"min_response_time"`
	TopEndpoints    []EndpointStats `json:"top_endpoints"`
}

type EndpointStats struct {
	Endpoint        string  `json:"endpoint"`
	CallCount       int64   `json:"call_count"`
	SuccessRate     float64 `json:"success_rate"`
	AvgResponseTime float64 `json:"avg_response_time"`
}
