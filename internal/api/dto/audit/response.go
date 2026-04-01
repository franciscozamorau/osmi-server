// internal/api/dto/audit/response.go
package audit

import "time"

type AuditResponse struct {
	ID            int64                  `json:"id"`
	TableName     string                 `json:"table_name"`
	RecordID      int64                  `json:"record_id"`
	Operation     string                 `json:"operation"`
	OldData       map[string]interface{} `json:"old_data,omitempty"`
	NewData       map[string]interface{} `json:"new_data,omitempty"`
	ChangedFields []string               `json:"changed_fields,omitempty"`
	UserID        string                 `json:"user_id,omitempty"`
	UserEmail     string                 `json:"user_email,omitempty"`
	IPAddress     string                 `json:"ip_address,omitempty"`
	UserAgent     string                 `json:"user_agent,omitempty"`
	RequestPath   string                 `json:"request_path,omitempty"`
	ChangedAt     time.Time              `json:"changed_at"`
}

type SecurityLogResponse struct {
	ID           int64                  `json:"id"`
	EventType    string                 `json:"event_type"`
	Severity     string                 `json:"severity"`
	Description  string                 `json:"description"`
	UserID       string                 `json:"user_id,omitempty"`
	UserEmail    string                 `json:"user_email,omitempty"`
	TargetUserID string                 `json:"target_user_id,omitempty"`
	TargetEmail  string                 `json:"target_email,omitempty"`
	IPAddress    string                 `json:"ip_address,omitempty"`
	UserAgent    string                 `json:"user_agent,omitempty"`
	RequestPath  string                 `json:"request_path,omitempty"`
	Details      map[string]interface{} `json:"details,omitempty"`
	OccurredAt   time.Time              `json:"occurred_at"`
}

type AuditStatsResponse struct {
	TotalChanges     int64            `json:"total_changes"`
	Inserts          int64            `json:"inserts"`
	Updates          int64            `json:"updates"`
	Deletes          int64            `json:"deletes"`
	ChangesByTable   map[string]int64 `json:"changes_by_table"`
	ChangesByUser    map[string]int64 `json:"changes_by_user"`
	ChangesLast7Days []DailyChange    `json:"changes_last_7_days"`
}

type SecurityStatsResponse struct {
	TotalEvents     int64            `json:"total_events"`
	CriticalEvents  int64            `json:"critical_events"`
	HighEvents      int64            `json:"high_events"`
	MediumEvents    int64            `json:"medium_events"`
	LowEvents       int64            `json:"low_events"`
	EventsByType    map[string]int64 `json:"events_by_type"`
	EventsByUser    map[string]int64 `json:"events_by_user"`
	EventsLast7Days []DailyEvent     `json:"events_last_7_days"`
}

type DailyChange struct {
	Date    string `json:"date"`
	Inserts int64  `json:"inserts"`
	Updates int64  `json:"updates"`
	Deletes int64  `json:"deletes"`
	Total   int64  `json:"total"`
}

type DailyEvent struct {
	Date     string `json:"date"`
	Critical int64  `json:"critical"`
	High     int64  `json:"high"`
	Medium   int64  `json:"medium"`
	Low      int64  `json:"low"`
	Total    int64  `json:"total"`
}

type SecurityAlert struct {
	ID          string    `json:"id"`
	Severity    string    `json:"severity"`
	EventType   string    `json:"event_type"`
	Description string    `json:"description"`
	UserEmail   string    `json:"user_email,omitempty"`
	IPAddress   string    `json:"ip_address,omitempty"`
	OccurredAt  time.Time `json:"occurred_at"`
	Resolved    bool      `json:"resolved"`
}

// ============================================================================
// TIPOS ADICIONALES PARA REPOSITORIOS
// ============================================================================

// ActivityPoint - punto de actividad en timeline
type ActivityPoint struct {
	Timestamp string `json:"timestamp"`
	Hour      int    `json:"hour"`
	Count     int64  `json:"count"`
}

// TableActivity - actividad por tabla
type TableActivity struct {
	TableName string `json:"table_name"`
	Reads     int64  `json:"reads"`
	Writes    int64  `json:"writes"`
	Deletes   int64  `json:"deletes"`
}

// UserActivity - actividad por usuario
type UserActivity struct {
	UserID       int64  `json:"user_id"`
	UserName     string `json:"user_name"`
	EventCount   int64  `json:"event_count"`
	LastActivity string `json:"last_activity"`
}

// SecurityEventDistribution - distribución de eventos de seguridad
type SecurityEventDistribution struct {
	EventType  string  `json:"event_type"`
	Count      int64   `json:"count"`
	Percentage float64 `json:"percentage"`
}

// ChangeFrequency - frecuencia de cambios
type ChangeFrequency struct {
	TableName   string `json:"table_name"`
	ChangeCount int64  `json:"change_count"`
	LastChange  string `json:"last_change"`
}
