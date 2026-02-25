package enums

// AuditSeverity representa el nivel de severidad de un evento de auditoría
// Valores alineados con el CHECK constraint de la tabla audit.security_logs
type AuditSeverity string

const (
	// AuditSeverityLow - Evento informativo, sin impacto en seguridad
	AuditSeverityLow AuditSeverity = "low"
	// AuditSeverityMedium - Evento de precaución, posible problema menor
	AuditSeverityMedium AuditSeverity = "medium"
	// AuditSeverityHigh - Evento importante, requiere atención
	AuditSeverityHigh AuditSeverity = "high"
	// AuditSeverityCritical - Evento crítico, acción inmediata requerida
	AuditSeverityCritical AuditSeverity = "critical"
)

// IsValid verifica si el valor del enum es válido
func (as AuditSeverity) IsValid() bool {
	switch as {
	case AuditSeverityLow, AuditSeverityMedium, AuditSeverityHigh, AuditSeverityCritical:
		return true
	}
	return false
}

// Level devuelve el nivel numérico de severidad (1-4)
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

// RequiresImmediateAction indica si la severidad requiere acción inmediata
func (as AuditSeverity) RequiresImmediateAction() bool {
	return as == AuditSeverityHigh || as == AuditSeverityCritical
}

// ShouldAlert indica si se debe generar una alerta
func (as AuditSeverity) ShouldAlert() bool {
	return as == AuditSeverityCritical
}

// IsAuditable indica si este nivel debe ser registrado en auditoría
// Todos los niveles son auditables por defecto
func (as AuditSeverity) IsAuditable() bool {
	return as.IsValid()
}

// Color devuelve un código de color para representación visual
func (as AuditSeverity) Color() string {
	switch as {
	case AuditSeverityLow:
		return "#28a745" // Verde
	case AuditSeverityMedium:
		return "#ffc107" // Amarillo
	case AuditSeverityHigh:
		return "#fd7e14" // Naranja
	case AuditSeverityCritical:
		return "#dc3545" // Rojo
	default:
		return "#6c757d" // Gris
	}
}

// String devuelve la representación string de la severidad
func (as AuditSeverity) String() string {
	return string(as)
}

// SeverityFromLevel convierte un nivel numérico (1-4) a severidad
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

// GetAllSeverities devuelve todos los niveles de severidad posibles
func GetAllSeverities() []AuditSeverity {
	return []AuditSeverity{
		AuditSeverityLow,
		AuditSeverityMedium,
		AuditSeverityHigh,
		AuditSeverityCritical,
	}
}

// GetSeveritiesAbove devuelve todos los niveles de severidad mayores o iguales al dado
func GetSeveritiesAbove(severity AuditSeverity) []AuditSeverity {
	threshold := severity.Level()
	all := GetAllSeverities()
	result := make([]AuditSeverity, 0)

	for _, s := range all {
		if s.Level() >= threshold {
			result = append(result, s)
		}
	}

	return result
}

// GetSeveritiesBelow devuelve todos los niveles de severidad menores o iguales al dado
func GetSeveritiesBelow(severity AuditSeverity) []AuditSeverity {
	threshold := severity.Level()
	all := GetAllSeverities()
	result := make([]AuditSeverity, 0)

	for _, s := range all {
		if s.Level() <= threshold {
			result = append(result, s)
		}
	}

	return result
}

// MarshalJSON implementa la interfaz json.Marshaler
func (as AuditSeverity) MarshalJSON() ([]byte, error) {
	return []byte(`"` + string(as) + `"`), nil
}

// UnmarshalJSON implementa la interfaz json.Unmarshaler
func (as *AuditSeverity) UnmarshalJSON(data []byte) error {
	// Remover comillas
	str := string(data)
	if len(str) >= 2 {
		str = str[1 : len(str)-1]
	}

	severity := AuditSeverity(str)
	if !severity.IsValid() {
		return &InvalidAuditSeverityError{Severity: str}
	}

	*as = severity
	return nil
}

// InvalidAuditSeverityError error para valores inválidos
type InvalidAuditSeverityError struct {
	Severity string
}

func (e *InvalidAuditSeverityError) Error() string {
	return "invalid audit severity: " + e.Severity
}
