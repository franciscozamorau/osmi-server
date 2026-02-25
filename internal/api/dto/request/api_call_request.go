package request

type APICallRequest struct {
	Provider       string                 `json:"provider" validate:"required"`
	Endpoint       string                 `json:"endpoint" validate:"required"`
	Method         string                 `json:"method" validate:"required,oneof=GET POST PUT DELETE PATCH"`
	RequestBody    map[string]interface{} `json:"request_body,omitempty"`
	RequestHeaders map[string]string      `json:"request_headers,omitempty"`
	RetryCount     int                    `json:"retry_count,omitempty" validate:"omitempty,min=0"`
	UserID         string                 `json:"user_id,omitempty" validate:"omitempty,uuid4"`
}

type APICallFilter struct {
	Provider        string `json:"provider,omitempty"`
	Endpoint        string `json:"endpoint,omitempty"`
	Method          string `json:"method,omitempty"`
	Success         *bool  `json:"success,omitempty"`
	DateFrom        string `json:"date_from,omitempty" validate:"omitempty,date"`
	DateTo          string `json:"date_to,omitempty" validate:"omitempty,date"`
	MinResponseTime int    `json:"min_response_time,omitempty" validate:"omitempty,min=0"`
	MaxResponseTime int    `json:"max_response_time,omitempty" validate:"omitempty,min=0"`
}
