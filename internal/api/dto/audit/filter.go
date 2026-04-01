// internal/api/dto/audit/filter.go
package audit

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
