package entities

import (
	"strings"
	"time"
)

// DataChange representa un cambio en datos
// Mapea exactamente la tabla audit.data_changes
type DataChange struct {
	ID int64 `json:"id" db:"id"`
	// CORREGIDO: La tabla usa table_name, no TableName
	TableName     string                  `json:"table_name" db:"table_name"`
	RecordID      int64                   `json:"record_id" db:"record_id"`
	Operation     string                  `json:"operation" db:"operation"` // INSERT, UPDATE, DELETE
	OldData       *map[string]interface{} `json:"old_data,omitempty" db:"old_data,type:jsonb"`
	NewData       *map[string]interface{} `json:"new_data,omitempty" db:"new_data,type:jsonb"`
	ChangedFields []string                `json:"changed_fields" db:"changed_fields,type:text[]"`
	UserID        *int64                  `json:"user_id,omitempty" db:"user_id"`
	IPAddress     *string                 `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent     *string                 `json:"user_agent,omitempty" db:"user_agent"`
	RequestPath   *string                 `json:"request_path,omitempty" db:"request_path"`
	ChangedAt     time.Time               `json:"changed_at" db:"changed_at"`
}

// SecurityLog representa un registro de seguridad
// Mapea exactamente la tabla audit.security_logs
type SecurityLog struct {
	ID           int64                   `json:"id" db:"id"`
	EventType    string                  `json:"event_type" db:"event_type"`
	Severity     string                  `json:"severity" db:"severity"` // low, medium, high, critical
	Description  string                  `json:"description" db:"description"`
	UserID       *int64                  `json:"user_id,omitempty" db:"user_id"`
	TargetUserID *int64                  `json:"target_user_id,omitempty" db:"target_user_id"`
	IPAddress    *string                 `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent    *string                 `json:"user_agent,omitempty" db:"user_agent"`
	RequestPath  *string                 `json:"request_path,omitempty" db:"request_path"`
	Details      *map[string]interface{} `json:"details,omitempty" db:"details,type:jsonb"`
	OccurredAt   time.Time               `json:"occurred_at" db:"occurred_at"`
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

// ContainsSensitiveData verifica si el cambio contiene datos sensibles
func (d *DataChange) ContainsSensitiveData() bool {
	sensitiveFields := []string{
		"password", "token", "secret", "key",
		"credit_card", "cvv", "ssn", "sin",
		"authorization", "api_key", "private_key",
	}

	// Verificar en old_data y new_data
	allData := []*map[string]interface{}{d.OldData, d.NewData}

	for _, data := range allData {
		if data == nil {
			continue
		}
		for key := range *data {
			for _, sensitive := range sensitiveFields {
				if containsIgnoreCase(key, sensitive) {
					return true
				}
			}
		}
	}
	return false
}

// ShouldAlert determina si el evento de seguridad debe generar una alerta
func (s *SecurityLog) ShouldAlert() bool {
	return s.Severity == "high" || s.Severity == "critical"
}

// IsLoginRelated verifica si el evento está relacionado con login
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
