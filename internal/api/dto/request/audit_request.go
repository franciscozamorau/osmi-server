package dto

type AuditRequest struct {
	TableName     string                 `json:"table_name" validate:"required"`
	RecordID      int64                  `json:"record_id" validate:"required"`
	Operation     string                 `json:"operation" validate:"required,oneof=INSERT UPDATE DELETE"`
	OldData       map[string]interface{} `json:"old_data,omitempty"`
	NewData       map[string]interface{} `json:"new_data,omitempty"`
	ChangedFields []string               `json:"changed_fields,omitempty"`
	UserID        string                 `json:"user_id,omitempty" validate:"omitempty,uuid4"`
	IPAddress     string                 `json:"ip_address,omitempty"`
	UserAgent     string                 `json:"user_agent,omitempty"`
	RequestPath   string                 `json:"request_path,omitempty"`
}

type SecurityLogRequest struct {
	EventType    string                 `json:"event_type" validate:"required"`
	Severity     string                 `json:"severity" validate:"required,oneof=low medium high critical"`
	Description  string                 `json:"description" validate:"required"`
	UserID       string                 `json:"user_id,omitempty" validate:"omitempty,uuid4"`
	TargetUserID string                 `json:"target_user_id,omitempty" validate:"omitempty,uuid4"`
	IPAddress    string                 `json:"ip_address,omitempty"`
	UserAgent    string                 `json:"user_agent,omitempty"`
	RequestPath  string                 `json:"request_path,omitempty"`
	Details      map[string]interface{} `json:"details,omitempty"`
}

type AuditFilter struct {
	TableName string `json:"table_name,omitempty"`
	RecordID  int64  `json:"record_id,omitempty"`
	Operation string `json:"operation,omitempty"`
	UserID    string `json:"user_id,omitempty"`
	DateFrom  string `json:"date_from,omitempty" validate:"omitempty,date"`
	DateTo    string `json:"date_to,omitempty" validate:"omitempty,date"`
}

type SecurityLogFilter struct {
	EventType    string `json:"event_type,omitempty"`
	Severity     string `json:"severity,omitempty"`
	UserID       string `json:"user_id,omitempty"`
	TargetUserID string `json:"target_user_id,omitempty"`
	DateFrom     string `json:"date_from,omitempty" validate:"omitempty,date"`
	DateTo       string `json:"date_to,omitempty" validate:"omitempty,date"`
}
