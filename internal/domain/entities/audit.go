package entities

import (
	"strings"
	"time"
)

// AuditLog representa un registro de auditoría del sistema
type AuditLog struct {
	ID             int64   `json:"id" db:"id"`
	PublicID       string  `json:"public_id" db:"public_id"`
	Action         string  `json:"action" db:"action"`           // create, update, delete, login, logout, etc.
	EntityType     string  `json:"entity_type" db:"entity_type"` // user, ticket, event, etc.
	EntityID       int64   `json:"entity_id" db:"entity_id"`
	EntityPublicID *string `json:"entity_public_id,omitempty" db:"entity_public_id"`

	UserID       *int64  `json:"user_id,omitempty" db:"user_id"`
	UserPublicID *string `json:"user_public_id,omitempty" db:"user_public_id"`
	UserName     *string `json:"user_name,omitempty" db:"user_name"`
	UserEmail    *string `json:"user_email,omitempty" db:"user_email"`
	UserRole     *string `json:"user_role,omitempty" db:"user_role"`

	IPAddress     *string `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent     *string `json:"user_agent,omitempty" db:"user_agent"`
	RequestPath   *string `json:"request_path,omitempty" db:"request_path"`
	RequestMethod *string `json:"request_method,omitempty" db:"request_method"`

	Changes map[string]interface{} `json:"changes,omitempty" db:"changes"`
	OldData map[string]interface{} `json:"old_data,omitempty" db:"old_data"`
	NewData map[string]interface{} `json:"new_data,omitempty" db:"new_data"`

	Severity string  `json:"severity" db:"severity"` // info, warning, error, critical
	Status   string  `json:"status" db:"status"`     // success, failed
	Message  string  `json:"message" db:"message"`
	Error    *string `json:"error,omitempty" db:"error"`

	Metadata map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	Tags     []string               `json:"tags,omitempty" db:"tags"`

	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty" db:"updated_at"`
}

