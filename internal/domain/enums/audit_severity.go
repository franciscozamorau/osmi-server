package enums

type AuditSeverity string

const (
	AuditSeverityLow      AuditSeverity = "low"
	AuditSeverityMedium   AuditSeverity = "medium"
	AuditSeverityHigh     AuditSeverity = "high"
	AuditSeverityCritical AuditSeverity = "critical"
)

func (as AuditSeverity) IsValid() bool {
	switch as {
	case AuditSeverityLow, AuditSeverityMedium, AuditSeverityHigh, AuditSeverityCritical:
		return true
	}
	return false
}

func (as AuditSeverity) Level() int {
	switch as {
	case AuditSeverityLow:
		return 1
	case AuditSeverityMedium:
		return 2
	case AuditSeverityHigh:
		return 3
	case AuditSeverityCritical:
		return 4
	default:
		return 0
	}
}

func (as AuditSeverity) RequiresImmediateAction() bool {
	return as == AuditSeverityHigh || as == AuditSeverityCritical
}

func (as AuditSeverity) ShouldAlert() bool {
	return as == AuditSeverityCritical
}

func (as AuditSeverity) String() string {
	return string(as)
}

// SeverityFromLevel convierte un nivel numérico a severidad
func SeverityFromLevel(level int) AuditSeverity {
	switch {
	case level >= 4:
		return AuditSeverityCritical
	case level == 3:
		return AuditSeverityHigh
	case level == 2:
		return AuditSeverityMedium
	default:
		return AuditSeverityLow
	}
}