// SecurityLog representa un registro de seguridad
type SecurityLog struct {
	ID        int64  `json:"id" db:"id"`
	PublicID  string `json:"public_id" db:"public_id"`
	EventType string `json:"event_type" db:"event_type"` // login_failed, password_changed, mfa_enabled, etc.

	UserID             *int64  `json:"user_id,omitempty" db:"user_id"`
	UserPublicID       *string `json:"user_public_id,omitempty" db:"user_public_id"`
	TargetUserID       *int64  `json:"target_user_id,omitempty" db:"target_user_id"`
	TargetUserPublicID *string `json:"target_user_public_id,omitempty" db:"target_user_public_id"`

	IPAddress string  `json:"ip_address" db:"ip_address"`
	UserAgent *string `json:"user_agent,omitempty" db:"user_agent"`
	Location  *string `json:"location,omitempty" db:"location"`

	Severity  string `json:"severity" db:"severity"` // low, medium, high, critical
	RiskScore int    `json:"risk_score" db:"risk_score"`

	Description string                 `json:"description" db:"description"`
	Details     map[string]interface{} `json:"details,omitempty" db:"details"`

	IsSuspicious bool `json:"is_suspicious" db:"is_suspicious"`
	IsBlocked    bool `json:"is_blocked" db:"is_blocked"`
	IsResolved   bool `json:"is_resolved" db:"is_resolved"`

	ActionTaken  *string    `json:"action_taken,omitempty" db:"action_taken"`
	ResolvedBy   *int64     `json:"resolved_by,omitempty" db:"resolved_by"`
	ResolvedAt   *time.Time `json:"resolved_at,omitempty" db:"resolved_at"`
	ResolvedNote *string    `json:"resolved_note,omitempty" db:"resolved_note"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// DataChange representa un cambio en datos
type DataChange struct {
	ID       int64  `json:"id" db:"id"`
	PublicID string `json:"public_id" db:"public_id"`

	TableName      string  `json:"table_name" db:"table_name"`
	RecordID       int64   `json:"record_id" db:"record_id"`
	RecordPublicID *string `json:"record_public_id,omitempty" db:"record_public_id"`

	Operation string    `json:"operation" db:"operation"` // INSERT, UPDATE, DELETE
	ChangedAt time.Time `json:"changed_at" db:"changed_at"`

	UserID       *int64  `json:"user_id,omitempty" db:"user_id"`
	UserPublicID *string `json:"user_public_id,omitempty" db:"user_public_id"`

	OldData       map[string]interface{} `json:"old_data,omitempty" db:"old_data"`
	NewData       map[string]interface{} `json:"new_data,omitempty" db:"new_data"`
	ChangedFields []string               `json:"changed_fields" db:"changed_fields"`
	Diff          map[string]interface{} `json:"diff,omitempty" db:"diff"`

	IPAddress *string `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent *string `json:"user_agent,omitempty" db:"user_agent"`
	RequestID *string `json:"request_id,omitempty" db:"request_id"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// AuditConfig configuración de auditoría
type AuditConfig struct {
	Enabled          bool     `json:"enabled"`
	LogLevels        []string `json:"log_levels"` // info, warning, error, critical
	RetentionDays    int      `json:"retention_days"`
	AutoArchive      bool     `json:"auto_archive"`
	ArchiveAfterDays int      `json:"archive_after_days"`
	CompressArchives bool     `json:"compress_archives"`

	// Qué eventos auditar
	AuditLogins      bool `json:"audit_logins"`
	AuditLogouts     bool `json:"audit_logouts"`
	AuditDataChanges bool `json:"audit_data_changes"`
	AuditPermissions bool `json:"audit_permissions"`
	AuditPayments    bool `json:"audit_payments"`
	AuditTickets     bool `json:"audit_tickets"`

	// Excepciones
	ExcludedUsers     []int64  `json:"excluded_users,omitempty"`
	ExcludedIPs       []string `json:"excluded_ips,omitempty"`
	ExcludedEndpoints []string `json:"excluded_endpoints,omitempty"`

	// Alertas
	EnableAlerts   bool     `json:"enable_alerts"`
	AlertThreshold int      `json:"alert_threshold"` // Eventos por minuto
	AlertEmails    []string `json:"alert_emails,omitempty"`

	// Compliance
	GDPRCompliant   bool `json:"gdpr_compliant"`
	PCIDSSCompliant bool `json:"pci_dss_compliant"`
	SOXCompliant    bool `json:"sox_compliant"`

	// Tiempos
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AuditStats estadísticas de auditoría
type AuditStats struct {
	Date           time.Time `json:"date"`
	TotalLogs      int64     `json:"total_logs"`
	SecurityLogs   int64     `json:"security_logs"`
	DataChangeLogs int64     `json:"data_change_logs"`
	LoginLogs      int64     `json:"login_logs"`
	FailedLogins   int64     `json:"failed_logins"`

	BySeverity map[string]int64 `json:"by_severity"`
	ByEntity   map[string]int64 `json:"by_entity"`
	ByUser     map[int64]int64  `json:"by_user"`
	ByIP       map[string]int64 `json:"by_ip"`

	AvgResponseTime      float64 `json:"avg_response_time"`
	PeakHour             string  `json:"peak_hour"`
	SuspiciousActivities int64   `json:"suspicious_activities"`
	BlockedIPs           int64   `json:"blocked_ips"`

	CreatedAt time.Time `json:"created_at"`
}

// Métodos de utilidad
func (a *AuditLog) IsHighSeverity() bool {
	return a.Severity == "error" || a.Severity == "critical"
}

func (a *AuditLog) ContainsSensitiveData() bool {
	sensitiveFields := []string{
		"password", "token", "secret", "key",
		"credit_card", "cvv", "ssn", "sin",
		"authorization", "api_key", "private_key",
	}

	// Verificar en changes, old_data, new_data
	allData := []map[string]interface{}{a.Changes, a.OldData, a.NewData}

	for _, data := range allData {
		if data == nil {
			continue
		}
		for key := range data {
			for _, sensitive := range sensitiveFields {
				if containsIgnoreCase(key, sensitive) {
					return true
				}
			}
		}
	}
	return false
}

func (s *SecurityLog) ShouldAlert() bool {
	return s.Severity == "high" || s.Severity == "critical" || s.RiskScore >= 80
}

func (s *SecurityLog) IsLoginRelated() bool {
	loginEvents := []string{
		"login_failed", "login_success", "login_attempt",
		"password_reset", "mfa_attempt", "account_lockout",
	}

	for _, event := range loginEvents {
		if s.EventType == event {
			return true
		}
	}
	return false
}

// Helper functions
func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
